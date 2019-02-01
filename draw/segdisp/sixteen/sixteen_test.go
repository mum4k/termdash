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
			desc:       "smallest display 6x5, all segments",
			cellCanvas: image.Rect(0, 0, 6, 5),
			/*
				update: func(d *Display) error {
					for _, s := range AllSegments() {
						if err := d.SetSegment(s); err != nil {
							return err
						}
					}
					return nil
				},*/
			update: func(d *Display) error {
				return d.SetCharacter('w')
			},
		},
		{
			desc:       "8x6, all segments",
			cellCanvas: image.Rect(0, 0, 8, 6),
			/*
				update: func(d *Display) error {
					for _, s := range AllSegments() {
						if err := d.SetSegment(s); err != nil {
							return err
						}
					}
					return nil
				},*/
			update: func(d *Display) error {
				return d.SetCharacter('W')
			},
		},
		{
			desc:       "16x12, all segments",
			cellCanvas: image.Rect(0, 0, 16, 12),
			update: func(d *Display) error {
				for _, s := range AllSegments() {
					if err := d.SetSegment(s); err != nil {
						return err
					}
				}
				return nil
			},
			/*
				update: func(d *Display) error {
					return d.SetCharacter('W')
				},*/
		},
		{
			desc:       "32x24, all segments",
			cellCanvas: image.Rect(0, 0, 32, 24),
			/*
				update: func(d *Display) error {
					for _, s := range AllSegments() {
						if err := d.SetSegment(s); err != nil {
							return err
						}
					}
					return nil
				},*/
			update: func(d *Display) error {
				return d.SetCharacter('W')
			},
		},
		{
			desc:       "64x48, all segments",
			cellCanvas: image.Rect(0, 0, 64, 48),
			/*
				update: func(d *Display) error {
					for _, s := range AllSegments() {
						if err := d.SetSegment(s); err != nil {
							return err
						}
					}
					return nil
				},*/
			update: func(d *Display) error {
				return d.SetCharacter('W')
			},
		},
		{
			desc:       "80x64, all segments",
			cellCanvas: image.Rect(0, 0, 80, 64),
			update: func(d *Display) error {
				for _, s := range AllSegments() {
					if err := d.SetSegment(s); err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			desc:       "80x64, W",
			cellCanvas: image.Rect(0, 0, 80, 64),
			update: func(d *Display) error {
				return d.SetCharacter('W')
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
