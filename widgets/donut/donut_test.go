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

package donut

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/braille/testbraille"
	"github.com/mum4k/termdash/canvas/testcanvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/draw/testdraw"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

func TestDonut(t *testing.T) {
	tests := []struct {
		desc          string
		opts          []Option
		update        func(*Donut) error // update gets called before drawing of the widget.
		canvas        image.Rectangle
		want          func(size image.Point) *faketerm.Terminal
		wantNewErr    bool
		wantUpdateErr bool // whether to expect an error on a call to the update function
		wantDrawErr   bool
	}{
		{
			desc: "New fails on negative donut hole percent",
			opts: []Option{
				HolePercent(-1),
			},
			canvas:     image.Rect(0, 0, 3, 3),
			wantNewErr: true,
		},
		{
			desc: "New fails on too large donut hole percent",
			opts: []Option{
				HolePercent(101),
			},
			canvas:     image.Rect(0, 0, 3, 3),
			wantNewErr: true,
		},
		{
			desc: "New fails on too small start angle",
			opts: []Option{
				StartAngle(-1),
			},
			canvas:     image.Rect(0, 0, 3, 3),
			wantNewErr: true,
		},
		{
			desc: "New fails on too large start angle",
			opts: []Option{
				StartAngle(360),
			},
			canvas:     image.Rect(0, 0, 3, 3),
			wantNewErr: true,
		},
		{
			desc:   "Percent fails on too small start angle",
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Percent(100, StartAngle(-1))
			},
			wantUpdateErr: true,
		},
		{
			desc:   "Percent fails on negative percent",
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Percent(-1)
			},
			wantUpdateErr: true,
		},
		{
			desc:   "Percent fails on value too large",
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Percent(101)
			},
			wantUpdateErr: true,
		},
		{
			desc:   "Absolute fails on too small start angle",
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Absolute(100, 100, StartAngle(-1))
			},
			wantUpdateErr: true,
		},
		{
			desc:   "Absolute fails on done to small",
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Absolute(-1, 100)
			},
			wantUpdateErr: true,
		},
		{
			desc:   "Absolute fails on total to small",
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Absolute(0, 0)
			},
			wantUpdateErr: true,
		},
		{
			desc:   "Absolute fails on done greater than total",
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Absolute(2, 1)
			},
			wantUpdateErr: true,
		},

		{
			desc:   "draws empty for no data points",
			canvas: image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc: "fails when canvas too small to draw a circle",
			update: func(d *Donut) error {
				return d.Percent(100)
			},
			canvas:      image.Rect(0, 0, 1, 1),
			wantDrawErr: true,
		},
		{
			desc:   "smallest valid donut, 100% progress",
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Percent(100)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleCircle(bc, image.Point{2, 5}, 2, draw.BrailleCircleFilled())

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "New sets donut options",
			opts: []Option{
				CellOpts(
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorBlue),
				),
			},
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Percent(100)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleCircle(bc, image.Point{2, 5}, 2,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleCellOpts(
						cell.FgColor(cell.ColorRed),
						cell.BgColor(cell.ColorBlue),
					),
				)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "Percent sets donut options",
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Percent(100,
					CellOpts(
						cell.FgColor(cell.ColorRed),
						cell.BgColor(cell.ColorBlue),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleCircle(bc, image.Point{2, 5}, 2,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleCellOpts(
						cell.FgColor(cell.ColorRed),
						cell.BgColor(cell.ColorBlue),
					),
				)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "Absolute sets donut options",
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Absolute(100, 100,
					CellOpts(
						cell.FgColor(cell.ColorRed),
						cell.BgColor(cell.ColorBlue),
					),
				)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleCircle(bc, image.Point{2, 5}, 2,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleCellOpts(
						cell.FgColor(cell.ColorRed),
						cell.BgColor(cell.ColorBlue),
					),
				)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "smallest valid donut, 100 absolute",
			canvas: image.Rect(0, 0, 3, 3),
			update: func(d *Donut) error {
				return d.Absolute(100, 100)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleCircle(bc, image.Point{2, 5}, 2, draw.BrailleCircleFilled())

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "smallest valid donut with a hole",
			canvas: image.Rect(0, 0, 6, 6),
			update: func(d *Donut) error {
				return d.Percent(100)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 5, draw.BrailleCircleFilled())
				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 2,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "draws a larger hole",
			canvas: image.Rect(0, 0, 6, 6),
			update: func(d *Donut) error {
				return d.Percent(100, HolePercent(50))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 5, draw.BrailleCircleFilled())
				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 3,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "hole as large as donut",
			canvas: image.Rect(0, 0, 6, 6),
			update: func(d *Donut) error {
				return d.Percent(100, HolePercent(100), HideTextProgress())
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 5, draw.BrailleCircleFilled())
				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 5,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:   "displays 100% progress",
			canvas: image.Rect(0, 0, 7, 7),
			update: func(d *Donut) error {
				return d.Percent(100, HolePercent(80))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				bc := testbraille.MustNew(c.Area())

				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 6, draw.BrailleCircleFilled())
				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 5,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)
				testbraille.MustCopyTo(bc, c)

				testdraw.MustText(c, "100%", image.Point{2, 3})

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "sets text cell options",
			canvas: image.Rect(0, 0, 7, 7),
			update: func(d *Donut) error {
				return d.Percent(100, HolePercent(80), TextCellOpts(
					cell.FgColor(cell.ColorGreen),
					cell.BgColor(cell.ColorYellow),
				))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				bc := testbraille.MustNew(c.Area())

				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 6, draw.BrailleCircleFilled())
				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 5,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)
				testbraille.MustCopyTo(bc, c)

				testdraw.MustText(c, "100%", image.Point{2, 3}, draw.TextCellOpts(
					cell.FgColor(cell.ColorGreen),
					cell.BgColor(cell.ColorYellow),
				))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "shows text again when hidden previously",
			opts: []Option{
				HideTextProgress(),
			},
			canvas: image.Rect(0, 0, 7, 7),
			update: func(d *Donut) error {
				return d.Percent(100, HolePercent(80), ShowTextProgress())
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				bc := testbraille.MustNew(c.Area())

				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 6, draw.BrailleCircleFilled())
				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 5,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)
				testbraille.MustCopyTo(bc, c)

				testdraw.MustText(c, "100%", image.Point{2, 3})

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "hides text when requested",
			canvas: image.Rect(0, 0, 7, 7),
			update: func(d *Donut) error {
				return d.Percent(100, HolePercent(80), HideTextProgress())
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				bc := testbraille.MustNew(c.Area())

				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 6, draw.BrailleCircleFilled())
				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 5,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "hides text when hole is too small",
			canvas: image.Rect(0, 0, 7, 7),
			update: func(d *Donut) error {
				return d.Percent(100, HolePercent(50))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				bc := testbraille.MustNew(c.Area())

				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 6, draw.BrailleCircleFilled())
				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 3,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)
				testbraille.MustCopyTo(bc, c)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "displays 1% progress",
			canvas: image.Rect(0, 0, 7, 7),
			update: func(d *Donut) error {
				return d.Percent(1, HolePercent(80))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				bc := testbraille.MustNew(c.Area())

				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 6,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleArcOnly(89, 90),
				)
				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 5,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)
				testbraille.MustCopyTo(bc, c)

				testdraw.MustText(c, "1%", image.Point{3, 3})

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "displays 25% progress, clockwise",
			canvas: image.Rect(0, 0, 7, 7),
			update: func(d *Donut) error {
				return d.Percent(25, HolePercent(80), Clockwise())
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				bc := testbraille.MustNew(c.Area())

				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 6,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleArcOnly(0, 90),
				)
				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 5,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)
				testbraille.MustCopyTo(bc, c)

				testdraw.MustText(c, "25%", image.Point{2, 3})

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "displays 25% progress, counter-clockwise",
			canvas: image.Rect(0, 0, 7, 7),
			update: func(d *Donut) error {
				return d.Percent(25, HolePercent(80), CounterClockwise())
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				bc := testbraille.MustNew(c.Area())

				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 6,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleArcOnly(90, 180),
				)
				testdraw.MustBrailleCircle(bc, image.Point{6, 13}, 5,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)
				testbraille.MustCopyTo(bc, c)

				testdraw.MustText(c, "25%", image.Point{2, 3})

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "displays 10/10 absolute progress",
			canvas: image.Rect(0, 0, 8, 8),
			update: func(d *Donut) error {
				return d.Absolute(10, 10, HolePercent(80))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				bc := testbraille.MustNew(c.Area())

				testdraw.MustBrailleCircle(bc, image.Point{8, 17}, 7, draw.BrailleCircleFilled())
				testdraw.MustBrailleCircle(bc, image.Point{8, 17}, 6,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)
				testbraille.MustCopyTo(bc, c)

				testdraw.MustText(c, "10/10", image.Point{2, 4})

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "displays 1/10 absolute progress",
			canvas: image.Rect(0, 0, 8, 8),
			update: func(d *Donut) error {
				return d.Absolute(1, 10, HolePercent(80))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				bc := testbraille.MustNew(c.Area())

				testdraw.MustBrailleCircle(bc, image.Point{8, 17}, 7,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleArcOnly(54, 90),
				)
				testdraw.MustBrailleCircle(bc, image.Point{8, 17}, 6,
					draw.BrailleCircleFilled(),
					draw.BrailleCircleClearPixels(),
				)
				testbraille.MustCopyTo(bc, c)

				testdraw.MustText(c, "1/10", image.Point{2, 4})

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			d, err := New(tc.opts...)
			if (err != nil) != tc.wantNewErr {
				t.Errorf("New => unexpected error: %v, wantNewErr: %v", err, tc.wantNewErr)
			}
			if err != nil {
				return
			}

			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			if tc.update != nil {
				err = tc.update(d)
				if (err != nil) != tc.wantUpdateErr {
					t.Errorf("update => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)

				}
				if err != nil {
					return
				}
			}

			err = d.Draw(c)
			if (err != nil) != tc.wantDrawErr {
				t.Errorf("Draw => unexpected error: %v, wantDrawErr: %v", err, tc.wantDrawErr)
			}
			if err != nil {
				return
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
		})
	}
}

func TestKeyboard(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := d.Keyboard(&terminalapi.Keyboard{}); err == nil {
		t.Errorf("Keyboard => got nil err, wanted one")
	}
}

func TestMouse(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := d.Mouse(&terminalapi.Mouse{}); err == nil {
		t.Errorf("Mouse => got nil err, wanted one")
	}
}

func TestOptions(t *testing.T) {
	d, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}

	got := d.Options()
	want := widgetapi.Options{
		Ratio:        image.Point{4, 2},
		MinimumSize:  image.Point{3, 3},
		WantKeyboard: widgetapi.KeyScopeNone,
		WantMouse: widgetapi.MouseScopeNone,
	}
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
	}

}
