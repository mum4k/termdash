package container

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/draw/testdraw"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/widget"
	"github.com/mum4k/termdash/widgets/fakewidget"
)

func TestDrawWidget(t *testing.T) {
	tests := []struct {
		desc      string
		termSize  image.Point
		container func(ft *faketerm.Terminal) *Container
		want      func(size image.Point) *faketerm.Terminal
		wantErr   bool
	}{
		{
			desc:     "draws widget with container border",
			termSize: image.Point{9, 5},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widget.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBox(
					cvs,
					cvs.Area(),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)

				// Fake widget border.
				testdraw.MustBox(cvs, image.Rect(1, 1, 8, 4), draw.LineStyleLight)
				tb := draw.TextBounds{Start: image.Point{2, 2}}
				testdraw.MustText(cvs, "(7,3)", tb)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "draws widget without container border",
			termSize: image.Point{9, 5},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widget.Options{})),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				// Fake widget border.
				testdraw.MustBox(cvs, image.Rect(0, 0, 9, 5), draw.LineStyleLight)
				tb := draw.TextBounds{Start: image.Point{1, 1}}
				testdraw.MustText(cvs, "(9,5)", tb)
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "widget.Draw returns an error",
			termSize: image.Point{5, 5}, // Too small for the widget's box.
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widget.Options{})),
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
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustText(cvs, "⇄", draw.TextBounds{})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "container without the requested space for its widget isn't drawn",
			termSize: image.Point{1, 1},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					PlaceWidget(fakewidget.New(widget.Options{
						MinimumSize: image.Point{2, 2}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustText(cvs, "⇄", draw.TextBounds{})
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "widget gets the requested aspect ratio",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					PlaceWidget(fakewidget.New(widget.Options{
						Ratio: image.Point{1, 2}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBox(
					cvs,
					cvs.Area(),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)

				// Fake widget border.
				testdraw.MustBox(cvs, image.Rect(1, 1, 11, 21), draw.LineStyleLight)
				tb := draw.TextBounds{Start: image.Point{2, 2}}
				testdraw.MustText(cvs, "(10,20)", tb)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal left align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					HorizontalAlignLeft(),
					PlaceWidget(fakewidget.New(widget.Options{
						Ratio: image.Point{1, 2}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBox(
					cvs,
					cvs.Area(),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)

				// Fake widget border.
				testdraw.MustBox(cvs, image.Rect(1, 1, 11, 21), draw.LineStyleLight)
				tb := draw.TextBounds{Start: image.Point{2, 2}}
				testdraw.MustText(cvs, "(10,20)", tb)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal center align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					HorizontalAlignCenter(),
					PlaceWidget(fakewidget.New(widget.Options{
						Ratio: image.Point{1, 2}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBox(
					cvs,
					cvs.Area(),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)

				// Fake widget border.
				testdraw.MustBox(cvs, image.Rect(6, 1, 16, 21), draw.LineStyleLight)
				tb := draw.TextBounds{Start: image.Point{7, 2}}
				testdraw.MustText(cvs, "(10,20)", tb)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "horizontal right align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					HorizontalAlignRight(),
					PlaceWidget(fakewidget.New(widget.Options{
						Ratio: image.Point{1, 2}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBox(
					cvs,
					cvs.Area(),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)

				// Fake widget border.
				testdraw.MustBox(cvs, image.Rect(11, 1, 21, 21), draw.LineStyleLight)
				tb := draw.TextBounds{Start: image.Point{12, 2}}
				testdraw.MustText(cvs, "(10,20)", tb)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical top align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					VerticalAlignTop(),
					PlaceWidget(fakewidget.New(widget.Options{
						Ratio: image.Point{2, 1}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBox(
					cvs,
					cvs.Area(),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)

				// Fake widget border.
				testdraw.MustBox(cvs, image.Rect(1, 1, 21, 11), draw.LineStyleLight)
				tb := draw.TextBounds{Start: image.Point{2, 2}}
				testdraw.MustText(cvs, "(20,10)", tb)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical middle align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					VerticalAlignMiddle(),
					PlaceWidget(fakewidget.New(widget.Options{
						Ratio: image.Point{2, 1}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBox(
					cvs,
					cvs.Area(),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)

				// Fake widget border.
				testdraw.MustBox(cvs, image.Rect(1, 6, 21, 16), draw.LineStyleLight)
				tb := draw.TextBounds{Start: image.Point{2, 7}}
				testdraw.MustText(cvs, "(20,10)", tb)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc:     "vertical bottom align for the widget",
			termSize: image.Point{22, 22},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					VerticalAlignBottom(),
					PlaceWidget(fakewidget.New(widget.Options{
						Ratio: image.Point{2, 1}},
					)),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				// Container border.
				testdraw.MustBox(
					cvs,
					cvs.Area(),
					draw.LineStyleLight,
					cell.FgColor(cell.ColorYellow),
				)

				// Fake widget border.
				testdraw.MustBox(cvs, image.Rect(1, 11, 21, 21), draw.LineStyleLight)
				tb := draw.TextBounds{Start: image.Point{2, 12}}
				testdraw.MustText(cvs, "(20,10)", tb)

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := faketerm.MustNew(tc.termSize)
			c := tc.container(got)
			err := c.Draw()
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
