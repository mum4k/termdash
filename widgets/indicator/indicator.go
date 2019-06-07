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

// Package Indicator is a widget that displays the status of a background operation
// or shows whether a value is (on/off)
package indicator

import (
	"errors"
	"fmt"
	"image"
	"math"
	"sync"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/internal/alignfor"
	"github.com/mum4k/termdash/internal/area"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/canvas/braille"
	"github.com/mum4k/termdash/internal/draw"
	"github.com/mum4k/termdash/internal/runewidth"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Indicator displays the progress of an operation by filling a partial circle and
// The circle can has a "hole" in the middle, which is where the name comes from.
// Implements widgetapi.Widget. This object is thread-safe.
type Indicator struct {
	// status is the current status (on/off) that will be displayed.
	status bool
	// mu protects the Indicator.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new Indicator.
func New(opts ...Option) (*Indicator, error) {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return &Indicator{
		opts: opt,
	}, nil
}

func (i *Indicator) On() error {
	i.status = true
	return nil
}

func (i *Indicator) Off() error {
	i.status = false
	return nil
}

func (i *Indicator) Toggle() error {
	if i.status == true {
		i.status = false
	} else {
		i.status = true
	}
	return nil
}

func (i *Indicator) drawLabel(cvs *canvas.Canvas, labelAr image.Rectangle) error {
	start, err := alignfor.Text(labelAr, i.opts.label, i.opts.labelAlign, align.VerticalBottom)
	if err != nil {
		return err
	}
	if err := draw.Text(
		cvs, i.opts.label, start,
		draw.TextOverrunMode(draw.OverrunModeThreeDot),
		draw.TextMaxX(labelAr.Max.X),
		draw.TextCellOpts(i.opts.labelColor...),
	); err != nil {
		return err
	}
	return nil
}

// innerholeRadius calculates the radius of the "inner hole" in the indicator.
// Returns zero if no hole should be drawn.
func (i *Indicator) innerHoleRadius(indicatorRadius int) int {
	r := int(math.Round(float64(indicatorRadius) / 100 * float64(90)))
	if r < 2 { // Smallest possible circle radius.
		return 0
	}
	return r
}

// Draw draws the Indicator widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (i *Indicator) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	needAr, err := area.FromSize(i.minSize())
	if err != nil {
		return err
	}
	if !needAr.In(cvs.Area()) {
		return draw.ResizeNeeded(cvs)
	}
	var indAr, labelAr image.Rectangle
	if len(i.opts.label) > 0 {
		i, l, err := indicatorAndLabel(cvs.Area())
		if err != nil {
			return err
		}
		indAr = i
		labelAr = l

	} else {
		indAr = cvs.Area()
	}

	var t string
	if indAr.Dx() < minScaledSize.X || indAr.Dy() < minScaledSize.Y {
		if i.status == true {
			t = "\u25C9"
		} else {
			t = "\u25EF"
		}
		needCells := runewidth.StringWidth(t)

		ar := image.Rect(minSize.X, minSize.Y, minSize.X+2, minSize.Y+1)
		start, err := alignfor.Text(ar, t, align.HorizontalCenter, align.VerticalMiddle)
		if err != nil {
			return fmt.Errorf("alignfor.Text => %v", err)
		}
		if err := draw.Text(cvs, t, start, draw.TextMaxX(start.X+needCells), draw.TextCellOpts(i.opts.color...)); err != nil {
			return fmt.Errorf("draw.Text => %v", err)
		}
	} else {
		bc, err := braille.New(indAr)
		if err != nil {
			return fmt.Errorf("braille.New => %v", err)
		}

		mid, r := midAndRadius(bc.Area())
		if r > i.opts.maxDiameter {
			r = i.opts.maxDiameter
		}
		if err := draw.BrailleCircle(bc, mid, r,
			draw.BrailleCircleCellOpts(i.opts.color...),
		); err != nil {
			return fmt.Errorf("failed to draw the outer circle: %v", err)
		}
		if i.status == true {
			if err := draw.BrailleCircle(bc, mid, int(float64(r)*float64(.9)),
				draw.BrailleCircleFilled(),
				draw.BrailleCircleCellOpts(i.opts.color...),
			); err != nil {
				return fmt.Errorf("failed to draw the outer circle: %v", err)
			}
		}
		innerHoleR := i.innerHoleRadius(r)
		if innerHoleR != 0 {
			if err := draw.BrailleCircle(bc, mid, innerHoleR,
				draw.BrailleCircleClearPixels(),
			); err != nil {
				return fmt.Errorf("failed to draw the outer circle: %v", err)
			}
		}

		if err := bc.CopyTo(cvs); err != nil {
			return err
		}
		///
	}

	if indAr.Dx() < minSize.X || indAr.Dy() < minSize.Y {
		// Reserving area for the label might have resulted in indAr being
		// too small.
		return draw.ResizeNeeded(cvs)
	}
	if !labelAr.Empty() {
		if err := i.drawLabel(cvs, labelAr); err != nil {
			return err
		}
	}

	return nil
}

// minSize determines the minimum required size of the canvas.
func (i *Indicator) minSize() image.Point {
	minWidth := 1  // Shorter indicator than this cannot display anything.
	minHeight := 1 // At least 1 for the indicator itself.
	return image.Point{minWidth, minHeight}
}

// Keyboard input isn't supported on the Indicator widget.
func (*Indicator) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("the Indicator widget doesn't support keyboard events")
}

// Mouse input isn't supported on the Indicator widget.
func (*Indicator) Mouse(m *terminalapi.Mouse) error {
	return errors.New("the Indicator widget doesn't support mouse events")
}

// minSize is the smallest area we can draw indicator on.
var minSize = image.Point{1, 1}
var minScaledSize = image.Point{3, 3}

// Options implements widgetapi.Widget.Options.
func (s *Indicator) Options() widgetapi.Options {
	return widgetapi.Options{
		// We are drawing a circle, ensure equal ratio of rows and columns.
		// This is adjusted for the inequality of the braille canvas.
		Ratio: image.Point{braille.RowMult, braille.ColMult},
		// The smallest circle that "looks" like a circle on the canvas.
		MinimumSize:  image.Point{1, 1},
		WantKeyboard: widgetapi.KeyScopeNone,
		WantMouse:    widgetapi.MouseScopeNone,
	}
}

// indicatorAndLabel splits the canvas area into an area for the indicator and an
// area under the indicator for the text label.
func indicatorAndLabel(cvsAr image.Rectangle) (indAr, labelAr image.Rectangle, err error) {
	height := cvsAr.Dy()
	// Two lines for the text label at the bottom.
	// One for the text itself and one for visual space between the donut and
	// the label.
	indAr, labelAr, err = area.HSplitCells(cvsAr, height-2)
	if err != nil {
		return image.ZR, image.ZR, err
	}
	return indAr, labelAr, nil
}
