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
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/faketerm"
)

func TestHVLines(t *testing.T) {
	tests := []struct {
		desc    string
		canvas  image.Rectangle // Size of the canvas for the test.
		lines   []HVLine
		opts    []HVLineOption
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:   "fails when line isn't horizontal or vertical",
			canvas: image.Rect(0, 0, 2, 2),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{1, 1},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "fails when start isn't in the canvas",
			canvas: image.Rect(0, 0, 1, 1),
			lines: []HVLine{
				{
					Start: image.Point{2, 0},
					End:   image.Point{0, 0},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "fails when end isn't in the canvas",
			canvas: image.Rect(0, 0, 1, 1),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 2},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "fails when the line has zero length",
			canvas: image.Rect(0, 0, 1, 1),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 0},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "draws single horizontal line",
			canvas: image.Rect(0, 0, 3, 1),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{2, 0},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{2, 0}, parts[hLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "respects line style set explicitly",
			canvas: image.Rect(0, 0, 3, 1),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{2, 0},
				},
			},
			opts: []HVLineOption{
				HVLineStyle(linestyle.Light),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{2, 0}, parts[hLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "respects cell options",
			canvas: image.Rect(0, 0, 3, 1),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{2, 0},
				},
			},
			opts: []HVLineOption{
				HVLineCellOpts(
					cell.FgColor(cell.ColorYellow),
					cell.BgColor(cell.ColorBlue),
				),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[hLine],
					cell.FgColor(cell.ColorYellow),
					cell.BgColor(cell.ColorBlue),
				)
				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[hLine],
					cell.FgColor(cell.ColorYellow),
					cell.BgColor(cell.ColorBlue),
				)
				testcanvas.MustSetCell(c, image.Point{2, 0}, parts[hLine],
					cell.FgColor(cell.ColorYellow),
					cell.BgColor(cell.ColorBlue),
				)

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws single horizontal line, supplied in reverse direction",
			canvas: image.Rect(0, 0, 3, 1),
			lines: []HVLine{
				{
					Start: image.Point{1, 0},
					End:   image.Point{0, 0},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[hLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws single vertical line",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{1, 0},
					End:   image.Point{1, 2},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{1, 1}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{1, 2}, parts[vLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws single vertical line, supplied in reverse direction",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{1, 1},
					End:   image.Point{1, 0},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{1, 1}, parts[vLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "parallel horizontal lines don't affect each other",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{2, 0},
				},
				{
					Start: image.Point{0, 1},
					End:   image.Point{2, 1},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{2, 0}, parts[hLine])

				testcanvas.MustSetCell(c, image.Point{0, 1}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{1, 1}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{2, 1}, parts[hLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "parallel vertical lines don't affect each other",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 2},
				},
				{
					Start: image.Point{1, 0},
					End:   image.Point{1, 2},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{0, 1}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, parts[vLine])

				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{1, 1}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{1, 2}, parts[vLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "perpendicular lines that don't cross don't affect each other",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 2},
				},
				{
					Start: image.Point{1, 1},
					End:   image.Point{2, 1},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{0, 1}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, parts[vLine])

				testcanvas.MustSetCell(c, image.Point{1, 1}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{2, 1}, parts[hLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws top left corner",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 2},
				},
				{
					Start: image.Point{0, 0},
					End:   image.Point{2, 0},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, parts[vLine])

				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{2, 0}, parts[hLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws top right corner",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{2, 0},
					End:   image.Point{2, 2},
				},
				{
					Start: image.Point{0, 0},
					End:   image.Point{2, 0},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{2, 0}, parts[topRightCorner])
				testcanvas.MustSetCell(c, image.Point{2, 1}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{2, 2}, parts[vLine])

				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[hLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws bottom left corner",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 2},
				},
				{
					Start: image.Point{0, 2},
					End:   image.Point{2, 2},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{0, 1}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, parts[bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 2}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{2, 2}, parts[hLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws bottom right corner",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{2, 0},
					End:   image.Point{2, 2},
				},
				{
					Start: image.Point{0, 2},
					End:   image.Point{2, 2},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{2, 0}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{2, 1}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{2, 2}, parts[bottomRightCorner])

				testcanvas.MustSetCell(c, image.Point{0, 2}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{1, 2}, parts[hLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws T horizontal and up",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{0, 2},
					End:   image.Point{2, 2},
				},
				{
					Start: image.Point{1, 0},
					End:   image.Point{1, 2},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 2}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{1, 2}, parts[hAndUp])
				testcanvas.MustSetCell(c, image.Point{2, 2}, parts[hLine])

				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{1, 1}, parts[vLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws T horizontal and down",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{0, 0},
					End:   image.Point{2, 0},
				},
				{
					Start: image.Point{1, 0},
					End:   image.Point{1, 2},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[hAndDown])
				testcanvas.MustSetCell(c, image.Point{2, 0}, parts[hLine])

				testcanvas.MustSetCell(c, image.Point{1, 1}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{1, 2}, parts[vLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws T vertical and left",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{0, 1},
					End:   image.Point{2, 1},
				},
				{
					Start: image.Point{2, 0},
					End:   image.Point{2, 2},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 1}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{1, 1}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{2, 1}, parts[vAndLeft])

				testcanvas.MustSetCell(c, image.Point{2, 0}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{2, 2}, parts[vLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws T vertical and right",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{0, 1},
					End:   image.Point{2, 1},
				},
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 2},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 1}, parts[vAndRight])
				testcanvas.MustSetCell(c, image.Point{1, 1}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{2, 1}, parts[hLine])

				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, parts[vLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws a cross",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				{
					Start: image.Point{0, 1},
					End:   image.Point{2, 1},
				},
				{
					Start: image.Point{1, 0},
					End:   image.Point{1, 2},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 1}, parts[hLine])
				testcanvas.MustSetCell(c, image.Point{1, 1}, parts[vAndH])
				testcanvas.MustSetCell(c, image.Point{2, 1}, parts[hLine])

				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[vLine])
				testcanvas.MustSetCell(c, image.Point{1, 2}, parts[vLine])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws multiple crossings",
			canvas: image.Rect(0, 0, 3, 3),
			lines: []HVLine{
				// Three horizontal lines.
				{
					Start: image.Point{0, 0},
					End:   image.Point{2, 0},
				},
				{
					Start: image.Point{0, 1},
					End:   image.Point{2, 1},
				},
				{
					Start: image.Point{0, 2},
					End:   image.Point{2, 2},
				},
				// Three vertical lines.
				{
					Start: image.Point{0, 0},
					End:   image.Point{0, 2},
				},
				{
					Start: image.Point{1, 0},
					End:   image.Point{1, 2},
				},
				{
					Start: image.Point{2, 0},
					End:   image.Point{2, 2},
				},
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				parts := lineStyleChars[linestyle.Light]
				testcanvas.MustSetCell(c, image.Point{0, 0}, parts[topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{1, 0}, parts[hAndDown])
				testcanvas.MustSetCell(c, image.Point{2, 0}, parts[topRightCorner])

				testcanvas.MustSetCell(c, image.Point{0, 1}, parts[vAndRight])
				testcanvas.MustSetCell(c, image.Point{1, 1}, parts[vAndH])
				testcanvas.MustSetCell(c, image.Point{2, 1}, parts[vAndLeft])

				testcanvas.MustSetCell(c, image.Point{0, 2}, parts[bottomLeftCorner])
				testcanvas.MustSetCell(c, image.Point{1, 2}, parts[hAndUp])
				testcanvas.MustSetCell(c, image.Point{2, 2}, parts[bottomRightCorner])

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

			err = HVLines(c, tc.lines, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("HVLines => unexpected error: %v, wantErr: %v", err, tc.wantErr)
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
				t.Errorf("HVLines => %v", diff)
			}
		})
	}
}
