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

package grid

import (
	"context"
	"image"
	"testing"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/draw/testdraw"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/private/fakewidget"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/barchart"
)

// Shows how to create a simple 4x4 grid with four widgets.
// All the cells in the grid contain the same widget in this example.
func Example() {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	bc, err := barchart.New()
	if err != nil {
		panic(err)
	}

	builder := New()
	builder.Add(
		RowHeightPerc(
			50,
			ColWidthPerc(50, Widget(bc)),
			ColWidthPerc(50, Widget(bc)),
		),
		RowHeightPerc(
			50,
			ColWidthPerc(50, Widget(bc)),
			ColWidthPerc(50, Widget(bc)),
		),
	)
	gridOpts, err := builder.Build()
	if err != nil {
		panic(err)
	}

	cont, err := container.New(t, gridOpts...)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := termdash.Run(ctx, t, cont); err != nil {
		panic(err)
	}
}

// Shows how to create rows iteratively. Each row contains two columns and each
// column contains the same widget.
func Example_iterative() {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	bc, err := barchart.New()
	if err != nil {
		panic(err)
	}

	builder := New()
	for i := 0; i < 5; i++ {
		builder.Add(
			RowHeightPerc(
				20,
				ColWidthPerc(50, Widget(bc)),
				ColWidthPerc(50, Widget(bc)),
			),
		)
	}
	gridOpts, err := builder.Build()
	if err != nil {
		panic(err)
	}

	cont, err := container.New(t, gridOpts...)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := termdash.Run(ctx, t, cont); err != nil {
		panic(err)
	}
}

// mirror returns a new fake widget.
func mirror() *fakewidget.Mirror {
	return fakewidget.New(widgetapi.Options{})
}

// mustHSplit splits the area or panics.
func mustHSplit(ar image.Rectangle, heightPerc int) (top image.Rectangle, bottom image.Rectangle) {
	t, b, err := area.HSplit(ar, heightPerc)
	if err != nil {
		panic(err)
	}
	return t, b
}

// mustVSplit splits the area or panics.
func mustVSplit(ar image.Rectangle, widthPerc int) (left image.Rectangle, right image.Rectangle) {
	l, r, err := area.VSplit(ar, widthPerc)
	if err != nil {
		panic(err)
	}
	return l, r
}

func TestBuilder(t *testing.T) {
	tests := []struct {
		desc     string
		termSize image.Point
		builder  *Builder
		want     func(size image.Point) *faketerm.Terminal
		wantErr  bool
	}{
		{
			desc:     "fails when Widget is mixed with Rows and Columns at top level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(50),
					Widget(mirror()),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Widget is mixed with Rows and Columns at sub level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(
						50,
						RowHeightPerc(50),
						Widget(mirror()),
					),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Row heightPerc is too low at top level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(0),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Row heightPerc is too low at sub level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(
						50,
						RowHeightPerc(0),
					),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Row heightPerc is too high at top level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(100),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Row heightPerc is too high at sub level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(
						50,
						RowHeightPerc(100),
					),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Row heightPerc used under Row heightFixed",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightFixed(
						5,
						RowHeightPerc(10),
					),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Row heightPerc used under Col widthFixed",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					ColWidthFixed(
						5,
						RowHeightPerc(10),
					),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Col widthPerc is too low at top level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					ColWidthPerc(0),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Col widthPerc is too low at sub level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					ColWidthPerc(
						50,
						ColWidthPerc(0),
					),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Col widthPerc is too high at top level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					ColWidthPerc(100),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Col widthPerc is too high at sub level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					ColWidthPerc(
						50,
						ColWidthPerc(100),
					),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Col widthPerc used under Col widthFixed",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					ColWidthFixed(
						5,
						ColWidthPerc(10),
					),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when Col widthPerc used under Row heightFixed",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightFixed(
						5,
						ColWidthPerc(10),
					),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when height sum is too large at top level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(50),
					RowHeightPerc(50),
					RowHeightPerc(1),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when height sum is too large at sub level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(
						50,
						RowHeightPerc(50),
						RowHeightPerc(50),
						RowHeightPerc(1),
					),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when width sum is too large at top level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					ColWidthPerc(50),
					ColWidthPerc(50),
					ColWidthPerc(1),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "fails when width sum is too large at sub level",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(
					ColWidthPerc(
						50,
						ColWidthPerc(50),
						ColWidthPerc(50),
						ColWidthPerc(1),
					),
				)
				return b
			}(),
			wantErr: true,
		},
		{
			desc:     "empty container when nothing is added",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				return New()
			}(),
		},
		{
			desc:     "widget in the outer most container",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(Widget(mirror()))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				fakewidget.MustDraw(ft, cvs, &widgetapi.Meta{Focused: true}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two equal rows",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(RowHeightPerc(50, Widget(mirror())))
				b.Add(RowHeightPerc(50, Widget(mirror())))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				top, bot := mustHSplit(ft.Area(), 50)
				fakewidget.MustDraw(ft, testcanvas.MustNew(top), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(bot), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two equal rows, fixed size",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(RowHeightFixed(5, Widget(mirror())))
				b.Add(RowHeightFixed(5, Widget(mirror())))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				top, bot := mustHSplit(ft.Area(), 50)
				fakewidget.MustDraw(ft, testcanvas.MustNew(top), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(bot), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two equal rows with options",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(RowHeightPercWithOpts(
					50,
					[]container.Option{
						container.Border(linestyle.Double),
					},
					Widget(mirror()),
				))
				b.Add(RowHeightPercWithOpts(
					50,
					[]container.Option{
						container.Border(linestyle.Double),
					},
					Widget(mirror()),
				))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				top, bot := mustHSplit(ft.Area(), 50)
				topCvs := testcanvas.MustNew(top)
				botCvs := testcanvas.MustNew(bot)
				testdraw.MustBorder(topCvs, topCvs.Area(), draw.BorderLineStyle(linestyle.Double))
				testdraw.MustBorder(botCvs, botCvs.Area(), draw.BorderLineStyle(linestyle.Double))
				testcanvas.MustApply(topCvs, ft)
				testcanvas.MustApply(botCvs, ft)

				topWidget := testcanvas.MustNew(area.ExcludeBorder(top))
				botWidget := testcanvas.MustNew(area.ExcludeBorder(bot))
				fakewidget.MustDraw(ft, topWidget, &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, botWidget, &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two equal rows with options, fixed size",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(RowHeightFixedWithOpts(
					5,
					[]container.Option{
						container.Border(linestyle.Double),
					},
					Widget(mirror()),
				))
				b.Add(RowHeightFixedWithOpts(
					5,
					[]container.Option{
						container.Border(linestyle.Double),
					},
					Widget(mirror()),
				))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				top, bot := mustHSplit(ft.Area(), 50)
				topCvs := testcanvas.MustNew(top)
				botCvs := testcanvas.MustNew(bot)
				testdraw.MustBorder(topCvs, topCvs.Area(), draw.BorderLineStyle(linestyle.Double))
				testdraw.MustBorder(botCvs, botCvs.Area(), draw.BorderLineStyle(linestyle.Double))
				testcanvas.MustApply(topCvs, ft)
				testcanvas.MustApply(botCvs, ft)

				topWidget := testcanvas.MustNew(area.ExcludeBorder(top))
				botWidget := testcanvas.MustNew(area.ExcludeBorder(bot))
				fakewidget.MustDraw(ft, topWidget, &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, botWidget, &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two unequal rows",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(RowHeightPerc(20, Widget(mirror())))
				b.Add(RowHeightPerc(80, Widget(mirror())))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				top, bot := mustHSplit(ft.Area(), 20)
				fakewidget.MustDraw(ft, testcanvas.MustNew(top), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(bot), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two unequal rows, fixed size",
			termSize: image.Point{10, 10},
			builder: func() *Builder {
				b := New()
				b.Add(RowHeightFixed(2, Widget(mirror())))
				b.Add(RowHeightFixed(8, Widget(mirror())))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				top, bot := mustHSplit(ft.Area(), 20)
				fakewidget.MustDraw(ft, testcanvas.MustNew(top), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(bot), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two equal columns",
			termSize: image.Point{20, 10},
			builder: func() *Builder {
				b := New()
				b.Add(ColWidthPerc(50, Widget(mirror())))
				b.Add(ColWidthPerc(50, Widget(mirror())))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				left, right := mustVSplit(ft.Area(), 50)
				fakewidget.MustDraw(ft, testcanvas.MustNew(left), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(right), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two equal columns, fixed size",
			termSize: image.Point{20, 10},
			builder: func() *Builder {
				b := New()
				b.Add(ColWidthFixed(10, Widget(mirror())))
				b.Add(ColWidthFixed(10, Widget(mirror())))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				left, right := mustVSplit(ft.Area(), 50)
				fakewidget.MustDraw(ft, testcanvas.MustNew(left), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(right), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two equal columns with options",
			termSize: image.Point{20, 10},
			builder: func() *Builder {
				b := New()
				b.Add(ColWidthPercWithOpts(
					50,
					[]container.Option{
						container.Border(linestyle.Double),
					},
					Widget(mirror()),
				))
				b.Add(ColWidthPercWithOpts(
					50,
					[]container.Option{
						container.Border(linestyle.Double),
					},
					Widget(mirror()),
				))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				left, right := mustVSplit(ft.Area(), 50)
				leftCvs := testcanvas.MustNew(left)
				rightCvs := testcanvas.MustNew(right)
				testdraw.MustBorder(leftCvs, leftCvs.Area(), draw.BorderLineStyle(linestyle.Double))
				testdraw.MustBorder(rightCvs, rightCvs.Area(), draw.BorderLineStyle(linestyle.Double))
				testcanvas.MustApply(leftCvs, ft)
				testcanvas.MustApply(rightCvs, ft)

				leftWidget := testcanvas.MustNew(area.ExcludeBorder(left))
				rightWidget := testcanvas.MustNew(area.ExcludeBorder(right))
				fakewidget.MustDraw(ft, leftWidget, &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, rightWidget, &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two equal columns with options, fixed size",
			termSize: image.Point{20, 10},
			builder: func() *Builder {
				b := New()
				b.Add(ColWidthFixedWithOpts(
					10,
					[]container.Option{
						container.Border(linestyle.Double),
					},
					Widget(mirror()),
				))
				b.Add(ColWidthFixedWithOpts(
					10,
					[]container.Option{
						container.Border(linestyle.Double),
					},
					Widget(mirror()),
				))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)

				left, right := mustVSplit(ft.Area(), 50)
				leftCvs := testcanvas.MustNew(left)
				rightCvs := testcanvas.MustNew(right)
				testdraw.MustBorder(leftCvs, leftCvs.Area(), draw.BorderLineStyle(linestyle.Double))
				testdraw.MustBorder(rightCvs, rightCvs.Area(), draw.BorderLineStyle(linestyle.Double))
				testcanvas.MustApply(leftCvs, ft)
				testcanvas.MustApply(rightCvs, ft)

				leftWidget := testcanvas.MustNew(area.ExcludeBorder(left))
				rightWidget := testcanvas.MustNew(area.ExcludeBorder(right))
				fakewidget.MustDraw(ft, leftWidget, &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, rightWidget, &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two unequal columns",
			termSize: image.Point{40, 10},
			builder: func() *Builder {
				b := New()
				b.Add(ColWidthPerc(20, Widget(mirror())))
				b.Add(ColWidthPerc(80, Widget(mirror())))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				left, right := mustVSplit(ft.Area(), 20)
				fakewidget.MustDraw(ft, testcanvas.MustNew(left), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(right), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "two unequal columns, fixed size",
			termSize: image.Point{40, 10},
			builder: func() *Builder {
				b := New()
				b.Add(ColWidthFixed(8, Widget(mirror())))
				b.Add(ColWidthFixed(32, Widget(mirror())))
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				left, right := mustVSplit(ft.Area(), 20)
				fakewidget.MustDraw(ft, testcanvas.MustNew(left), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(right), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "rows with columns (equal)",
			termSize: image.Point{20, 20},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(
						50,
						ColWidthPerc(50, Widget(mirror())),
						ColWidthPerc(50, Widget(mirror())),
					),
					RowHeightPerc(
						50,
						ColWidthPerc(50, Widget(mirror())),
						ColWidthPerc(50, Widget(mirror())),
					),
				)
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				top, bot := mustHSplit(ft.Area(), 50)

				topLeft, topRight := mustVSplit(top, 50)
				botLeft, botRight := mustVSplit(bot, 50)
				fakewidget.MustDraw(ft, testcanvas.MustNew(topLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(topRight), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botRight), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "rows with columns (unequal)",
			termSize: image.Point{40, 20},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(
						20,
						ColWidthPerc(20, Widget(mirror())),
						ColWidthPerc(80, Widget(mirror())),
					),
					RowHeightPerc(
						80,
						ColWidthPerc(80, Widget(mirror())),
						ColWidthPerc(20, Widget(mirror())),
					),
				)
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				top, bot := mustHSplit(ft.Area(), 20)

				topLeft, topRight := mustVSplit(top, 20)
				botLeft, botRight := mustVSplit(bot, 80)
				fakewidget.MustDraw(ft, testcanvas.MustNew(topLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(topRight), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botRight), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "rows with columns (unequal), fixed and relative sizes mixed",
			termSize: image.Point{40, 20},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightFixed(
						4,
						ColWidthFixed(8, Widget(mirror())),
						ColWidthFixed(32, Widget(mirror())),
					),
					RowHeightPerc(
						80,
						ColWidthPerc(80, Widget(mirror())),
						ColWidthPerc(20, Widget(mirror())),
					),
				)
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				top, bot := mustHSplit(ft.Area(), 20)

				topLeft, topRight := mustVSplit(top, 20)
				botLeft, botRight := mustVSplit(bot, 80)
				fakewidget.MustDraw(ft, testcanvas.MustNew(topLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(topRight), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botRight), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "columns with rows (equal)",
			termSize: image.Point{20, 20},
			builder: func() *Builder {
				b := New()
				b.Add(
					ColWidthPerc(
						50,
						RowHeightPerc(50, Widget(mirror())),
						RowHeightPerc(50, Widget(mirror())),
					),
					ColWidthPerc(
						50,
						RowHeightPerc(50, Widget(mirror())),
						RowHeightPerc(50, Widget(mirror())),
					),
				)
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				top, bot := mustHSplit(ft.Area(), 50)

				topLeft, topRight := mustVSplit(top, 50)
				botLeft, botRight := mustVSplit(bot, 50)
				fakewidget.MustDraw(ft, testcanvas.MustNew(topLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(topRight), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botRight), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "columns with rows (unequal)",
			termSize: image.Point{40, 20},
			builder: func() *Builder {
				b := New()
				b.Add(
					ColWidthPerc(
						20,
						RowHeightPerc(20, Widget(mirror())),
						RowHeightPerc(80, Widget(mirror())),
					),
					ColWidthPerc(
						80,
						RowHeightPerc(80, Widget(mirror())),
						RowHeightPerc(20, Widget(mirror())),
					),
				)
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				left, right := mustVSplit(ft.Area(), 20)

				topLeft, topRight := mustHSplit(left, 20)
				botLeft, botRight := mustHSplit(right, 80)
				fakewidget.MustDraw(ft, testcanvas.MustNew(topLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(topRight), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botRight), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "rows with rows with columns",
			termSize: image.Point{40, 40},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(
						50,
						RowHeightPerc(
							50,
							ColWidthPerc(50, Widget(mirror())),
							ColWidthPerc(50, Widget(mirror())),
						),
						RowHeightPerc(
							50,
							ColWidthPerc(50, Widget(mirror())),
							ColWidthPerc(50, Widget(mirror())),
						),
					),
					RowHeightPerc(
						50,
						RowHeightPerc(
							50,
							ColWidthPerc(50, Widget(mirror())),
							ColWidthPerc(50, Widget(mirror())),
						),
						RowHeightPerc(
							50,
							ColWidthPerc(50, Widget(mirror())),
							ColWidthPerc(50, Widget(mirror())),
						),
					),
				)
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				top, bot := mustHSplit(ft.Area(), 50)
				topTop, topBot := mustHSplit(top, 50)
				botTop, botBot := mustHSplit(bot, 50)

				topTopLeft, topTopRight := mustVSplit(topTop, 50)
				topBotLeft, topBotRight := mustVSplit(topBot, 50)
				botTopLeft, botTopRight := mustVSplit(botTop, 50)
				botBotLeft, botBotRight := mustVSplit(botBot, 50)
				fakewidget.MustDraw(ft, testcanvas.MustNew(topTopLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(topTopRight), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(topBotLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(topBotRight), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botTopLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botTopRight), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botBotLeft), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botBotRight), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "rows mixed with columns at top level",
			termSize: image.Point{40, 30},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(20, Widget(mirror())),
					ColWidthPerc(20, Widget(mirror())),
					RowHeightPerc(20, Widget(mirror())),
					ColWidthPerc(20, Widget(mirror())),
				)
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				top, bot := mustHSplit(ft.Area(), 20)

				left, right := mustVSplit(bot, 20)
				topRight, botRight := mustHSplit(right, 25)
				fakewidget.MustDraw(ft, testcanvas.MustNew(top), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(left), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(topRight), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botRight), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "rows mixed with columns at sub level",
			termSize: image.Point{40, 30},
			builder: func() *Builder {
				b := New()
				b.Add(
					RowHeightPerc(
						50,
						RowHeightPerc(20, Widget(mirror())),
						ColWidthPerc(20, Widget(mirror())),
						RowHeightPerc(20, Widget(mirror())),
						ColWidthPerc(20, Widget(mirror())),
					),
					RowHeightPerc(50, Widget(mirror())),
				)
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				top, bot := mustHSplit(ft.Area(), 50)
				fakewidget.MustDraw(ft, testcanvas.MustNew(bot), &widgetapi.Meta{}, widgetapi.Options{})

				topTop, topBot := mustHSplit(top, 20)
				left, right := mustVSplit(topBot, 20)
				topRight, botRight := mustHSplit(right, 25)
				fakewidget.MustDraw(ft, testcanvas.MustNew(topTop), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(left), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(topRight), &widgetapi.Meta{}, widgetapi.Options{})
				fakewidget.MustDraw(ft, testcanvas.MustNew(botRight), &widgetapi.Meta{}, widgetapi.Options{})
				return ft
			},
		},
		{
			desc:     "widget's container can have options",
			termSize: image.Point{20, 20},
			builder: func() *Builder {
				b := New()
				b.Add(
					Widget(
						mirror(),
						container.Border(linestyle.Double),
					),
				)
				return b
			}(),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())
				testdraw.MustBorder(
					cvs,
					cvs.Area(),
					draw.BorderLineStyle(linestyle.Double),
					draw.BorderCellOpts(cell.FgColor(cell.ColorYellow)),
				)
				wCvs := testcanvas.MustNew(area.ExcludeBorder(cvs.Area()))
				fakewidget.MustDraw(ft, wCvs, &widgetapi.Meta{Focused: true}, widgetapi.Options{})
				testcanvas.MustCopyTo(wCvs, cvs)
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

			gridOpts, err := tc.builder.Build()
			if (err != nil) != tc.wantErr {
				t.Errorf("tc.builder => unexpected error: %v, wantErr:%v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			cont, err := container.New(got, gridOpts...)
			if err != nil {
				t.Fatalf("container.New => unexpected error: %v", err)
			}
			if err := cont.Draw(); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			var want *faketerm.Terminal
			if tc.want != nil {
				want = tc.want(tc.termSize)
			} else {
				w, err := faketerm.New(tc.termSize)
				if err != nil {
					t.Fatalf("faketerm.New => unexpected error: %v", err)
				}
				want = w
			}
			if diff := faketerm.Diff(want, got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}
}
