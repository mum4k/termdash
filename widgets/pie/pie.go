// Copyright 2026 Google Inc.
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

package pie

import (
	"errors"
	"fmt"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/braille"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"image"
	"sync"
)

// Pie is the widget that displays a pie chart.
type Pie struct {
	mu     sync.Mutex
	values []int
	total  int
	colors []cell.Color
	opts   *options
}

// New returns a new Pie widget.
func New(opts ...Option) (*Pie, error) {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	return &Pie{
		opts: opt,
	}, nil
}

// Values must be provided before calling Draw.
func (p *Pie) Values(values []int, opts ...Option) error {
	// The values must be non-negative and a color must be provided for each value.
	// If not enough colors are provided, they will be reused.
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(values) == 0 {
		return errors.New("values cannot be empty")
	}

	for _, opt := range opts {
		opt.set(p.opts)
	}

	p.values = values
	p.total = 0
	if len(p.colors) == 0 {
		p.colors = DefaultColors
	}
	for _, v := range values {
		if v < 0 {
			return errors.New("all values must be non-negative")
		}
		p.total += v
	}

	return nil
}

// it returns the center point and horizontal and vertical radii.
func pieChartMidAndRadii(ar image.Rectangle) (image.Point, int) {
	width := ar.Dx() * braille.ColMult
	height := ar.Dy() * braille.RowMult

	radiusX := width/2 - 2
	if radiusX < 1 {
		radiusX = 1
	}
	mid := image.Point{
		X: ar.Min.X*braille.ColMult + width/2,
		Y: ar.Min.Y*braille.RowMult + height/2,
	}
	return mid, radiusX
}

// Draw renders the Pie widget onto the provided canvas. It calculates the
// pie chart slices based on the values and colors defined in the Pie struct.
// Each slice is drawn as a series of radial lines from the inner radius to
// the outer radius. The method ensures thread safety by locking the Pie's
// mutex during the drawing process.
//
// The number of colors in the list is not significant. If there are more values than
// colors, the colors will be reused in a round-robin fashion. This ensures that all
// segments are assigned a color, even if the number of values exceeds the number of
// available colors in the list.
//
// Parameters:
//   - cvs: The canvas onto which the pie chart will be drawn.
//   - meta: Metadata about the widget's environment.
//
// Returns:
//   - error: An error if the drawing process fails, or nil if successful.
func (p *Pie) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.total <= 0 {
		return nil
	}

	bc, err := braille.New(cvs.Area())
	if err != nil {
		return fmt.Errorf("braille.New => %v", err)
	}

	mid, radiusX := pieChartMidAndRadii(cvs.Area())

	innerRadiusX := int(float64(radiusX) * 0.6)

	currentAngle := 0.0
	for i, value := range p.values {
		endAngle := currentAngle + float64(value)/float64(p.total)*360 // Convert to degrees
		color := p.colors[i%len(p.colors)]

		for j := innerRadiusX; j <= radiusX; j++ {
			// Draw the arc for the current slice.
			if err := draw.BrailleCircle(
				bc,
				mid,
				j,
				draw.BrailleCircleArcOnly(int(currentAngle), int(endAngle)),
				draw.BrailleCircleCellOpts(cell.FgColor(color)),
			); err != nil {
				return fmt.Errorf("failed to draw pie slice arc: %v", err)
			}
		}

		currentAngle = endAngle
	}

	if err := bc.CopyTo(cvs); err != nil {
		return err
	}

	return nil
}

// Keyboard input isn't supported on the Pie widget.
func (*Pie) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return errors.New("the Pie widget doesn't support keyboard events")
}

// Mouse input isn't supported on the Pie widget.
func (*Pie) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	return errors.New("the Pie widget doesn't support mouse events")
}

// Options implements widgetapi.Widget.Options.
func (p *Pie) Options() widgetapi.Options {
	return widgetapi.Options{
		Ratio:        image.Point{braille.RowMult, braille.ColMult},
		MinimumSize:  image.Point{5, 5},
		WantKeyboard: widgetapi.KeyScopeNone,
		WantMouse:    widgetapi.MouseScopeNone,
	}
}
