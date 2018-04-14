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

package area

import (
	"fmt"
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestSize(t *testing.T) {
	tests := []struct {
		desc string
		area image.Rectangle
		want image.Point
	}{
		{
			desc: "zero area",
			area: image.Rect(0, 0, 0, 0),
			want: image.Point{0, 0},
		},
		{
			desc: "1-D on X axis",
			area: image.Rect(0, 0, 1, 0),
			want: image.Point{1, 0},
		},
		{
			desc: "1-D on Y axis",
			area: image.Rect(0, 0, 0, 1),
			want: image.Point{0, 1},
		},
		{
			desc: "area with a single cell",
			area: image.Rect(0, 0, 1, 1),
			want: image.Point{1, 1},
		},
		{
			desc: "a rectangle",
			area: image.Rect(0, 0, 2, 3),
			want: image.Point{2, 3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := Size(tc.area)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Size => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestFromSize(t *testing.T) {
	tests := []struct {
		desc    string
		size    image.Point
		want    image.Rectangle
		wantErr bool
	}{
		{
			desc:    "negative size on X axis",
			size:    image.Point{-1, 0},
			wantErr: true,
		},
		{
			desc:    "negative size on Y axis",
			size:    image.Point{0, -1},
			wantErr: true,
		},
		{
			desc: "zero size",
		},
		{
			desc: "1-D on X axis",
			size: image.Point{1, 0},
			want: image.Rect(0, 0, 1, 0),
		},
		{
			desc: "1-D on Y axis",
			size: image.Point{0, 1},
			want: image.Rect(0, 0, 0, 1),
		},
		{
			desc: "a rectangle",
			size: image.Point{2, 3},
			want: image.Rect(0, 0, 2, 3),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := FromSize(tc.size)
			if (err != nil) != tc.wantErr {
				t.Fatalf("FromSize => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("FromSize => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestHSplit(t *testing.T) {
	tests := []struct {
		desc  string
		area  image.Rectangle
		want1 image.Rectangle
		want2 image.Rectangle
	}{
		{
			desc:  "zero area to begin with",
			area:  image.ZR,
			want1: image.ZR,
			want2: image.ZR,
		},
		{
			desc:  "splitting results in zero height area",
			area:  image.Rect(1, 1, 2, 2),
			want1: image.ZR,
			want2: image.ZR,
		},
		{
			desc:  "splits area with even height",
			area:  image.Rect(1, 1, 3, 3),
			want1: image.Rect(1, 1, 3, 2),
			want2: image.Rect(1, 2, 3, 3),
		},
		{
			desc:  "splits area with odd height",
			area:  image.Rect(1, 1, 4, 4),
			want1: image.Rect(1, 1, 4, 2),
			want2: image.Rect(1, 2, 4, 4),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got1, got2 := HSplit(tc.area)
			if diff := pretty.Compare(tc.want1, got1); diff != "" {
				t.Errorf("HSplit => first value unexpected diff (-want, +got):\n%s", diff)
			}
			if diff := pretty.Compare(tc.want2, got2); diff != "" {
				t.Errorf("HSplit => second value unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestVSplit(t *testing.T) {
	tests := []struct {
		desc  string
		area  image.Rectangle
		want1 image.Rectangle
		want2 image.Rectangle
	}{
		{
			desc:  "zero area to begin with",
			area:  image.ZR,
			want1: image.ZR,
			want2: image.ZR,
		},
		{
			desc:  "splitting results in zero width area",
			area:  image.Rect(1, 1, 2, 2),
			want1: image.ZR,
			want2: image.ZR,
		},
		{
			desc:  "splits area with even width",
			area:  image.Rect(1, 1, 3, 3),
			want1: image.Rect(1, 1, 2, 3),
			want2: image.Rect(2, 1, 3, 3),
		},
		{
			desc:  "splits area with odd width",
			area:  image.Rect(1, 1, 4, 4),
			want1: image.Rect(1, 1, 2, 4),
			want2: image.Rect(2, 1, 4, 4),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got1, got2 := VSplit(tc.area)
			if diff := pretty.Compare(tc.want1, got1); diff != "" {
				t.Errorf("VSplit => first value unexpected diff (-want, +got):\n%s", diff)
			}
			if diff := pretty.Compare(tc.want2, got2); diff != "" {
				t.Errorf("VSplit => second value unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestExcludeBorder(t *testing.T) {
	tests := []struct {
		desc string
		area image.Rectangle
		want image.Rectangle
	}{
		{
			desc: "zero area to begin with",
			area: image.ZR,
			want: image.ZR,
		},
		{
			desc: "not enough space to exclude the border",
			area: image.Rect(1, 1, 2, 2),
			want: image.ZR,
		},
		{
			desc: "excluding the border results in zero size area",
			area: image.Rect(2, 2, 4, 4),
			want: image.Rect(3, 3, 3, 3),
		},
		{
			desc: "excludes the border",
			area: image.Rect(1, 1, 3, 4),
			want: image.Rect(2, 2, 2, 3),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := ExcludeBorder(tc.area)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("ExcludeBorder => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestFindGCF(t *testing.T) {
	tests := []struct {
		a    int
		b    int
		want int
	}{
		{0, 0, 0},
		{0, 1, 0},
		{1, 0, 0},
		{1, 1, 1},
		{2, 2, 2},
		{50, 35, 5},
		{16, 88, 8},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("findGCF(%d,%d)", tc.a, tc.b), func(t *testing.T) {
			if got := findGCF(tc.a, tc.b); got != tc.want {
				t.Errorf("findGCF(%d,%d) => got %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestWithRatio(t *testing.T) {
	tests := []struct {
		desc  string
		area  image.Rectangle
		ratio image.Point
		want  image.Rectangle
	}{
		{
			desc:  "area is zero",
			area:  image.ZR,
			ratio: image.Point{1, 1},
			want:  image.ZR,
		},
		{
			desc:  "ratio is zero",
			area:  image.Rect(0, 0, 1, 1),
			ratio: image.ZP,
			want:  image.ZR,
		},
		{
			desc:  "there is no smaller area in the ratio",
			area:  image.Rect(0, 0, 1, 1),
			ratio: image.Point{1, 2},
			want:  image.ZR,
		},
		{
			desc:  "already has the target ratio",
			area:  image.Rect(0, 0, 10, 5),
			ratio: image.Point{2, 1},
			want:  image.Rect(0, 0, 10, 5),
		},
		{
			desc:  "scales to closest height",
			area:  image.Rect(0, 0, 50, 25), // Ratio 2:1.
			ratio: image.Point{3, 1},
			want:  image.Rect(0, 0, 48, 16),
		},
		{
			desc:  "scales to closest width of a non-zero base area",
			area:  image.Rect(1, 2, 21, 62), // Ratio 2:3.
			ratio: image.Point{3, 1},
			want:  image.Rect(1, 2, 19, 8),
		},
		{
			desc:  "scales to closest height",
			area:  image.Rect(0, 0, 50, 25), // Ratio 2:1.
			ratio: image.Point{1, 3},
			want:  image.Rect(0, 0, 8, 24),
		},
		{
			desc:  "scales to closest height of a non-zero base area",
			area:  image.Rect(2, 4, 7, 29), // Ratio 1:5.
			ratio: image.Point{3, 2},
			want:  image.Rect(2, 4, 5, 6),
		},
		{
			desc:  "non-simplified ratio",
			area:  image.Rect(0, 0, 50, 25), // Ratio 2:1.
			ratio: image.Point{300, 100},    // Ratio 3:1.
			want:  image.Rect(0, 0, 48, 16),
		},
		{
			desc:  "square ratio",
			area:  image.Rect(0, 0, 50, 25), // Ratio 2:1.
			ratio: image.Point{1, 1},        // Ratio 3:1.
			want:  image.Rect(0, 0, 25, 25),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := WithRatio(tc.area, tc.ratio)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("WithRatio => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
