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

package linechart

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/widgetapi"
)

func TestLineChartDraws(t *testing.T) {
	t.Skip() // Unimplemented.
	tests := []struct {
		desc         string
		canvas       image.Rectangle
		opts         []Option
		writes       func(*LineChart) error
		want         func(size image.Point) *faketerm.Terminal
		wantWriteErr bool
	}{
		{
			desc:   "empty without series",
			canvas: image.Rect(0, 0, 1, 1),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := canvas.New(tc.canvas)
			if err != nil {
				t.Fatalf("canvas.New => unexpected error: %v", err)
			}

			widget := New(tc.opts...)
			if tc.writes != nil {
				err := tc.writes(widget)
				if (err != nil) != tc.wantWriteErr {
					t.Errorf("Series => unexpected error: %v, wantWriteErr: %v", err, tc.wantWriteErr)
				}
				if err != nil {
					return
				}
			}

			if err := widget.Draw(c); err != nil {
				t.Fatalf("Draw => unexpected error: %v", err)
			}

			got, err := faketerm.New(c.Size())
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			if err := c.Apply(got); err != nil {
				t.Fatalf("Apply => unexpected error: %v", err)
			}

			want := faketerm.MustNew(c.Size())
			if tc.want != nil {
				want = tc.want(c.Size())
			}
			if diff := faketerm.Diff(want, got); diff != "" {
				t.Errorf("Draw => %v", diff)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		desc string
		// if not nil, executed before obtaining the options.
		addSeries func(*LineChart) error
		want      widgetapi.Options
	}{
		{
			desc: "reserves space for axis without series",
			want: widgetapi.Options{
				MinimumSize: image.Point{3, 3},
			},
		},
		{
			desc: "reserves space for longer Y labels",
			addSeries: func(lc *LineChart) error {
				return lc.Series("series", []float64{0, 100})
			},
			want: widgetapi.Options{
				MinimumSize: image.Point{5, 3},
			},
		},
		{
			desc: "reserves space for negative Y labels",
			addSeries: func(lc *LineChart) error {
				return lc.Series("series", []float64{-100, 100})
			},
			want: widgetapi.Options{
				MinimumSize: image.Point{6, 3},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			lc := New()

			if tc.addSeries != nil {
				if err := tc.addSeries(lc); err != nil {
					t.Fatalf("tc.addSeries => %v", err)
				}
			}
			got := lc.Options()
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("Options => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
