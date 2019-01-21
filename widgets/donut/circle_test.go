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

package donut

import (
	"image"
	"testing"
)

func TestStartEndAngles(t *testing.T) {
	tests := []struct {
		desc       string
		current    int
		total      int
		startAngle int
		direction  int
		wantStart  int
		wantEnd    int
	}{
		{
			desc:       "zero angle without current or total",
			current:    0,
			total:      0,
			startAngle: 90,
			direction:  -1,
			wantStart:  90,
			wantEnd:    90,
		},
		{
			desc:       "zero angle without current",
			current:    0,
			total:      100,
			startAngle: 90,
			direction:  -1,
			wantStart:  90,
			wantEnd:    90,
		},
		{
			desc:       "25% current, start at 90, clockwise",
			current:    25,
			total:      100,
			startAngle: 90,
			direction:  -1,
			wantStart:  0,
			wantEnd:    90,
		},
		{
			desc:       "25% current, start at 90, counter-clockwise",
			current:    25,
			total:      100,
			startAngle: 90,
			direction:  1,
			wantStart:  90,
			wantEnd:    180,
		},
		{
			desc:       "50% current, start at 90, clockwise",
			current:    50,
			total:      100,
			startAngle: 90,
			direction:  -1,
			wantStart:  270,
			wantEnd:    90,
		},
		{
			desc:       "50% current, start at 90, counter-clockwise",
			current:    50,
			total:      100,
			startAngle: 90,
			direction:  1,
			wantStart:  90,
			wantEnd:    270,
		},
		{
			desc:       "75% current, start at 90, clockwise",
			current:    75,
			total:      100,
			startAngle: 90,
			direction:  -1,
			wantStart:  180,
			wantEnd:    90,
		},
		{
			desc:       "75% current, start at 90, counter-clockwise",
			current:    75,
			total:      100,
			startAngle: 90,
			direction:  1,
			wantStart:  90,
			wantEnd:    360,
		},
		{
			desc:       "100% current, start at 90, clockwise",
			current:    100,
			total:      100,
			startAngle: 90,
			direction:  -1,
			wantStart:  0,
			wantEnd:    360,
		},
		{
			desc:       "100% current, start at 90, counter-clockwise",
			current:    100,
			total:      100,
			startAngle: 90,
			direction:  1,
			wantStart:  0,
			wantEnd:    360,
		},
		{
			desc:       "25% current, start at 0, clockwise",
			current:    25,
			total:      100,
			startAngle: 0,
			direction:  -1,
			wantStart:  270,
			wantEnd:    360,
		},
		{
			desc:       "25% current, start at 0, counter-clockwise",
			current:    25,
			total:      100,
			startAngle: 0,
			direction:  1,
			wantStart:  0,
			wantEnd:    90,
		},
		{
			desc:       "50% current, start at 0, clockwise",
			current:    50,
			total:      100,
			startAngle: 0,
			direction:  -1,
			wantStart:  180,
			wantEnd:    360,
		},
		{
			desc:       "50% current, start at 0, counter-clockwise",
			current:    50,
			total:      100,
			startAngle: 0,
			direction:  1,
			wantStart:  0,
			wantEnd:    180,
		},
		{
			desc:       "75% current, start at 0, clockwise",
			current:    75,
			total:      100,
			startAngle: 0,
			direction:  -1,
			wantStart:  90,
			wantEnd:    360,
		},
		{
			desc:       "75% current, start at 0, counter-clockwise",
			current:    75,
			total:      100,
			startAngle: 0,
			direction:  1,
			wantStart:  0,
			wantEnd:    270,
		},
		{
			desc:       "100% current, start at 0, clockwise",
			current:    100,
			total:      100,
			startAngle: 0,
			direction:  -1,
			wantStart:  0,
			wantEnd:    360,
		},
		{
			desc:       "100% current, start at 0, counter-clockwise",
			current:    100,
			total:      100,
			startAngle: 0,
			direction:  1,
			wantStart:  0,
			wantEnd:    360,
		},
		{
			desc:       "25% current, start at 270, clockwise",
			current:    25,
			total:      100,
			startAngle: 270,
			direction:  -1,
			wantStart:  180,
			wantEnd:    270,
		},
		{
			desc:       "25% current, start at 270, counter-clockwise",
			current:    25,
			total:      100,
			startAngle: 270,
			direction:  1,
			wantStart:  270,
			wantEnd:    360,
		},
		{
			desc:       "50% current, start at 270, clockwise",
			current:    50,
			total:      100,
			startAngle: 270,
			direction:  -1,
			wantStart:  90,
			wantEnd:    270,
		},
		{
			desc:       "50% current, start at 270, counter-clockwise",
			current:    50,
			total:      100,
			startAngle: 270,
			direction:  1,
			wantStart:  270,
			wantEnd:    90,
		},
		{
			desc:       "75% current, start at 270, clockwise",
			current:    75,
			total:      100,
			startAngle: 270,
			direction:  -1,
			wantStart:  0,
			wantEnd:    270,
		},
		{
			desc:       "75% current, start at 270, counter-clockwise",
			current:    75,
			total:      100,
			startAngle: 270,
			direction:  1,
			wantStart:  270,
			wantEnd:    180,
		},
		{
			desc:       "100% current, start at 270, clockwise",
			current:    100,
			total:      100,
			startAngle: 270,
			direction:  -1,
			wantStart:  0,
			wantEnd:    360,
		},
		{
			desc:       "100% current, start at 270, counter-clockwise",
			current:    100,
			total:      100,
			startAngle: 270,
			direction:  1,
			wantStart:  0,
			wantEnd:    360,
		},
		{
			desc:       "25% current, start at 180, clockwise",
			current:    25,
			total:      100,
			startAngle: 180,
			direction:  -1,
			wantStart:  90,
			wantEnd:    180,
		},
		{
			desc:       "25% current, start at 180, counter-clockwise",
			current:    25,
			total:      100,
			startAngle: 180,
			direction:  1,
			wantStart:  180,
			wantEnd:    270,
		},
		{
			desc:       "50% current, start at 180, clockwise",
			current:    50,
			total:      100,
			startAngle: 180,
			direction:  -1,
			wantStart:  0,
			wantEnd:    180,
		},
		{
			desc:       "50% current, start at 180, counter-clockwise",
			current:    50,
			total:      100,
			startAngle: 180,
			direction:  1,
			wantStart:  180,
			wantEnd:    360,
		},
		{
			desc:       "75% current, start at 180, clockwise",
			current:    75,
			total:      100,
			startAngle: 180,
			direction:  -1,
			wantStart:  270,
			wantEnd:    180,
		},
		{
			desc:       "75% current, start at 180, counter-clockwise",
			current:    75,
			total:      100,
			startAngle: 180,
			direction:  1,
			wantStart:  180,
			wantEnd:    90,
		},
		{
			desc:       "100% current, start at 180, clockwise",
			current:    100,
			total:      100,
			startAngle: 180,
			direction:  -1,
			wantStart:  0,
			wantEnd:    360,
		},
		{
			desc:       "100% current, start at 180, counter-clockwise",
			current:    100,
			total:      100,
			startAngle: 180,
			direction:  1,
			wantStart:  0,
			wantEnd:    360,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotStart, gotEnd := startEndAngles(tc.current, tc.total, tc.startAngle, tc.direction)
			if gotStart != tc.wantStart || gotEnd != tc.wantEnd {
				t.Errorf("startEndAngles => %v, %v, want %v, %v", gotStart, gotEnd, tc.wantStart, tc.wantEnd)
			}
		})
	}
}

func TestMidAndRadius(t *testing.T) {
	tests := []struct {
		desc      string
		pixelArea image.Rectangle
		wantMid   image.Point
		wantR     int
	}{
		{
			desc:      "middle on X falls on beginning of cell",
			pixelArea: image.Rect(0, 0, 4, 3),
			wantMid:   image.Point{2, 1},
			wantR:     1,
		},
		{
			desc:      "middle on X falls on end of cell and is adjusted",
			pixelArea: image.Rect(0, 0, 3, 3),
			wantMid:   image.Point{0, 1},
			wantR:     1,
		},
		{
			desc:      "middle on Y falls on 1st cell pixel, adjusted",
			pixelArea: image.Rect(0, 0, 4, 16),
			wantMid:   image.Point{2, 9},
			wantR:     1,
		},
		{
			desc:      "middle on Y falls on 2nd cell pixel, left as is",
			pixelArea: image.Rect(0, 0, 4, 10),
			wantMid:   image.Point{2, 5},
			wantR:     1,
		},
		{
			desc:      "middle on Y falls on 3rd cell pixel, adjusted",
			pixelArea: image.Rect(0, 0, 4, 12),
			wantMid:   image.Point{2, 5},
			wantR:     1,
		},
		{
			desc:      "middle on Y falls on 4th cell pixel, adjusted",
			pixelArea: image.Rect(0, 0, 4, 30),
			wantMid:   image.Point{2, 13},
			wantR:     1,
		},
		{
			desc:      "Dx less than Dy, mid falls before half",
			pixelArea: image.Rect(0, 0, 14, 40),
			wantMid:   image.Point{6, 21},
			wantR:     6,
		},
		{
			desc:      "Dx less than Dy, mid falls on half",
			pixelArea: image.Rect(0, 0, 20, 40),
			wantMid:   image.Point{10, 21},
			wantR:     9,
		},
		{
			desc:      "Dy less than Dx, mid falls before half",
			pixelArea: image.Rect(0, 0, 20, 20),
			wantMid:   image.Point{10, 9},
			wantR:     9,
		},
		{
			desc:      "Dy less than Dx, mid falls on half",
			pixelArea: image.Rect(0, 0, 20, 18),
			wantMid:   image.Point{10, 9},
			wantR:     8,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotMid, gotR := midAndRadius(tc.pixelArea)
			if gotMid != tc.wantMid || gotR != tc.wantR {
				t.Errorf("midAndRadius => %v, %v, want %v, %v", gotMid, gotR, tc.wantMid, tc.wantR)
			}
		})
	}
}

func TestAvailableCells(t *testing.T) {
	tests := []struct {
		desc      string
		mid       image.Point
		radius    int
		wantCells int
		wantFirst image.Point
	}{
		{
			desc:      "radius too small",
			mid:       image.Point{1, 0},
			radius:    2,
			wantCells: 0,
			wantFirst: image.Point{0, 0},
		},
		{
			desc:      "radius of three",
			mid:       image.Point{2, 1},
			radius:    3,
			wantCells: 2,
			wantFirst: image.Point{0, 0},
		},
		{
			desc:      "radius of four",
			mid:       image.Point{20, 10},
			radius:    4,
			wantCells: 3,
			wantFirst: image.Point{8, 2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotCells, gotFirst := availableCells(tc.mid, tc.radius)
			if gotCells != tc.wantCells || !gotFirst.Eq(tc.wantFirst) {
				t.Errorf("availableCells => %v, %v, want %v, %v", gotCells, gotFirst, tc.wantCells, tc.wantFirst)
			}

		})
	}
}
