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

package trig

import (
	"fmt"
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestCirclePointAtAngleAndAngle(t *testing.T) {
	tests := []struct {
		degrees int
		mid     image.Point
		radius  int
		want    image.Point
	}{
		{0, image.Point{0, 0}, 1, image.Point{1, 0}},
		{90, image.Point{0, 0}, 1, image.Point{0, -1}},
		{180, image.Point{0, 0}, 1, image.Point{-1, 0}},
		{270, image.Point{0, 0}, 1, image.Point{0, 1}},

		// Non-zero mid point.
		{0, image.Point{5, 5}, 1, image.Point{6, 5}},
		{90, image.Point{5, 5}, 1, image.Point{5, 4}},
		{180, image.Point{5, 5}, 1, image.Point{4, 5}},
		{270, image.Point{5, 5}, 1, image.Point{5, 6}},
		{0, image.Point{1, 1}, 1, image.Point{2, 1}},
		{90, image.Point{1, 1}, 1, image.Point{1, 0}},
		{180, image.Point{1, 1}, 1, image.Point{0, 1}},
		{270, image.Point{1, 1}, 1, image.Point{1, 2}},

		// Larger radius.
		{0, image.Point{0, 0}, 11, image.Point{11, 0}},
		{90, image.Point{0, 0}, 11, image.Point{0, -11}},
		{180, image.Point{0, 0}, 11, image.Point{-11, 0}},
		{270, image.Point{0, 0}, 11, image.Point{0, 11}},

		// Other angles.
		{27, image.Point{0, 0}, 11, image.Point{10, -5}},
		{68, image.Point{0, 0}, 11, image.Point{4, -10}},
		{333, image.Point{2, 2}, 2, image.Point{4, 3}},
		{153, image.Point{2, 2}, 2, image.Point{0, 1}},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("CirclePointAtAngle %v %v %v", tc.degrees, tc.mid, tc.radius), func(t *testing.T) {
			got := CirclePointAtAngle(tc.degrees, tc.mid, tc.radius)
			if got != tc.want {
				t.Errorf("CirclePointAtAngle(%v, %v, %v) => %v, want %v", tc.degrees, tc.mid, tc.radius, got, tc.want)
			}
		})
		t.Run(fmt.Sprintf("CircleAngleAtPoint %v %v", tc.want, tc.mid), func(t *testing.T) {
			got := CircleAngleAtPoint(tc.want, tc.mid)
			want := tc.degrees
			if got != want {
				t.Errorf("CircleAngleAtPoint(%v, %v) => %v, want %v", tc.want, tc.mid, got, want)
			}
		})
	}
}

func TestPointIsIn(t *testing.T) {
	tests := []struct {
		desc  string
		point image.Point
		shape []image.Point
		want  bool
	}{
		{
			desc:  "no points provided",
			point: image.Point{0, 0},
			shape: nil,
			want:  false,
		},
		{
			desc:  "point is on the shape",
			point: image.Point{0, 0},
			shape: []image.Point{
				{0, 0},
			},
			want: false,
		},
		{
			desc:  "point is left of the shape",
			point: image.Point{0, 1},
			shape: []image.Point{
				{1, 0}, {2, 0}, {3, 0},
				{1, 1}, {3, 1},
				{1, 2}, {2, 2}, {3, 2},
			},
			want: false,
		},
		{
			desc:  "point is in a shape whose border gets crossed once",
			point: image.Point{2, 1},
			shape: []image.Point{
				{1, 0}, {2, 0}, {3, 0},
				{1, 1}, {3, 1},
				{1, 2}, {2, 2}, {3, 2},
			},
			want: true,
		},
		{
			desc:  "point is in an U shape whose border gets crossed multiple times",
			point: image.Point{1, 1},
			shape: []image.Point{
				{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}, {5, 0}, {6, 0},
				{0, 1}, {2, 1}, {4, 1}, {6, 1},
				{0, 2}, {1, 2}, {2, 2}, {3, 2}, {4, 2}, {5, 2},
			},
			want: true,
		},
		{
			desc:  "point is in a triangle",
			point: image.Point{3, 1},
			shape: []image.Point{
				{3, 0},
				{2, 1}, {4, 1},
				{1, 2}, {2, 2}, {3, 2}, {4, 2}, {5, 2},
			},
			want: true,
		},
		{
			desc:  "ignores multiple shape points on the same row",
			point: image.Point{2, 1},
			shape: []image.Point{
				{1, 0}, {2, 0}, {3, 0},
				{1, 1}, {3, 1}, {4, 1}, {5, 1},
				{1, 2}, {2, 2}, {3, 2},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := PointIsIn(tc.point, tc.shape)
			if got != tc.want {
				t.Errorf("PointIsIn => %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRangeSize(t *testing.T) {
	tests := []struct {
		desc    string
		start   int
		end     int
		want    int
		wantErr bool
	}{
		{
			desc:    "invalid start, too small",
			start:   MinAngle - 1,
			end:     MaxAngle,
			wantErr: true,
		},
		{
			desc:    "invalid start, too large",
			start:   MaxAngle + 1,
			end:     MaxAngle,
			wantErr: true,
		},
		{
			desc:    "invalid end, too small",
			start:   MinAngle,
			end:     MinAngle - 1,
			wantErr: true,
		},
		{
			desc:    "invalid end, too large",
			start:   MinAngle,
			end:     MaxAngle + 1,
			wantErr: true,
		},
		{
			desc:  "zero range starting at zero",
			start: 0,
			end:   0,
			want:  0,
		},
		{
			desc:  "zero range starting at max angle",
			start: 360,
			end:   360,
			want:  0,
		},
		{
			desc:  "range with size of one",
			start: 1,
			end:   2,
			want:  1,
		},
		{
			desc:  "reverse range with size of 359",
			start: 2,
			end:   1,
			want:  359,
		},
		{
			desc:  "range that crosses 360",
			start: 350,
			end:   10,
			want:  20,
		},
		{
			desc:  "reverse range that doesn't cross 360",
			start: 10,
			end:   350,
			want:  340,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := RangeSize(tc.start, tc.end)
			if (err != nil) != tc.wantErr {
				t.Errorf("RangeSize => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if got != tc.want {
				t.Errorf("RangeSize => %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRangeMid(t *testing.T) {
	tests := []struct {
		desc    string
		start   int
		end     int
		want    int
		wantErr bool
	}{
		{
			desc:    "invalid start, too small",
			start:   MinAngle - 1,
			end:     MaxAngle,
			wantErr: true,
		},
		{
			desc:    "invalid start, too large",
			start:   MaxAngle + 1,
			end:     MaxAngle,
			wantErr: true,
		},
		{
			desc:    "invalid end, too small",
			start:   MinAngle,
			end:     MinAngle - 1,
			wantErr: true,
		},
		{
			desc:    "invalid end, too large",
			start:   MinAngle,
			end:     MaxAngle + 1,
			wantErr: true,
		},
		{
			desc:  "zero range",
			start: 0,
			end:   0,
			want:  0,
		},
		{
			desc:  "one degree range",
			start: 0,
			end:   1,
			want:  0,
		},
		{
			desc:  "three degree range",
			start: 0,
			end:   3,
			want:  1,
		},
		{
			desc:  "range that crosses 360, mid isn't 360",
			start: 351,
			end:   11,
			want:  1,
		},
		{
			desc:  "range that crosses 360, mid is 360",
			start: 350,
			end:   10,
			want:  0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := RangeMid(tc.start, tc.end)
			if (err != nil) != tc.wantErr {
				t.Errorf("RangeMid => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if got != tc.want {
				t.Errorf("RangeMid => %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFilterByAngle(t *testing.T) {
	tests := []struct {
		desc    string
		points  []image.Point
		mid     image.Point
		start   int
		end     int
		want    []image.Point
		wantErr bool
	}{
		{
			desc:    "invalid mid, negative X coordinate",
			mid:     image.Point{-1, 0},
			start:   MinAngle,
			end:     MaxAngle,
			wantErr: true,
		},
		{
			desc:    "invalid mid, negative Y coordinate",
			mid:     image.Point{0, -1},
			start:   MinAngle,
			end:     MaxAngle,
			wantErr: true,
		},
		{
			desc:    "invalid start, too small",
			start:   MinAngle - 1,
			end:     MaxAngle,
			wantErr: true,
		},
		{
			desc:    "invalid start, too large",
			start:   MaxAngle + 1,
			end:     MaxAngle,
			wantErr: true,
		},
		{
			desc:    "invalid end, too small",
			start:   MinAngle,
			end:     MinAngle - 1,
			wantErr: true,
		},
		{
			desc:    "invalid end, too large",
			start:   MinAngle,
			end:     MaxAngle + 1,
			wantErr: true,
		},
		{
			desc: "full first quadrant",
			points: []image.Point{
				{0, 0}, {1, 0}, {2, 0},
				{0, 1}, {1, 1}, {2, 1},
				{0, 2}, {1, 2}, {2, 2},
			},
			mid:   image.Point{1, 1},
			start: 0,
			end:   90,
			want: []image.Point{
				{1, 0}, {2, 0},
				{1, 1}, {2, 1},
			},
		},
		{
			desc: "partial second quadrant",
			points: []image.Point{
				{0, 0}, {1, 0}, {2, 0},
				{0, 1}, {1, 1}, {2, 1},
				{0, 2}, {1, 2}, {2, 2},
			},
			mid:   image.Point{1, 1},
			start: 130,
			end:   140,
			want: []image.Point{
				{0, 0},
			},
		},
		{
			desc: "range crosses 360",
			points: []image.Point{
				{0, 0}, {1, 0}, {2, 0},
				{0, 1}, {1, 1}, {2, 1},
				{0, 2}, {1, 2}, {2, 2},
			},
			mid:   image.Point{1, 1},
			start: 310,
			end:   50,
			want: []image.Point{
				{2, 0},
				{1, 1}, {2, 1},
				{2, 2},
			},
		},
		{
			desc: "full circle",
			points: []image.Point{
				{0, 0}, {1, 0}, {2, 0},
				{0, 1}, {1, 1}, {2, 1},
				{0, 2}, {1, 2}, {2, 2},
			},
			mid:   image.Point{1, 1},
			start: 0,
			end:   360,
			want: []image.Point{
				{0, 0}, {1, 0}, {2, 0},
				{0, 1}, {1, 1}, {2, 1},
				{0, 2}, {1, 2}, {2, 2},
			},
		},
		{
			desc: "full circle in reverse",
			points: []image.Point{
				{0, 0}, {1, 0}, {2, 0},
				{0, 1}, {1, 1}, {2, 1},
				{0, 2}, {1, 2}, {2, 2},
			},
			mid:   image.Point{1, 1},
			start: 360,
			end:   0,
			want: []image.Point{
				{0, 0}, {1, 0}, {2, 0},
				{0, 1}, {1, 1}, {2, 1},
				{0, 2}, {1, 2}, {2, 2},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := FilterByAngle(tc.points, tc.mid, tc.start, tc.end)
			if (err != nil) != tc.wantErr {
				t.Errorf("FilterByAngle => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("FilterByAngle => unexpected diff (-want, +got):\n%s", diff)
			}

		})
	}
}
