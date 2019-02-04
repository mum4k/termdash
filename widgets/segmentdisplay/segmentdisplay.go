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
	"image"
	"sync"

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
		opts: opt,
	}, nil
}

// Write writes text for the widget to display. Multiple calls append
// additional text.
// The provided write options determine the behavior when text contains
// unsupported characters.
func (sd *SegmentDisplay) Write(text string /* TODO wOpts ...WriteOption */) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return nil
}

// Reset resets the widget back to empty content.
func (sd *SegmentDisplay) Reset() {
	sd.mu.Lock()
	defer sd.mu.Unlock()
}

// Draw draws the SegmentDisplay widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (sd *SegmentDisplay) Draw(cvs *canvas.Canvas) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	return errors.New("unimplemented")
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
		//
		// TODO: Return size required based on the text length and options.
		MinimumSize:  image.Point{sixteen.MinCols, sixteen.MinRows},
		WantKeyboard: false,
		WantMouse:    false,
	}
}
