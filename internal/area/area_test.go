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
		desc       string
		area       image.Rectangle
		heightPerc int
		wantTop    image.Rectangle
		wantBot    image.Rectangle
		wantErr    bool
	}{
		{
			desc:       "fails on heightPerc too small",
			area:       image.Rect(1, 1, 2, 2),
			heightPerc: -1,
			wantErr:    true,
		},
		{
			desc:       "fails on heightPerc too large",
			area:       image.Rect(1, 1, 2, 2),
			heightPerc: 101,
			wantErr:    true,
		},
		{
			desc:       "zero area to begin with",
			area:       image.ZR,
			heightPerc: 50,
			wantTop:    image.ZR,
			wantBot:    image.ZR,
		},
		{
			desc:       "splitting results in zero height area on the top",
			area:       image.Rect(1, 1, 2, 2),
			heightPerc: 0,
			wantTop:    image.ZR,
			wantBot:    image.Rect(1, 1, 2, 2),
		},
		{
			desc:       "splitting results in zero height area on the bottom",
			area:       image.Rect(1, 1, 2, 2),
			heightPerc: 100,
			wantTop:    image.Rect(1, 1, 2, 2),
			wantBot:    image.ZR,
		},
		{
			desc:       "splits area with even height",
			area:       image.Rect(1, 1, 3, 3),
			heightPerc: 50,
			wantTop:    image.Rect(1, 1, 3, 2),
			wantBot:    image.Rect(1, 2, 3, 3),
		},
		{
			desc:       "splits area with odd height",
			area:       image.Rect(1, 1, 4, 4),
			heightPerc: 50,
			wantTop:    image.Rect(1, 1, 4, 2),
			wantBot:    image.Rect(1, 2, 4, 4),
		},
		{
			desc:       "splits to unequal areas",
			area:       image.Rect(0, 0, 4, 4),
			heightPerc: 25,
			wantTop:    image.Rect(0, 0, 4, 1),
			wantBot:    image.Rect(0, 1, 4, 4),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotTop, gotBot, err := HSplit(tc.area, tc.heightPerc)
			if (err != nil) != tc.wantErr {
				t.Errorf("VSplit => unexpected error:%v, wantErr:%v", err, tc.wantErr)
			}
			if diff := pretty.Compare(tc.wantTop, gotTop); diff != "" {
				t.Errorf("HSplit => first value unexpected diff (-want, +got):\n%s", diff)
			}
			if diff := pretty.Compare(tc.wantBot, gotBot); diff != "" {
				t.Errorf("HSplit => second value unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestVSplit(t *testing.T) {
	tests := []struct {
		desc      string
		area      image.Rectangle
		widthPerc int
		wantLeft  image.Rectangle
		wantRight image.Rectangle
		wantErr   bool
	}{
		{
			desc:      "fails on widthPerc too small",
			area:      image.Rect(1, 1, 2, 2),
			widthPerc: -1,
			wantErr:   true,
		},
		{
			desc:      "fails on widthPerc too large",
			area:      image.Rect(1, 1, 2, 2),
			widthPerc: 101,
			wantErr:   true,
		},
		{
			desc:      "zero area to begin with",
			area:      image.ZR,
			widthPerc: 50,
			wantLeft:  image.ZR,
			wantRight: image.ZR,
		},
		{
			desc:      "splitting results in zero width area on the left",
			area:      image.Rect(1, 1, 2, 2),
			widthPerc: 0,
			wantLeft:  image.ZR,
			wantRight: image.Rect(1, 1, 2, 2),
		},
		{
			desc:      "splitting results in zero width area on the right",
			area:      image.Rect(1, 1, 2, 2),
			widthPerc: 100,
			wantLeft:  image.Rect(1, 1, 2, 2),
			wantRight: image.ZR,
		},
		{
			desc:      "splits area with even width",
			area:      image.Rect(1, 1, 3, 3),
			widthPerc: 50,
			wantLeft:  image.Rect(1, 1, 2, 3),
			wantRight: image.Rect(2, 1, 3, 3),
		},
		{
			desc:      "splits area with odd width",
			area:      image.Rect(1, 1, 4, 4),
			widthPerc: 50,
			wantLeft:  image.Rect(1, 1, 2, 4),
			wantRight: image.Rect(2, 1, 4, 4),
		},
		{
			desc:      "splits to unequal areas",
			area:      image.Rect(0, 0, 4, 4),
			widthPerc: 25,
			wantLeft:  image.Rect(0, 0, 1, 4),
			wantRight: image.Rect(1, 0, 4, 4),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotLeft, gotRight, err := VSplit(tc.area, tc.widthPerc)
			if (err != nil) != tc.wantErr {
				t.Errorf("VSplit => unexpected error:%v, wantErr:%v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.wantLeft, gotLeft); diff != "" {
				t.Errorf("VSplit => left value unexpected diff (-want, +got):\n%s", diff)
			}
			if diff := pretty.Compare(tc.wantRight, gotRight); diff != "" {
				t.Errorf("VSplit => right value unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestVSplitCells(t *testing.T) {
	tests := []struct {
		desc      string
		area      image.Rectangle
		cells     int
		wantLeft  image.Rectangle
		wantRight image.Rectangle
		wantErr   bool
	}{
		{
			desc:    "fails on negative cells",
			area:    image.Rect(1, 1, 2, 2),
			cells:   -1,
			wantErr: true,
		},
		{
			desc:      "returns area as left on cells too large",
			area:      image.Rect(1, 1, 2, 2),
			cells:     2,
			wantLeft:  image.Rect(1, 1, 2, 2),
			wantRight: image.ZR,
		},
		{
			desc:      "returns area as left on cells equal area width",
			area:      image.Rect(1, 1, 2, 2),
			cells:     1,
			wantLeft:  image.Rect(1, 1, 2, 2),
			wantRight: image.ZR,
		},
		{
			desc:      "returns area as right on zero cells",
			area:      image.Rect(1, 1, 2, 2),
			cells:     0,
			wantRight: image.Rect(1, 1, 2, 2),
			wantLeft:  image.ZR,
		},
		{
			desc:      "zero area to begin with",
			area:      image.ZR,
			cells:     0,
			wantLeft:  image.ZR,
			wantRight: image.ZR,
		},
		{
			desc:      "splits area with even width",
			area:      image.Rect(1, 1, 3, 3),
			cells:     1,
			wantLeft:  image.Rect(1, 1, 2, 3),
			wantRight: image.Rect(2, 1, 3, 3),
		},
		{
			desc:      "splits area with odd width",
			area:      image.Rect(1, 1, 4, 4),
			cells:     1,
			wantLeft:  image.Rect(1, 1, 2, 4),
			wantRight: image.Rect(2, 1, 4, 4),
		},
		{
			desc:      "splits to unequal areas",
			area:      image.Rect(0, 0, 4, 4),
			cells:     3,
			wantLeft:  image.Rect(0, 0, 3, 4),
			wantRight: image.Rect(3, 0, 4, 4),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotLeft, gotRight, err := VSplitCells(tc.area, tc.cells)
			if (err != nil) != tc.wantErr {
				t.Errorf("VSplitCells => unexpected error:%v, wantErr:%v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.wantLeft, gotLeft); diff != "" {
				t.Errorf("VSplitCells => left value unexpected diff (-want, +got):\n%s", diff)
			}
			if diff := pretty.Compare(tc.wantRight, gotRight); diff != "" {
				t.Errorf("VSplitCells => right value unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestHSplitCells(t *testing.T) {
	tests := []struct {
		desc       string
		area       image.Rectangle
		cells      int
		wantTop    image.Rectangle
		wantBottom image.Rectangle
		wantErr    bool
	}{
		{
			desc:    "fails on negative cells",
			area:    image.Rect(1, 1, 2, 2),
			cells:   -1,
			wantErr: true,
		},
		{
			desc:       "returns area as top on cells too large",
			area:       image.Rect(1, 1, 2, 2),
			cells:      2,
			wantTop:    image.Rect(1, 1, 2, 2),
			wantBottom: image.ZR,
		},
		{
			desc:       "returns area as top on cells equal area width",
			area:       image.Rect(1, 1, 2, 2),
			cells:      1,
			wantTop:    image.Rect(1, 1, 2, 2),
			wantBottom: image.ZR,
		},
		{
			desc:       "returns area as bottom on zero cells",
			area:       image.Rect(1, 1, 2, 2),
			cells:      0,
			wantBottom: image.Rect(1, 1, 2, 2),
			wantTop:    image.ZR,
		},
		{
			desc:       "zero area to begin with",
			area:       image.ZR,
			cells:      0,
			wantTop:    image.ZR,
			wantBottom: image.ZR,
		},
		{
			desc:       "splits area with even height",
			area:       image.Rect(1, 1, 3, 3),
			cells:      1,
			wantTop:    image.Rect(1, 1, 3, 2),
			wantBottom: image.Rect(1, 2, 3, 3),
		},
		{
			desc:       "splits area with odd width",
			area:       image.Rect(1, 1, 4, 4),
			cells:      1,
			wantTop:    image.Rect(1, 1, 4, 2),
			wantBottom: image.Rect(1, 2, 4, 4),
		},
		{
			desc:       "splits to unequal areas",
			area:       image.Rect(0, 0, 4, 4),
			cells:      3,
			wantTop:    image.Rect(0, 0, 4, 3),
			wantBottom: image.Rect(0, 3, 4, 4),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotTop, gotBottom, err := HSplitCells(tc.area, tc.cells)
			if (err != nil) != tc.wantErr {
				t.Errorf("HSplitCells => unexpected error:%v, wantErr:%v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if diff := pretty.Compare(tc.wantTop, gotTop); diff != "" {
				t.Errorf("HSplitCells => left value unexpected diff (-want, +got):\n%s", diff)
			}
			if diff := pretty.Compare(tc.wantBottom, gotBottom); diff != "" {
				t.Errorf("HSplitCells => right value unexpected diff (-want, +got):\n%s", diff)
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

func TestShrink(t *testing.T) {
	tests := []struct {
		desc                     string
		area                     image.Rectangle
		top, right, bottom, left int
		want                     image.Rectangle
		wantErr                  bool
	}{
		{
			desc:    "fails for negative top",
			area:    image.Rect(0, 0, 1, 1),
			top:     -1,
			right:   0,
			bottom:  0,
			left:    0,
			wantErr: true,
		},
		{
			desc:    "fails for negative right",
			area:    image.Rect(0, 0, 1, 1),
			top:     0,
			right:   -1,
			bottom:  0,
			left:    0,
			wantErr: true,
		},
		{
			desc:    "fails for negative bottom",
			area:    image.Rect(0, 0, 1, 1),
			top:     0,
			right:   0,
			bottom:  -1,
			left:    0,
			wantErr: true,
		},
		{
			desc:    "fails for negative left",
			area:    image.Rect(0, 0, 1, 1),
			top:     0,
			right:   0,
			bottom:  0,
			left:    -1,
			wantErr: true,
		},
		{
			desc:   "area unchanged when all zero",
			area:   image.Rect(7, 8, 9, 10),
			top:    0,
			right:  0,
			bottom: 0,
			left:   0,
			want:   image.Rect(7, 8, 9, 10),
		},
		{
			desc:   "shrinks top",
			area:   image.Rect(7, 8, 17, 18),
			top:    1,
			right:  0,
			bottom: 0,
			left:   0,
			want:   image.Rect(7, 9, 17, 18),
		},
		{
			desc:   "zero area when top too large",
			area:   image.Rect(7, 8, 17, 18),
			top:    10,
			right:  0,
			bottom: 0,
			left:   0,
			want:   image.ZR,
		},
		{
			desc:   "shrinks bottom",
			area:   image.Rect(7, 8, 17, 18),
			top:    0,
			right:  0,
			bottom: 1,
			left:   0,
			want:   image.Rect(7, 8, 17, 17),
		},
		{
			desc:   "zero area when bottom too large",
			area:   image.Rect(7, 8, 17, 18),
			top:    0,
			right:  0,
			bottom: 10,
			left:   0,
			want:   image.ZR,
		},
		{
			desc:   "zero area when top and bottom cross",
			area:   image.Rect(7, 8, 17, 18),
			top:    5,
			right:  0,
			bottom: 5,
			left:   0,
			want:   image.ZR,
		},
		{
			desc:   "zero area when top and bottom overrun",
			area:   image.Rect(7, 8, 17, 18),
			top:    50,
			right:  0,
			bottom: 50,
			left:   0,
			want:   image.ZR,
		},
		{
			desc:   "shrinks right",
			area:   image.Rect(7, 8, 17, 18),
			top:    0,
			right:  1,
			bottom: 0,
			left:   0,
			want:   image.Rect(7, 8, 16, 18),
		},
		{
			desc:   "zero area when right too large",
			area:   image.Rect(7, 8, 17, 18),
			top:    0,
			right:  10,
			bottom: 0,
			left:   0,
			want:   image.ZR,
		},
		{
			desc:   "shrinks left",
			area:   image.Rect(7, 8, 17, 18),
			top:    0,
			right:  0,
			bottom: 0,
			left:   1,
			want:   image.Rect(8, 8, 17, 18),
		},
		{
			desc:   "zero area when left too large",
			area:   image.Rect(7, 8, 17, 18),
			top:    0,
			right:  0,
			bottom: 0,
			left:   10,
			want:   image.ZR,
		},
		{
			desc:   "zero area when right and left cross",
			area:   image.Rect(7, 8, 17, 18),
			top:    0,
			right:  5,
			bottom: 0,
			left:   5,
			want:   image.ZR,
		},
		{
			desc:   "zero area when right and left overrun",
			area:   image.Rect(7, 8, 17, 18),
			top:    0,
			right:  50,
			bottom: 0,
			left:   50,
			want:   image.ZR,
		},
		{
			desc:   "shrinks from all sides",
			area:   image.Rect(7, 8, 17, 18),
			top:    1,
			right:  2,
			bottom: 3,
			left:   4,
			want:   image.Rect(11, 9, 15, 15),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := Shrink(tc.area, tc.top, tc.right, tc.bottom, tc.left)
			if (err != nil) != tc.wantErr {
				t.Errorf("Shrink => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Shrink => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestShrinkPercent(t *testing.T) {
	tests := []struct {
		desc                     string
		area                     image.Rectangle
		top, right, bottom, left int
		want                     image.Rectangle
		wantErr                  bool
	}{
		{
			desc:    "fails on top too low",
			top:     -1,
			wantErr: true,
		},
		{
			desc:    "fails on top too high",
			top:     101,
			wantErr: true,
		},
		{
			desc:    "fails on right too low",
			right:   -1,
			wantErr: true,
		},
		{
			desc:    "fails on right too high",
			right:   101,
			wantErr: true,
		},
		{
			desc:    "fails on bottom too low",
			bottom:  -1,
			wantErr: true,
		},
		{
			desc:    "fails on bottom too high",
			bottom:  101,
			wantErr: true,
		},
		{
			desc:    "fails on left too low",
			left:    -1,
			wantErr: true,
		},
		{
			desc:    "fails on left too high",
			left:    101,
			wantErr: true,
		},
		{
			desc: "shrinks to zero area for top too large",
			area: image.Rect(0, 0, 100, 100),
			top:  100,
			want: image.ZR,
		},
		{
			desc:   "shrinks to zero area for bottom too large",
			area:   image.Rect(0, 0, 100, 100),
			bottom: 100,
			want:   image.ZR,
		},
		{
			desc:   "shrinks to zero area top and bottom that meet",
			area:   image.Rect(0, 0, 100, 100),
			top:    50,
			bottom: 50,
			want:   image.ZR,
		},
		{
			desc:  "shrinks to zero area for right too large",
			area:  image.Rect(0, 0, 100, 100),
			right: 100,
			want:  image.ZR,
		},
		{
			desc: "shrinks to zero area for left too large",
			area: image.Rect(0, 0, 100, 100),
			left: 100,
			want: image.ZR,
		},
		{
			desc:  "shrinks to zero area right and left that meet",
			area:  image.Rect(0, 0, 100, 100),
			right: 50,
			left:  50,
			want:  image.ZR,
		},
		{
			desc:   "shrinks from all sides",
			area:   image.Rect(0, 0, 100, 100),
			top:    10,
			right:  20,
			bottom: 30,
			left:   40,
			want:   image.Rect(40, 10, 80, 70),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := ShrinkPercent(tc.area, tc.top, tc.right, tc.bottom, tc.left)
			if (err != nil) != tc.wantErr {
				t.Errorf("ShrinkPercent => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("ShrinkPercent => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestMoveUp(t *testing.T) {
	tests := []struct {
		desc    string
		area    image.Rectangle
		cells   int
		want    image.Rectangle
		wantErr bool
	}{
		{
			desc:    "fails on negative cells",
			area:    image.Rect(0, 0, 1, 1),
			cells:   -1,
			wantErr: true,
		},
		{
			desc:    "zero area cannot be moved",
			area:    image.ZR,
			cells:   1,
			wantErr: true,
		},
		{
			desc:    "cannot move area beyond zero Y coordinate",
			area:    image.Rect(0, 5, 1, 10),
			cells:   6,
			wantErr: true,
		},
		{
			desc:  "move by zero cells is idempotent",
			area:  image.Rect(0, 5, 1, 10),
			cells: 0,
			want:  image.Rect(0, 5, 1, 10),
		},
		{
			desc:  "moves area up",
			area:  image.Rect(0, 5, 1, 10),
			cells: 3,
			want:  image.Rect(0, 2, 1, 7),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := MoveUp(tc.area, tc.cells)
			if (err != nil) != tc.wantErr {
				t.Errorf("MoveUp => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("MoveUp => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestMoveDown(t *testing.T) {
	tests := []struct {
		desc    string
		area    image.Rectangle
		cells   int
		want    image.Rectangle
		wantErr bool
	}{
		{
			desc:    "fails on negative cells",
			area:    image.Rect(0, 0, 1, 1),
			cells:   -1,
			wantErr: true,
		},
		{
			desc:  "moves zero area",
			area:  image.ZR,
			cells: 1,
			want:  image.Rect(0, 1, 0, 1),
		},
		{
			desc:  "move by zero cells is idempotent",
			area:  image.Rect(0, 5, 1, 10),
			cells: 0,
			want:  image.Rect(0, 5, 1, 10),
		},
		{
			desc:  "moves area down",
			area:  image.Rect(0, 5, 1, 10),
			cells: 3,
			want:  image.Rect(0, 8, 1, 13),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := MoveDown(tc.area, tc.cells)
			if (err != nil) != tc.wantErr {
				t.Errorf("MoveDown => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("MoveDown => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
