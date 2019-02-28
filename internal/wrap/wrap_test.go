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

package wrap

import (
	"fmt"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/internal/canvas/buffer"
)

func TestNeeded(t *testing.T) {
	tests := []struct {
		desc  string
		r     rune
		posX  int
		width int
		mode  Mode
		want  bool
	}{
		{
			desc:  "half-width rune, falls within canvas",
			r:     'a',
			posX:  2,
			width: 3,
			want:  false,
		},
		{
			desc:  "full-width rune, falls within canvas",
			r:     '世',
			posX:  1,
			width: 3,
			want:  false,
		},
		{
			desc:  "half-width rune, falls outside of canvas, wrapping not configured",
			r:     'a',
			posX:  3,
			width: 3,
			want:  false,
		},
		{
			desc:  "full-width rune, starts outside of canvas, wrapping not configured",
			r:     '世',
			posX:  3,
			width: 3,
			want:  false,
		},
		{
			desc:  "half-width rune, falls outside of canvas, wrapping configured",
			r:     'a',
			posX:  3,
			width: 3,
			mode:  AtRunes,
			want:  true,
		},
		{
			desc:  "full-width rune, starts in and falls outside of canvas, wrapping configured",
			r:     '世',
			posX:  3,
			width: 3,
			mode:  AtRunes,
			want:  true,
		},
		{
			desc:  "full-width rune, starts outside of canvas, wrapping configured",
			r:     '世',
			posX:  3,
			width: 3,
			mode:  AtRunes,
			want:  true,
		},
		{
			desc:  "doesn't wrap for newline characters",
			r:     '\n',
			posX:  3,
			width: 3,
			mode:  AtRunes,
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := needed(tc.r, tc.posX, tc.width, tc.mode)
			if got != tc.want {
				t.Errorf("needed => got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCells(t *testing.T) {
	tests := []struct {
		desc  string
		cells []*buffer.Cell
		// width is the width of the canvas.
		width int
		mode  Mode
		want  [][]*buffer.Cell
	}{
		{
			desc:  "zero text",
			width: 1,
		},
		{
			desc:  "zero canvas width",
			cells: buffer.NewCells("hello"),
			width: 0,
			want:  nil,
		},
		{
			desc:  "wrapping disabled, no newlines, fits in canvas width",
			cells: buffer.NewCells("hello"),
			width: 5,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
			},
		},
		{
			desc:  "wrapping disabled, no newlines, doesn't fits in canvas width",
			cells: buffer.NewCells("hello"),
			width: 4,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
			},
		},
		{
			desc:  "wrapping disabled, newlines, fits in canvas width",
			cells: buffer.NewCells("hello\nworld"),
			width: 5,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
				buffer.NewCells("world"),
			},
		},
		{
			desc:  "wrapping disabled, newlines, doesn't fit in canvas width",
			cells: buffer.NewCells("hello\nworld"),
			width: 4,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
				buffer.NewCells("world"),
			},
		},
		{
			desc:  "wrapping enabled, no newlines, fits in canvas width",
			cells: buffer.NewCells("hello"),
			width: 5,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
			},
		},
		{
			desc:  "wrapping enabled, no newlines, doesn't fit in canvas width",
			cells: buffer.NewCells("hello"),
			width: 4,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("hell"),
				buffer.NewCells("o"),
			},
		},
		{
			desc:  "wrapping enabled, newlines, fits in canvas width",
			cells: buffer.NewCells("hello\nworld"),
			width: 5,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
				buffer.NewCells("world"),
			},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in canvas width",
			cells: buffer.NewCells("hello\nworld"),
			width: 4,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("hell"),
				buffer.NewCells("o"),
				buffer.NewCells("worl"),
				buffer.NewCells("d"),
			},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in canvas width, unicode characters",
			cells: buffer.NewCells("⇧\n…\n⇩"),
			width: 1,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("⇧"),
				buffer.NewCells("…"),
				buffer.NewCells("⇩"),
			},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in width, full-width unicode characters",
			cells: buffer.NewCells("你好\n世界"),
			width: 2,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("你"),
				buffer.NewCells("好"),
				buffer.NewCells("世"),
				buffer.NewCells("界"),
			},
		},
		{
			desc:  "wraps before a full-width character that starts in and falls out",
			cells: buffer.NewCells("a你b"),
			width: 2,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("a"),
				buffer.NewCells("你"),
				buffer.NewCells("b"),
			},
		},
		{
			desc:  "wraps before a full-width character that falls out",
			cells: buffer.NewCells("ab你b"),
			width: 2,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("ab"),
				buffer.NewCells("你"),
				buffer.NewCells("b"),
			},
		},
		{
			desc:  "handles leading and trailing newlines",
			cells: buffer.NewCells("\n\n\nhello\n\n\n"),
			width: 4,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells("hell"),
				buffer.NewCells("o"),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells(""),
			},
		},
		{
			desc:  "handles multiple newlines in the middle",
			cells: buffer.NewCells("hello\n\n\nworld"),
			width: 5,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("hello"),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells("world"),
			},
		},
		{
			desc:  "handles multiple newlines in the middle and wraps",
			cells: buffer.NewCells("hello\n\n\nworld"),
			width: 2,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells("he"),
				buffer.NewCells("ll"),
				buffer.NewCells("o"),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells("wo"),
				buffer.NewCells("rl"),
				buffer.NewCells("d"),
			},
		},
		{
			desc:  "contains only newlines",
			cells: buffer.NewCells("\n\n\n"),
			width: 4,
			mode:  AtRunes,
			want: [][]*buffer.Cell{
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells(""),
				buffer.NewCells(""),
			},
		},
		/*
			{
				desc:  "wraps at words, all fit individually",
				text:  "aaa bb cc ddddd",
				width: 5,
				mode:  AtRunes,
				want:  []int{0, 4, 7, 10},
			},*/

		// wraps at words - handles newline characters
		// wraps at words - handles leading and trailing newlines
		// wraps at words - handles continuous newlines
		// wraps at words, no need to wrap
		// wraps at words all individually fit within width
		// wraps at words, no spaces so goes back to AtRunes
		// wraps at words, one doesn't fit so goes back to AtRunes
		// wraps at words - full width runes - fit exactly
		// wraps at words - full width runes - cause a wrap
		// weird cases with multiple spaces between words
		// preserves cell options
		// Inserted cells have the same cell options ?
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Logf(fmt.Sprintf("Mode: %v", tc.mode))
			got := Cells(tc.cells, tc.width, tc.mode)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Cells =>\n got:%v\nwant:%v\nunexpected diff (-want, +got):\n%s", got, tc.want, diff)
			}
		})
	}

}
