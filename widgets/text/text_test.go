package text

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/draw/testdraw"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/terminalapi"
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
		wantWriteErr bool
	}{
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
			desc:   "multiple writes append",
			canvas: image.Rect(0, 0, 12, 1),
			writes: func(widget *Text) error {
				if err := widget.Write("hello"); err != nil {
					return err
				}
				if err := widget.Write(" "); err != nil {
					return err
				}
				if err := widget.Write("world"); err != nil {
					return err
				}
				return nil
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
				if err := widget.Write("world"); err != nil {
					return err
				}
				return nil
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
				if err := widget.Write("blue\n", WriteCellOpts(cell.FgColor(cell.ColorBlue))); err != nil {
					return err
				}
				return nil
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
			canvas: image.Rect(0, 0, 10, 3),
			writes: func(widget *Text) error {
				return widget.Write("hello world\nshort\nand long again")
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testdraw.MustText(c, "hello wor…", image.Point{0, 0})
				testdraw.MustText(c, "short", image.Point{0, 1})
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
			desc:   "wraps lines at rune boundaries",
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

			widget := New(tc.opts...)
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
				WantKeyboard: true,
				WantMouse:    true,
			},
		},
		{
			desc: "disabling scrolling removes keyboard and mouse",
			opts: []Option{
				DisableScrolling(),
			},
			want: widgetapi.Options{
				MinimumSize:  image.Point{1, 1},
				WantKeyboard: false,
				WantMouse:    false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			text := New(tc.opts...)
			got := text.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
