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

// Package textinput implements a widget that accepts text input.
package textinput

import (
	"image"
	"strings"
	"sync"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/alignfor"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/private/wrap"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// TextInput accepts text input from the user.
//
// Displays an input field and an optional text label. The input field allows
// the user to edit and submit text.
//
// The text can be submitted by pressing enter or read at any time by calling
// Read. The text input field can be navigated using arrows, the Home and End
// button and using mouse.
//
// Implements widgetapi.Widget. This object is thread-safe.
type TextInput struct {
	// mu protects the widget.
	mu sync.Mutex

	// editor tracks the edits and the state of the text input field.
	editor *fieldEditor

	// forField is the area that was occupied by the text input field last
	// time Draw() was called.
	forField image.Rectangle

	// opts are the provided options.
	opts *options
}

// New returns a new TextInput.
func New(opts ...Option) (*TextInput, error) {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}
	ti := &TextInput{
		editor: newFieldEditor(),
		opts:   opt,
	}

	for _, r := range ti.opts.defaultText {
		ti.editor.insert(r)
	}
	return ti, nil
}

// Vars to be replaced from tests.
var (
	// textFieldRune is the rune used in cells reserved for the text input
	// field if no text is present.
	// Changed from tests to provide readable test failures.
	textFieldRune rune

	// cursorRune is rune that represents the cursor position.
	cursorRune rune
)

// Read reads the content of the text input field.
func (ti *TextInput) Read() string {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	return ti.editor.content()
}

// ReadAndClear reads the content of the text input field and clears it.
func (ti *TextInput) ReadAndClear() string {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	c := ti.editor.content()
	ti.editor.reset()
	return c
}

// drawLabel draws the text label in the area.
func (ti *TextInput) drawLabel(cvs *canvas.Canvas, labelAr image.Rectangle) error {
	start, err := alignfor.Text(labelAr, ti.opts.label, ti.opts.labelAlign, align.VerticalMiddle)
	if err != nil {
		return err
	}
	return draw.Text(
		cvs, ti.opts.label, start,
		draw.TextOverrunMode(draw.OverrunModeThreeDot),
		draw.TextMaxX(labelAr.Max.X),
		draw.TextCellOpts(ti.opts.labelCellOpts...),
	)
}

// drawField draws the text input field.
func (ti *TextInput) drawField(cvs *canvas.Canvas, text string) error {
	if err := cvs.SetAreaCells(ti.forField, textFieldRune, cell.BgColor(ti.opts.fillColor)); err != nil {
		return err
	}

	if ti.opts.hideTextWith != 0 {
		text = hideText(text, ti.opts.hideTextWith)
	}

	return draw.Text(
		cvs, text, ti.forField.Min,
		draw.TextMaxX(ti.forField.Max.X),
		draw.TextCellOpts(cell.FgColor(ti.opts.textColor)),
	)
}

// drawCursor draws the cursor within the text input field.
func (ti *TextInput) drawCursor(cvs *canvas.Canvas, curPos int) error {
	p := image.Point{
		curPos + ti.forField.Min.X,
		ti.forField.Min.Y,
	}
	if err := cvs.SetCellOpts(
		p,
		cell.FgColor(ti.opts.highlightedColor),
		cell.BgColor(ti.opts.cursorColor),
	); err != nil {
		return err
	}
	if cursorRune != 0 {
		if _, err := cvs.SetCell(p, cursorRune); err != nil {
			return err
		}
	}
	return nil
}

// Draw draws the TextInput widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (ti *TextInput) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	labelAr, textAr, err := split(cvs.Area(), ti.opts.label, ti.opts.widthPerc)
	if err != nil {
		return err
	}

	if ti.opts.border != linestyle.None {
		ti.forField = area.ExcludeBorder(textAr)
	} else {
		ti.forField = textAr
	}

	if ti.forField.Dx() < minFieldWidth || ti.forField.Dy() < minFieldHeight {
		return draw.ResizeNeeded(cvs)
	}

	if !labelAr.Eq(image.ZR) {
		if err := ti.drawLabel(cvs, labelAr); err != nil {
			return err
		}
	}

	if ti.opts.border != linestyle.None {
		if err := draw.Border(cvs, textAr, draw.BorderCellOpts(cell.FgColor(ti.opts.borderColor))); err != nil {
			return err
		}
	}

	text, curPos, err := ti.editor.viewFor(ti.forField.Dx())
	if err != nil {
		return err
	}

	if err := ti.drawField(cvs, text); err != nil {
		return err
	}

	if meta.Focused {
		if err := ti.drawCursor(cvs, curPos); err != nil {
			return err
		}
	} else if ti.opts.placeHolder != "" && text == "" {
		if err := draw.Text(
			cvs, ti.opts.placeHolder, ti.forField.Min,
			draw.TextMaxX(ti.forField.Max.X),
			draw.TextCellOpts(cell.FgColor(ti.opts.placeHolderColor)),
		); err != nil {
			return err
		}
	}
	return nil
}

// keyboard processes keyboard events.
// Returns a bool indicating if the content was submitted and the text in the
// field at submission time.
// Implements widgetapi.Widget.Keyboard.
func (ti *TextInput) keyboard(k *terminalapi.Keyboard) (bool, string) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	switch k.Key {
	case keyboard.KeyBackspace, keyboard.KeyBackspace2:
		ti.editor.deleteBefore()

	case keyboard.KeyDelete:
		ti.editor.delete()

	case keyboard.KeyArrowLeft:
		ti.editor.cursorLeft()

	case keyboard.KeyArrowRight:
		ti.editor.cursorRight()

	case keyboard.KeyHome, keyboard.KeyCtrlA:
		ti.editor.cursorStart()

	case keyboard.KeyEnd, keyboard.KeyCtrlE:
		ti.editor.cursorEnd()

	case keyboard.KeyEnter:
		text := ti.editor.content()
		if ti.opts.clearOnSubmit {
			ti.editor.reset()
		}
		if ti.opts.onSubmit != nil {
			return true, text
		}

	default:
		if err := wrap.ValidText(string(k.Key)); err != nil {
			// Ignore unsupported runes.
			return false, ""
		}
		if ti.opts.filter != nil && !ti.opts.filter(rune(k.Key)) {
			// Ignore filtered runes.
			return false, ""
		}
		ti.editor.insert(rune(k.Key))
	}

	return false, ""
}

// Keyboard processes keyboard events.
// Implements widgetapi.Widget.Keyboard.
func (ti *TextInput) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	if submitted, text := ti.keyboard(k); submitted {
		// Mutex must be released when calling the callback.
		// Users might call container methods from the callback like the
		// Container.Update, see #205.
		return ti.opts.onSubmit(text)
	}
	return nil
}

// Mouse processes mouse events.
// Implements widgetapi.Widget.Mouse.
func (ti *TextInput) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	if m.Button != mouse.ButtonLeft || !m.Position.In(ti.forField) {
		return nil
	}

	cellIdx := m.Position.X - ti.forField.Min.X
	ti.editor.cursorRelCell(cellIdx)
	return nil
}

// minFieldHeight is the minimum height in cells needed for the text input field.
const minFieldHeight = 1

// Options implements widgetapi.Widget.Options.
func (ti *TextInput) Options() widgetapi.Options {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	needWidth := minFieldWidth
	if lw := runewidth.StringWidth(ti.opts.label); lw > 0 {
		needWidth += lw
	}

	needHeight := minFieldHeight
	if ti.opts.border != linestyle.None {
		needWidth += 2
		needHeight += 2
	}

	maxWidth := 0
	if ti.opts.maxWidthCells != nil {
		additional := *ti.opts.maxWidthCells - minFieldWidth
		maxWidth = needWidth + additional
	}

	return widgetapi.Options{
		MinimumSize: image.Point{
			needWidth,
			needHeight,
		},
		MaximumSize: image.Point{
			maxWidth,
			needHeight,
		},
		WantKeyboard:             widgetapi.KeyScopeFocused,
		WantMouse:                widgetapi.MouseScopeWidget,
		ExclusiveKeyboardOnFocus: ti.opts.exclusiveKeyboardOnFocus,
	}
}

// split splits the available area into label and text input areas according to
// configuration. The returned labelAr might be image.ZR if no label was
// configured.
func split(cvsAr image.Rectangle, label string, widthPerc *int) (labelAr, textAr image.Rectangle, err error) {
	switch {
	case widthPerc != nil:
		splitP := 100 - *widthPerc
		labelAr, textAr, err := area.VSplit(cvsAr, splitP)
		if err != nil {
			return image.ZR, image.ZR, err
		}
		if len(label) == 0 {
			labelAr = image.ZR
		}
		return labelAr, textAr, nil

	case len(label) > 0:
		cells := runewidth.StringWidth(label)
		labelAr, textAr, err := area.VSplitCells(cvsAr, cells)
		if err != nil {
			return image.ZR, image.ZR, err
		}
		return labelAr, textAr, nil

	default:
		// Neither a label nor width percentage specified.
		return image.ZR, cvsAr, nil
	}
}

// hideText returns the text with all runes replaced with hr.
func hideText(text string, hr rune) string {
	var b strings.Builder

	i := 0
	sw := runewidth.StringWidth(text)
	for _, r := range text {
		rw := runewidth.RuneWidth(r)
		switch {
		case i == 0 && r == '⇦':
			b.WriteRune(r)

		case i == sw-1 && r == '⇨':
			b.WriteRune(r)

		default:
			b.WriteString(strings.Repeat(string(hr), rw))
		}
		i++
	}
	return b.String()
}
