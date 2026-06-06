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

import "math"

const (
	defaultStereoAttack = 0.68
	defaultStereoDecay  = 0.18
	defaultRightAttack  = 0.64
	defaultRightDecay   = 0.16
	defaultHalfAttack   = 0.72
	defaultHalfDecay    = 0.21
)

// Synth generates smoothed synthetic activity suitable for driving spectrum
// widgets in demos and quick prototypes.
//
// The output slices are reused between calls to Step. Treat them as read-only
// and copy if they need to be retained.
type Synth struct {
	phase float64
	max   int

	left        []int
	right       []int
	half        []int
	leftTarget  []int
	rightTarget []int
	halfTarget  []int
}

// NewSynth allocates a reusable synthetic feed generator.
func NewSynth(stereoBins, halfDuplexBins, max int) *Synth {
	if stereoBins < 1 {
		stereoBins = 1
	}
	if halfDuplexBins < 1 {
		halfDuplexBins = 1
	}
	if max < 1 {
		max = 1
	}
	return &Synth{
		max:         max,
		left:        make([]int, stereoBins),
		right:       make([]int, stereoBins),
		half:        make([]int, halfDuplexBins),
		leftTarget:  make([]int, stereoBins),
		rightTarget: make([]int, stereoBins),
		halfTarget:  make([]int, halfDuplexBins),
	}
}

// Step advances internal phase and returns the current stereo and half-duplex
// samples.
func (s *Synth) Step() (left, right, half []int) {
	s.phase += 0.14
	fillSpectrumTargets(s.leftTarget, s.phase, 0.0, s.max)
	fillSpectrumTargets(s.rightTarget, s.phase, 0.38, s.max)
	fillPulseTargets(s.halfTarget, s.phase, s.max)
	smoothSeries(s.left, s.leftTarget, defaultStereoAttack, defaultStereoDecay, s.max)
	smoothSeries(s.right, s.rightTarget, defaultRightAttack, defaultRightDecay, s.max)
	smoothSeries(s.half, s.halfTarget, defaultHalfAttack, defaultHalfDecay, s.max)
	return s.left, s.right, s.half
}

// fillSpectrumTargets generates a band-swept stereo-like target waveform.
func fillSpectrumTargets(dst []int, phase, offset float64, max int) {
	if len(dst) == 0 {
		return
	}
	last := float64(len(dst) - 1)
	for i := range dst {
		band := float64(i) / maxFloat(last, 1)
		// Build a wavefield across bands so the rendered profile reads like a
		// moving sine contour rather than pure random noise.
		waveA := 0.5 + 0.5*math.Sin(phase*0.42+band*math.Pi*3.2+offset)
		waveB := 0.5 + 0.5*math.Sin(phase*0.77-band*math.Pi*5.1+offset*1.7)
		waveC := 0.5 + 0.5*math.Sin(phase*1.31+band*math.Pi*8.7+offset*0.9)

		// Add controlled harshness with quantized comb accents and burst gates.
		comb := 0.5 + 0.5*math.Sin(phase*4.9+band*35.0+offset*2.4)
		burst := math.Pow(0.5+0.5*math.Sin(phase*0.56+offset+band*4.2), 8)
		picket := math.Pow(0.5+0.5*math.Sin(phase*1.72+band*19.0+offset*1.2), 5)

		envelope := 0.16 + 0.32*waveA + 0.28*waveB + 0.24*waveC
		edged := envelope + 0.18*comb + 0.24*burst + 0.14*picket
		v := int(edged * float64(max))
		// Quantization gives the signal a more digital/professional analyzer
		// cadence instead of an overly smooth analog look.
		v = (v / 4) * 4
		dst[i] = clampSample(v, max)
	}
}

// fillPulseTargets generates bursty telemetry-style half-duplex activity.
func fillPulseTargets(dst []int, phase float64, max int) {
	if len(dst) == 0 {
		return
	}
	last := float64(len(dst) - 1)
	for i := range dst {
		band := float64(i) / maxFloat(last, 1)
		baseline := 0.20 + 0.10*(0.5+0.5*math.Sin(phase*0.37+band*4.1))
		jitter := 0.09 + 0.14*(0.5+0.5*math.Sin(phase*5.6+band*13.7))
		spikeEnvelope := math.Pow(0.5+0.5*math.Sin(phase*1.9+band*8.4), 10)
		spike := spikeEnvelope * (0.62 + 0.38*(0.5+0.5*math.Sin(phase*11.2+band*17.9)))
		dropEnvelope := math.Pow(0.5+0.5*math.Sin(phase*0.73+band*6.2+1.4), 13)
		drop := dropEnvelope * 0.22
		pathSweep := math.Exp(-math.Pow(band-(0.50+0.34*math.Sin(phase*0.69)), 2) / 0.010)
		energy := baseline + jitter + spike*0.56 + pathSweep*0.22 - drop
		v := int(energy * float64(max))
		v = (v / 6) * 6
		dst[i] = clampSample(v, max)
	}
}

// smoothSeries applies fast attack and slower decay to reduce flicker.
func smoothSeries(dst, target []int, attack, decay float64, max int) {
	for i := range dst {
		diff := target[i] - dst[i]
		rate := decay
		if diff > 0 {
			rate = attack
		}
		next := float64(dst[i]) + float64(diff)*rate
		if math.Abs(float64(diff)) < 1 {
			dst[i] = target[i]
			continue
		}
		dst[i] = clampSample(int(next+0.5), max)
	}
}

// clampSample keeps generated values inside the configured spectrum range.
func clampSample(v, max int) int {
	switch {
	case v < 0:
		return 0
	case v > max:
		return max
	default:
		return v
	}
}

// maxFloat returns the larger of two float64 values.
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
