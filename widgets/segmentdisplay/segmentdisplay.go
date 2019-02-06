// Copyright 2019 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package segmentdisplay is a widget that displays text by simulating a
// segment display.
package segmentdisplay

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"sync"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/attrrange"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/draw/segdisp/sixteen"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// SegmentDisplay displays ASCII content by simulating a segment display.
//
// Automatically determines the size of individual segments with goal of
// maximizing the segment size or with fitting the entire text depending on the
// provided options.
//
// Segment displays support only a subset of ASCII characters, provided options
// determine the behavior when an unsupported character is encountered.
//
// Implements widgetapi.Widget. This object is thread-safe.
type SegmentDisplay struct {
	// buff contains the text to be displayed.
	buff bytes.Buffer

	// givenWOpts are write options given for the text in buff.
	givenWOpts []*writeOptions
	// wOptsTracker tracks the positions in a buff to which the givenWOpts apply.
	wOptsTracker *attrrange.Tracker

	// mu protects the widget.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new SegmentDisplay.
func New(opts ...Option) (*SegmentDisplay, error) {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return &SegmentDisplay{
		wOptsTracker: attrrange.NewTracker(),
		opts:         opt,
	}, nil
}

// TextChunk is a part of or the full text that will be displayed.
type TextChunk struct {
	text  string
	wOpts *writeOptions
}

// NewChunk creates a new text chunk.
func NewChunk(text string, wOpts ...WriteOption) *TextChunk {
	return &TextChunk{
		text:  text,
		wOpts: newWriteOptions(wOpts...),
	}
}

// Write writes text for the widget to display. Subsequent calls replace text
// written previously. All the provided text chunks are broken into characters
// and each character is displayed in one segment.
//
// The provided write options determine the behavior when text contains
// unsupported characters and set cell options for cells that contain
// individual display segments.
//
// Each of the text chunks can have its own options. At least one chunk must be
// specified.
//
// Any provided options override options given to New.
func (sd *SegmentDisplay) Write(chunks []*TextChunk, opts ...Option) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	for _, o := range opts {
		o.set(sd.opts)
	}
	if err := sd.opts.validate(); err != nil {
		return err
	}

	if len(chunks) == 0 {
		return errors.New("at least one text chunk must be specified")
	}
	sd.reset()

	for i, tc := range chunks {
		if tc.text == "" {
			return fmt.Errorf("text chunk[%d] is empty, all chunks must contains some text", i)
		}
		if ok, badRunes := sixteen.SupportsChars(tc.text); !ok && tc.wOpts.errOnUnsupported {
			return fmt.Errorf("text chunk[%d] contains unsupported characters %v, clean the text or provide the WriteSanitize option", i, badRunes)
		}
		text := sixteen.Sanitize(tc.text)

		pos := sd.buff.Len()
		sd.givenWOpts = append(sd.givenWOpts, tc.wOpts)
		wOptsIdx := len(sd.givenWOpts) - 1
		if err := sd.wOptsTracker.Add(pos, pos+len(text), wOptsIdx); err != nil {
			return err
		}
		sd.buff.WriteString(text)
	}
	return nil
}

// Reset resets the widget back to empty content.
func (sd *SegmentDisplay) Reset() {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	sd.reset()
}

// reset is the implementation of Reset.
// Caller must hold sd.mu.
func (sd *SegmentDisplay) reset() {
	sd.buff.Reset()
	sd.givenWOpts = nil
	sd.wOptsTracker = attrrange.NewTracker()
}

// segArea contains information about the area that will contain the segments.
type segArea struct {
	// segment is the area for one segment.
	segment image.Rectangle
	// canFit is the number of segments we can fit on the canvas.
	canFit int
	// gapPixels is the size of gaps between segments in pixels.
	gapPixels int
	// gaps is the number of gaps that will be drawn.
	gaps int
}

// needArea returns the complete area required for all the segments that we can
// fit and any gaps.
func (sa *segArea) needArea() image.Rectangle {
	return image.Rect(
		0,
		0,
		sa.segment.Dx()*sa.canFit+sa.gaps*sa.gapPixels,
		sa.segment.Dy(),
	)
}

// segArea calculates size and number of segments that can fit onto the
// specified area.
func (sd *SegmentDisplay) segArea(ar image.Rectangle) (*segArea, error) {
	segAr, err := sixteen.Required(ar)
	if err != nil {
		return nil, fmt.Errorf("sixteen.Required => %v", err)
	}
	gapPixels := segAr.Dy() * sd.opts.gapPercent / 100

	var (
		gaps   int
		canFit int
		taken  int
	)
	for i := 0; i < sd.buff.Len(); i++ {
		taken += segAr.Dx()

		if taken > ar.Dx() {
			break
		}
		canFit++

		// Don't insert gaps after the last segment in the text or the last
		// segment we can fit.
		if gapPixels == 0 || i == sd.buff.Len()-1 {
			continue
		}

		remaining := ar.Dx() - taken
		// Only insert gaps if we can still fit one more segment with the gap.
		if remaining >= gapPixels+segAr.Dx() {
			taken += gapPixels
			gaps++
		} else {
			// Gap is needed but doesn't fit together with the next segment.
			// So insert neither.
			break
		}
	}
	return &segArea{
		segment:   segAr,
		canFit:    canFit,
		gapPixels: gapPixels,
		gaps:      gaps,
	}, nil
}

// maximizeFit finds the largest individual segment size that enables us to fit
// the most characters onto a canvas with the provided area. Returns the area
// required for a single segment and the number of segments we can fit.
func (sd *SegmentDisplay) maximizeFit(ar image.Rectangle) (*segArea, error) {
	var bestSegAr *segArea
	need := sd.buff.Len()
	for height := ar.Dy(); height >= sixteen.MinRows; height-- {
		ar := image.Rect(ar.Min.X, ar.Min.Y, ar.Max.X, ar.Min.Y+height)
		segAr, err := sd.segArea(ar)
		if err != nil {
			return nil, err
		}

		if segAr.canFit >= need {
			return segAr, nil
		}
		bestSegAr = segAr
	}
	return bestSegAr, nil
}

// preprocess determines the size of individual segments maximizing their
// height or the amount of displayed characters based on the specified options.
// Returns the area required for a single segment, the text that we can fit and
// size of gaps between segments in cells.
func (sd *SegmentDisplay) preprocess(cvsAr image.Rectangle) (*segArea, error) {
	segAr, err := sd.segArea(cvsAr)
	if err != nil {
		return nil, err
	}

	need := sd.buff.Len()
	if need <= segAr.canFit || sd.opts.maximizeSegSize {
		return segAr, nil
	}

	bestAr, err := sd.maximizeFit(cvsAr)
	if err != nil {
		return nil, err
	}
	return bestAr, nil
}

// Draw draws the SegmentDisplay widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (sd *SegmentDisplay) Draw(cvs *canvas.Canvas) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	if sd.buff.Len() == 0 {
		return nil
	}

	segAr, err := sd.preprocess(cvs.Area())
	if err != nil {
		return err
	}

	text := sd.buff.String()
	aligned, err := align.Rectangle(cvs.Area(), segAr.needArea(), sd.opts.hAlign, sd.opts.vAlign)
	if err != nil {
		return fmt.Errorf("align.Rectangle => %v", err)
	}

	optRange, err := sd.wOptsTracker.ForPosition(0) // Text options for the current byte.
	if err != nil {
		return err
	}

	gaps := segAr.gaps
	startX := aligned.Min.X
	for i, c := range text {
		if i >= segAr.canFit {
			break
		}

		disp := sixteen.New()
		if err := disp.SetCharacter(c); err != nil {
			return fmt.Errorf("disp.SetCharacter => %v", err)
		}

		endX := startX + segAr.segment.Dx()
		ar := image.Rect(startX, aligned.Min.Y, endX, aligned.Max.Y)
		startX = endX
		if gaps > 0 {
			startX += segAr.gapPixels
			gaps--
		}

		dCvs, err := canvas.New(ar)
		if err != nil {
			return fmt.Errorf("canvas.New => %v", err)
		}

		if i >= optRange.High { // Get the next write options.
			or, err := sd.wOptsTracker.ForPosition(i)
			if err != nil {
				return err
			}
			optRange = or
		}
		wOpts := sd.givenWOpts[optRange.AttrIdx]

		if err := disp.Draw(dCvs, sixteen.CellOpts(wOpts.cellOpts...)); err != nil {
			return fmt.Errorf("disp.Draw => %v", err)
		}

		if err := dCvs.CopyTo(cvs); err != nil {
			return fmt.Errorf("dCvs.CopyTo => %v", err)
		}
	}
	return nil
}

// Keyboard input isn't supported on the SegmentDisplay widget.
func (*SegmentDisplay) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("the SegmentDisplay widget doesn't support keyboard events")
}

// Mouse input isn't supported on the SegmentDisplay widget.
func (*SegmentDisplay) Mouse(m *terminalapi.Mouse) error {
	return errors.New("the SegmentDisplay widget doesn't support mouse events")
}

// Options implements widgetapi.Widget.Options.
func (sd *SegmentDisplay) Options() widgetapi.Options {
	return widgetapi.Options{
		// The smallest supported size of a display segment.
		MinimumSize:  image.Point{sixteen.MinCols, sixteen.MinRows},
		WantKeyboard: false,
		WantMouse:    false,
	}
}
