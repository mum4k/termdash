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

package spectrum

import "testing"

// TestNewSynthNormalizesInputs ensures minimal valid generator dimensions.
func TestNewSynthNormalizesInputs(t *testing.T) {
	s := NewSynth(0, -3, 0)
	if len(s.left) != 1 || len(s.right) != 1 || len(s.half) != 1 {
		t.Fatalf("unexpected synth sizes: left=%d right=%d half=%d", len(s.left), len(s.right), len(s.half))
	}
	if s.max != 1 {
		t.Fatalf("max = %d, want 1", s.max)
	}
}

// TestSynthStepProducesBoundedSamples ensures frame output stays in range.
func TestSynthStepProducesBoundedSamples(t *testing.T) {
	s := NewSynth(8, 10, 600)
	left, right, half := s.Step()

	if s.phase <= 0 {
		t.Fatalf("phase = %v, want > 0", s.phase)
	}
	for _, series := range [][]int{left, right, half} {
		for i, sample := range series {
			if sample < 0 || sample > s.max {
				t.Fatalf("sample[%d] = %d, want 0 <= sample <= %d", i, sample, s.max)
			}
		}
	}
}

// TestSynthStepReusesBuffers ensures no per-frame slice reallocation.
func TestSynthStepReusesBuffers(t *testing.T) {
	s := NewSynth(6, 7, 500)
	left1, right1, half1 := s.Step()
	left2, right2, half2 := s.Step()

	if &left1[0] != &left2[0] {
		t.Fatal("left channel buffer was reallocated between steps")
	}
	if &right1[0] != &right2[0] {
		t.Fatal("right channel buffer was reallocated between steps")
	}
	if &half1[0] != &half2[0] {
		t.Fatal("half channel buffer was reallocated between steps")
	}
}

// TestSynthHelpers verifies helper math used by the synthetic signal generator.
func TestSynthHelpers(t *testing.T) {
	dst := make([]int, 6)
	fillSpectrumTargets(dst, 0.4, 1.2, 600)
	for i, sample := range dst {
		if sample < 0 || sample > 600 {
			t.Fatalf("fillSpectrumTargets sample[%d] = %d, want bounded sample", i, sample)
		}
	}

	fillPulseTargets(dst, 0.9, 600)
	for i, sample := range dst {
		if sample < 0 || sample > 600 {
			t.Fatalf("fillPulseTargets sample[%d] = %d, want bounded sample", i, sample)
		}
	}

	cur := []int{0, 100, 300}
	target := []int{600, 0, 450}
	smoothSeries(cur, target, 0.5, 0.25, 600)
	if cur[0] <= 0 || cur[0] >= target[0] {
		t.Fatalf("attack smoothing = %d, want 0 < sample < %d", cur[0], target[0])
	}
	if cur[1] >= 100 {
		t.Fatalf("decay smoothing = %d, want lower value", cur[1])
	}
	if got := clampSample(-1, 600); got != 0 {
		t.Fatalf("clampSample(-1) = %d, want 0", got)
	}
	if got := clampSample(700, 600); got != 600 {
		t.Fatalf("clampSample(high) = %d, want 600", got)
	}
	if got := maxFloat(2, 4); got != 4 {
		t.Fatalf("maxFloat = %v, want 4", got)
	}
}
