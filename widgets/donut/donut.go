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
	"math"
	"sync"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/private/alignfor"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/braille"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
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
	r := int(math.Round(float64(donutRadius) / 100 * float64(d.opts.donutHolePercent)))
	if r < 2 { // Smallest possible circle radius.
		return 0
	}
	return r
}

// drawText draws the text label showing the progress.
// The text is only drawn if the radius of the donut "hole" is large enough to
// accommodate it.
// The mid point addresses coordinates in pixels on a braille canvas.
func (d *Donut) drawText(cvs *canvas.Canvas, mid image.Point, holeR int) error {
	cells, first := availableCells(mid, holeR)
	t := d.progressText()
	needCells := runewidth.StringWidth(t)
	if cells < needCells {
		return nil
	}

	ar := image.Rect(first.X, first.Y, first.X+cells+2, first.Y+1)
	start, err := alignfor.Text(ar, t, align.HorizontalCenter, align.VerticalMiddle)
	if err != nil {
		return fmt.Errorf("alignfor.Text => %v", err)
	}
	if err := draw.Text(cvs, t, start, draw.TextMaxX(start.X+needCells), draw.TextCellOpts(d.opts.textCellOpts...)); err != nil {
		return fmt.Errorf("draw.Text => %v", err)
	}
	return nil
}

// drawLabel draws the text label in the area.
func (d *Donut) drawLabel(cvs *canvas.Canvas, labelAr image.Rectangle) error {
	start, err := alignfor.Text(labelAr, d.opts.label, d.opts.labelAlign, align.VerticalBottom)
	if err != nil {
		return err
	}
	return draw.Text(
		cvs, d.opts.label, start,
		draw.TextOverrunMode(draw.OverrunModeThreeDot),
		draw.TextMaxX(labelAr.Max.X),
		draw.TextCellOpts(d.opts.labelCellOpts...),
	)
}

// Draw draws the Donut widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (d *Donut) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	startA, endA := startEndAngles(d.current, d.total, d.opts.startAngle, d.opts.direction)
	if startA == endA {
		// No progress recorded, so nothing to do.
		return nil
	}

	var donutAr, labelAr image.Rectangle
	if len(d.opts.label) > 0 {
		d, l, err := donutAndLabel(cvs.Area())
		if err != nil {
			return err
		}
		donutAr = d
		labelAr = l

	} else {
		donutAr = cvs.Area()
	}

	if donutAr.Dx() < minSize.X || donutAr.Dy() < minSize.Y {
		// Reserving area for the label might have resulted in donutAr being
		// too small.
		return draw.ResizeNeeded(cvs)
	}

	bc, err := braille.New(donutAr)
	if err != nil {
		return fmt.Errorf("braille.New => %v", err)
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
		if err := d.drawText(cvs, mid, holeR); err != nil {
			return err
		}
	}

	if !labelAr.Empty() {
		if err := d.drawLabel(cvs, labelAr); err != nil {
			return err
		}
	}
	return nil
}

// Keyboard input isn't supported on the Donut widget.
func (*Donut) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return errors.New("the Donut widget doesn't support keyboard events")
}

// Mouse input isn't supported on the Donut widget.
func (*Donut) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	return errors.New("the Donut widget doesn't support mouse events")
}

// minSize is the smallest area we can draw donut on.
var minSize = image.Point{3, 3}

// Options implements widgetapi.Widget.Options.
func (d *Donut) Options() widgetapi.Options {
	return widgetapi.Options{
		// We are drawing a circle, ensure equal ratio of rows and columns.
		// This is adjusted for the inequality of the braille canvas.
		Ratio: image.Point{braille.RowMult, braille.ColMult},

		// The smallest circle that "looks" like a circle on the canvas.
		MinimumSize:  minSize,
		WantKeyboard: widgetapi.KeyScopeNone,
		WantMouse:    widgetapi.MouseScopeNone,
	}
}

// donutAndLabel splits the canvas area into an area for the donut and an
// area under the donut for the text label.
func donutAndLabel(cvsAr image.Rectangle) (donAr, labelAr image.Rectangle, err error) {
	height := cvsAr.Dy()
	// Two lines for the text label at the bottom.
	// One for the text itself and one for visual space between the donut and
	// the label.
	donAr, labelAr, err = area.HSplitCells(cvsAr, height-2)
	if err != nil {
		return image.ZR, image.ZR, err
	}
	return donAr, labelAr, nil
}
