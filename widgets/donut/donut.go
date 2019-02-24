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

// Package donut is a widget that displays the progress of an operation as a
// partial or full circle.
package donut

import (
	"errors"
	"fmt"
	"image"
	"sync"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell/runewidth"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/canvas/braille"
	"github.com/mum4k/termdash/internal/draw"
	"github.com/mum4k/termdash/internal/numbers"
	"github.com/mum4k/termdash/internal/widgetapi"
	"github.com/mum4k/termdash/terminalapi"
)

// progressType indicates how was the current progress provided by the caller.
type progressType int

// String implements fmt.Stringer()
func (pt progressType) String() string {
	if n, ok := progressTypeNames[pt]; ok {
		return n
	}
	return "progressTypeUnknown"
}

// progressTypeNames maps progressType values to human readable names.
var progressTypeNames = map[progressType]string{
	progressTypePercent:  "progressTypePercent",
	progressTypeAbsolute: "progressTypeAbsolute",
}

const (
	progressTypePercent = iota
	progressTypeAbsolute
)

// Donut displays the progress of an operation by filling a partial circle and
// eventually by completing a full circle. The circle can have a "hole" in the
// middle, which is where the name comes from.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Donut struct {
	// pt indicates how current and total are interpreted.
	pt progressType
	// current is the current progress that will be drawn.
	current int
	// total is the value that represents completion.
	// For progressTypePercent, this is 100, for progressTypeAbsolute this is
	// the total provided by the caller.
	total int
	// mu protects the Donut.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new Donut.
func New(opts ...Option) (*Donut, error) {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return &Donut{
		opts: opt,
	}, nil
}

// Absolute sets the progress in absolute numbers, e.g. 7 out of 10.
// The total amount must be a non-zero positive integer. The done amount must
// be a zero or a positive integer such that done <= total.
// Provided options override values set when New() was called.
func (d *Donut) Absolute(done, total int, opts ...Option) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if done < 0 || total < 1 || done > total {
		return fmt.Errorf("invalid progress, done(%d) must be <= total(%d), done must be zero or positive "+
			"and total must be a non-zero positive number", done, total)
	}

	for _, opt := range opts {
		opt.set(d.opts)
	}
	if err := d.opts.validate(); err != nil {
		return err
	}

	d.pt = progressTypeAbsolute
	d.current = done
	d.total = total
	return nil
}

// Percent sets the current progress in percentage.
// The provided value must be between 0 and 100.
// Provided options override values set when New() was called.
func (d *Donut) Percent(p int, opts ...Option) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if p < 0 || p > 100 {
		return fmt.Errorf("invalid percentage, p(%d) must be 0 <= p <= 100", p)
	}

	for _, opt := range opts {
		opt.set(d.opts)
	}
	if err := d.opts.validate(); err != nil {
		return err
	}

	d.pt = progressTypePercent
	d.current = p
	d.total = 100
	return nil
}

// progressText returns the textual representation of the current progress.
func (d *Donut) progressText() string {
	switch d.pt {
	case progressTypePercent:
		return fmt.Sprintf("%d%%", d.current)
	case progressTypeAbsolute:
		return fmt.Sprintf("%d/%d", d.current, d.total)
	default:
		return ""
	}
}

// holeRadius calculates the radius of the "hole" in the donut.
// Returns zero if no hole should be drawn.
func (d *Donut) holeRadius(donutRadius int) int {
	r := int(numbers.Round(float64(donutRadius) / 100 * float64(d.opts.donutHolePercent)))
	if r < 2 { // Smallest possible circle radius.
		return 0
	}
	return r
}

// drawText draws the text label showing the progress.
// The text is only drawn if the radius of the donut "hole" is large enough to
// accommodate it.
func (d *Donut) drawText(cvs *canvas.Canvas, mid image.Point, holeR int) error {
	cells, first := availableCells(mid, holeR)
	t := d.progressText()
	needCells := runewidth.StringWidth(t)
	if cells < needCells {
		return nil
	}

	ar := image.Rect(first.X, first.Y, first.X+cells+2, first.Y+1)
	start, err := align.Text(ar, t, align.HorizontalCenter, align.VerticalMiddle)
	if err != nil {
		return fmt.Errorf("align.Text => %v", err)
	}
	if err := draw.Text(cvs, t, start, draw.TextMaxX(start.X+needCells), draw.TextCellOpts(d.opts.textCellOpts...)); err != nil {
		return fmt.Errorf("draw.Text => %v", err)
	}
	return nil
}

// Draw draws the Donut widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (d *Donut) Draw(cvs *canvas.Canvas) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	bc, err := braille.New(cvs.Area())
	if err != nil {
		return fmt.Errorf("braille.New => %v", err)
	}

	startA, endA := startEndAngles(d.current, d.total, d.opts.startAngle, d.opts.direction)
	if startA == endA {
		// No progress recorded, so nothing to do.
		return nil
	}

	mid, r := midAndRadius(bc.Area())
	if err := draw.BrailleCircle(bc, mid, r,
		draw.BrailleCircleFilled(),
		draw.BrailleCircleArcOnly(startA, endA),
		draw.BrailleCircleCellOpts(d.opts.cellOpts...),
	); err != nil {
		return fmt.Errorf("failed to draw the outer circle: %v", err)
	}

	holeR := d.holeRadius(r)
	if holeR != 0 {
		if err := draw.BrailleCircle(bc, mid, holeR,
			draw.BrailleCircleFilled(),
			draw.BrailleCircleClearPixels(),
		); err != nil {
			return fmt.Errorf("failed to draw the outer circle: %v", err)
		}
	}
	if err := bc.CopyTo(cvs); err != nil {
		return err
	}

	if !d.opts.hideTextProgress {
		return d.drawText(cvs, mid, holeR)
	}
	return nil
}

// Keyboard input isn't supported on the Donut widget.
func (*Donut) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("the Donut widget doesn't support keyboard events")
}

// Mouse input isn't supported on the Donut widget.
func (*Donut) Mouse(m *terminalapi.Mouse) error {
	return errors.New("the Donut widget doesn't support mouse events")
}

// Options implements widgetapi.Widget.Options.
func (d *Donut) Options() widgetapi.Options {
	return widgetapi.Options{
		// We are drawing a circle, ensure equal ratio of rows and columns.
		// This is adjusted for the inequality of the braille canvas.
		Ratio: image.Point{braille.RowMult, braille.ColMult},

		// The smallest circle that "looks" like a circle on the canvas.
		MinimumSize:  image.Point{3, 3},
		WantKeyboard: widgetapi.KeyScopeNone,
		WantMouse:    widgetapi.MouseScopeNone,
	}
}
