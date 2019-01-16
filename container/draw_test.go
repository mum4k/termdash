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
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/fakewidget"
)

func TestDrawWidget(t *testing.T) {
	tests := []struct {
		desc      string
		termSize  image.Point
		container func(ft *faketerm.Terminal) (*Container, error)
		want      func(size image.Point) *faketerm.Terminal
		wantErr   bool
	}{
		{
			desc:     "draws widget with container border",
			termSize: image.Point{9, 5},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(1, 1, 8, 4))
				testdraw.MustText(cvs, "(7,3)", image.Point{2, 2})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "draws widget with container border and title aligned on the left",
			termSize: image.Point{9, 5},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					BorderTitle("ab"),
					BorderTitleAlignLeft(),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
					draw.BorderTitle(
						"ab",
						draw.OverrunModeThreeDot,
						cell.FgColor(cell.ColorYellow),
					),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(1, 1, 8, 4))
				testdraw.MustText(cvs, "(7,3)", image.Point{2, 2})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "draws widget with container border and title aligned in the center",
			termSize: image.Point{9, 5},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					BorderTitle("ab"),
					BorderTitleAlignCenter(),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
					draw.BorderTitle(
						"ab",
						draw.OverrunModeThreeDot,
						cell.FgColor(cell.ColorYellow),
					),
					draw.BorderTitleAlign(align.HorizontalCenter),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(1, 1, 8, 4))
				testdraw.MustText(cvs, "(7,3)", image.Point{2, 2})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "draws widget with container border and title aligned on the right",
			termSize: image.Point{9, 5},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					BorderTitle("ab"),
					BorderTitleAlignRight(),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
					draw.BorderTitle(
						"ab",
						draw.OverrunModeThreeDot,
						cell.FgColor(cell.ColorYellow),
					),
					draw.BorderTitleAlign(align.HorizontalRight),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(1, 1, 8, 4))
				testdraw.MustText(cvs, "(7,3)", image.Point{2, 2})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "draws widget with container border and title that is trimmed",
			termSize: image.Point{9, 5},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					BorderTitle("abcdefgh"),
					BorderTitleAlignRight(),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
					draw.BorderTitle(
						"abcdefgh",
						draw.OverrunModeThreeDot,
						cell.FgColor(cell.ColorYellow),
					),
					draw.BorderTitleAlign(align.HorizontalRight),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(1, 1, 8, 4))
				testdraw.MustText(cvs, "(7,3)", image.Point{2, 2})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "draws widget without container border",
			termSize: image.Point{9, 5},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(0, 0, 9, 5))
				testdraw.MustText(cvs, "(9,5)", image.Point{1, 1})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "widget.Draw returns an error",
			termSize: image.Point{5, 5}, // Too small for the widget's box.
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widgetapi.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:     "container with border and no space isn't drawn",
			termSize: image.Point{1, 1},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustText(cvs, "⇄", image.Point{0, 0})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "container without the requested space for its widget isn't drawn",
			termSize: image.Point{1, 1},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widgetapi.Options{
						MinimumSize: image.Point{2, 2}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustText(cvs, "⇄", image.Point{0, 0})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "widget's canvas is limited to the requested maximum size",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widgetapi.Options{
						MaximumSize: image.Point{10, 10},
					})),
					AlignHorizontal(align.HorizontalLeft),
					AlignVertical(align.VerticalTop),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				contCvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					contCvs,
					contCvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(contCvs, ft)

				// Fake widget.
				cvs := testcanvas.MustNew(image.Rect(1, 1, 11, 11))
				fakewidget.MustDraw(ft, cvs, widgetapi.Options{})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "widget's canvas is limited to the requested maximum width",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widgetapi.Options{
						MaximumSize: image.Point{10, 0},
					})),
					AlignHorizontal(align.HorizontalLeft),
					AlignVertical(align.VerticalTop),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				contCvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					contCvs,
					contCvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(contCvs, ft)

				// Fake widget.
				cvs := testcanvas.MustNew(image.Rect(1, 1, 11, 21))
				fakewidget.MustDraw(ft, cvs, widgetapi.Options{})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "widget's canvas is limited to the requested maximum height",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widgetapi.Options{
						MaximumSize: image.Point{0, 10},
					})),
					AlignHorizontal(align.HorizontalLeft),
					AlignVertical(align.VerticalTop),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				contCvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					contCvs,
					contCvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(contCvs, ft)

				// Fake widget.
				cvs := testcanvas.MustNew(image.Rect(1, 1, 21, 11))
				fakewidget.MustDraw(ft, cvs, widgetapi.Options{})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "widget gets the requested aspect ratio",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widgetapi.Options{
						Ratio: image.Point{1, 2}},
					)),
					AlignHorizontal(align.HorizontalLeft),
					AlignVertical(align.VerticalTop),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(1, 1, 11, 21))
				testdraw.MustText(cvs, "(10,20)", image.Point{2, 2})

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "widget's canvas is limited to the requested maximum size and ratio",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widgetapi.Options{
						MaximumSize: image.Point{20, 19},
						Ratio:       image.Point{1, 1},
					})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				contCvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					contCvs,
					contCvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				testcanvas.MustApply(contCvs, ft)

				// Fake widget.
				cvs := testcanvas.MustNew(image.Rect(1, 1, 20, 20))
				fakewidget.MustDraw(ft, cvs, widgetapi.Options{})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},

		{
			desc:     "horizontal left align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					AlignHorizontal(align.HorizontalLeft),
					PlaceWidget(fakewidget.New(widgetapi.Options{
						Ratio: image.Point{1, 2}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(1, 1, 11, 21))
				testdraw.MustText(cvs, "(10,20)", image.Point{2, 2})

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal center align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					AlignHorizontal(align.HorizontalCenter),
					PlaceWidget(fakewidget.New(widgetapi.Options{
						Ratio: image.Point{1, 2}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(6, 1, 16, 21))
				testdraw.MustText(cvs, "(10,20)", image.Point{7, 2})

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal right align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					AlignHorizontal(align.HorizontalRight),
					PlaceWidget(fakewidget.New(widgetapi.Options{
						Ratio: image.Point{1, 2}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(11, 1, 21, 21))
				testdraw.MustText(cvs, "(10,20)", image.Point{12, 2})

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical top align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					AlignVertical(align.VerticalTop),
					PlaceWidget(fakewidget.New(widgetapi.Options{
						Ratio: image.Point{2, 1}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(1, 1, 21, 11))
				testdraw.MustText(cvs, "(20,10)", image.Point{2, 2})

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical middle align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					AlignVertical(align.VerticalMiddle),
					PlaceWidget(fakewidget.New(widgetapi.Options{
						Ratio: image.Point{2, 1}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(1, 6, 21, 16))
				testdraw.MustText(cvs, "(20,10)", image.Point{2, 7})

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical bottom align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) (*Container, error) {
				return New(
					ft,
					Border(draw.LineStyleLight),
					AlignVertical(align.VerticalBottom),
					PlaceWidget(fakewidget.New(widgetapi.Options{
						Ratio: image.Point{2, 1}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)

				// Fake widget border.
				testdraw.MustBorder(cvs, image.Rect(1, 11, 21, 21))
				testdraw.MustText(cvs, "(20,10)", image.Point{2, 12})

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := faketerm.MustNew(tc.termSize)
			c, err := tc.container(got)
			if err != nil {
				t.Fatalf("tc.container => unexpected error: %v", err)
			}
			err = c.Draw()
			if (err != nil) != tc.wantErr {
				t.Errorf("Draw => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := faketerm.Diff(tc.want(got.Size()), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}
}

func TestDrawHandlesTerminalResize(t *testing.T) {
	termSize := image.Point{60, 10}
	got, err := faketerm.New(termSize)
	if err != nil {
		t.Errorf("faketerm.New => unexpected error: %v", err)
	}

	cont, err := New(
		got,
		SplitVertical(
			Left(
				SplitHorizontal(
					Top(
						PlaceWidget(fakewidget.New(widgetapi.Options{})),
					),
					Bottom(
						PlaceWidget(fakewidget.New(widgetapi.Options{})),
					),
				),
			),
			Right(
				SplitVertical(
					Left(
						PlaceWidget(fakewidget.New(widgetapi.Options{})),
					),
					Right(
						PlaceWidget(fakewidget.New(widgetapi.Options{})),
					),
				),
			),
		),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	// The following tests aren't hermetic, they all access the same container
	// and fake terminal in order to retain state between resizes.
	tests := []struct {
		desc   string
		resize *image.Point // if not nil, the fake terminal will be resized.
		want   func(size image.Point) *faketerm.Terminal
	}{
		{
			desc: "handles the initial draw request",
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 30, 5)),
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 30, 10)),
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(30, 0, 45, 10)),
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(45, 0, 60, 10)),
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:   "increase in size",
			resize: &image.Point{80, 10},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 40, 5)),
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 40, 10)),
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(40, 0, 60, 10)),
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(60, 0, 80, 10)),
					widgetapi.Options{},
				)
				return ft
			},
		},
		{
			desc:   "decrease in size",
			resize: &image.Point{50, 10},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 0, 25, 5)),
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(0, 5, 25, 10)),
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(25, 0, 37, 10)),
					widgetapi.Options{},
				)
				fakewidget.MustDraw(
					ft,
					testcanvas.MustNew(image.Rect(37, 0, 50, 10)),
					widgetapi.Options{},
				)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.resize != nil {
				if err := got.Resize(*tc.resize); err != nil {
					t.Fatalf("Resize => unexpected error: %v", err)
				}
			}
			if err := cont.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			if diff := faketerm.Diff(tc.want(got.Size()), got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}
}
