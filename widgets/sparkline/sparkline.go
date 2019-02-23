// Copyright 2018 Google Inc.
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

// Package sparkline is a widget that draws a graph showing a series of values as vertical bars.
package sparkline

import (
	"errors"
	"fmt"
	"image"
	"sync"

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// SparkLine draws a graph showing a series of values as vertical bars.
//
// Bars can have sub-cell height. The graphs scale adjusts dynamically based on
// the largest visible value.
//
// Implements widgetapi.Widget. This object is thread-safe.
type SparkLine struct {
	// data are the data points the SparkLine displays.
	data []int

	// mu protects the SparkLine.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new SparkLine.
func New(opts ...Option) (*SparkLine, error) {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}

	return &SparkLine{
		opts: opt,
	}, nil
}

// Draw draws the SparkLine widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (sl *SparkLine) Draw(cvs *canvas.Canvas) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	needAr, err := area.FromSize(sl.minSize())
	if err != nil {
		return err
	}
	if !needAr.In(cvs.Area()) {
		return draw.ResizeNeeded(cvs)
	}

	ar := sl.area(cvs)
	visible, max := visibleMax(sl.data, ar.Dx())
	var curX int
	if len(visible) < ar.Dx() {
		curX = ar.Max.X - len(visible)
	} else {
		curX = ar.Min.X
	}

	for _, v := range visible {
		blocks := toBlocks(v, max, ar.Dy())
		curY := ar.Max.Y - 1
		for i := 0; i < blocks.full; i++ {
			if _, err := cvs.SetCell(
				image.Point{curX, curY},
				sparks[len(sparks)-1], // Last spark represents full cell.
				cell.FgColor(sl.opts.color),
			); err != nil {
				return err
			}

			curY--
		}

		if blocks.partSpark != 0 {
			if _, err := cvs.SetCell(
				image.Point{curX, curY},
				blocks.partSpark,
				cell.FgColor(sl.opts.color),
			); err != nil {
				return err
			}
		}

		curX++
	}

	if sl.opts.label != "" {
		// Label is placed immediately above the SparkLine.
		lStart := image.Point{ar.Min.X, ar.Min.Y - 1}
		if err := draw.Text(cvs, sl.opts.label, lStart,
			draw.TextCellOpts(sl.opts.labelCellOpts...),
			draw.TextOverrunMode(draw.OverrunModeThreeDot),
		); err != nil {
			return err
		}
	}
	return nil
}

// Add adds data points to the SparkLine.
// Each data point is represented by one bar on the SparkLine. Zero value data
// points are valid and are represented by an empty space on the SparkLine
// (i.e. a missing bar).
//
// At least one data point must be provided. All data points must be positive
// integers.
//
// The last added data point will be the one displayed all the way on the right
// of the SparkLine. If there are more data points than we can fit bars to the
// width of the SparkLine, only the last n data points that fit will be
// visible.
//
// Provided options override values set when New() was called.
func (sl *SparkLine) Add(data []int, opts ...Option) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	for _, opt := range opts {
		opt.set(sl.opts)
	}

	for i, d := range data {
		if d < 0 {
			return fmt.Errorf("data point[%d]: %v must be a positive integer", i, d)
		}
	}
	sl.data = append(sl.data, data...)
	return nil
}

// Clear removes all the data points in the SparkLine, effectively returning to
// an empty graph.
func (sl *SparkLine) Clear() {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.data = nil
}

// Keyboard input isn't supported on the SparkLine widget.
func (*SparkLine) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("the SparkLine widget doesn't support keyboard events")
}

// Mouse input isn't supported on the SparkLine widget.
func (*SparkLine) Mouse(m *terminalapi.Mouse) error {
	return errors.New("the SparkLine widget doesn't support mouse events")
}

// area returns the area of the canvas available to the SparkLine.
func (sl *SparkLine) area(cvs *canvas.Canvas) image.Rectangle {
	cvsAr := cvs.Area()
	maxY := cvsAr.Max.Y

	// Height is determined based on options (fixed height / label).
	var minY int
	if sl.opts.height > 0 {
		minY = maxY - sl.opts.height
	} else {
		minY = cvsAr.Min.Y

		if sl.opts.label != "" {
			minY++ // Reserve one line for the label.
		}
	}
	return image.Rect(
		cvsAr.Min.X,
		minY,
		cvsAr.Max.X,
		maxY,
	)
}

// minSize returns the minimum canvas size for the SparkLine based on the options.
func (sl *SparkLine) minSize() image.Point {
	const minWidth = 1 // At least one data point.

	var minHeight int
	if sl.opts.height > 0 {
		minHeight = sl.opts.height
	} else {
		minHeight = 1 // At least one line of characters.
	}

	if sl.opts.label != "" {
		minHeight++ // One line for the text label.
	}
	return image.Point{minWidth, minHeight}
}

// Options implements widgetapi.Widget.Options.
func (sl *SparkLine) Options() widgetapi.Options {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	min := sl.minSize()
	var max image.Point
	if sl.opts.height > 0 {
		max = min // Fix the height to the one specified.
	}

	return widgetapi.Options{
		MinimumSize:  min,
		MaximumSize:  max,
		WantKeyboard: widgetapi.KeyScopeNone,
		WantMouse: widgetapi.MouseScopeNone,
	}
}
