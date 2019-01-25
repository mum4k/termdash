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

package container

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/align"
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
	if _, err := New(
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
					SplitPercent(30),
				),
			),
			Right(
				Border(draw.LineStyleLight),
				PlaceWidget(fakewidget.New(widgetapi.Options{})),
			),
		),
	); err != nil {
		panic(err)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		desc             string
		termSize         image.Point
		container        func(ft *faketerm.Terminal) (*Container, error)
		wantContainerErr bool
		want             func(size image.Point) *faketerm.Terminal
	}{
		{
			desc:     "empty container",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(ft)
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "container with a border",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal split, children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
				testdraw.MustBorder(cvs, image.Rect(0, 0, 10, 5))
				testdraw.MustBorder(cvs, image.Rect(0, 5, 10, 10))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "fails on horizontal split too small",
			termSize: image.Point{10, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(draw.LineStyleLight),
						),
						Bottom(
							Border(draw.LineStyleLight),
						),
						SplitPercent(0),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "fails on horizontal split too large",
			termSize: image.Point{10, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(draw.LineStyleLight),
						),
						Bottom(
							Border(draw.LineStyleLight),
						),
						SplitPercent(100),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "horizontal unequal split",
			termSize: image.Point{10, 20},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitHorizontal(
						Top(
							Border(draw.LineStyleLight),
						),
						Bottom(
							Border(draw.LineStyleLight),
						),
						SplitPercent(20),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 10, 4))
				testdraw.MustBorder(cvs, image.Rect(0, 4, 10, 20))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal split, parent and children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testdraw.MustBorder(cvs, image.Rect(1, 1, 9, 5))
				testdraw.MustBorder(cvs, image.Rect(1, 5, 9, 9))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical split, children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
				testdraw.MustBorder(cvs, image.Rect(0, 0, 5, 10))
				testdraw.MustBorder(cvs, image.Rect(5, 0, 10, 10))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "fails on vertical split too small",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
						SplitPercent(0),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "fails on vertical split too large",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
						SplitPercent(100),
					),
				)
			},
			wantContainerErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:     "vertical unequal split",
			termSize: image.Point{20, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					SplitVertical(
						Left(
							Border(draw.LineStyleLight),
						),
						Right(
							Border(draw.LineStyleLight),
						),
						SplitPercent(20),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(cvs, image.Rect(0, 0, 4, 10))
				testdraw.MustBorder(cvs, image.Rect(4, 0, 20, 10))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical split, parent and children have borders",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testdraw.MustBorder(cvs, image.Rect(1, 1, 5, 9))
				testdraw.MustBorder(cvs, image.Rect(5, 1, 9, 9))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "multi level split",
			termSize: image.Point{10, 16},
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
				testdraw.MustBorder(cvs, image.Rect(0, 0, 5, 8))
				testdraw.MustBorder(cvs, image.Rect(0, 8, 5, 12))
				testdraw.MustBorder(cvs, image.Rect(0, 12, 5, 16))
				testdraw.MustBorder(cvs, image.Rect(5, 0, 10, 16))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "inherits border and focused color",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
				testdraw.MustBorder(
					cvs,
					image.Rect(0, 0, 10, 10),
					draw.BorderCellOpts(cell.FgColor(cell.ColorBlue)),
				)
				testdraw.MustBorder(
					cvs,
					image.Rect(1, 1, 5, 9),
					draw.BorderCellOpts(cell.FgColor(cell.ColorRed)),
				)
				testdraw.MustBorder(
					cvs,
					image.Rect(5, 1, 9, 9),
					draw.BorderCellOpts(cell.FgColor(cell.ColorRed)),
				)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "splitting a container removes the widget",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
				testdraw.MustBorder(
					cvs,
					ft.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testdraw.MustBorder(cvs, image.Rect(1, 1, 5, 9))
				testdraw.MustBorder(cvs, image.Rect(5, 1, 9, 9))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "draw widget with padding",
			termSize: image.Point{11, 11},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PaddingBottom(1),
					PaddingRight(1),
					PaddingLeft(1),
					PaddingTop(1),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					ft.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testdraw.MustBorder(cvs, image.Rect(2, 2, 9, 9))
				testdraw.MustText(cvs, "(7,7)", image.Pt(3, 3))
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "placing a widget removes container split",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
				testdraw.MustBorder(cvs, image.Rect(0, 0, 10, 10))
				testdraw.MustText(cvs, "(10,10)", image.Point{1, 1})
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

			cont, err := tc.container(got)
			if (err != nil) != tc.wantContainerErr {
				t.Errorf("tc.container => unexpected error:%v, wantErr:%v", err, tc.wantContainerErr)
			}
			if err != nil {
				return
			}
			if err := cont.Draw(); err != nil {
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
		container func(ft *faketerm.Terminal) (*Container, error)
		events    []terminalapi.Event
		want      func(size image.Point) *faketerm.Terminal
		wantErr   bool
	}{
		{
			desc:     "event not forwarded if container has no widget",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
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

			c, err := tc.container(got)
			if err != nil {
				t.Fatalf("tc.container => unexpected error: %v", err)
			}
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
		container func(ft *faketerm.Terminal) (*Container, error)
		events    []terminalapi.Event
		want      func(size image.Point) *faketerm.Terminal
		wantErr   bool
	}{
		{
			desc:     "mouse click outside of the terminal is ignored",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
				testdraw.MustBorder(
					cvs,
					ft.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: true,
							Ratio:     image.Point{2, 1},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: true,
							Ratio:     image.Point{2, 1},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(
						fakewidget.New(widgetapi.Options{
							WantMouse: true,
							Ratio:     image.Point{9, 10},
						}),
					),
					AlignVertical(align.VerticalMiddle),
					AlignHorizontal(align.HorizontalCenter),
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
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

			c, err := tc.container(got)
			if err != nil {
				t.Fatalf("tc.container => unexpected error: %v", err)
			}
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
