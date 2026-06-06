// Copyright 2026 Google Inc.
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
	"math"
	"testing"

	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/widgetapi"
)

// TestBrailleOnlyUsesFullCanvas verifies the plot can occupy the full widget.
func TestBrailleOnlyUsesFullCanvas(t *testing.T) {
	lc, err := New(BrailleOnly())
	if err != nil {
		t.Fatalf("New(BrailleOnly) => unexpected error: %v", err)
	}
	if got, want := lc.minSize(), (image.Point{X: 1, Y: 1}); got != want {
		t.Fatalf("minSize = %v, want %v", got, want)
	}
	if err := lc.Series("series", []float64{0, 1, 2, 3, 4, 5}); err != nil {
		t.Fatalf("Series => unexpected error: %v", err)
	}

	ft := faketerm.MustNew(image.Point{X: 4, Y: 2})
	cvs := testcanvas.MustNew(ft.Area())
	if err := lc.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	if got, want := lc.ValueCapacity(), 8; got != want {
		t.Fatalf("ValueCapacity = %d, want %d", got, want)
	}
}

// TestSampleVisibleSeriesLTTBPreservesPeak verifies LTTB keeps sharp spikes.
func TestSampleVisibleSeriesLTTBPreservesPeak(t *testing.T) {
	values := []float64{1, 1, 1, 1, 100, 1, 1, 1, 1}
	segments := sampleVisibleSeries(values, 0, len(values)-1, 4, downsamplerModeLTTB)
	if got, want := len(segments), 1; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}

	var sawPeak bool
	for _, point := range segments[0] {
		if point.index == 4 && point.value == 100 {
			sawPeak = true
			break
		}
	}
	if !sawPeak {
		t.Fatalf("downsampled points = %v, want to preserve the peak", segments[0])
	}
}

// TestSampleVisibleSeriesPreservesNaNGaps verifies NaN gaps split segments.
func TestSampleVisibleSeriesPreservesNaNGaps(t *testing.T) {
	values := []float64{0, 1, 2, math.NaN(), 4, 5}
	segments := sampleVisibleSeries(values, 0, len(values)-1, 4, downsamplerModeLTTB)
	if got, want := len(segments), 2; got != want {
		t.Fatalf("segment count = %d, want %d", got, want)
	}
	if got, want := segments[0][0].index, 0; got != want {
		t.Fatalf("first segment starts at %d, want %d", got, want)
	}
	if got, want := segments[1][0].index, 4; got != want {
		t.Fatalf("second segment starts at %d, want %d", got, want)
	}
}
