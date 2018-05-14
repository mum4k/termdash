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

// Package gauge implements a widget that displays the progress of an operation.
package gauge

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"sync"
	"unicode/utf8"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"golang.org/x/exp/utf8string"
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

// Gauge displays the progress of an operation.
//
// Draws a rectangle, a progress bar with optional display of percentage and /
// or text label.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Gauge struct {
	// pt indicates how current and total are interpreted.
	pt progressType
	// current is the current progress that will be drawn.
	current int
	// total is the value that represents completion.
	// For progressTypePercent, this is 100, for progressTypeAbsolute this is
	// the total provided by the caller.
	total int
	// mu protects the Gauge.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new Gauge.
func New(opts ...Option) *Gauge {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	return &Gauge{
		opts: opt,
	}
}

// Absolute sets the progress in absolute numbers, i.e. 7 out of 10.
// The total amount must be a non-zero positive integer. The done amount must
// be a zero or a positive integer such that done <= total.
// Provided options override values set when New() was called.
func (g *Gauge) Absolute(done, total int, opts ...Option) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if done < 0 || total < 1 || done > total {
		return fmt.Errorf("invalid progress, done(%d) must be <= total(%d), done must be zero or positive "+
			"and total must be a non-zero positive number", done, total)
	}

	for _, opt := range opts {
		opt.set(g.opts)
	}

	g.pt = progressTypeAbsolute
	g.current = done
	g.total = total
	return nil
}

// Percent sets the current progress in percentage.
// The provided value must be between 0 and 100.
// Provided options override values set when New() was called.
func (g *Gauge) Percent(p int, opts ...Option) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if p < 0 || p > 100 {
		return fmt.Errorf("invalid percentage, p(%d) must be 0 <= p <= 100", p)
	}

	for _, opt := range opts {
		opt.set(g.opts)
	}

	g.pt = progressTypePercent
	g.current = p
	g.total = 100
	return nil
}

// width determines the required width of the gauge drawn on the provided area
// in order to represent the current progress.
func (g *Gauge) width(ar image.Rectangle) int {
	mult := float32(g.current) / float32(g.total)
	width := float32(ar.Dx()) * mult
	return int(width)
}

// hasBorder determines of the gauge has a border.
func (g *Gauge) hasBorder() bool {
	return g.opts.border != draw.LineStyleNone
}

// usable determines the usable area for the gauge itself.
func (g *Gauge) usable(cvs *canvas.Canvas) image.Rectangle {
	if g.hasBorder() {
		return area.ExcludeBorder(cvs.Area())
	}
	return cvs.Area()
}

// progressText returns the textual representation of the current progress.
func (g *Gauge) progressText() string {
	if g.opts.hideTextProgress {
		return ""
	}

	switch g.pt {
	case progressTypePercent:
		return fmt.Sprintf("%d%%", g.current)
	case progressTypeAbsolute:
		return fmt.Sprintf("%d/%d", g.current, g.total)
	default:
		return ""
	}
}

// gaugeText returns full text to be displayed within the gauge, i.e. the
// progress text and the optional label.
func (g *Gauge) gaugeText() string {
	var b bytes.Buffer
	b.WriteString(g.progressText())
	if g.opts.textLabel != "" {
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		b.WriteString(fmt.Sprintf("(%s)", g.opts.textLabel))
	}
	return b.String()
}

// drawText draws the text enumerating the progress and the text label.
func (g *Gauge) drawText(cvs *canvas.Canvas) error {
	text := g.gaugeText()
	if text == "" {
		return nil
	}

	ar := g.usable(cvs)
	textStart, err := align.Text(ar, text, g.opts.hTextAlign, g.opts.vTextAlign)
	if err != nil {
		return err
	}
	textEndX := textStart.X + utf8.RuneCountInString(text)
	if textEndX >= ar.Max.X { // The text will be trimmed.
		textEndX = ar.Max.X - 1
	}
	gaugeEndX := g.width(ar)

	switch {
	case gaugeEndX < textStart.X:
		// The text entirely falls outside of the drawn gauge.
		_, err := draw.Text(cvs, text, textStart,
			draw.TextOverrunMode(draw.OverrunModeThreeDot),
			draw.TextCellOpts(cell.FgColor(g.opts.emptyTextColor)),
			draw.TextMaxX(ar.Max.X),
		)
		return err

	case gaugeEndX >= textEndX:
		// The text entirely falls inside of the drawn gauge.
		_, err := draw.Text(cvs, text, textStart,
			draw.TextOverrunMode(draw.OverrunModeThreeDot),
			draw.TextCellOpts(cell.FgColor(g.opts.filledTextColor)),
			draw.TextMaxX(ar.Max.X),
		)
		return err

	default:
		// Part of the text falls inside of the drawn gauge and part outside.
		utfText := utf8string.NewString(text)
		insideCount := ar.Min.X + gaugeEndX - textStart.X
		insideText := utfText.Slice(0, insideCount)
		outsideText := utfText.Slice(insideCount, utfText.RuneCount())

		_, err := draw.Text(cvs, insideText, textStart,
			draw.TextOverrunMode(draw.OverrunModeTrim),
			draw.TextCellOpts(cell.FgColor(g.opts.filledTextColor)),
		)
		if err != nil {
			return err
		}

		outsideStart := image.Point{textStart.X + insideCount, textStart.Y}
		if outsideStart.In(ar) {
			if _, err := draw.Text(cvs, outsideText, outsideStart,
				draw.TextOverrunMode(draw.OverrunModeThreeDot),
				draw.TextCellOpts(cell.FgColor(g.opts.emptyTextColor)),
				draw.TextMaxX(ar.Max.X),
			); err != nil {
				return err
			}
		}
	}
	return nil
}

// Draw draws the Gauge widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (g *Gauge) Draw(cvs *canvas.Canvas) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.hasBorder() {
		if err := draw.Border(cvs, cvs.Area(),
			draw.BorderLineStyle(g.opts.border),
			draw.BorderTitle(g.opts.borderTitle, draw.OverrunModeThreeDot, g.opts.borderCellOpts...),
			draw.BorderTitleAlign(g.opts.borderTitleHAlign),
			draw.BorderCellOpts(g.opts.borderCellOpts...),
		); err != nil {
			return err
		}
	}

	usable := g.usable(cvs)
	progress := image.Rect(
		usable.Min.X,
		usable.Min.Y,
		usable.Min.X+g.width(usable),
		usable.Max.Y,
	)
	if progress.Dx() > 0 {
		if err := draw.Rectangle(cvs, progress,
			draw.RectChar(g.opts.gaugeChar),
			draw.RectCellOpts(cell.BgColor(g.opts.color)),
		); err != nil {
			return err
		}
	}
	return g.drawText(cvs)
}

// Keyboard input isn't supported on the Gauge widget.
func (g *Gauge) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("the Gauge widget doesn't support keyboard events")
}

// Mouse input isn't supported on the Gauge widget.
func (g *Gauge) Mouse(m *terminalapi.Mouse) error {
	return errors.New("the Gauge widget doesn't support mouse events")
}

// maxSize determines the maximum size of the canvas.
func (g *Gauge) maxSize() image.Point {
	maxHeight := g.opts.height
	if g.hasBorder() {
		// Add the required space for the border.
		maxHeight += 2
	}
	return image.Point{0, maxHeight}
}

// minSize determines the minimum required size of the canvas.
func (g *Gauge) minSize() image.Point {
	minWidth := 1  // Shorter gauge than this cannot display anything.
	minHeight := 1 // At least one line for the gauge itself.
	if g.hasBorder() {
		// Add the required space for the border.
		minWidth += 2
		minHeight += 2
	}
	return image.Point{minWidth, minHeight}
}

// Options implements widgetapi.Widget.Options.
func (g *Gauge) Options() widgetapi.Options {
	g.mu.Lock()
	defer g.mu.Unlock()
	return widgetapi.Options{
		MaximumSize:  g.maxSize(),
		MinimumSize:  g.minSize(),
		WantKeyboard: false,
		WantMouse:    false,
	}
}
