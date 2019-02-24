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

package cell

import (
	"fmt"
	"testing"
)

func TestColorNumber(t *testing.T) {
	tests := []struct {
		desc   string
		number int
		want   Color
	}{
		{
			desc:   "default when too small",
			number: -1,
			want:   ColorDefault,
		},
		{
			desc:   "default when too large",
			number: 256,
			want:   ColorDefault,
		},
		{
			desc:   "translates system color",
			number: 0,
			want:   ColorBlack,
		},
		{
			desc:   "adds one to the value",
			number: 42,
			want:   Color(43),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Logf(fmt.Sprintf("color: %v", tc.want))
			got := ColorNumber(tc.number)
			if got != tc.want {
				t.Errorf("ColorNumber(%v) => %v, want %v", tc.number, got, tc.want)
			}
		})
	}
}

func TestColorRGB6(t *testing.T) {
	tests := []struct {
		desc    string
		r, g, b int
		want    Color
	}{
		{
			desc: "default when r too small",
			r:    -1,
			g:    0,
			b:    0,
			want: ColorDefault,
		},
		{
			desc: "default when r too large",
			r:    6,
			g:    0,
			b:    0,
			want: ColorDefault,
		},
		{
			desc: "default when g too small",
			r:    0,
			g:    -1,
			b:    0,
			want: ColorDefault,
		},
		{
			desc: "default when g too large",
			r:    0,
			g:    6,
			b:    0,
			want: ColorDefault,
		},
		{
			desc: "default when b too small",
			r:    0,
			g:    0,
			b:    -1,
			want: ColorDefault,
		},
		{
			desc: "default when b too large",
			r:    0,
			g:    0,
			b:    6,
			want: ColorDefault,
		},
		{
			desc: "translates black",
			r:    0,
			g:    0,
			b:    0,
			want: Color(17),
		},
		{
			desc: "adds one to value",
			r:    2,
			g:    1,
			b:    3,
			want: Color(98),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := ColorRGB6(tc.r, tc.g, tc.b)
			if got != tc.want {
				t.Errorf("ColorRGB6(%v, %v, %v) => %v, want %v", tc.r, tc.g, tc.b, got, tc.want)
			}
		})
	}
}

func TestColorRGB24(t *testing.T) {
	tests := []struct {
		desc    string
		r, g, b int
		want    Color
	}{
		{
			desc: "default when r too small",
			r:    -1,
			g:    0,
			b:    0,
			want: ColorDefault,
		},
		{
			desc: "default when r too large",
			r:    256,
			g:    0,
			b:    0,
			want: ColorDefault,
		},
		{
			desc: "default when g too small",
			r:    0,
			g:    -1,
			b:    0,
			want: ColorDefault,
		},
		{
			desc: "default when g too large",
			r:    0,
			g:    256,
			b:    0,
			want: ColorDefault,
		},
		{
			desc: "default when b too small",
			r:    0,
			g:    0,
			b:    -1,
			want: ColorDefault,
		},
		{
			desc: "default when b too large",
			r:    0,
			g:    0,
			b:    256,
			want: ColorDefault,
		},
		{
			desc: "translates black",
			r:    0,
			g:    0,
			b:    0,
			want: Color(17),
		},
		{
			desc: "adds one to value",
			r:    95,
			g:    255,
			b:    135,
			want: Color(85),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := ColorRGB24(tc.r, tc.g, tc.b)
			if got != tc.want {
				t.Errorf("ColorRGB24(%v, %v, %v) => %v, want %v", tc.r, tc.g, tc.b, got, tc.want)
			}
		})
	}
}
