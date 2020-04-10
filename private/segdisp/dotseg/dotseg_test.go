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

package dotseg

import (
	"image"
	"sort"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/braille/testbraille"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/private/segdisp"
	"github.com/mum4k/termdash/private/segdisp/segment"
	"github.com/mum4k/termdash/private/segdisp/segment/testsegment"
)

func TestSegmentString(t *testing.T) {
	tests := []struct {
		desc string
		seg  Segment
		want string
	}{
		{
			desc: "known segment",
			seg:  D1,
			want: "D1",
		},
		{
			desc: "unknown segment",
			seg:  Segment(-1),
			want: "SegmentUnknown",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.seg.String()
			if got != tc.want {
				t.Errorf("String => %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDraw(t *testing.T) {
	tests := []struct {
		desc       string
		opts       []Option
		drawOpts   []Option
		cellCanvas image.Rectangle
		// If not nil, it is called before Draw is called and can set, clear or
		// toggle segments or characters.
		update        func(*Display) error
		want          func(size image.Point) *faketerm.Terminal
		wantErr       bool
		wantUpdateErr bool
	}{
		{
			desc:       "fails for area not wide enough",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols-1, segdisp.MinRows),
			wantErr:    true,
		},
		{
			desc:       "fails for area not tall enough",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows-1),
			wantErr:    true,
		},
		{
			desc:       "fails to set invalid segment (too small)",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				return d.SetSegment(Segment(-1))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "fails to set invalid segment (too large)",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				return d.SetSegment(Segment(segmentMax))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "fails to clear invalid segment (too small)",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				return d.ClearSegment(Segment(-1))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "fails to clear invalid segment (too large)",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				return d.ClearSegment(Segment(segmentMax))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "fails to toggle invalid segment (too small)",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				return d.ToggleSegment(Segment(-1))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "fails to toggle invalid segment (too large)",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				return d.ToggleSegment(Segment(segmentMax))
			},
			wantUpdateErr: true,
		},
		{
			desc:       "empty when no segments set",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
		},
		{
			desc:       "smallest valid display 6x5, all segments",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				for _, seg := range AllSegments() {
					if err := d.SetSegment(seg); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(5, 6, 7, 8), segment.Horizontal)   // D1
				testsegment.MustHV(bc, image.Rect(5, 12, 7, 14), segment.Horizontal) // D2
				testsegment.MustHV(bc, image.Rect(5, 15, 7, 17), segment.Horizontal) // D3
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "smallest valid display 6x5, all segments, New sets cell options",
			opts: []Option{
				CellOpts(
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorGreen),
				),
			},
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				for _, seg := range AllSegments() {
					if err := d.SetSegment(seg); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				opts := []segment.Option{
					segment.CellOpts(
						cell.FgColor(cell.ColorRed),
						cell.BgColor(cell.ColorGreen),
					),
				}
				testsegment.MustHV(bc, image.Rect(5, 6, 7, 8), segment.Horizontal, opts...)   // D1
				testsegment.MustHV(bc, image.Rect(5, 12, 7, 14), segment.Horizontal, opts...) // D2
				testsegment.MustHV(bc, image.Rect(5, 15, 7, 17), segment.Horizontal, opts...) // D5
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc: "smallest valid display 6x5, all segments, Draw sets cell options",
			drawOpts: []Option{
				CellOpts(
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorGreen),
				),
			},
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				for _, seg := range AllSegments() {
					if err := d.SetSegment(seg); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				opts := []segment.Option{
					segment.CellOpts(
						cell.FgColor(cell.ColorRed),
						cell.BgColor(cell.ColorGreen),
					),
				}
				testsegment.MustHV(bc, image.Rect(5, 6, 7, 8), segment.Horizontal, opts...)   // D1
				testsegment.MustHV(bc, image.Rect(5, 12, 7, 14), segment.Horizontal, opts...) // D2
				testsegment.MustHV(bc, image.Rect(5, 15, 7, 17), segment.Horizontal, opts...) // D5
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, D1",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				return d.SetSegment(D1)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(5, 6, 7, 8), segment.Horizontal) // D1
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, D2",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				return d.SetSegment(D2)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(5, 12, 7, 14), segment.Horizontal) // D2
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "smallest valid display 6x5, D3",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				return d.SetSegment(D3)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(5, 15, 7, 17), segment.Horizontal) // D3
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "clears segment",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				for _, seg := range AllSegments() {
					if err := d.SetSegment(seg); err != nil {
						return err
					}
				}
				return d.ClearSegment(D1)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(5, 12, 7, 14), segment.Horizontal) // D2
				testsegment.MustHV(bc, image.Rect(5, 15, 7, 17), segment.Horizontal) // D3
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "clears the display",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				for _, seg := range AllSegments() {
					if err := d.SetSegment(seg); err != nil {
						return err
					}
				}
				d.Clear()
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				return ft
			},
		},
		{
			desc:       "clear sets new cell options",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				d.Clear(CellOpts(
					cell.FgColor(cell.ColorRed),
					cell.BgColor(cell.ColorGreen),
				))
				for _, seg := range AllSegments() {
					if err := d.SetSegment(seg); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				opts := []segment.Option{
					segment.CellOpts(
						cell.FgColor(cell.ColorRed),
						cell.BgColor(cell.ColorGreen),
					),
				}
				testsegment.MustHV(bc, image.Rect(5, 6, 7, 8), segment.Horizontal, opts...)   // D1
				testsegment.MustHV(bc, image.Rect(5, 12, 7, 14), segment.Horizontal, opts...) // D2
				testsegment.MustHV(bc, image.Rect(5, 15, 7, 17), segment.Horizontal, opts...) // D5
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "toggles segment off",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				for _, seg := range AllSegments() {
					if err := d.SetSegment(seg); err != nil {
						return err
					}
				}
				return d.ToggleSegment(D1)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(5, 12, 7, 14), segment.Horizontal) // D2
				testsegment.MustHV(bc, image.Rect(5, 15, 7, 17), segment.Horizontal) // D3
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
		{
			desc:       "toggles segment on",
			cellCanvas: image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows),
			update: func(d *Display) error {
				for _, seg := range AllSegments() {
					if err := d.SetSegment(seg); err != nil {
						return err
					}
				}
				if err := d.ToggleSegment(D1); err != nil {
					return err
				}
				return d.ToggleSegment(D1)
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(5, 6, 7, 8), segment.Horizontal)   // D1
				testsegment.MustHV(bc, image.Rect(5, 12, 7, 14), segment.Horizontal) // D2
				testsegment.MustHV(bc, image.Rect(5, 15, 7, 17), segment.Horizontal) // D3
				testbraille.MustApply(bc, ft)
				return ft
			},
		},

		{
			desc:       "larger display 18x15, all segments",
			cellCanvas: image.Rect(0, 0, 3*segdisp.MinCols, 3*segdisp.MinRows),
			update: func(d *Display) error {
				for _, seg := range AllSegments() {
					if err := d.SetSegment(seg); err != nil {
						return err
					}
				}
				return nil
			},
			want: func(size image.Point) *faketerm.Terminal {
				ft := faketerm.MustNew(size)
				bc := testbraille.MustNew(ft.Area())

				testsegment.MustHV(bc, image.Rect(15, 18, 21, 24), segment.Horizontal) // D1
				testsegment.MustHV(bc, image.Rect(15, 36, 21, 42), segment.Horizontal) // D2
				testsegment.MustHV(bc, image.Rect(15, 51, 21, 57), segment.Horizontal) // D3
				testbraille.MustApply(bc, ft)
				return ft
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			d := New(tc.opts...)
			if tc.update != nil {
				err := tc.update(d)
				if (err != nil) != tc.wantUpdateErr {
					t.Errorf("tc.update => unexpected error: %v, wantUpdateErr: %v", err, tc.wantUpdateErr)
				}
				if err != nil {
					return
				}
			}

			cvs, err := canvas.New(tc.cellCanvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			{
				err := d.Draw(cvs, tc.drawOpts...)
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

// mustDrawSegments returns a fake terminal of the specified size with the
// segments drawn on it or panics.
func mustDrawSegments(size image.Point, seg ...Segment) *faketerm.Terminal {
	ft := faketerm.MustNew(size)
	cvs := testcanvas.MustNew(ft.Area())

	d := New()
	for _, s := range seg {
		if err := d.SetSegment(s); err != nil {
			panic(err)
		}
	}

	if err := d.Draw(cvs); err != nil {
		panic(err)
	}

	testcanvas.MustApply(cvs, ft)
	return ft
}

func TestSetCharacter(t *testing.T) {
	tests := []struct {
		desc string
		char rune
		// If not nil, it is called before Draw is called and can set, clear or
		// toggle segments or characters.
		update  func(*Display) error
		want    func(size image.Point) *faketerm.Terminal
		wantErr bool
	}{
		{
			desc:    "fails on unsupported character",
			char:    'A',
			wantErr: true,
		},
		{
			desc: "doesn't clear the display",
			update: func(d *Display) error {
				return d.SetSegment(D3)
			},
			char: ':',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, D1, D2, D3)
			},
		},
		{
			desc: "displays '.'",
			char: '.',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, D3)
			},
		},
		{
			desc: "displays ':'",
			char: ':',
			want: func(size image.Point) *faketerm.Terminal {
				return mustDrawSegments(size, D1, D2)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			d := New()
			if tc.update != nil {
				err := tc.update(d)
				if err != nil {
					t.Fatalf("tc.update => unexpected error: %v", err)
				}
			}

			{
				err := d.SetCharacter(tc.char)
				if (err != nil) != tc.wantErr {
					t.Errorf("SetCharacter => unexpected error: %v, wantErr: %v", err, tc.wantErr)
				}
				if err != nil {
					return
				}
			}

			ar := image.Rect(0, 0, segdisp.MinCols, segdisp.MinRows)
			cvs, err := canvas.New(ar)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			if err := d.Draw(cvs); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			size := area.Size(ar)
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
				t.Fatalf("SetCharacter => %v", diff)
			}
		})
	}
}

func TestAllSegments(t *testing.T) {
	want := []Segment{D1, D2, D3}
	got := AllSegments()
	sort.Slice(got, func(i, j int) bool {
		return int(got[i]) < int(got[j])
	})
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("AllSegments => unexpected diff (-want, +got):\n%s", diff)
	}
}

func TestSupportedsChars(t *testing.T) {
	want := []rune{'.', ':'}

	gotStr := SupportedChars()
	var got []rune
	for _, r := range gotStr {
		got = append(got, r)
	}
	sort.Slice(got, func(i, j int) bool {
		return int(got[i]) < int(got[j])
	})
	sort.Slice(want, func(i, j int) bool {
		return int(want[i]) < int(want[j])
	})
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("SupportedChars => unexpected diff (-want, +got):\n%s", diff)
	}
}
