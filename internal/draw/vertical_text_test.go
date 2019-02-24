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

package draw

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/internal/canvas"
	"github.com/mum4k/termdash/internal/canvas/testcanvas"
	"github.com/mum4k/termdash/internal/cell"
	"github.com/mum4k/termdash/internal/terminal/faketerm"
)

func TestVerticalText(t *testing.T) {
	tests := []struct {
		desc    string
		canvas  image.Rectangle
		text    string
		start   image.Point
		opts    []VerticalTextOption
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:   "start falls outside of the canvas",
			canvas: image.Rect(0, 0, 2, 2),
			start:  image.Point{2, 2},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "unsupported overrun mode specified",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "ab",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextOverrunMode(OverrunMode(-1)),
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "zero text",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "",
			start:  image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:   "text falls outside of the canvas on OverrunModeStrict",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "ab",
			start:  image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "text falls outside of the canvas because the rune is full-width on OverrunModeStrict",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "界",
			start:  image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "text falls outside of the canvas on OverrunModeTrim",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "ab",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextOverrunMode(OverrunModeTrim),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, 'a')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "text falls outside of the canvas because the rune is full-width on OverrunModeTrim",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "界",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextOverrunMode(OverrunModeTrim),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "OverrunModeTrim trims longer text",
			canvas: image.Rect(0, 0, 1, 2),
			text:   "abcdef",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextOverrunMode(OverrunModeTrim),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, 'a')
				testcanvas.MustSetCell(c, image.Point{0, 1}, 'b')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "OverrunModeTrim trims longer text with full-width runes, trim falls before the rune",
			canvas: image.Rect(0, 0, 1, 2),
			text:   "ab界",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextOverrunMode(OverrunModeTrim),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, 'a')
				testcanvas.MustSetCell(c, image.Point{0, 1}, 'b')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "OverrunModeTrim trims longer text with full-width runes, trim falls on the rune",
			canvas: image.Rect(0, 0, 1, 2),
			text:   "a界",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextOverrunMode(OverrunModeTrim),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, 'a')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "text falls outside of the canvas on OverrunModeThreeDot",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "ab",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextOverrunMode(OverrunModeThreeDot),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '…')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "text falls outside of the canvas because the rune is full-width on OverrunModeThreeDot",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "界",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextOverrunMode(OverrunModeThreeDot),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '…')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "OverrunModeThreeDot trims longer text",
			canvas: image.Rect(0, 0, 1, 2),
			text:   "abcdef",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextOverrunMode(OverrunModeThreeDot),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, 'a')
				testcanvas.MustSetCell(c, image.Point{0, 1}, '…')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "OverrunModeThreeDot trims longer text with full-width runes, trim falls before the rune",
			canvas: image.Rect(0, 0, 1, 2),
			text:   "ab界",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextOverrunMode(OverrunModeThreeDot),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, 'a')
				testcanvas.MustSetCell(c, image.Point{0, 1}, '…')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "OverrunModeThreeDot trims longer text with full-width runes, trim falls on the rune",
			canvas: image.Rect(0, 0, 1, 2),
			text:   "a界",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextOverrunMode(OverrunModeThreeDot),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, 'a')
				testcanvas.MustSetCell(c, image.Point{0, 1}, '…')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "requested MaxY is negative",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextMaxY(-1),
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "requested MaxY is greater than canvas height",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "",
			start:  image.Point{0, 0},
			opts: []VerticalTextOption{
				VerticalTextMaxY(2),
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "text falls outside of requested MaxY",
			canvas: image.Rect(0, 0, 2, 3),
			text:   "ab",
			start:  image.Point{1, 1},
			opts: []VerticalTextOption{
				VerticalTextMaxY(2),
			},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
			wantErr: true,
		},
		{
			desc:   "text is empty, nothing to do",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "",
			start:  image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				return faketerm.MustNew(size)
			},
		},
		{
			desc:   "draws text",
			canvas: image.Rect(0, 0, 2, 3),
			text:   "ab",
			start:  image.Point{1, 1},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{1, 1}, 'a')
				testcanvas.MustSetCell(c, image.Point{1, 2}, 'b')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws text with cell options",
			canvas: image.Rect(0, 0, 2, 3),
			text:   "ab",
			start:  image.Point{1, 1},
			opts: []VerticalTextOption{
				VerticalTextCellOpts(cell.FgColor(cell.ColorRed)),
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{1, 1}, 'a', cell.FgColor(cell.ColorRed))
				testcanvas.MustSetCell(c, image.Point{1, 2}, 'b', cell.FgColor(cell.ColorRed))
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws a half-width unicode character",
			canvas: image.Rect(0, 0, 1, 1),
			text:   "⇄",
			start:  image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⇄')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws multiple half-width unicode characters",
			canvas: image.Rect(0, 0, 3, 3),
			text:   "⇄࿃°",
			start:  image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '⇄')
				testcanvas.MustSetCell(c, image.Point{0, 1}, '࿃')
				testcanvas.MustSetCell(c, image.Point{0, 2}, '°')
				testcanvas.MustApply(c, ft)
				return ft
			},
		},
		{
			desc:   "draws multiple full-width unicode characters",
			canvas: image.Rect(0, 0, 3, 10),
			text:   "你好，世界",
			start:  image.Point{0, 0},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				c := testcanvas.MustNew(ft.Area())

				testcanvas.MustSetCell(c, image.Point{0, 0}, '你')
				testcanvas.MustSetCell(c, image.Point{0, 2}, '好')
				testcanvas.MustSetCell(c, image.Point{0, 4}, '，')
				testcanvas.MustSetCell(c, image.Point{0, 6}, '世')
				testcanvas.MustSetCell(c, image.Point{0, 8}, '界')
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

			err = VerticalText(c, tc.text, tc.start, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("VerticalText => unexpected error: %v, wantErr: %v", err, tc.wantErr)
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
				t.Errorf("VerticalText => %v", diff)
			}
		})
	}
}
