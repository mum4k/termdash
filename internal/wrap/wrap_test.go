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
	"testing"

	"github.com/kylelemons/godebug/pretty"
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
			got := Needed(tc.r, tc.posX, tc.width, tc.mode)
			if got != tc.want {
				t.Errorf("Needed => got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLines(t *testing.T) {
	tests := []struct {
		desc string
		text string
		// width is the width of the canvas.
		width int
		mode  Mode
		want  []int
	}{
		{
			desc:  "zero text",
			text:  "",
			width: 1,
			want:  nil,
		},
		{
			desc:  "zero canvas width",
			text:  "hello",
			width: 0,
			want:  nil,
		},
		{
			desc:  "wrapping disabled, no newlines, fits in canvas width",
			text:  "hello",
			width: 5,
			want:  []int{0},
		},
		{
			desc:  "wrapping disabled, no newlines, doesn't fits in canvas width",
			text:  "hello",
			width: 4,
			want:  []int{0},
		},
		{
			desc:  "wrapping disabled, newlines, fits in canvas width",
			text:  "hello\nworld",
			width: 5,
			want:  []int{0, 6},
		},
		{
			desc:  "wrapping disabled, newlines, doesn't fit in canvas width",
			text:  "hello\nworld",
			width: 4,
			want:  []int{0, 6},
		},
		{
			desc:  "wrapping enabled, no newlines, fits in canvas width",
			text:  "hello",
			width: 5,
			mode:  AtRunes,
			want:  []int{0},
		},
		{
			desc:  "wrapping enabled, no newlines, doesn't fit in canvas width",
			text:  "hello",
			width: 4,
			mode:  AtRunes,
			want:  []int{0, 4},
		},
		{
			desc:  "wrapping enabled, newlines, fits in canvas width",
			text:  "hello\nworld",
			width: 5,
			mode:  AtRunes,
			want:  []int{0, 6},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in canvas width",
			text:  "hello\nworld",
			width: 4,
			mode:  AtRunes,
			want:  []int{0, 4, 6, 10},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in canvas width, unicode characters",
			text:  "⇧\n…\n⇩",
			width: 1,
			mode:  AtRunes,
			want:  []int{0, 4, 8},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in width, full-width unicode characters",
			text:  "你好\n世界",
			width: 2,
			mode:  AtRunes,
			want:  []int{0, 3, 7, 10},
		},
		{
			desc:  "wraps before a full-width character that starts in and falls out",
			text:  "a你b",
			width: 2,
			mode:  AtRunes,
			want:  []int{0, 1, 4},
		},
		{
			desc:  "wraps before a full-width character that falls out",
			text:  "ab你b",
			width: 2,
			mode:  AtRunes,
			want:  []int{0, 2, 5},
		},
		{
			desc:  "handles leading and trailing newlines",
			text:  "\n\n\nhello\n\n\n",
			width: 4,
			mode:  AtRunes,
			want:  []int{0, 1, 2, 3, 7, 9, 10},
		},
		{
			desc:  "handles multiple newlines in the middle",
			text:  "hello\n\n\nworld",
			width: 5,
			mode:  AtRunes,
			want:  []int{0, 6, 7, 8},
		},
		{
			desc:  "handles multiple newlines in the middle and wraps",
			text:  "hello\n\n\nworld",
			width: 2,
			mode:  AtRunes,
			want:  []int{0, 2, 4, 6, 7, 8, 10, 12},
		},
		{
			desc:  "contains only newlines",
			text:  "\n\n\n",
			width: 4,
			mode:  AtRunes,
			want:  []int{0, 1, 2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := Lines(tc.text, tc.width, tc.mode)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Lines => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}

}
