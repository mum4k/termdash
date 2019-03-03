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

package text

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/canvas/testcanvas"
	"github.com/mum4k/termdash/internal/draw"
	"github.com/mum4k/termdash/internal/draw/testdraw"
	"github.com/mum4k/termdash/internal/faketerm"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

func TestTextDraws(t *testing.T) {
	tests := []struct {
		desc         string
		canvas       image.Rectangle
		opts         []Option
		writes       func(*Text) error
		events       func(*Text)
		want         func(size image.Point) *faketerm.Terminal
		wantErr      bool
		wantWriteErr bool
	}{
		{
			desc: "fails when scroll keys aren't unique",
			opts: []Option{
				ScrollKeys('a', 'a', 'a', 'a'),
			},
			canvas: image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "fails when scroll mouse buttons aren't unique",
			opts: []Option{
				ScrollMouseButtons(mouse.ButtonLeft, mouse.ButtonLeft),
			},
			canvas: image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "empty when no written text",
			canvas: image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:   "write fails for invalid text",
			canvas: image.Rect(0, 0, 1, 1),
			writes: func(widget *Text) error {
				return widget.Write("\thello")
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantWriteErr: true,
		},
		{
			desc:   "draws line of text",
			canvas: image.Rect(0, 0, 10, 1),
			writes: func(widget *Text) error {
				return widget.Write("hello")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "hello", image.Point{0, 0})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws line of full-width runes",
			canvas: image.Rect(0, 0, 10, 1),
			writes: func(widget *Text) error {
				return widget.Write("你好，世界")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "你好，世界", image.Point{0, 0})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "multiple writes append",
			canvas: image.Rect(0, 0, 12, 1),
			writes: func(widget *Text) error {
				if err := widget.Write("hello"); err != nil {
					return err
				}
				if err := widget.Write(" "); err != nil {
					return err
				}
				return widget.Write("world")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "hello world", image.Point{0, 0})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "multiple writes replace when requested",
			canvas: image.Rect(0, 0, 12, 1),
			writes: func(widget *Text) error {
				if err := widget.Write("hello", WriteReplace()); err != nil {
					return err
				}
				return widget.Write("world", WriteReplace())
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "world", image.Point{0, 0})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "reset clears the content",
			canvas: image.Rect(0, 0, 12, 1),
			writes: func(widget *Text) error {
				if err := widget.Write("hello", WriteCellOpts(cell.FgColor(cell.ColorRed))); err != nil {
					return err
				}
				if err := widget.Write(" "); err != nil {
					return err
				}
				widget.Reset()
				return widget.Write("world")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "world", image.Point{0, 0})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "respects newlines in the input text",
			canvas: image.Rect(0, 0, 10, 10),
			writes: func(widget *Text) error {
				return widget.Write("\n\nhello\n\nworld\n\n")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "hello", image.Point{0, 2})
				testdraw.MustText(c, "world", image.Point{0, 4})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "respects write options",
			canvas: image.Rect(0, 0, 10, 3),
			writes: func(widget *Text) error {
				if err := widget.Write("default\n"); err != nil {
					return err
				}
				if err := widget.Write("red\n", WriteCellOpts(cell.FgColor(cell.ColorRed))); err != nil {
					return err
				}
				return widget.Write("blue", WriteCellOpts(cell.FgColor(cell.ColorBlue)))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "default", image.Point{0, 0})
				testdraw.MustText(c, "red", image.Point{0, 1}, draw.TextCellOpts(cell.FgColor(cell.ColorRed)))
				testdraw.MustText(c, "blue", image.Point{0, 2}, draw.TextCellOpts(cell.FgColor(cell.ColorBlue)))
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "trims long lines",
			canvas: image.Rect(0, 0, 10, 4),
			writes: func(widget *Text) error {
				return widget.Write("hello world\nshort\nexactly 10\nand long again")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "hello wor…", image.Point{0, 0})
				testdraw.MustText(c, "short", image.Point{0, 1})
				testdraw.MustText(c, "exactly 10", image.Point{0, 2})
				testdraw.MustText(c, "and long …", image.Point{0, 3})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "trims long lines with full-width runes",
			canvas: image.Rect(0, 0, 10, 3),
			writes: func(widget *Text) error {
				return widget.Write("hello wor你\nhello wor你d\nand long 世")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "hello wor…", image.Point{0, 0})
				testdraw.MustText(c, "hello wor…", image.Point{0, 1})
				testdraw.MustText(c, "and long …", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "trims content when longer than canvas, no scroll marker on small canvas",
			canvas: image.Rect(0, 0, 10, 2),
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "line0", image.Point{0, 0})
				testdraw.MustText(c, "line1", image.Point{0, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "trims content when longer than canvas and draws bottom scroll marker",
			canvas: image.Rect(0, 0, 10, 3),
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "line0", image.Point{0, 0})
				testdraw.MustText(c, "line1", image.Point{0, 1})
				testdraw.MustText(c, "⇩", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls down on mouse wheel down a line at a time",
			canvas: image.Rect(0, 0, 10, 3),
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3")
			},
			events: func(widget *Text) {
				widget.Mouse(&terminalapi.Mouse{
					Button: mouse.ButtonWheelDown,
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "⇧", image.Point{0, 0})
				testdraw.MustText(c, "line2", image.Point{0, 1})
				testdraw.MustText(c, "line3", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "doesn't draw the scroll up marker on small canvas",
			canvas: image.Rect(0, 0, 10, 2),
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3")
			},
			events: func(widget *Text) {
				widget.Mouse(&terminalapi.Mouse{
					Button: mouse.ButtonWheelDown,
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "line1", image.Point{0, 0})
				testdraw.MustText(c, "line2", image.Point{0, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls down on down arrow a line at a time",
			canvas: image.Rect(0, 0, 10, 3),
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3")
			},
			events: func(widget *Text) {
				widget.Keyboard(&terminalapi.Keyboard{
					Key: keyboard.KeyArrowDown,
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "⇧", image.Point{0, 0})
				testdraw.MustText(c, "line2", image.Point{0, 1})
				testdraw.MustText(c, "line3", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls down on pageDn a page at a time",
			canvas: image.Rect(0, 0, 10, 3),
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3\nline4\nline5\nline6")
			},
			events: func(widget *Text) {
				widget.Keyboard(&terminalapi.Keyboard{
					Key: keyboard.KeyPgDn,
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "⇧", image.Point{0, 0})
				testdraw.MustText(c, "line4", image.Point{0, 1})
				testdraw.MustText(c, "⇩", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls down using custom mouse button a line at a time",
			canvas: image.Rect(0, 0, 10, 3),
			opts: []Option{
				ScrollMouseButtons(mouse.ButtonLeft, mouse.ButtonRight),
			},
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3")
			},
			events: func(widget *Text) {
				widget.Mouse(&terminalapi.Mouse{
					Button: mouse.ButtonRight,
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "⇧", image.Point{0, 0})
				testdraw.MustText(c, "line2", image.Point{0, 1})
				testdraw.MustText(c, "line3", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls down using custom key a line at a time",
			canvas: image.Rect(0, 0, 10, 3),
			opts: []Option{
				ScrollKeys('u', 'd', 'k', 'l'),
			},
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3")
			},
			events: func(widget *Text) {
				widget.Keyboard(&terminalapi.Keyboard{
					Key: 'd',
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "⇧", image.Point{0, 0})
				testdraw.MustText(c, "line2", image.Point{0, 1})
				testdraw.MustText(c, "line3", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls down using custom key a page at a time",
			canvas: image.Rect(0, 0, 10, 3),
			opts: []Option{
				ScrollKeys('u', 'd', 'k', 'l'),
			},
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3\nline4\nline5\nline6")
			},
			events: func(widget *Text) {
				widget.Keyboard(&terminalapi.Keyboard{
					Key: 'l',
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "⇧", image.Point{0, 0})
				testdraw.MustText(c, "line4", image.Point{0, 1})
				testdraw.MustText(c, "⇩", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "wraps lines at half-width rune boundaries",
			canvas: image.Rect(0, 0, 10, 5),
			opts: []Option{
				WrapAtRunes(),
			},
			writes: func(widget *Text) error {
				return widget.Write("hello world\nshort\nand long again")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "hello worl", image.Point{0, 0})
				testdraw.MustText(c, "d", image.Point{0, 1})
				testdraw.MustText(c, "short", image.Point{0, 2})
				testdraw.MustText(c, "and long a", image.Point{0, 3})
				testdraw.MustText(c, "gain", image.Point{0, 4})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "wraps lines at full-width rune boundaries",
			canvas: image.Rect(0, 0, 10, 6),
			opts: []Option{
				WrapAtRunes(),
			},
			writes: func(widget *Text) error {
				return widget.Write("hello wor你\nhello wor你d\nand long 世")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "hello wor", image.Point{0, 0})
				testdraw.MustText(c, "你", image.Point{0, 1})
				testdraw.MustText(c, "hello wor", image.Point{0, 2})
				testdraw.MustText(c, "你d", image.Point{0, 3})
				testdraw.MustText(c, "and long ", image.Point{0, 4})
				testdraw.MustText(c, "世", image.Point{0, 5})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "wraps lines at word boundaries",
			canvas: image.Rect(0, 0, 10, 6),
			opts: []Option{
				WrapAtWords(),
			},
			writes: func(widget *Text) error {
				return widget.Write("hello wor你\nhello wor你d\nand long 世")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "hello", image.Point{0, 0})
				testdraw.MustText(c, "wor你", image.Point{0, 1})
				testdraw.MustText(c, "hello", image.Point{0, 2})
				testdraw.MustText(c, "wor你d", image.Point{0, 3})
				testdraw.MustText(c, "and long", image.Point{0, 4})
				testdraw.MustText(c, "世", image.Point{0, 5})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "wraps lines at word boundaries, inserts dash for long words",
			canvas: image.Rect(0, 0, 10, 6),
			opts: []Option{
				WrapAtWords(),
			},
			writes: func(widget *Text) error {
				return widget.Write("hello thisisalongword world")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "hello", image.Point{0, 0})
				testdraw.MustText(c, "thisisalo-", image.Point{0, 1})
				testdraw.MustText(c, "ngword", image.Point{0, 2})
				testdraw.MustText(c, "world", image.Point{0, 3})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "rolls content upwards and trims lines",
			canvas: image.Rect(0, 0, 10, 2),
			opts: []Option{
				RollContent(),
			},
			writes: func(widget *Text) error {
				return widget.Write("hello world\nshort\nand long again")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "short", image.Point{0, 0})
				testdraw.MustText(c, "and long …", image.Point{0, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "rolls content upwards and draws an up scroll marker",
			canvas: image.Rect(0, 0, 10, 3),
			opts: []Option{
				RollContent(),
			},
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "⇧", image.Point{0, 0})
				testdraw.MustText(c, "line2", image.Point{0, 1})
				testdraw.MustText(c, "line3", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "rolls content upwards and wraps lines at rune boundaries",
			canvas: image.Rect(0, 0, 10, 2),
			opts: []Option{
				RollContent(),
				WrapAtRunes(),
			},
			writes: func(widget *Text) error {
				return widget.Write("hello world\nshort\nand long again")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "and long a", image.Point{0, 0})
				testdraw.MustText(c, "gain", image.Point{0, 1})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls up on mouse wheel up a line at a time",
			canvas: image.Rect(0, 0, 10, 3),
			opts: []Option{
				RollContent(),
			},
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3\nline4")
			},
			events: func(widget *Text) {
				// Draw once to roll the content all the way down before we scroll.
				if err := widget.Draw(testcanvas.MustNew(image.Rect(0, 0, 10, 3))); err != nil {
					panic(err)
				}
				widget.Mouse(&terminalapi.Mouse{
					Button: mouse.ButtonWheelUp,
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "⇧", image.Point{0, 0})
				testdraw.MustText(c, "line2", image.Point{0, 1})
				testdraw.MustText(c, "⇩", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls up on up arrow a line at a time",
			canvas: image.Rect(0, 0, 10, 3),
			opts: []Option{
				RollContent(),
			},
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3\nline4")
			},
			events: func(widget *Text) {
				// Draw once to roll the content all the way down before we scroll.
				if err := widget.Draw(testcanvas.MustNew(image.Rect(0, 0, 10, 3))); err != nil {
					panic(err)
				}
				widget.Keyboard(&terminalapi.Keyboard{
					Key: keyboard.KeyArrowUp,
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "⇧", image.Point{0, 0})
				testdraw.MustText(c, "line2", image.Point{0, 1})
				testdraw.MustText(c, "⇩", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls up on pageUp a page at a time",
			canvas: image.Rect(0, 0, 10, 3),
			opts: []Option{
				RollContent(),
			},
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3\nline4")
			},
			events: func(widget *Text) {
				// Draw once to roll the content all the way down before we scroll.
				if err := widget.Draw(testcanvas.MustNew(image.Rect(0, 0, 10, 3))); err != nil {
					panic(err)
				}
				widget.Keyboard(&terminalapi.Keyboard{
					Key: keyboard.KeyPgUp,
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "line0", image.Point{0, 0})
				testdraw.MustText(c, "line1", image.Point{0, 1})
				testdraw.MustText(c, "⇩", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls up using custom mouse button a line at a time",
			canvas: image.Rect(0, 0, 10, 3),
			opts: []Option{
				RollContent(),
				ScrollMouseButtons(mouse.ButtonLeft, mouse.ButtonRight),
			},
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3\nline4")
			},
			events: func(widget *Text) {
				// Draw once to roll the content all the way down before we scroll.
				if err := widget.Draw(testcanvas.MustNew(image.Rect(0, 0, 10, 3))); err != nil {
					panic(err)
				}
				widget.Mouse(&terminalapi.Mouse{
					Button: mouse.ButtonLeft,
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "⇧", image.Point{0, 0})
				testdraw.MustText(c, "line2", image.Point{0, 1})
				testdraw.MustText(c, "⇩", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls up using custom key a line at a time",
			canvas: image.Rect(0, 0, 10, 3),
			opts: []Option{
				ScrollKeys('u', 'd', 'k', 'l'),
				RollContent(),
			},
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3\nline4")
			},
			events: func(widget *Text) {
				// Draw once to roll the content all the way down before we scroll.
				if err := widget.Draw(testcanvas.MustNew(image.Rect(0, 0, 10, 3))); err != nil {
					panic(err)
				}
				widget.Keyboard(&terminalapi.Keyboard{
					Key: 'u',
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "⇧", image.Point{0, 0})
				testdraw.MustText(c, "line2", image.Point{0, 1})
				testdraw.MustText(c, "⇩", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "scrolls up using custom key a page at a time",
			canvas: image.Rect(0, 0, 10, 3),
			opts: []Option{
				ScrollKeys('u', 'd', 'k', 'l'),
				RollContent(),
			},
			writes: func(widget *Text) error {
				return widget.Write("line0\nline1\nline2\nline3\nline4")
			},
			events: func(widget *Text) {
				// Draw once to roll the content all the way down before we scroll.
				if err := widget.Draw(testcanvas.MustNew(image.Rect(0, 0, 10, 3))); err != nil {
					panic(err)
				}
				widget.Keyboard(&terminalapi.Keyboard{
					Key: 'k',
				})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "line0", image.Point{0, 0})
				testdraw.MustText(c, "line1", image.Point{0, 1})
				testdraw.MustText(c, "⇩", image.Point{0, 2})
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			widget, err := New(tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("New => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if tc.writes != nil {
				err := tc.writes(widget)
				if (err != nil) != tc.wantWriteErr {
					t.Errorf("Write => unexpected error: %v, wantWriteErr: %v", err, tc.wantWriteErr)
				}
				if err != nil {
					return
				}
			}

			if tc.events != nil {
				tc.events(widget)
			}

			if err := widget.Draw(c); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			got, err := faketerm.New(c.Size())
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			if err := c.Apply(got); err != nil {
				t.Fatalf("Apply => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(c.Size()), got); diff != "" {
				t.Errorf("Draw => %v", diff)
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
			desc: "minimum size for one character",
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: widgetapi.KeyScopeFocused,
				WantMouse:    widgetapi.MouseScopeWidget,
			},
		},
		{
			desc: "disabling scrolling removes keyboard and mouse",
			opts: []Option{
				DisableScrolling(),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: widgetapi.KeyScopeNone,
				WantMouse:    widgetapi.MouseScopeNone,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			text, err := New(tc.opts...)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			got := text.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
