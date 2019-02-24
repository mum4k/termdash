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

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/canvas/testcanvas"
	"github.com/mum4k/termdash/internal/faketerm"
	"github.com/mum4k/termdash/linestyle"
)

func TestBorder(t *testing.T) {
	tests := []struct {
		desc    string
		canvas  image.Rectangle
		border  image.Rectangle
		opts    []BorderOption
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:   "border is larger than canvas",
			canvas: image.Rect(0, 0, 1, 1),
			border: image.Rect(0, 0, 2, 2),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "border is too small",
			canvas: image.Rect(0, 0, 2, 2),
			border: image.Rect(0, 0, 1, 1),
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "unsupported line style",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(0, 0, 2, 2),
			opts: []BorderOption{
				BorderLineStyle(linestyle.LineStyle(-1)),
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "draws border around the canvas",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(0, 0, 4, 4),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, lineStyleChars[linestyle.Light][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 3}, lineStyleChars[linestyle.Light][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 0}, lineStyleChars[linestyle.Light][hLine])
				testcanvas.MustSetCell(c, image.Point{1, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{2, 0}, lineStyleChars[linestyle.Light][hLine])
				testcanvas.MustSetCell(c, image.Point{2, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{3, 0}, lineStyleChars[linestyle.Light][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{3, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 3}, lineStyleChars[linestyle.Light][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws double border around the canvas",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(0, 0, 4, 4),
			opts: []BorderOption{
				BorderLineStyle(linestyle.Double),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, lineStyleChars[linestyle.Double][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, lineStyleChars[linestyle.Double][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, lineStyleChars[linestyle.Double][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 3}, lineStyleChars[linestyle.Double][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 0}, lineStyleChars[linestyle.Double][hLine])
				testcanvas.MustSetCell(c, image.Point{1, 3}, lineStyleChars[linestyle.Double][hLine])

				testcanvas.MustSetCell(c, image.Point{2, 0}, lineStyleChars[linestyle.Double][hLine])
				testcanvas.MustSetCell(c, image.Point{2, 3}, lineStyleChars[linestyle.Double][hLine])

				testcanvas.MustSetCell(c, image.Point{3, 0}, lineStyleChars[linestyle.Double][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{3, 1}, lineStyleChars[linestyle.Double][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 2}, lineStyleChars[linestyle.Double][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 3}, lineStyleChars[linestyle.Double][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws round border around the canvas",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(0, 0, 4, 4),
			opts: []BorderOption{
				BorderLineStyle(linestyle.Round),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, lineStyleChars[linestyle.Round][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, lineStyleChars[linestyle.Round][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, lineStyleChars[linestyle.Round][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 3}, lineStyleChars[linestyle.Round][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 0}, lineStyleChars[linestyle.Round][hLine])
				testcanvas.MustSetCell(c, image.Point{1, 3}, lineStyleChars[linestyle.Round][hLine])

				testcanvas.MustSetCell(c, image.Point{2, 0}, lineStyleChars[linestyle.Round][hLine])
				testcanvas.MustSetCell(c, image.Point{2, 3}, lineStyleChars[linestyle.Round][hLine])

				testcanvas.MustSetCell(c, image.Point{3, 0}, lineStyleChars[linestyle.Round][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{3, 1}, lineStyleChars[linestyle.Round][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 2}, lineStyleChars[linestyle.Round][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 3}, lineStyleChars[linestyle.Round][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws border in the canvas",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(1, 1, 3, 3),
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{1, 1}, lineStyleChars[linestyle.Light][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{1, 2}, lineStyleChars[linestyle.Light][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{2, 1}, lineStyleChars[linestyle.Light][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{2, 2}, lineStyleChars[linestyle.Light][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws border with cell options",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(1, 1, 3, 3),
			opts: []BorderOption{
				BorderCellOpts(cell.FgColor(cell.ColorRed)),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{1, 1},
					lineStyleChars[linestyle.Light][topLeftCorner], cell.FgColor(cell.ColorRed))
				testcanvas.MustSetCell(c, image.Point{1, 2},
					lineStyleChars[linestyle.Light][bottomLeftCorner], cell.FgColor(cell.ColorRed))

				testcanvas.MustSetCell(c, image.Point{2, 1},
					lineStyleChars[linestyle.Light][topRightCorner], cell.FgColor(cell.ColorRed))
				testcanvas.MustSetCell(c, image.Point{2, 2},
					lineStyleChars[linestyle.Light][bottomRightCorner], cell.FgColor(cell.ColorRed))

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws border with a title",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(0, 0, 4, 4),
			opts: []BorderOption{
				BorderTitle("ab", OverrunModeStrict),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, lineStyleChars[linestyle.Light][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 3}, lineStyleChars[linestyle.Light][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 0}, 'a')
				testcanvas.MustSetCell(c, image.Point{1, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{2, 0}, 'b')
				testcanvas.MustSetCell(c, image.Point{2, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{3, 0}, lineStyleChars[linestyle.Light][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{3, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 3}, lineStyleChars[linestyle.Light][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws border with a title and cell options",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(0, 0, 4, 4),
			opts: []BorderOption{
				BorderTitle("ab", OverrunModeStrict, cell.FgColor(cell.ColorRed)),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, lineStyleChars[linestyle.Light][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 3}, lineStyleChars[linestyle.Light][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 0}, 'a', cell.FgColor(cell.ColorRed))
				testcanvas.MustSetCell(c, image.Point{1, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{2, 0}, 'b', cell.FgColor(cell.ColorRed))
				testcanvas.MustSetCell(c, image.Point{2, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{3, 0}, lineStyleChars[linestyle.Light][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{3, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 3}, lineStyleChars[linestyle.Light][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "fails to draw border with a too long title in strict mode",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(0, 0, 4, 4),
			opts: []BorderOption{
				BorderTitle("abc", OverrunModeStrict),
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "doesn't draw the border if there isn't enough space",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(1, 1, 3, 3),
			opts: []BorderOption{
				BorderTitle("abc", OverrunModeStrict),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{1, 1}, lineStyleChars[linestyle.Light][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{1, 2}, lineStyleChars[linestyle.Light][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{2, 1}, lineStyleChars[linestyle.Light][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{2, 2}, lineStyleChars[linestyle.Light][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws border with a trimmed title",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(0, 0, 4, 4),
			opts: []BorderOption{
				BorderTitle("abc", OverrunModeTrim),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, lineStyleChars[linestyle.Light][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 3}, lineStyleChars[linestyle.Light][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 0}, 'a')
				testcanvas.MustSetCell(c, image.Point{1, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{2, 0}, 'b')
				testcanvas.MustSetCell(c, image.Point{2, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{3, 0}, lineStyleChars[linestyle.Light][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{3, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 3}, lineStyleChars[linestyle.Light][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws border with a title shortened using the horizontal ellipsis rune",
			canvas: image.Rect(0, 0, 4, 4),
			border: image.Rect(0, 0, 4, 4),
			opts: []BorderOption{
				BorderTitle("abc", OverrunModeThreeDot),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, lineStyleChars[linestyle.Light][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 3}, lineStyleChars[linestyle.Light][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 0}, 'a')
				testcanvas.MustSetCell(c, image.Point{1, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{2, 0}, '…')
				testcanvas.MustSetCell(c, image.Point{2, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{3, 0}, lineStyleChars[linestyle.Light][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{3, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{3, 3}, lineStyleChars[linestyle.Light][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "aligns the title to the left",
			canvas: image.Rect(0, 0, 6, 4),
			border: image.Rect(0, 0, 6, 4),
			opts: []BorderOption{
				BorderTitle("ab", OverrunModeStrict),
				BorderTitleAlign(align.HorizontalLeft),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, lineStyleChars[linestyle.Light][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 3}, lineStyleChars[linestyle.Light][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 0}, 'a')
				testcanvas.MustSetCell(c, image.Point{1, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{2, 0}, 'b')
				testcanvas.MustSetCell(c, image.Point{2, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{3, 0}, lineStyleChars[linestyle.Light][hLine])
				testcanvas.MustSetCell(c, image.Point{3, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{4, 0}, lineStyleChars[linestyle.Light][hLine])
				testcanvas.MustSetCell(c, image.Point{4, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{5, 0}, lineStyleChars[linestyle.Light][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{5, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{5, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{5, 3}, lineStyleChars[linestyle.Light][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "aligns the title in the center",
			canvas: image.Rect(0, 0, 6, 4),
			border: image.Rect(0, 0, 6, 4),
			opts: []BorderOption{
				BorderTitle("ab", OverrunModeStrict),
				BorderTitleAlign(align.HorizontalCenter),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, lineStyleChars[linestyle.Light][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 3}, lineStyleChars[linestyle.Light][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 0}, lineStyleChars[linestyle.Light][hLine])
				testcanvas.MustSetCell(c, image.Point{1, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{2, 0}, 'a')
				testcanvas.MustSetCell(c, image.Point{2, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{3, 0}, 'b')
				testcanvas.MustSetCell(c, image.Point{3, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{4, 0}, lineStyleChars[linestyle.Light][hLine])
				testcanvas.MustSetCell(c, image.Point{4, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{5, 0}, lineStyleChars[linestyle.Light][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{5, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{5, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{5, 3}, lineStyleChars[linestyle.Light][bottomRightCorner])

				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "aligns the title to the right",
			canvas: image.Rect(0, 0, 6, 4),
			border: image.Rect(0, 0, 6, 4),
			opts: []BorderOption{
				BorderTitle("ab", OverrunModeStrict),
				BorderTitleAlign(align.HorizontalRight),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, lineStyleChars[linestyle.Light][topLeftCorner])
				testcanvas.MustSetCell(c, image.Point{0, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{0, 3}, lineStyleChars[linestyle.Light][bottomLeftCorner])

				testcanvas.MustSetCell(c, image.Point{1, 0}, lineStyleChars[linestyle.Light][hLine])
				testcanvas.MustSetCell(c, image.Point{1, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{2, 0}, lineStyleChars[linestyle.Light][hLine])
				testcanvas.MustSetCell(c, image.Point{2, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{3, 0}, 'a')
				testcanvas.MustSetCell(c, image.Point{3, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{4, 0}, 'b')
				testcanvas.MustSetCell(c, image.Point{4, 3}, lineStyleChars[linestyle.Light][hLine])

				testcanvas.MustSetCell(c, image.Point{5, 0}, lineStyleChars[linestyle.Light][topRightCorner])
				testcanvas.MustSetCell(c, image.Point{5, 1}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{5, 2}, lineStyleChars[linestyle.Light][vLine])
				testcanvas.MustSetCell(c, image.Point{5, 3}, lineStyleChars[linestyle.Light][bottomRightCorner])

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

			err = Border(c, tc.border, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("Border => unexpected error: %v, wantErr: %v", err, tc.wantErr)
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
				t.Errorf("Border => %v", diff)
			}
		})
	}
}
