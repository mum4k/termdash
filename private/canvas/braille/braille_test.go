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

package braille

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/faketerm"
)

func Example_copiedToCanvas() {
	// Given a parent canvas the widget receives from the infrastructure:
	parent, err := canvas.New(image.Rect(0, 0, 3, 3))
	if err != nil {
		panic(err)
	}

	// The widget can create a braille canvas with the same or smaller area:
	braille, err := New(parent.Area())
	if err != nil {
		panic(err)
	}

	// After setting / clearing / toggling of pixels on the braille canvas, it
	// can be copied back to the parent canvas.
	if err := braille.SetPixel(image.Point{0, 0}); err != nil {
		panic(err)
	}
	if err := braille.CopyTo(parent); err != nil {
		panic(err)
	}
}

func Example_appliedToTerminal() {
	// When working with a terminal directly:
	ft, err := faketerm.New(image.Point{3, 3})
	if err != nil {
		panic(err)
	}

	// The widget can create a braille canvas with the same or smaller area:
	braille, err := New(ft.Area())
	if err != nil {
		panic(err)
	}

	// After setting / clearing / toggling of pixels on the braille canvas, it
	// can be applied to the terminal.
	if err := braille.SetPixel(image.Point{0, 0}); err != nil {
		panic(err)
	}
	if err := braille.Apply(ft); err != nil {
		panic(err)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		desc         string
		ar           image.Rectangle
		wantSize     image.Point
		wantArea     image.Rectangle
		wantCellArea image.Rectangle
		wantErr      bool
	}{
		{
			desc:    "fails on a negative area",
			ar:      image.Rect(-1, -1, -2, -2),
			wantErr: true,
		},
		{
			desc:         "braille from zero-based single-cell area",
			ar:           image.Rect(0, 0, 1, 1),
			wantSize:     image.Point{2, 4},
			wantArea:     image.Rect(0, 0, 2, 4),
			wantCellArea: image.Rect(0, 0, 1, 1),
		},
		{
			desc:         "braille from non-zero-based single-cell area",
			ar:           image.Rect(3, 3, 4, 4),
			wantSize:     image.Point{2, 4},
			wantArea:     image.Rect(0, 0, 2, 4),
			wantCellArea: image.Rect(0, 0, 1, 1),
		},
		{
			desc:         "braille from zero-based multi-cell area",
			ar:           image.Rect(0, 0, 3, 3),
			wantSize:     image.Point{6, 12},
			wantArea:     image.Rect(0, 0, 6, 12),
			wantCellArea: image.Rect(0, 0, 3, 3),
		},
		{
			desc:         "braille from non-zero-based multi-cell area",
			ar:           image.Rect(6, 6, 9, 9),
			wantSize:     image.Point{6, 12},
			wantArea:     image.Rect(0, 0, 6, 12),
			wantCellArea: image.Rect(0, 0, 3, 3),
		},
		{
			desc:         "braille from non-zero-based multi-cell rectangular area",
			ar:           image.Rect(6, 6, 9, 10),
			wantSize:     image.Point{6, 16},
			wantArea:     image.Rect(0, 0, 6, 16),
			wantCellArea: image.Rect(0, 0, 3, 4),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := New(tc.ar)
			if (err != nil) != tc.wantErr {
				t.Errorf("New => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			gotSize := got.Size()
			if diff := pretty.Compare(tc.wantSize, gotSize); diff != "" {
				t.Errorf("Size => unexpected diff (-want, +got):\n%s", diff)
			}

			gotArea := got.Area()
			if diff := pretty.Compare(tc.wantArea, gotArea); diff != "" {
				t.Errorf("Area => unexpected diff (-want, +got):\n%s", diff)
			}

			gotCellArea := got.CellArea()
			if diff := pretty.Compare(tc.wantCellArea, gotCellArea); diff != "" {
				t.Errorf("CellArea => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestBraille(t *testing.T) {
	tests := []struct {
		desc     string
		ar       image.Rectangle
		pixelOps func(*Canvas) error
		want     func(size image.Point) *faketerm.Terminal
		wantErr  bool
	}{
		{
			desc: "fails on pixel with negative X",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{-1, 0})
			},
			wantErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc: "fails on pixel with negative Y",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{0, -1})
			},
			wantErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc: "SetCellOptions fails on a cell outside of the braille canvas",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetCellOpts(image.Point{0, -1})
			},
			wantErr: true,
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc: "SetCellOptions sets options on cell with no options",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetCellOpts(image.Point{0, 0}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				c := testcanvas.MustCell(cvs, image.Point{0, 0})
				testcanvas.MustSetCell(cvs, image.Point{0, 0}, c.Rune, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "SetCellOptions preserves the cell rune",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				return c.SetCellOpts(image.Point{0, 0}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(cvs, image.Point{0, 0}, '⠁', cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "SetCellOptions overwrites options set previously",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
					return err
				}
				return c.SetCellOpts(image.Point{0, 0}, cell.FgColor(cell.ColorGreen), cell.BgColor(cell.ColorYellow))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(cvs, image.Point{0, 0}, '⠁', cell.FgColor(cell.ColorGreen), cell.BgColor(cell.ColorYellow))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "SetCellOptions sets default options when no options provided",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
					return err
				}
				return c.SetCellOpts(image.Point{0, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(cvs, image.Point{0, 0}, '⠁')

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "SetCellOptions is idempotent",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetCellOpts(image.Point{0, 0}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue)); err != nil {
					return err
				}
				return c.SetCellOpts(image.Point{0, 0}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				c := testcanvas.MustCell(cvs, image.Point{0, 0})
				testcanvas.MustSetCell(cvs, image.Point{0, 0}, c.Rune, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))

				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "SetAreaCellOptions fails on area too large",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetAreaCellOpts(image.Rect(0, 0, 2, 2), cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			wantErr: true,
		},
		{
			desc: "SetAreaCellOptions sets the cell options in full area",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetAreaCellOpts(image.Rect(0, 0, 1, 1), cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, p := range []image.Point{
					{0, 0},
				} {
					c := testcanvas.MustCell(cvs, p)
					testcanvas.MustSetCell(cvs, p, c.Rune, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
				}
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "SetAreaCellOptions sets the cell options in a sub-area",
			ar:   image.Rect(0, 0, 3, 3),
			pixelOps: func(c *Canvas) error {
				return c.SetAreaCellOpts(image.Rect(0, 0, 2, 2), cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				cvs := testcanvas.MustNew(ft.Area())

				for _, p := range []image.Point{
					{0, 0},
					{0, 1},
					{1, 0},
					{1, 1},
				} {
					c := testcanvas.MustCell(cvs, p)
					testcanvas.MustSetCell(cvs, p, c.Rune, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorBlue))
				}
				testcanvas.MustApply(cvs, ft)
				return ft
			},
		},
		{
			desc: "set pixel 0,0",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{0, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠁')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "set is idempotent",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				return c.SetPixel(image.Point{0, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠁')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "set pixel 1,0",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{1, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠈')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "set pixel 0,1",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{0, 1})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠂')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "set pixel 1,1",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{1, 1})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠐')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "set pixel 0,2",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{0, 2})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠄')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "set pixel 1,2",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{1, 2})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠠')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "set pixel 0,3",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{0, 3})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⡀')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "set pixel 1,3",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{1, 3})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⢀')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "fails on point outside of the canvas",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{2, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "clears the canvas",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				return c.Clear()
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc: "sets multiple pixels",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 0}); err != nil {
					return err
				}
				return c.SetPixel(image.Point{0, 1})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠋')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "sets all the pixels in a cell",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{0, 1}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 1}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{0, 2}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 2}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{0, 3}); err != nil {
					return err
				}
				return c.SetPixel(image.Point{1, 3})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⣿')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "set cell options",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.SetPixel(image.Point{0, 0}, cell.FgColor(cell.ColorRed))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠁', cell.FgColor(cell.ColorRed))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "set pixels in multiple cells",
			ar:   image.Rect(0, 0, 2, 2),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{2, 2}); err != nil {
					return err
				}
				return c.SetPixel(image.Point{1, 7})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠁') // pixel 0,0
				testcanvas.MustSetCell(c, image.Point{1, 0}, '⠄') // pixel 0,2
				testcanvas.MustSetCell(c, image.Point{0, 1}, '⢀') // pixel 1,3

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "clear pixels in multiple cells",
			ar:   image.Rect(0, 0, 2, 2),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{2, 2}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 7}); err != nil {
					return err
				}
				if err := c.ClearPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.ClearPixel(image.Point{2, 2}); err != nil {
					return err
				}
				return c.ClearPixel(image.Point{1, 7})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠀')
				testcanvas.MustSetCell(c, image.Point{1, 0}, '⠀')
				testcanvas.MustSetCell(c, image.Point{0, 1}, '⠀')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "toggle pixels in multiple cells",
			ar:   image.Rect(0, 0, 2, 2),
			pixelOps: func(c *Canvas) error {
				if err := c.TogglePixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.TogglePixel(image.Point{2, 2}); err != nil {
					return err
				}
				if err := c.TogglePixel(image.Point{1, 7}); err != nil {
					return err
				}
				if err := c.TogglePixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.TogglePixel(image.Point{2, 2}); err != nil {
					return err
				}
				return c.TogglePixel(image.Point{1, 7})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠀')
				testcanvas.MustSetCell(c, image.Point{1, 0}, '⠀')
				testcanvas.MustSetCell(c, image.Point{0, 1}, '⠀')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "sets and clears one pixel",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{0, 1}); err != nil {
					return err
				}
				return c.ClearPixel(image.Point{0, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠊')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "sets and clears all pixels",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{0, 1}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 1}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{0, 2}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 2}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{0, 3}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 3}); err != nil {
					return err
				}

				if err := c.ClearPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.ClearPixel(image.Point{1, 0}); err != nil {
					return err
				}
				if err := c.ClearPixel(image.Point{0, 1}); err != nil {
					return err
				}
				if err := c.ClearPixel(image.Point{1, 1}); err != nil {
					return err
				}
				if err := c.ClearPixel(image.Point{0, 2}); err != nil {
					return err
				}
				if err := c.ClearPixel(image.Point{1, 2}); err != nil {
					return err
				}
				if err := c.ClearPixel(image.Point{0, 3}); err != nil {
					return err
				}
				return c.ClearPixel(image.Point{1, 3})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠀')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "clear is idempotent when cell doesn't contain braille pattern character",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.ClearPixel(image.Point{0, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc: "clear fails on point outside of the canvas",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.ClearPixel(image.Point{3, 1})
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc: "clear is idempotent when the pixel is already cleared",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.ClearPixel(image.Point{0, 0}); err != nil {
					return err
				}
				return c.ClearPixel(image.Point{0, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠀')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "clearing a pixel sets options on the cell",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{0, 1}); err != nil {
					return err
				}
				return c.ClearPixel(image.Point{0, 0}, cell.FgColor(cell.ColorBlue))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠊', cell.FgColor(cell.ColorBlue))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "toggles a pixel which adds the first braille pattern character",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.TogglePixel(image.Point{0, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠁')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "toggles a pixel on",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				return c.TogglePixel(image.Point{1, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠉')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "toggles a pixel on and sets options",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				return c.TogglePixel(image.Point{1, 0}, cell.FgColor(cell.ColorBlue))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠉', cell.FgColor(cell.ColorBlue))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "toggles a pixel off",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 0}); err != nil {
					return err
				}
				return c.TogglePixel(image.Point{0, 0})
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠈')

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "toggles a pixel off and sets options",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				if err := c.SetPixel(image.Point{0, 0}); err != nil {
					return err
				}
				if err := c.SetPixel(image.Point{1, 0}); err != nil {
					return err
				}
				return c.TogglePixel(image.Point{0, 0}, cell.FgColor(cell.ColorBlue))
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⠈', cell.FgColor(cell.ColorBlue))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc: "toggle fails on point outside of the canvas",
			ar:   image.Rect(0, 0, 1, 1),
			pixelOps: func(c *Canvas) error {
				return c.TogglePixel(image.Point{3, 3})
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			bc, err := New(tc.ar)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			err = tc.pixelOps(bc)
			if (err != nil) != tc.wantErr {
				t.Errorf("pixelOps => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			size := area.Size(tc.ar)
			gotApplied, err := faketerm.New(size)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}
			if err := bc.Apply(gotApplied); err != nil {
				t.Fatalf("bc.Apply => unexpected error: %v", err)
			}
			if diff := faketerm.Diff(tc.want(size), gotApplied); diff != "" {
				t.Fatalf("Direct Apply => %v", diff)
			}

			// When copied to another another canvas, the result on the
			// terminal must be the same.
			rc, err := canvas.New(tc.ar)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			if err := bc.CopyTo(rc); err != nil {
				t.Fatalf("CopyTo => unexpected error: %v", err)
			}

			gotCopied, err := faketerm.New(size)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}
			if err := rc.Apply(gotCopied); err != nil {
				t.Fatalf("rc.Apply => unexpected error: %v", err)
			}
			if diff := faketerm.Diff(tc.want(size), gotCopied); diff != "" {
				t.Fatalf("Copy then Apply => %v", diff)
			}
		})
	}
}
