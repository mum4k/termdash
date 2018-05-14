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

package text

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestWrapNeeded(t *testing.T) {
	tests := []struct {
		desc  string
		r     rune
		point image.Point
		width int
		opts  *options
		want  bool
	}{
		{
			desc:  "point within canvas",
			r:     'a',
			point: image.Point{2, 0},
			width: 3,
			opts:  &options{},
			want:  false,
		},
		{
			desc:  "point outside of canvas, wrapping not configured",
			r:     'a',
			point: image.Point{3, 0},
			width: 3,
			opts:  &options{},
			want:  false,
		},
		{
			desc:  "point outside of canvas, wrapping configured",
			r:     'a',
			point: image.Point{3, 0},
			width: 3,
			opts: &options{
				wrapAtRunes: true,
			},
			want: true,
		},
		{
			desc:  "doesn't wrap for newline characters",
			r:     '\n',
			point: image.Point{3, 0},
			width: 3,
			opts: &options{
				wrapAtRunes: true,
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := wrapNeeded(tc.r, tc.point.X, tc.width, tc.opts)
			if got != tc.want {
				t.Errorf("wrapNeeded => got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFindLines(t *testing.T) {
	tests := []struct {
		desc  string
		text  string
		width int
		opts  *options
		want  []int
	}{
		{
			desc:  "zero text",
			text:  "",
			width: 1,
			opts:  &options{},
			want:  nil,
		},
		{
			desc:  "zero width",
			text:  "hello",
			width: 0,
			opts:  &options{},
			want:  nil,
		},
		{
			desc:  "wrapping disabled, no newlines, fits in width",
			text:  "hello",
			width: 5,
			opts:  &options{},
			want:  []int{0},
		},
		{
			desc:  "wrapping disabled, no newlines, doesn't fits in width",
			text:  "hello",
			width: 4,
			opts:  &options{},
			want:  []int{0},
		},
		{
			desc:  "wrapping disabled, newlines, fits in width",
			text:  "hello\nworld",
			width: 5,
			opts:  &options{},
			want:  []int{0, 6},
		},
		{
			desc:  "wrapping disabled, newlines, doesn't fit in width",
			text:  "hello\nworld",
			width: 4,
			opts:  &options{},
			want:  []int{0, 6},
		},
		{
			desc:  "wrapping enabled, no newlines, fits in width",
			text:  "hello",
			width: 5,
			opts: &options{
				wrapAtRunes: true,
			},
			want: []int{0},
		},
		{
			desc:  "wrapping enabled, no newlines, doesn't fit in width",
			text:  "hello",
			width: 4,
			opts: &options{
				wrapAtRunes: true,
			},
			want: []int{0, 4},
		},
		{
			desc:  "wrapping enabled, newlines, fits in width",
			text:  "hello\nworld",
			width: 5,
			opts: &options{
				wrapAtRunes: true,
			},
			want: []int{0, 6},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in width",
			text:  "hello\nworld",
			width: 4,
			opts: &options{
				wrapAtRunes: true,
			},
			want: []int{0, 4, 6, 10},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in width, unicode characters",
			text:  "⇧\n…\n⇩",
			width: 1,
			opts: &options{
				wrapAtRunes: true,
			},
			want: []int{0, 4, 8},
		},
		{
			desc:  "wrapping enabled, newlines, doesn't fit in width, wide unicode characters",
			text:  "你好\n世界",
			width: 1,
			opts: &options{
				wrapAtRunes: true,
			},
			want: []int{0, 3, 7, 10},
		},

		{
			desc:  "handles leading and trailing newlines",
			text:  "\n\n\nhello\n\n\n",
			width: 4,
			opts: &options{
				wrapAtRunes: true,
			},
			want: []int{0, 1, 2, 3, 7, 9, 10},
		},
		{
			desc:  "handles multiple newlines in the middle",
			text:  "hello\n\n\nworld",
			width: 5,
			opts: &options{
				wrapAtRunes: true,
			},
			want: []int{0, 6, 7, 8},
		},
		{
			desc:  "handles multiple newlines in the middle and wraps",
			text:  "hello\n\n\nworld",
			width: 2,
			opts: &options{
				wrapAtRunes: true,
			},
			want: []int{0, 2, 4, 6, 7, 8, 10, 12},
		},
		{
			desc:  "contains only newlines",
			text:  "\n\n\n",
			width: 4,
			opts: &options{
				wrapAtRunes: true,
			},
			want: []int{0, 1, 2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := findLines(tc.text, tc.width, tc.opts)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("findLines => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}

}
