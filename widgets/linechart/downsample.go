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

import "math"

// samplePoint represents one visible point on the plotted line.
type samplePoint struct {
	index int
	value float64
}

// sampleVisibleSeries returns the visible, finite segments for a series.
//
// NaN values split the series into disconnected segments. When the selected
// downsampler is enabled and the visible data exceeds the target resolution,
// each segment is reduced before drawing.
func sampleVisibleSeries(values []float64, minIndex, maxIndex, target int, mode downsamplerMode) [][]samplePoint {
	if len(values) == 0 || maxIndex < minIndex {
		return nil
	}
	if minIndex < 0 {
		minIndex = 0
	}
	if maxIndex >= len(values) {
		maxIndex = len(values) - 1
	}

	segments := collectVisibleSegments(values, minIndex, maxIndex)
	if len(segments) == 0 || mode != downsamplerModeLTTB || target < 3 {
		return segments
	}

	total := 0
	for _, segment := range segments {
		total += len(segment)
	}
	if total <= target {
		return segments
	}

	out := make([][]samplePoint, 0, len(segments))
	for _, segment := range segments {
		share := int(math.Round(float64(len(segment)) / float64(total) * float64(target)))
		if share < 2 {
			share = 2
		}
		if share > len(segment) {
			share = len(segment)
		}
		out = append(out, lttbDownsample(segment, share))
	}
	return out
}

// collectVisibleSegments splits the visible range into finite point segments.
func collectVisibleSegments(values []float64, minIndex, maxIndex int) [][]samplePoint {
	var (
		segments [][]samplePoint
		current  []samplePoint
	)
	for i := minIndex; i <= maxIndex; i++ {
		value := values[i]
		if math.IsNaN(value) {
			if len(current) > 0 {
				segments = append(segments, current)
				current = nil
			}
			continue
		}
		current = append(current, samplePoint{index: i, value: value})
	}
	if len(current) > 0 {
		segments = append(segments, current)
	}
	return segments
}

// lttbDownsample applies Largest-Triangle-Three-Buckets to one point segment.
func lttbDownsample(points []samplePoint, threshold int) []samplePoint {
	if len(points) <= threshold || threshold < 3 {
		return append([]samplePoint(nil), points...)
	}

	sampled := make([]samplePoint, 0, threshold)
	sampled = append(sampled, points[0])

	every := float64(len(points)-2) / float64(threshold-2)
	a := 0
	for i := 0; i < threshold-2; i++ {
		avgStart := int(math.Floor(float64(i+1)*every)) + 1
		avgEnd := int(math.Floor(float64(i+2)*every)) + 1
		if avgEnd > len(points) {
			avgEnd = len(points)
		}
		avgX, avgY := averagePoint(points, avgStart, avgEnd)

		rangeStart := int(math.Floor(float64(i)*every)) + 1
		rangeEnd := int(math.Floor(float64(i+1)*every)) + 1
		if rangeEnd > len(points)-1 {
			rangeEnd = len(points) - 1
		}

		nextA := rangeStart
		maxArea := -1.0
		for j := rangeStart; j < rangeEnd; j++ {
			area := triangleArea(points[a], points[j], avgX, avgY)
			if area > maxArea {
				maxArea = area
				nextA = j
			}
		}
		sampled = append(sampled, points[nextA])
		a = nextA
	}

	sampled = append(sampled, points[len(points)-1])
	return sampled
}

// averagePoint returns the centroid of the provided point range.
func averagePoint(points []samplePoint, start, end int) (float64, float64) {
	if start >= end {
		last := points[len(points)-1]
		return float64(last.index), last.value
	}

	var sumX, sumY float64
	for _, point := range points[start:end] {
		sumX += float64(point.index)
		sumY += point.value
	}
	count := float64(end - start)
	return sumX / count, sumY / count
}

// triangleArea returns the area of the triangle formed by A, B, and C.
func triangleArea(a, b samplePoint, cx, cy float64) float64 {
	ax := float64(a.index)
	bx := float64(b.index)
	return math.Abs((ax-cx)*(b.value-a.value)-(ax-bx)*(cy-a.value)) * 0.5
}
