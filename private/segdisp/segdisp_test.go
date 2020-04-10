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

package segdisp

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/braille"
	"github.com/mum4k/termdash/private/canvas/braille/testbraille"
)

func TestRequired(t *testing.T) {
	tests := []struct {
		desc     string
		cellArea image.Rectangle
		want     image.Rectangle
		wantErr  bool
	}{
		{
			desc:     "fails when area isn't wide enough",
			cellArea: image.Rect(0, 0, MinCols-1, MinRows),
			wantErr:  true,
		},
		{
			desc:     "fails when area isn't tall enough",
			cellArea: image.Rect(0, 0, MinCols, MinRows-1),
			wantErr:  true,
		},
		{
			desc:     "returns same area when no adjustment needed",
			cellArea: image.Rect(0, 0, MinCols, MinRows),
			want:     image.Rect(0, 0, MinCols, MinRows),
		},
		{
			desc:     "adjusts width to aspect ratio",
			cellArea: image.Rect(0, 0, MinCols+100, MinRows),
			want:     image.Rect(0, 0, MinCols, MinRows),
		},
		{
			desc:     "adjusts height to aspect ratio",
			cellArea: image.Rect(0, 0, MinCols, MinRows+100),
			want:     image.Rect(0, 0, MinCols, MinRows),
		},
		{
			desc:     "adjusts larger area to aspect ratio",
			cellArea: image.Rect(0, 0, MinCols*2, MinRows*4),
			want:     image.Rect(0, 0, 12, 10),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := Required(tc.cellArea)
			if (err != nil) != tc.wantErr {
				t.Errorf("Required => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Required => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestToBraille(t *testing.T) {
	tests := []struct {
		desc     string
		cellArea image.Rectangle
		wantBC   *braille.Canvas
		wantAr   image.Rectangle
		wantErr  bool
	}{
		{
			desc:     "fails when area isn't wide enough",
			cellArea: image.Rect(0, 0, MinCols-1, MinRows),
			wantErr:  true,
		},
		{
			desc:     "canvas creates braille with the desired aspect ratio",
			cellArea: image.Rect(0, 0, MinCols, MinRows),
			wantBC:   testbraille.MustNew(image.Rect(0, 0, MinCols, MinRows)),
			wantAr:   image.Rect(0, 0, MinCols*braille.ColMult, MinRows*braille.RowMult),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			cvs, err := canvas.New(tc.cellArea)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			gotBC, gotAr, err := ToBraille(cvs)
			if (err != nil) != tc.wantErr {
				t.Errorf("ToBraille => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.wantBC, gotBC); diff != "" {
				t.Errorf("ToBraille => unexpected braille canvas, diff (-want, +got):\n%s", diff)
			}
			if diff := pretty.Compare(tc.wantAr, gotAr); diff != "" {
				t.Errorf("ToBraille => unexpected area, diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestSegmentSize(t *testing.T) {
	tests := []struct {
		desc string
		ar   image.Rectangle
		want int
	}{
		{
			desc: "zero area",
			ar:   image.ZR,
			want: 0,
		},
		{
			desc: "smallest segment size",
			ar:   image.Rect(0, 0, 15, 1),
			want: 1,
		},
		{
			desc: "allows even size of two",
			ar:   image.Rect(0, 0, 22, 1),
			want: 2,
		},
		{
			desc: "lands on even width, corrected to odd",
			ar:   image.Rect(0, 0, 44, 1),
			want: 5,
		},
		{
			desc: "lands on odd width",
			ar:   image.Rect(0, 0, 55, 1),
			want: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := SegmentSize(tc.ar)
			if got != tc.want {
				t.Errorf("SegmentSize => %d, want %d", got, tc.want)
			}
		})
	}
}
