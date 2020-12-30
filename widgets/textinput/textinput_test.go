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

package textinput

import (
	"errors"
	"image"
	"sync"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/draw/testdraw"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// callbackTracker tracks whether callback was called.
type callbackTracker struct {
	// wantErr when set to true, makes callback return an error.
	wantErr bool

	// text is the text received OnSubmit.
	text string

	// count is the number of times the callback was called.
	count int

	// mu protects the tracker.
	mu sync.Mutex
}

// submit is the callback function called OnSubmit.
func (ct *callbackTracker) submit(text string) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if ct.wantErr {
		return errors.New("ct.wantErr set to true")
	}

	ct.count++
	ct.text = text
	return nil
}

func TestTextInput(t *testing.T) {
	// Makes the empty text input field visible and cursor in test outputs.
	textFieldRune = '_'
	cursorRune = '█'

	tests := []struct {
		desc         string
		callback     *callbackTracker
		opts         []Option
		events       []terminalapi.Event
		canvas       image.Rectangle
		meta         *widgetapi.Meta
		want         func(size image.Point) *faketerm.Terminal
		wantCallback *callbackTracker
		wantNewErr   bool
		wantDrawErr  bool
		wantEventErr bool
	}{
		{
			desc: "fails on WidthPerc too low",
			opts: []Option{
				WidthPerc(0),
			},
			wantNewErr: true,
		},
		{
			desc: "fails on WidthPerc too high",
			opts: []Option{
				WidthPerc(101),
			},
			wantNewErr: true,
		},
		{
			desc: "fails on MaxWidthCells too low",
			opts: []Option{
				MaxWidthCells(3),
			},
			wantNewErr: true,
		},
		{
			desc: "fails on HideTextWith control rune",
			opts: []Option{
				HideTextWith(0x007f),
			},
			wantNewErr: true,
		},
		{
			desc: "fails on HideTextWith full-width rune",
			opts: []Option{
				HideTextWith('世'),
			},
			wantNewErr: true,
		},
		{
			desc: "fails on invalid DefaultText which has control characters",
			opts: []Option{
				DefaultText("\r"),
			},
			wantNewErr: true,
		},
		{
			desc: "fails on invalid DefaultText which has newline",
			opts: []Option{
				DefaultText("\n"),
			},
			wantNewErr: true,
		},
		{
			desc:   "takes all space without label",
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					cvs.Area(),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "text field with border",
			opts: []Option{
				Border(linestyle.Light),
			},
			canvas: image.Rect(0, 0, 10, 3),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testdraw.MustBorder(cvs, cvs.Area())
				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(1, 1, 9, 2),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets custom border color",
			opts: []Option{
				Border(linestyle.Light),
				BorderColor(cell.ColorRed),
			},
			canvas: image.Rect(0, 0, 10, 3),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testdraw.MustBorder(cvs, cvs.Area(), draw.BorderCellOpts(cell.FgColor(cell.ColorRed)))
				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(1, 1, 9, 2),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets custom fill color",
			opts: []Option{
				FillColor(cell.ColorRed),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					cvs.Area(),
					textFieldRune,
					cell.BgColor(cell.ColorRed),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "draws cursor when focused",
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					cvs.Area(),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{0, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "draws place holder text when empty and not focused",
			opts: []Option{
				PlaceHolder("holder"),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: false,
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					cvs.Area(),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"holder",
					image.Point{0, 0},
					draw.TextCellOpts(cell.FgColor(cell.ColorNumber(DefaultPlaceHolderColorNumber))),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets custom place holder text color",
			opts: []Option{
				PlaceHolder("holder"),
				PlaceHolderColor(cell.ColorRed),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: false,
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					cvs.Area(),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"holder",
					image.Point{0, 0},
					draw.TextCellOpts(cell.FgColor(cell.ColorRed)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets custom cursor color",
			opts: []Option{
				CursorColor(cell.ColorRed),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					cvs.Area(),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{0, 0},
					cursorRune,
					cell.BgColor(cell.ColorRed),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets width percentage, results in area too small",
			opts: []Option{
				WidthPerc(10),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustResizeNeeded(cvs)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets width percentage, field aligns right",
			opts: []Option{
				WidthPerc(50),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(5, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "automatically adjusts space for label, rest for text field",
			opts: []Option{
				Label("hi:"),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testdraw.MustText(cvs, "hi:", image.Point{0, 0})
				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(3, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "has label and border, not enough remaining height",
			opts: []Option{
				Label("hi:"),
				Border(linestyle.Light),
			},
			canvas: image.Rect(0, 0, 10, 2),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustResizeNeeded(cvs)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "draws label and border",
			opts: []Option{
				Label("hi:"),
				Border(linestyle.Light),
			},
			canvas: image.Rect(0, 0, 10, 3),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testdraw.MustText(cvs, "hi:", image.Point{0, 1})
				testdraw.MustBorder(cvs, image.Rect(3, 0, 10, 3))
				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(4, 1, 9, 2),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "draws resize needed if label makes text field too narrow",
			opts: []Option{
				Label("hello world:"),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustResizeNeeded(cvs)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets width percentage for text field, label gets the rest, aligns left by default",
			opts: []Option{
				Label("hi:"),
				WidthPerc(50),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testdraw.MustText(cvs, "hi:", image.Point{0, 0})
				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(5, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets width percentage for text field, label gets the rest, aligns left with option",
			opts: []Option{
				Label("hi:"),
				WidthPerc(50),
				LabelAlign(align.HorizontalLeft),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testdraw.MustText(cvs, "hi:", image.Point{0, 0})
				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(5, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets width percentage for text field, label gets the rest, aligns center with option",
			opts: []Option{
				Label("hi:"),
				WidthPerc(50),
				LabelAlign(align.HorizontalCenter),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testdraw.MustText(cvs, "hi:", image.Point{1, 0})
				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(5, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets width percentage for text field, label gets the rest, aligns right with option",
			opts: []Option{
				Label("hi:"),
				WidthPerc(50),
				LabelAlign(align.HorizontalRight),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testdraw.MustText(cvs, "hi:", image.Point{2, 0})
				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(5, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets label cell options",
			opts: []Option{
				Label(
					"hi:",
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorBlue),
				),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testdraw.MustText(
					cvs,
					"hi:",
					image.Point{0, 0},
					draw.TextCellOpts(
						cell.FgColor(cell.ColorRed),
						cell.BgColor(cell.ColorBlue),
					),
				)
				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(3, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "displays default text",
			opts: []Option{
				DefaultText("text"),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"text",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "default text can be edited",
			opts: []Option{
				DefaultText("text"),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyBackspace},
				&terminalapi.Keyboard{Key: 'a'},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"texa",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "displays written text",
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "submits written text on enter",
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			callback: &callbackTracker{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{
				text:  "abc",
				count: 1,
			},
		},
		{
			desc:   "forwards error returned by SubmitFn",
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			callback: &callbackTracker{
				wantErr: true,
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{
				text:  "abc",
				count: 1,
			},
			wantEventErr: true,
		},
		{
			desc: "submits written text on enter and clears the text input field",
			opts: []Option{
				ClearOnSubmit(),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			callback: &callbackTracker{},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
			wantCallback: &callbackTracker{
				text:  "abc",
				count: 1,
			},
		},
		{
			desc: "clears the text input field when enter is pressed and ClearOnSubmit option given",
			opts: []Option{
				ClearOnSubmit(),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "write ignores control or unsupported space runes",
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: '\t'},
				&terminalapi.Keyboard{Key: 0x007f},
				&terminalapi.Keyboard{Key: 'b'},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"ab",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "write filters runes with the provided FilterFn",
			opts: []Option{
				Filter(func(r rune) bool {
					return r != 'b' && r != 'c'
				}),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: 'd'},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"ad",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},

		{
			desc:   "displays written text with full-width runes",
			canvas: image.Rect(0, 0, 4, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: '世'},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 4, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"⇦世",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "hides text when requested",
			opts: []Option{
				HideTextWith('*'),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"***",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "hides text, but doesn't hide scrolling arrows",
			opts: []Option{
				HideTextWith('*'),
			},
			canvas: image.Rect(0, 0, 4, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: 'd'},
				&terminalapi.Keyboard{Key: 'e'},
				&terminalapi.Keyboard{Key: keyboard.KeyArrowLeft},
				&terminalapi.Keyboard{Key: keyboard.KeyArrowLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 4, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"⇦**⇨",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "hides text hides scrolling arrows that are part of the text",
			opts: []Option{
				HideTextWith('*'),
			},
			canvas: image.Rect(0, 0, 4, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: '⇦'},
				&terminalapi.Keyboard{Key: '⇨'},
				&terminalapi.Keyboard{Key: 'e'},
				&terminalapi.Keyboard{Key: keyboard.KeyArrowLeft},
				&terminalapi.Keyboard{Key: keyboard.KeyArrowLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 4, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"⇦**⇨",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "hides full-width runes with two hide runes",
			opts: []Option{
				HideTextWith('*'),
			},
			canvas: image.Rect(0, 0, 4, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: '世'},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 4, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"⇦**",
					image.Point{0, 0},
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets custom text color",
			opts: []Option{
				TextColor(cell.ColorRed),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta:   &widgetapi.Meta{},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
					draw.TextCellOpts(cell.FgColor(cell.ColorRed)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "displays written text and cursor when focused",
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{3, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "moves cursor left",
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: keyboard.KeyArrowLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{2, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "sets custom highlight color",
			opts: []Option{
				HighlightedColor(cell.ColorRed),
			},
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: keyboard.KeyArrowLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{2, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorRed),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "moves cursor to start",
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: keyboard.KeyHome},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{0, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "moves cursor right",
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: keyboard.KeyHome},
				&terminalapi.Keyboard{Key: keyboard.KeyArrowRight},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{1, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "moves cursor to end",
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: keyboard.KeyHome},
				&terminalapi.Keyboard{Key: keyboard.KeyEnd},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{3, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "deletes rune the cursor is on",
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: keyboard.KeyHome},
				&terminalapi.Keyboard{Key: keyboard.KeyDelete},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"bc",
					image.Point{0, 0},
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{0, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "deletes rune just before the cursor",
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Keyboard{Key: keyboard.KeyBackspace},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"ab",
					image.Point{0, 0},
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{2, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "left mouse button moves the cursor",
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Mouse{
					Button:   mouse.ButtonLeft,
					Position: image.Point{1, 0},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{1, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "ignores other mouse buttons",
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Mouse{
					Button:   mouse.ButtonRight,
					Position: image.Point{1, 0},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{3, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:   "ignores mouse events outside of the text field",
			canvas: image.Rect(0, 0, 10, 1),
			meta: &widgetapi.Meta{
				Focused: true,
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
				&terminalapi.Mouse{
					Button:   mouse.ButtonLeft,
					Position: image.Point{5, 15},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetAreaCells(
					cvs,
					image.Rect(0, 0, 10, 1),
					textFieldRune,
					cell.BgColor(cell.ColorNumber(DefaultFillColorNumber)),
				)
				testdraw.MustText(
					cvs,
					"abc",
					image.Point{0, 0},
				)
				testcanvas.MustSetCell(
					cvs,
					image.Point{3, 0},
					cursorRune,
					cell.BgColor(cell.ColorNumber(DefaultCursorColorNumber)),
					cell.FgColor(cell.ColorNumber(DefaultHighlightedColorNumber)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotCallback := tc.callback
			if gotCallback != nil {
				tc.opts = append(tc.opts, OnSubmit(gotCallback.submit))
			}

			ti, err := New(tc.opts...)
			if (err != nil) != tc.wantNewErr {
				t.Errorf("New => unexpected error: %v, wantNewErr: %v", err, tc.wantNewErr)
			}
			if err != nil {
				return
			}

			{
				// Draw once so mouse events are acceptable.
				c, err := canvas.New(tc.canvas)
				if err != nil {
					t.Fatalf("canvas.New => unexpected error: %v", err)
				}

				err = ti.Draw(c, tc.meta)
				if (err != nil) != tc.wantDrawErr {
					t.Errorf("Draw => unexpected error: %v, wantDrawErr: %v", err, tc.wantDrawErr)
				}
				if err != nil {
					return
				}
			}

			for i, ev := range tc.events {
				switch e := ev.(type) {
				case *terminalapi.Mouse:
					err := ti.Mouse(e, &widgetapi.EventMeta{})
					// Only the last event in test cases is the one that triggers the callback.
					if i == len(tc.events)-1 {
						if (err != nil) != tc.wantEventErr {
							t.Errorf("Mouse => unexpected error: %v, wantEventErr: %v", err, tc.wantEventErr)
						}
						if err != nil {
							return
						}
					} else {
						if err != nil {
							t.Fatalf("Mouse => unexpected error: %v", err)
						}
					}

				case *terminalapi.Keyboard:
					err := ti.Keyboard(e, &widgetapi.EventMeta{})
					// Only the last event in test cases is the one that triggers the callback.
					if i == len(tc.events)-1 {
						if (err != nil) != tc.wantEventErr {
							t.Errorf("Keyboard => unexpected error: %v, wantEventErr: %v", err, tc.wantEventErr)
						}
						if err != nil {
							return
						}
					} else {
						if err != nil {
							t.Fatalf("Keyboard => unexpected error: %v", err)
						}
					}

				default:
					t.Fatalf("unsupported event type: %T", ev)
				}
			}

			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			{
				err = ti.Draw(c, tc.meta)
				if (err != nil) != tc.wantDrawErr {
					t.Errorf("Draw => unexpected error: %v, wantDrawErr: %v", err, tc.wantDrawErr)
				}
				if err != nil {
					return
				}
			}

			got, err := faketerm.New(c.Size())
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			if err := c.Apply(got); err != nil {
				t.Fatalf("Apply => unexpected error: %v", err)
			}

			var want *faketerm.Terminal
			if tc.want != nil {
				want = tc.want(c.Size())
			} else {
				want = faketerm.MustNew(c.Size())
			}

			if diff := faketerm.Diff(want, got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}

			if diff := pretty.Compare(tc.wantCallback, gotCallback); diff != "" {
				t.Errorf("CallbackFn => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestTextInputRead(t *testing.T) {
	tests := []struct {
		desc   string
		events []terminalapi.Event
		want   string
	}{
		{
			desc:   "reads empty without events",
			events: []terminalapi.Event{},
			want:   "",
		},
		{
			desc: "reads written text",
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: 'a'},
				&terminalapi.Keyboard{Key: 'b'},
				&terminalapi.Keyboard{Key: 'c'},
			},
			want: "abc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ti, err := New()
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			for _, ev := range tc.events {
				switch e := ev.(type) {
				case *terminalapi.Keyboard:
					err := ti.Keyboard(e, &widgetapi.EventMeta{})
					if err != nil {
						t.Fatalf("Keyboard => unexpected error: %v", err)
					}

				default:
					t.Fatalf("unsupported event type: %T", ev)
				}
			}

			got := ti.Read()
			if got != tc.want {
				t.Errorf("Read => %q, want %q", got, tc.want)
			}

			gotRC := ti.ReadAndClear()
			if gotRC != tc.want {
				t.Errorf("ReadAndClear after clearing => %q, want %q", gotRC, tc.want)
			}

			// Both should now return empty content.
			{
				want := ""
				got := ti.Read()
				if got != want {
					t.Errorf("Read after clearing => %q, want %q", got, want)
				}

				gotRC := ti.ReadAndClear()
				if gotRC != want {
					t.Errorf("ReadAndClear after clearing => %q, want %q", gotRC, want)
				}
			}
		})
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		desc string
		opts []Option
		want widgetapi.Options
	}{
		{
			desc: "no label and no border",
			want: widgetapi.Options{
				MinimumSize:  image.Point{4, 1},
				MaximumSize:  image.Point{0, 1},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "no label and no border, max width specified",
			opts: []Option{
				MaxWidthCells(5),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{4, 1},
				MaximumSize:  image.Point{5, 1},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "no label, has border",
			opts: []Option{
				Border(linestyle.Light),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{6, 3},
				MaximumSize:  image.Point{0, 3},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "no label, has border, max width specified",
			opts: []Option{
				Border(linestyle.Light),
				MaxWidthCells(5),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{6, 3},
				MaximumSize:  image.Point{7, 3},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "has label and no border",
			opts: []Option{
				Label("hello"),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{9, 1},
				MaximumSize:  image.Point{0, 1},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "has label and no border, max width specified",
			opts: []Option{
				Label("hello"),
				MaxWidthCells(5),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{9, 1},
				MaximumSize:  image.Point{10, 1},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "has label with full-width runes and no border",
			opts: []Option{
				Label("hello世"),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{11, 1},
				MaximumSize:  image.Point{0, 1},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "has label and border",
			opts: []Option{
				Label("hello"),
				Border(linestyle.Light),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{11, 3},
				MaximumSize:  image.Point{0, 3},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "has label and border, max width specified",
			opts: []Option{
				Label("hello"),
				Border(linestyle.Light),
				MaxWidthCells(5),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{11, 3},
				MaximumSize:  image.Point{12, 3},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "requests ExclusiveKeyboardOnFocus",
			opts: []Option{
				ExclusiveKeyboardOnFocus(),
			},
			want: widgetapi.Options{
				MinimumSize:              image.Point{4, 1},
				MaximumSize:              image.Point{0, 1},
				WantKeyboard:             widgetapi.KeyScopeFocused,
				WantMouse:                widgetapi.MouseScopeWidget,
				ExclusiveKeyboardOnFocus: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ti, err := New(tc.opts...)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			got := ti.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		desc        string
		cvsAr       image.Rectangle
		label       string
		widthPerc   *int
		wantLabelAr image.Rectangle
		wantTextAr  image.Rectangle
		wantErr     bool
	}{
		{
			desc:  "fails on invalid widthPerc",
			cvsAr: image.Rect(0, 0, 10, 1),
			widthPerc: func() *int {
				i := -1
				return &i
			}(),
			wantErr: true,
		},
		{
			desc:        "no label and no widthPerc, full area for text input field",
			cvsAr:       image.Rect(0, 0, 5, 1),
			wantLabelAr: image.ZR,
			wantTextAr:  image.Rect(0, 0, 5, 1),
		},
		{
			desc:  "widthPerc set, splits canvas area",
			cvsAr: image.Rect(0, 0, 10, 1),
			widthPerc: func() *int {
				i := 30
				return &i
			}(),
			wantLabelAr: image.ZR,
			wantTextAr:  image.Rect(7, 0, 10, 1),
		},
		{
			desc:  "widthPerc and label set",
			cvsAr: image.Rect(0, 0, 10, 1),
			widthPerc: func() *int {
				i := 30
				return &i
			}(),
			label:       "hello",
			wantLabelAr: image.Rect(0, 0, 7, 1),
			wantTextAr:  image.Rect(7, 0, 10, 1),
		},

		{
			desc:  "widthPerc set to 100, splits canvas area",
			cvsAr: image.Rect(0, 0, 10, 1),
			widthPerc: func() *int {
				i := 100
				return &i
			}(),
			wantLabelAr: image.ZR,
			wantTextAr:  image.Rect(0, 0, 10, 1),
		},
		{
			desc:  "widthPerc set to 1, splits canvas area",
			cvsAr: image.Rect(0, 0, 10, 1),
			widthPerc: func() *int {
				i := 1
				return &i
			}(),
			wantLabelAr: image.ZR,
			wantTextAr:  image.Rect(9, 0, 10, 1),
		},
		{
			desc:        "label set, half-width runes only",
			cvsAr:       image.Rect(0, 0, 10, 1),
			label:       "hello",
			wantLabelAr: image.Rect(0, 0, 5, 1),
			wantTextAr:  image.Rect(5, 0, 10, 1),
		},
		{
			desc:        "label set, full-width runes",
			cvsAr:       image.Rect(0, 0, 10, 1),
			label:       "hello世",
			wantLabelAr: image.Rect(0, 0, 7, 1),
			wantTextAr:  image.Rect(7, 0, 10, 1),
		},
		{
			desc:        "label longer than canvas width",
			cvsAr:       image.Rect(0, 0, 10, 1),
			label:       "helloworld1",
			wantLabelAr: image.Rect(0, 0, 10, 1),
			wantTextAr:  image.ZR,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotLabelAr, gotTextAr, err := split(tc.cvsAr, tc.label, tc.widthPerc)
			if (err != nil) != tc.wantErr {
				t.Errorf("split => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.wantLabelAr, gotLabelAr); diff != "" {
				t.Errorf("split => unexpected labelAr, diff (-want, +got):\n%s", diff)
			}
			if diff := pretty.Compare(tc.wantTextAr, gotTextAr); diff != "" {
				t.Errorf("split => unexpected labelAr, diff (-want, +got):\n%s", diff)
			}
		})
	}
}
