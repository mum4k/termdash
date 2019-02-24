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

package draw

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/canvas/testcanvas"
	"github.com/mum4k/termdash/internal/faketerm"
)

func TestRectangle(t *testing.T) {
	tests := []struct {
		desc    string
		canvas  image.Rectangle
		rect    image.Rectangle
		opts    []RectangleOption
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:   "draws a 1x1 rectangle",
			canvas: image.Rect(0, 0, 2, 2),
			rect:   image.Rect(0, 0, 1, 1),
			opts: []RectangleOption{
				RectChar('x'),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, 'x')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "fails when the rectangle character occupies multiple cells",
			canvas: image.Rect(0, 0, 2, 2),
			rect:   image.Rect(0, 0, 1, 1),
			opts: []RectangleOption{
				RectChar('界'),
			},
			wantErr: true,
		},
		{
			desc:   "sets cell options",
			canvas: image.Rect(0, 0, 2, 2),
			rect:   image.Rect(0, 0, 1, 1),
			opts: []RectangleOption{
				RectChar('x'),
				RectCellOpts(
					cell.FgColor(cell.ColorBlue),
					cell.BgColor(cell.ColorRed),
				),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(
					c,
					image.Point{0, 0},
					'x',
					cell.FgColor(cell.ColorBlue),
					cell.BgColor(cell.ColorRed),
				)
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws a larger rectangle",
			canvas: image.Rect(0, 0, 10, 10),
			rect:   image.Rect(0, 0, 3, 2),
			opts: []RectangleOption{
				RectChar('o'),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, 'o')
				testcanvas.MustSetCell(c, image.Point{1, 0}, 'o')
				testcanvas.MustSetCell(c, image.Point{2, 0}, 'o')
				testcanvas.MustSetCell(c, image.Point{0, 1}, 'o')
				testcanvas.MustSetCell(c, image.Point{1, 1}, 'o')
				testcanvas.MustSetCell(c, image.Point{2, 1}, 'o')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "rectangle not in the corner of the canvas",
			canvas: image.Rect(0, 0, 10, 10),
			rect:   image.Rect(2, 1, 4, 4),
			opts: []RectangleOption{
				RectChar('o'),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{2, 1}, 'o')
				testcanvas.MustSetCell(c, image.Point{3, 1}, 'o')
				testcanvas.MustSetCell(c, image.Point{2, 2}, 'o')
				testcanvas.MustSetCell(c, image.Point{3, 2}, 'o')
				testcanvas.MustSetCell(c, image.Point{2, 3}, 'o')
				testcanvas.MustSetCell(c, image.Point{3, 3}, 'o')
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

			err = Rectangle(c, tc.rect, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("Rectangle => unexpected error: %v, wantErr: %v", err, tc.wantErr)
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

			if diff := faketerm.Diff(tc.want(c.Size()), got); diff != "" {
				t.Errorf("Rectangle => %v", diff)
			}
		})
	}
}
