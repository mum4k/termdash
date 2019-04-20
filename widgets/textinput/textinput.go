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
	"sync"
	"unicode"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/alignfor"
	"github.com/mum4k/termdash/internal/area"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/draw"
	"github.com/mum4k/termdash/internal/runewidth"
	"github.com/mum4k/termdash/internal/wrap"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// FilterFn if provided can be used to filter runes that are allowed in the
// text input field. Any rune for which this function returns false will be
// rejected.
type FilterFn func(rune) bool

// SubmitFn if provided is called when the user submits the content of the text
// input field, the argument text contains all the text in the field.
// Submitting the input field clears its content.
//
// The callback function must be thread-safe as the keyboard event that
// triggers the submission comes from a separate goroutine.
type SubmitFn func(text string) error

// TextInput accepts text input from the user.
//
// Displays an input field where the user can edit text and an optional label.
//
// The text can be submitted by pressing enter or read at any time by calling
// Read.
//
// Implements widgetapi.Widget. This object is thread-safe.
type TextInput struct {
	// mu protects the widget.
	mu sync.Mutex

	// editor tracks the edits and the state of the text input field.
	editor *fieldEditor

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
	return &TextInput{
		editor: newFieldEditor(),
		opts:   opt,
	}, nil
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

// Draw draws the TextInput widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (ti *TextInput) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	labelAr, textAr, err := split(cvs.Area(), ti.opts.label, ti.opts.widthPerc)
	if err != nil {
		return err
	}

	var forField image.Rectangle
	if ti.opts.border != linestyle.None {
		forField = area.ExcludeBorder(textAr)
	} else {
		forField = textAr
	}

	if forField.Dx() < minFieldWidth {
		return draw.ResizeNeeded(cvs)
	}

	if !labelAr.Eq(image.ZR) {
		start, err := alignfor.Text(labelAr, ti.opts.label, ti.opts.labelAlign, align.VerticalMiddle)
		if err != nil {
			return err
		}
		if err := draw.Text(
			cvs, ti.opts.label, start,
			draw.TextOverrunMode(draw.OverrunModeThreeDot),
			draw.TextMaxX(labelAr.Max.X),
			draw.TextCellOpts(ti.opts.labelCellOpts...),
		); err != nil {
			return err
		}
	}

	if ti.opts.border != linestyle.None {
		if err := draw.Border(cvs, textAr, draw.BorderCellOpts(cell.FgColor(ti.opts.borderColor))); err != nil {
			return err
		}
	}

	if err := cvs.SetAreaCellOpts(forField, cell.BgColor(ti.opts.fillColor)); err != nil {
		return err
	}

	text, curPos, err := ti.editor.viewFor(forField.Dx())
	if err != nil {
		return err
	}
	if err := draw.Text(
		cvs, text, forField.Min,
		draw.TextMaxX(forField.Max.X),
		draw.TextCellOpts(cell.FgColor(ti.opts.textColor)),
	); err != nil {
		return err
	}

	if meta.Focused {
		p := image.Point{
			curPos + forField.Min.X,
			forField.Min.Y,
		}
		if err := cvs.SetCellOpts(
			p,
			cell.FgColor(ti.opts.highlightedColor),
			cell.BgColor(ti.opts.cursorColor),
		); err != nil {
			return err
		}
	}

	return nil
}

// Keyboard processes keyboard events.
// Implements widgetapi.Widget.Keyboard.
func (ti *TextInput) Keyboard(k *terminalapi.Keyboard) error {
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

	default:
		if err := wrap.ValidText(string(k.Key)); err != nil {
			// Ignore unsupported runes.
			return nil
		}
		if !unicode.IsPrint(rune(k.Key)) {
			return nil
		}
		ti.editor.insert(rune(k.Key))
	}

	return nil
}

// Mouse processes mouse events.
// Implements widgetapi.Widget.Mouse.
func (ti *TextInput) Mouse(m *terminalapi.Mouse) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	return nil
}

// Options implements widgetapi.Widget.Options.
func (ti *TextInput) Options() widgetapi.Options {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	needWidth := minFieldWidth
	if lw := runewidth.StringWidth(ti.opts.label); lw > 0 {
		needWidth += lw
	}

	needHeight := 1
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
		WantKeyboard: widgetapi.KeyScopeFocused,
		WantMouse:    widgetapi.MouseScopeWidget,
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
