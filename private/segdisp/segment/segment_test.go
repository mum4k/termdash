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

package segment

import (
	"fmt"
	"image"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas/braille"
	"github.com/mum4k/termdash/private/canvas/braille/testbraille"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/draw/testdraw"
	"github.com/mum4k/termdash/private/faketerm"
)

func TestHV(t *testing.T) {
	tests := []struct {
		desc       string
		opts       []Option
		cellCanvas image.Rectangle // Canvas in cells that will be converted to braille canvas for drawing.
		ar         image.Rectangle
		st         Type
		want       func(size image.Point) *faketerm.Terminal
		wantErr    bool
	}{
		{
			desc:       "fails on area with negative Min.X",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(-1, 0, 1, 1),
			st:         Horizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on area with negative Min.Y",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, -1, 1, 1),
			st:         Horizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on area with negative Max.X",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rectangle{image.Point{0, 0}, image.Point{-1, 1}},
			st:         Horizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on area with negative Max.Y",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rectangle{image.Point{0, 0}, image.Point{1, -1}},
			st:         Horizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on area with zero Dx()",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 0, 1),
			st:         Horizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on area with zero Dy()",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 0),
			st:         Horizontal,
			wantErr:    true,
		},
		{
			desc:       "fails on unsupported segment type (too small)",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 2),
			st:         Type(0),
			wantErr:    true,
		},
		{
			desc:       "fails on unsupported segment type (too large)",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 2),
			st:         Type(int(Vertical) + 1),
			wantErr:    true,
		},
		{
			desc:       "fails on area larger than the canvas",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 3, 1),
			st:         Horizontal,
			wantErr:    true,
		},
		{
			desc: "sets cell options",
			opts: []Option{
				CellOpts(
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorGreen),
				),
			},
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 1),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0}, cell.FgColor(cell.ColorRed), cell.BgColor(cell.ColorGreen))
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 1x1",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 1),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 1x2",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 2),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 1x3",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 3),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{0, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 1x4",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 4),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{0, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 1x5",
			cellCanvas: image.Rect(0, 0, 1, 2),
			ar:         image.Rect(0, 0, 1, 5),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{0, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 4}, image.Point{0, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 2x1",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 1),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 2x2",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 2),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 2x3",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 3),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 2x4",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 4),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{1, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 2x5",
			cellCanvas: image.Rect(0, 0, 1, 2),
			ar:         image.Rect(0, 0, 2, 5),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{1, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 4}, image.Point{1, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 3x1",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 1),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{2, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 3x2",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 2),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "horizontal, segment 3x2, reverse slopes",
			opts: []Option{
				ReverseSlopes(),
			},
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 2),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{1, 1}, image.Point{1, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 3x3",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 3),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{1, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "horizontal, segment 3x3, reverse slopes has no effect on larger height",
			opts: []Option{
				ReverseSlopes(),
			},
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 3),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{1, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "horizontal, segment 3x3, skip slopes",
			opts: []Option{
				SkipSlopesLTE(3),
			},
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 3),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "horizontal, segment 3x3, doesn't skip slopes because is taller",
			opts: []Option{
				SkipSlopesLTE(2),
			},
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 3),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{1, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 3x4",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 4),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{1, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 3x5",
			cellCanvas: image.Rect(0, 0, 2, 2),
			ar:         image.Rect(0, 0, 3, 5),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{1, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{1, 3})
				testdraw.MustBrailleLine(bc, image.Point{1, 4}, image.Point{1, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 4x1",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 1),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 4x2",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 2),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{3, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 4x3",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 3),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{2, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 4x4",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{2, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 4x5",
			cellCanvas: image.Rect(0, 0, 2, 2),
			ar:         image.Rect(0, 0, 4, 5),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{1, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{2, 3})
				testdraw.MustBrailleLine(bc, image.Point{1, 4}, image.Point{2, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 5x1",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 1),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{4, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 5x2",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 2),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 5x3",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 3),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{3, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 5x4",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 4),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{3, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "horizontal, segment 5x5",
			cellCanvas: image.Rect(0, 0, 3, 2),
			ar:         image.Rect(0, 0, 5, 5),
			st:         Horizontal,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{1, 1}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{3, 3})
				testdraw.MustBrailleLine(bc, image.Point{2, 4}, image.Point{2, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 1x1",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 1),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testbraille.MustSetPixel(bc, image.Point{0, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 1x2",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 2),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 1x3",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 3),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{0, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 1x4",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 4),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{0, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 1x5",
			cellCanvas: image.Rect(0, 0, 1, 2),
			ar:         image.Rect(0, 0, 1, 5),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{0, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 4}, image.Point{0, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 2x1",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 1),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 2x2",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 2),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 2x3",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 3),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{1, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "vertical, segment 2x3, reverse slopes",
			opts: []Option{
				ReverseSlopes(),
			},
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 3),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{0, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 2x4",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 4),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{1, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 2x5",
			cellCanvas: image.Rect(0, 0, 1, 2),
			ar:         image.Rect(0, 0, 2, 5),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{1, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{1, 3})
				testdraw.MustBrailleLine(bc, image.Point{1, 4}, image.Point{1, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 3x1",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 1),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{2, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 3x2",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 2),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 3x3",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 3),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{1, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "vertical, segment 3x3, reverse slopes has no effect on larger width",
			opts: []Option{
				ReverseSlopes(),
			},
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 3),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{1, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "vertical, segment 3x3, skips slopes",
			opts: []Option{
				SkipSlopesLTE(3),
			},
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 3),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "vertical, segment 3x3, doesn't skips slopes because is wider",
			opts: []Option{
				SkipSlopesLTE(2),
			},
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 3),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{1, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 3x4",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 3, 4),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{1, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 3x5",
			cellCanvas: image.Rect(0, 0, 2, 2),
			ar:         image.Rect(0, 0, 3, 5),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{1, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{2, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{2, 3})
				testdraw.MustBrailleLine(bc, image.Point{1, 4}, image.Point{1, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 4x1",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 1),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 4x2",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 2),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{3, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 4x3",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 3),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 2}, image.Point{2, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 4x4",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{2, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 4x5",
			cellCanvas: image.Rect(0, 0, 2, 2),
			ar:         image.Rect(0, 0, 4, 5),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{3, 3})
				testdraw.MustBrailleLine(bc, image.Point{1, 4}, image.Point{2, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 5x1",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 1),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{4, 0})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 5x2",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 2),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{4, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 5x3",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 3),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				testdraw.MustBrailleLine(bc, image.Point{2, 2}, image.Point{2, 2})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 5x4",
			cellCanvas: image.Rect(0, 0, 3, 1),
			ar:         image.Rect(0, 0, 5, 4),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{4, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2})
				testdraw.MustBrailleLine(bc, image.Point{2, 3}, image.Point{2, 3})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "vertical, segment 5x5",
			cellCanvas: image.Rect(0, 0, 3, 2),
			ar:         image.Rect(0, 0, 5, 5),
			st:         Vertical,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{2, 0})
				testdraw.MustBrailleLine(bc, image.Point{1, 1}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{4, 2})
				testdraw.MustBrailleLine(bc, image.Point{1, 3}, image.Point{3, 3})
				testdraw.MustBrailleLine(bc, image.Point{2, 4}, image.Point{2, 4})
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s st:%v", tc.desc, tc.st), func(t *testing.T) {
			bc, err := braille.New(tc.cellCanvas)
			if err != nil {
				t.Fatalf("braille.New => unexpected error: %v", err)
			}

			err = HV(bc, tc.ar, tc.st, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("HV => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			size := area.Size(tc.cellCanvas)
			want := faketerm.MustNew(size)
			if tc.want != nil {
				want = tc.want(size)
			}

			got, err := faketerm.New(size)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}
			if err := bc.Apply(got); err != nil {
				t.Fatalf("bc.Apply => unexpected error: %v", err)
			}
			if diff := faketerm.Diff(want, got); diff != "" {
				t.Fatalf("HV => %v", diff)
			}
		})
	}
}

// hvSegment is one horizontal or vertical segment.
type hvSegment struct {
	ar image.Rectangle
	st Type
}

// diagSegment is one diagonal segment.
type diagSegment struct {
	ar    image.Rectangle
	width int
	dt    DiagonalType
}

func TestAdjustHoriz(t *testing.T) {
	tests := []struct {
		desc      string
		start     image.Point
		end       image.Point
		segWidth  int
		adjust    int
		wantStart image.Point
		wantEnd   image.Point
	}{
		{
			desc: "no change for zero adjustment",
		},
		{
			desc:      "safe adjustments, points don't cross",
			start:     image.Point{0, 0},
			end:       image.Point{5, 0},
			segWidth:  6,
			adjust:    1,
			wantStart: image.Point{1, 0},
			wantEnd:   image.Point{4, 0},
		},
		{
			desc:      "safe adjustments, points land on each other",
			start:     image.Point{0, 0},
			end:       image.Point{4, 0},
			segWidth:  5,
			adjust:    2,
			wantStart: image.Point{2, 0},
			wantEnd:   image.Point{2, 0},
		},

		{
			desc:      "points cross, width divides evenly",
			start:     image.Point{0, 0},
			end:       image.Point{5, 0},
			segWidth:  6,
			adjust:    3,
			wantStart: image.Point{2, 0},
			wantEnd:   image.Point{3, 0},
		},
		{
			desc:      "points cross, width divides oddly",
			start:     image.Point{0, 0},
			end:       image.Point{6, 0},
			segWidth:  7,
			adjust:    4,
			wantStart: image.Point{3, 0},
			wantEnd:   image.Point{3, 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotStart, gotEnd := adjustHoriz(tc.start, tc.end, tc.segWidth, tc.adjust)
			if !gotStart.Eq(tc.wantStart) || !gotEnd.Eq(tc.wantEnd) {
				t.Errorf("adjustHoriz(%v, %v, %v, %v) => %v, %v, want %v, %v", tc.start, tc.end, tc.segWidth, tc.adjust, gotStart, gotEnd, tc.wantStart, tc.wantEnd)
			}

		})
	}
}

func TestAdjustVert(t *testing.T) {
	tests := []struct {
		desc      string
		start     image.Point
		end       image.Point
		segHeight int
		adjust    int
		wantStart image.Point
		wantEnd   image.Point
	}{
		{
			desc: "no change for zero adjustment",
		},
		{
			desc:      "safe adjustments, points don't cross",
			start:     image.Point{0, 0},
			end:       image.Point{0, 5},
			segHeight: 6,
			adjust:    1,
			wantStart: image.Point{0, 1},
			wantEnd:   image.Point{0, 4},
		},
		{
			desc:      "safe adjustments, points land on each other",
			start:     image.Point{0, 0},
			end:       image.Point{0, 4},
			segHeight: 5,
			adjust:    2,
			wantStart: image.Point{0, 2},
			wantEnd:   image.Point{0, 2},
		},

		{
			desc:      "points cross, width divides evenly",
			start:     image.Point{0, 0},
			end:       image.Point{0, 5},
			segHeight: 6,
			adjust:    3,
			wantStart: image.Point{0, 2},
			wantEnd:   image.Point{0, 3},
		},
		{
			desc:      "points cross, width divides oddly",
			start:     image.Point{0, 0},
			end:       image.Point{0, 6},
			segHeight: 7,
			adjust:    4,
			wantStart: image.Point{0, 3},
			wantEnd:   image.Point{0, 3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotStart, gotEnd := adjustVert(tc.start, tc.end, tc.segHeight, tc.adjust)
			if !gotStart.Eq(tc.wantStart) || !gotEnd.Eq(tc.wantEnd) {
				t.Errorf("adjustVert(%v, %v, %v, %v) => %v, %v, want %v, %v", tc.start, tc.end, tc.segHeight, tc.adjust, gotStart, gotEnd, tc.wantStart, tc.wantEnd)
			}

		})
	}
}

func TestDiagonal(t *testing.T) {
	tests := []struct {
		desc       string
		opts       []DiagonalOption
		cellCanvas image.Rectangle // Canvas in cells that will be converted to braille canvas for drawing.
		ar         image.Rectangle
		width      int
		dt         DiagonalType
		want       func(size image.Point) *faketerm.Terminal
		wantErr    bool
	}{
		{
			desc:       "fails on area with negative Min.X",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(-1, 0, 1, 1),
			width:      1,
			dt:         LeftToRight,
			wantErr:    true,
		},
		{
			desc:       "fails on area with negative Min.Y",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, -1, 1, 1),
			width:      1,
			dt:         LeftToRight,
			wantErr:    true,
		},
		{
			desc:       "fails on area with negative Max.X",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rectangle{image.Point{0, 0}, image.Point{-1, 1}},
			width:      1,
			dt:         LeftToRight,
			wantErr:    true,
		},
		{
			desc:       "fails on area with negative Max.Y",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rectangle{image.Point{0, 0}, image.Point{1, -1}},
			width:      1,
			dt:         LeftToRight,
			wantErr:    true,
		},
		{
			desc:       "fails on area with zero Dx()",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 0, 1),
			width:      1,
			dt:         LeftToRight,
			wantErr:    true,
		},
		{
			desc:       "fails on area with zero Dy()",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 1, 0),
			width:      1,
			dt:         LeftToRight,
			wantErr:    true,
		},
		{
			desc:       "fails on unsupported diagonal type (too small)",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 2),
			width:      1,
			dt:         DiagonalType(0),
			wantErr:    true,
		},
		{
			desc:       "fails on unsupported diagonal type (too large)",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 2, 2),
			width:      1,
			dt:         DiagonalType(int(RightToLeft) + 1),
			wantErr:    true,
		},
		{
			desc:       "fails on area larger than the canvas",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 3, 1),
			width:      1,
			dt:         LeftToRight,
			wantErr:    true,
		},
		{
			desc:       "fails on zero width",
			cellCanvas: image.Rect(0, 0, 1, 1),
			ar:         image.Rect(0, 0, 3, 1),
			width:      0,
			dt:         LeftToRight,
			wantErr:    true,
		},
		{
			desc:       "left to right, area 4x4, width 1",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         LeftToRight,
			width:      1,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "right to left, area 4x4, width 1",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         RightToLeft,
			width:      1,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{3, 0}, image.Point{0, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "left to right, area 4x4, width 2",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         LeftToRight,
			width:      2,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "right to left, area 4x4, width 2",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         RightToLeft,
			width:      2,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{3, 0}, image.Point{0, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "left to right, area 4x4, width 3",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         LeftToRight,
			width:      3,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "right to left, area 4x4, width 3",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         RightToLeft,
			width:      3,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{3, 0}, image.Point{0, 3})
				testdraw.MustBrailleLine(bc, image.Point{3, 1}, image.Point{1, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "left to right, area 8x4, width 3",
			cellCanvas: image.Rect(0, 0, 4, 1),
			ar:         image.Rect(0, 0, 8, 4),
			dt:         LeftToRight,
			width:      3,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{7, 3})
				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{7, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{6, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "right to left, area 8x4, width 3",
			cellCanvas: image.Rect(0, 0, 4, 1),
			ar:         image.Rect(0, 0, 8, 4),
			dt:         RightToLeft,
			width:      3,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{7, 0}, image.Point{0, 3})
				testdraw.MustBrailleLine(bc, image.Point{6, 0}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{7, 1}, image.Point{1, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "left to right, area 4x8, width 3",
			cellCanvas: image.Rect(0, 0, 2, 2),
			ar:         image.Rect(0, 0, 4, 8),
			dt:         LeftToRight,
			width:      3,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 7})
				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 6})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 7})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "right to left, area 4x8, width 3",
			cellCanvas: image.Rect(0, 0, 2, 2),
			ar:         image.Rect(0, 0, 4, 8),
			dt:         RightToLeft,
			width:      3,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{3, 0}, image.Point{0, 7})
				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{0, 6})
				testdraw.MustBrailleLine(bc, image.Point{3, 1}, image.Point{1, 7})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "left to right, area 4x4, width 4",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         LeftToRight,
			width:      4,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "right to left, area 4x4, width 4",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         RightToLeft,
			width:      4,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{3, 0}, image.Point{0, 3})
				testdraw.MustBrailleLine(bc, image.Point{3, 1}, image.Point{1, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "left to right, area 4x4, width 5",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         LeftToRight,
			width:      5,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "right to left, area 4x4, width 5",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         RightToLeft,
			width:      5,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{3, 0}, image.Point{0, 3})
				testdraw.MustBrailleLine(bc, image.Point{3, 1}, image.Point{1, 3})
				testdraw.MustBrailleLine(bc, image.Point{3, 2}, image.Point{2, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "left to right, area 4x4, width 6",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         LeftToRight,
			width:      6,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{3, 0}, image.Point{3, 0})
				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "right to left, area 4x4, width 6",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         RightToLeft,
			width:      6,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{3, 0}, image.Point{0, 3})
				testdraw.MustBrailleLine(bc, image.Point{3, 1}, image.Point{1, 3})
				testdraw.MustBrailleLine(bc, image.Point{3, 2}, image.Point{2, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "left to right, area 4x4, width 7",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         LeftToRight,
			width:      7,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{3, 0}, image.Point{3, 0})
				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{3, 1})
				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 2})
				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 1}, image.Point{2, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 2}, image.Point{1, 3})
				testdraw.MustBrailleLine(bc, image.Point{0, 3}, image.Point{3, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "right to left, area 4x4, width 7",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         RightToLeft,
			width:      7,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{0, 0})
				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{0, 1})
				testdraw.MustBrailleLine(bc, image.Point{2, 0}, image.Point{0, 2})
				testdraw.MustBrailleLine(bc, image.Point{3, 0}, image.Point{0, 3})
				testdraw.MustBrailleLine(bc, image.Point{3, 1}, image.Point{1, 3})
				testdraw.MustBrailleLine(bc, image.Point{3, 2}, image.Point{2, 3})
				testdraw.MustBrailleLine(bc, image.Point{3, 3}, image.Point{3, 3})

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "left to right, fails when width is larger than area",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         LeftToRight,
			width:      8,
			wantErr:    true,
		},
		{
			desc:       "right to left, fails when width is larger than area",
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         RightToLeft,
			width:      8,
			wantErr:    true,
		},
		{
			desc: "sets cell options",
			opts: []DiagonalOption{
				DiagonalCellOpts(
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorGreen),
				),
			},
			cellCanvas: image.Rect(0, 0, 2, 1),
			ar:         image.Rect(0, 0, 4, 4),
			dt:         LeftToRight,
			width:      2,
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				opts := []draw.BrailleLineOption{
					draw.BrailleLineCellOpts(
						cell.FgColor(cell.ColorRed),
						cell.BgColor(cell.ColorGreen),
					),
				}
				testdraw.MustBrailleLine(bc, image.Point{1, 0}, image.Point{3, 2}, opts...)
				testdraw.MustBrailleLine(bc, image.Point{0, 0}, image.Point{3, 3}, opts...)

				testbraille.MustApply(bc, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s dt:%v", tc.desc, tc.dt), func(t *testing.T) {
			bc, err := braille.New(tc.cellCanvas)
			if err != nil {
				t.Fatalf("braille.New => unexpected error: %v", err)
			}

			err = Diagonal(bc, tc.ar, tc.width, tc.dt, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("Diagonal => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			size := area.Size(tc.cellCanvas)
			want := faketerm.MustNew(size)
			if tc.want != nil {
				want = tc.want(size)
			}

			got, err := faketerm.New(size)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}
			if err := bc.Apply(got); err != nil {
				t.Fatalf("bc.Apply => unexpected error: %v", err)
			}
			if diff := faketerm.Diff(want, got); diff != "" {
				t.Fatalf("Diagonal => %v", diff)
			}
		})
	}
}
