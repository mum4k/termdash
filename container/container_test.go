package container

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/draw/testdraw"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/fakewidget"
)

// Example demonstrates how to use the Container API.
func Example() {
	New(
		/* terminal = */ nil,
		SplitVertical(
			Left(
				SplitHorizontal(
					Top(
						Border(draw.LineStyleLight),
					),
					Bottom(
						SplitHorizontal(
							Top(
								Border(draw.LineStyleLight),
							),
							Bottom(
								Border(draw.LineStyleLight),
							),
						),
					),
				),
			),
			Right(
				Border(draw.LineStyleLight),
			),
		),
	)

	// TODO(mum4k): Allow splits on different ratios.
	// TODO(mum4k): Include an example with a widget.
}

func TestNew(t *testing.T) {
	tests := []struct {
		desc      string
		termSize  image.Point
		container func(ft *faketerm.Terminal) *Container
		want      func(size image.Point) *faketerm.Terminal
	}{
		{
			desc:     "empty container",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(ft)
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "container with a border",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal split, children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(draw.LineStyleLight),
						),
						Bottom(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(cvs, image.Rect(0, 0, 10, 5), draw.LineStyleLight)
				testdraw.MustBox(cvs, image.Rect(0, 5, 10, 10), draw.LineStyleLight)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal split, parent and children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					SplitHorizontal(
						Top(
							Border(draw.LineStyleLight),
						),
						Bottom(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)
				testdraw.MustBox(cvs, image.Rect(1, 1, 9, 5), draw.LineStyleLight)
				testdraw.MustBox(cvs, image.Rect(1, 5, 9, 9), draw.LineStyleLight)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical split, children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(cvs, image.Rect(0, 0, 5, 10), draw.LineStyleLight)
				testdraw.MustBox(cvs, image.Rect(5, 0, 10, 10), draw.LineStyleLight)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical split, parent and children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)
				testdraw.MustBox(cvs, image.Rect(1, 1, 5, 9), draw.LineStyleLight)
				testdraw.MustBox(cvs, image.Rect(5, 1, 9, 9), draw.LineStyleLight)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "multi level split",
			termSize: image.Point{10, 16},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					SplitVertical(
						Left(
							SplitHorizontal(
								Top(
									Border(draw.LineStyleLight),
								),
								Bottom(
									SplitHorizontal(
										Top(
											Border(draw.LineStyleLight),
										),
										Bottom(
											Border(draw.LineStyleLight),
										),
									),
								),
							),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(cvs, image.Rect(0, 0, 5, 8), draw.LineStyleLight)
				testdraw.MustBox(cvs, image.Rect(0, 8, 5, 12), draw.LineStyleLight)
				testdraw.MustBox(cvs, image.Rect(0, 12, 5, 16), draw.LineStyleLight)
				testdraw.MustBox(cvs, image.Rect(5, 0, 10, 16), draw.LineStyleLight)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "inherits border and focused color",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					BorderColor(cell.ColorRed),
					FocusedColor(cell.ColorBlue),
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorBlue),
				)
				testdraw.MustBox(
					cvs,
					image.Rect(1, 1, 5, 9),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorRed),
				)
				testdraw.MustBox(
					cvs,
					image.Rect(5, 1, 9, 9),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorRed),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "splitting a container removes the widget",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(
					cvs,
					ft.Area(),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)
				testdraw.MustBox(cvs, image.Rect(1, 1, 5, 9), draw.LineStyleLight)
				testdraw.MustBox(cvs, image.Rect(5, 1, 9, 9), draw.LineStyleLight)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "placing a widget removes container split",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
					),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(cvs, image.Rect(0, 0, 10, 10), draw.LineStyleLight)
				testdraw.MustText(cvs, "(10,10)", draw.TextBounds{Start: image.Point{1, 1}})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			if err := tc.container(got).Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(tc.termSize), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}

}

func TestKeyboard(t *testing.T) {
	tests := []struct {
		desc      string
		termSize  image.Point
		container func(ft *faketerm.Terminal) *Container
		events    []terminalapi.Event
		want      func(size image.Point) *faketerm.Terminal
		wantErr   bool
	}{
		{
			desc:     "event not forwarded if container has no widget",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(ft)
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "event forwarded to focused container only",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					SplitVertical(
						Left(
							PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: true})),
						),
						Right(
							SplitHorizontal(
								Top(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: true})),
								),
								Bottom(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: true})),
								),
							),
						),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{39, 19}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{39, 19}, Button: mouse.ButtonRelease},
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				// Widgets that aren't focused don't get the key.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 20, 20)),
					widgetapi.Options{WantKeyboard: true},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(20, 0, 40, 10)),
					widgetapi.Options{WantKeyboard: true},
				)

				// The focused widget receives the key.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(20, 10, 40, 20)),
					widgetapi.Options{WantKeyboard: true},
					&terminalapi.Keyboard{Key: keyboard.KeyEnter},
				)
				return ft
			},
		},
		{
			desc:     "event not forwarded if the widget didn't request it",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: false})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEnter},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "widget returns an error when processing the event",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantKeyboard: true})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Keyboard{Key: keyboard.KeyEsc},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{},
				)
				return ft
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			c := tc.container(got)
			for _, ev := range tc.events {
				switch e := ev.(type) {
				case *terminalapi.Mouse:
					if err := c.Mouse(e); err != nil {
						t.Fatalf("Mouse => unexpected error: %v", err)
					}

				case *terminalapi.Keyboard:
					err := c.Keyboard(e)
					if (err != nil) != tc.wantErr {
						t.Fatalf("Keyboard => unexpected error: %v, wantErr: %v", err, tc.wantErr)
					}

				default:
					t.Fatalf("Unsupported event %T.", e)
				}
			}

			if err := c.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(tc.termSize), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}
}

func TestMouse(t *testing.T) {
	tests := []struct {
		desc      string
		termSize  image.Point
		container func(ft *faketerm.Terminal) *Container
		events    []terminalapi.Event
		want      func(size image.Point) *faketerm.Terminal
		wantErr   bool
	}{
		{
			desc:     "mouse click outside of the terminal is ignored",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: true})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{-1, -1}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{10, 10}, Button: mouse.ButtonRelease},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "event not forwarded if container has no widget",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(ft)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "event forwarded to container at that point",
			termSize: image.Point{50, 20},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					SplitVertical(
						Left(
							PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: true})),
						),
						Right(
							SplitHorizontal(
								Top(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: true})),
								),
								Bottom(
									PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: true})),
								),
							),
						),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{49, 9}, Button: mouse.ButtonLeft},
				&terminalapi.Mouse{Position: image.Point{49, 9}, Button: mouse.ButtonRelease},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				// Widgets that aren't focused don't get the mouse clicks.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 25, 20)),
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(25, 10, 50, 20)),
					widgetapi.Options{WantMouse: true},
					&terminalapi.Keyboard{},
				)

				// The focused widget receives the key.
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(25, 0, 50, 10)),
					widgetapi.Options{WantMouse: true},
					&terminalapi.Mouse{Position: image.Point{24, 9}, Button: mouse.ButtonLeft},
					&terminalapi.Mouse{Position: image.Point{24, 9}, Button: mouse.ButtonRelease},
				)
				return ft
			},
		},
		{
			desc:     "event not forwarded if the widget didn't request it",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: false})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "event not forwarded if it falls on the container's border",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(
						fakewidget.New(widgetapi.Options{WantMouse: true}),
					),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBox(
					cvs,
					ft.Area(),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)
				testcanvas.MustApply(cvs, ft)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(1, 1, 19, 19)),
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "event not forwarded if it falls outside of widget's canvas",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: true,
							Ratio:     image.Point{2, 1},
						}),
					),
					VerticalAlignMiddle(),
					HorizontalAlignCenter(),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 20, 15)),
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:     "mouse poisition adjusted relative to widget's canvas, vertical offset",
			termSize: image.Point{20, 20},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: true,
							Ratio:     image.Point{2, 1},
						}),
					),
					VerticalAlignMiddle(),
					HorizontalAlignCenter(),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 5}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 20, 15)),
					widgetapi.Options{WantMouse: true},
					&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				)
				return ft
			},
		},
		{
			desc:     "mouse poisition adjusted relative to widget's canvas, horizontal offset",
			termSize: image.Point{30, 20},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: true,
							Ratio:     image.Point{9, 10},
						}),
					),
					VerticalAlignMiddle(),
					HorizontalAlignCenter(),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{6, 0}, Button: mouse.ButtonLeft},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(6, 0, 24, 20)),
					widgetapi.Options{WantMouse: true},
					&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				)
				return ft
			},
		},
		{
			desc:     "widget returns an error when processing the event",
			termSize: image.Point{40, 20},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{WantMouse: true})),
				)
			},
			events: []terminalapi.Event{
				&terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRight},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(ft.Area()),
					widgetapi.Options{},
				)
				return ft
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			c := tc.container(got)
			for _, ev := range tc.events {
				switch e := ev.(type) {
				case *terminalapi.Mouse:
					err := c.Mouse(e)
					if (err != nil) != tc.wantErr {
						t.Fatalf("Mouse => unexpected error: %v, wantErr: %v", err, tc.wantErr)
					}

				case *terminalapi.Keyboard:
					if err := c.Keyboard(e); err != nil {
						t.Fatalf("Keyboard => unexpected error: %v", err)
					}

				default:
					t.Fatalf("Unsupported event %T.", e)
				}
			}

			if err := c.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(tc.termSize), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}
}
