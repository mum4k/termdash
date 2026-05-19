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

package fx

import (
	"image"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// FramedWidget wraps any widgetapi.Widget and paints its own border directly
// onto the widget canvas so that effect wrappers (EffectWidget,
// FocusEffectWidget) animate the border characters together with the inner
// content.
//
// The typical stack for a whole-container focus effect is:
//
//	inner  → FramedNew(inner, ...)     → fx.NewLooping(framed, effect)
//	       → fx.FocusNew(ew, in, out)
//
// Place the outermost widget in a container with linestyle.None so the
// container does not draw a second border on top.
//
// The border color can be updated at runtime (e.g. from AnimateBorderFocus)
// by calling SetBorderColor — it is safe to call from any goroutine.
//
// Create with FramedNew.
type FramedWidget struct {
	mu         sync.Mutex
	inner      widgetapi.Widget
	lineStyle  linestyle.LineStyle
	title      string
	titleOpts  []cell.Option
	borderOpts []cell.Option // base options (set at creation time)
	colorOpt   cell.Option   // runtime color override; nil = use borderOpts as-is

	tmp      *canvas.Canvas
	lastSize image.Point
}

// FramedOption configures a FramedWidget.
type FramedOption interface {
	set(*FramedWidget)
}

type framedOption func(*FramedWidget)

func (f framedOption) set(fw *FramedWidget) { f(fw) }

// FramedLineStyle sets the border line style.
// Default: linestyle.Round.
func FramedLineStyle(ls linestyle.LineStyle) FramedOption {
	return framedOption(func(fw *FramedWidget) { fw.lineStyle = ls })
}

// FramedTitle sets the title text drawn in the top border row.
// opts are applied to the title characters only.
func FramedTitle(title string, opts ...cell.Option) FramedOption {
	return framedOption(func(fw *FramedWidget) {
		fw.title = " " + title + " "
		fw.titleOpts = opts
	})
}

// FramedBorderOpts sets the initial cell options applied to all border cells
// (e.g. foreground color, bold).  These are used as a base; SetBorderColor
// layered on top without discarding the base options.
func FramedBorderOpts(opts ...cell.Option) FramedOption {
	return framedOption(func(fw *FramedWidget) {
		fw.borderOpts = append(fw.borderOpts[:0:0], opts...)
	})
}

// FramedNew wraps inner with a self-drawn border.
// The returned widget should be placed in a container with linestyle.None to
// avoid a double border.
func FramedNew(inner widgetapi.Widget, opts ...FramedOption) (*FramedWidget, error) {
	fw := &FramedWidget{
		inner:     inner,
		lineStyle: linestyle.Round,
	}
	for _, o := range opts {
		o.set(fw)
	}
	return fw, nil
}

// SetBorderColor updates the foreground color of all border cells.
// It overlays the color on top of any options set via FramedBorderOpts.
// Safe to call from any goroutine; takes effect on the next Draw call.
func (fw *FramedWidget) SetBorderColor(c cell.Color) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.colorOpt = cell.FgColor(c)
}

// Draw implements widgetapi.Widget.
func (fw *FramedWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	size := cvs.Size()

	// Need at least 3×3: a 1-cell border on each side plus a 1×1 inner area.
	if size.X < 3 || size.Y < 3 {
		return nil
	}

	innerSize := image.Point{X: size.X - 2, Y: size.Y - 2}

	// (Re)create the scratch canvas on resize.
	if fw.tmp == nil || fw.lastSize != innerSize {
		var err error
		fw.tmp, err = canvas.New(image.Rect(0, 0, innerSize.X, innerSize.Y))
		if err != nil {
			return err
		}
		fw.lastSize = innerSize
	}

	// Clear and draw the inner widget onto the scratch canvas.
	if err := fw.tmp.Clear(); err != nil {
		return err
	}
	if err := fw.inner.Draw(fw.tmp, meta); err != nil {
		return err
	}

	// Copy the inner canvas onto the outer canvas at offset (1, 1).
	innerAr := fw.tmp.Area()
	for y := innerAr.Min.Y; y < innerAr.Max.Y; y++ {
		for x := innerAr.Min.X; x < innerAr.Max.X; x++ {
			p := image.Point{X: x, Y: y}
			c, err := fw.tmp.Cell(p)
			if err != nil {
				continue
			}
			_, _ = cvs.SetCell(image.Point{X: x + 1, Y: y + 1}, c.Rune, c.Opts)
		}
	}

	// Assemble border draw options.
	var cellOpts []cell.Option
	cellOpts = append(cellOpts, fw.borderOpts...)
	if fw.colorOpt != nil {
		cellOpts = append(cellOpts, fw.colorOpt)
	}

	borderDrawOpts := []draw.BorderOption{
		draw.BorderLineStyle(fw.lineStyle),
	}
	if len(cellOpts) > 0 {
		borderDrawOpts = append(borderDrawOpts, draw.BorderCellOpts(cellOpts...))
	}
	if fw.title != "" {
		borderDrawOpts = append(borderDrawOpts, draw.BorderTitle(fw.title, draw.OverrunModeTrim, fw.titleOpts...))
	}

	return draw.Border(cvs, cvs.Area(), borderDrawOpts...)
}

// Keyboard implements widgetapi.Widget.
func (fw *FramedWidget) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return fw.inner.Keyboard(k, meta)
}

// Mouse implements widgetapi.Widget.
func (fw *FramedWidget) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	return fw.inner.Mouse(m, meta)
}

// Options implements widgetapi.Widget.
func (fw *FramedWidget) Options() widgetapi.Options {
	opts := fw.inner.Options()
	opts.MinimumSize.X += 2
	opts.MinimumSize.Y += 2
	if opts.MinimumSize.X < 3 {
		opts.MinimumSize.X = 3
	}
	if opts.MinimumSize.Y < 3 {
		opts.MinimumSize.Y = 3
	}
	return opts
}
