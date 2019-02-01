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

package sixteen

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/canvas/braille/testbraille"
	"github.com/mum4k/termdash/draw/segdisp/segment"
	"github.com/mum4k/termdash/draw/segdisp/segment/testsegment"
	"github.com/mum4k/termdash/terminal/faketerm"
)

func TestDraw(t *testing.T) {
	tests := []struct {
		desc       string
		opts       []Option
		drawOpts   []Option
		cellCanvas image.Rectangle
		// If not nil, called before Draw is called - can set, clear or toggle segments.
		update  func(*Display) error
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:       "empty when no segments set",
			cellCanvas: image.Rect(0, 0, 6, 5),
		},
		{
			desc:       "smallest valid display 6x5, all segments",
			cellCanvas: image.Rect(0, 0, 6, 5),
			update: func(d *Display) error {
				for _, s := range AllSegments() {
					if err := d.SetSegment(s); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(1, 0, 4, 1), segment.Horizontal) // A1
				testsegment.MustHV(bc, image.Rect(5, 0, 8, 1), segment.Horizontal) // A2

				testsegment.MustHV(bc, image.Rect(0, 1, 1, 8), segment.Vertical) // F
				testsegment.MustHV(bc, image.Rect(4, 1, 5, 8), segment.Vertical) // J
				testsegment.MustHV(bc, image.Rect(8, 1, 9, 8), segment.Vertical) // B

				testsegment.MustHV(bc, image.Rect(1, 8, 4, 9), segment.Horizontal) // G1
				testsegment.MustHV(bc, image.Rect(5, 8, 8, 9), segment.Horizontal) // G2

				testsegment.MustHV(bc, image.Rect(0, 9, 1, 16), segment.Vertical) // E
				testsegment.MustHV(bc, image.Rect(4, 9, 5, 16), segment.Vertical) // M
				testsegment.MustHV(bc, image.Rect(8, 9, 9, 16), segment.Vertical) // C

				testsegment.MustHV(bc, image.Rect(1, 16, 4, 17), segment.Horizontal) // D1
				testsegment.MustHV(bc, image.Rect(5, 16, 8, 17), segment.Horizontal) // D2

				testsegment.MustDiagonal(bc, image.Rect(1, 1, 4, 8), 1, segment.LeftToRight)  // H
				testsegment.MustDiagonal(bc, image.Rect(5, 1, 8, 8), 1, segment.RightToLeft)  // K
				testsegment.MustDiagonal(bc, image.Rect(1, 9, 4, 16), 1, segment.RightToLeft) // N
				testsegment.MustDiagonal(bc, image.Rect(5, 9, 8, 16), 1, segment.LeftToRight) // L
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			d := New(tc.opts...)
			if tc.update != nil {
				if err := tc.update(d); err != nil {
					t.Fatalf("tc.update => unexpected error: %v", err)
				}
			}

			cvs, err := canvas.New(tc.cellCanvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			{
				err := d.Draw(cvs)
				if (err != nil) != tc.wantErr {
					t.Errorf("Draw => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
				if err != nil {
					return
				}
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
			if err := cvs.Apply(got); err != nil {
				t.Fatalf("bc.Apply => unexpected error: %v", err)
			}
			if diff := faketerm.Diff(want, got); diff != "" {
				t.Fatalf("Draw => %v", diff)
			}

		})
	}
}
