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

package threed

// SpectrumAnalyzer converts normalized band values into a 3D bar analyzer.
func SpectrumAnalyzer(bands []float64, opts ...ModelOption) *Model {
	cfg := newModelOptions(opts...)
	values := normalizedBands(bands)
	if len(values) == 0 {
		values = normalizedBands([]float64{0.18, 0.35, 0.52, 0.74, 0.61, 0.48, 0.34, 0.22})
	}

	model := NewModel()
	totalWidth := cfg.size
	if totalWidth <= 0 {
		totalWidth = 1
	}
	step := totalWidth / float64(len(values))
	barWidth := step * 0.78
	depth := step * 0.46
	heightBase := totalWidth * 0.08
	heightMax := totalWidth * 0.86
	left := -totalWidth / 2
	floorY := -totalWidth * 0.42
	backY := floorY + totalWidth*0.10
	railY := floorY - depth*0.42
	backRailY := railY + totalWidth*0.04

	addSpectrumRail(model, totalWidth, railY, 0, spectrumRailColor())
	addSpectrumRail(model, totalWidth*0.94, backRailY, depth, spectrumRailColor())

	for i, value := range values {
		height := heightBase + value*heightMax
		x0 := left + float64(i)*step + (step-barWidth)/2
		x1 := x0 + barWidth
		topY := floorY + height
		backTopY := topY + totalWidth*0.045
		color := spectrumBandColor(value)
		if cfg.hasColor {
			color = cfg.color
		}
		addSpectrumBar(model, x0, x1, floorY, topY, backY, backTopY, depth, color)
	}

	if cfg.position != (Vector3D{}) {
		model.Move(cfg.position)
	}
	return model
}

// NetworkSpectrum converts download and upload rates into a split 3D spectrum.
func NetworkSpectrum(download, upload []float64, opts ...ModelOption) *Model {
	cfg := newModelOptions(opts...)
	bands := maxInt(len(download), len(upload))
	if bands == 0 {
		bands = 12
	}
	downValues, upValues := normalizedBandPair(download, upload, bands)

	model := NewModel()
	totalWidth := cfg.size
	if totalWidth <= 0 {
		totalWidth = 1
	}
	step := totalWidth / float64(bands)
	barWidth := step * 0.74
	depth := step * 0.42
	left := -totalWidth / 2
	topFloor := totalWidth * 0.06
	bottomFloor := -totalWidth * 0.48
	backOffset := totalWidth * 0.06
	heightBase := totalWidth * 0.025
	heightMax := totalWidth * 0.36
	topRailY := topFloor - depth*0.42
	bottomRailY := bottomFloor - depth*0.42

	addSpectrumRail(model, totalWidth, topRailY, 0, RGB(42, 120, 230))
	addSpectrumRail(model, totalWidth, bottomRailY, 0, RGB(160, 58, 92))

	for i := 0; i < bands; i++ {
		x0 := left + float64(i)*step + (step-barWidth)/2
		x1 := x0 + barWidth
		downHeight := heightBase + downValues[i]*heightMax
		uploadHeight := heightBase + upValues[i]*heightMax
		addSpectrumBar(model, x0, x1, topFloor, topFloor+downHeight, topFloor+backOffset, topFloor+backOffset+downHeight, depth, RGB(72, 210, 255))
		addSpectrumBar(model, x0, x1, bottomFloor, bottomFloor+uploadHeight, bottomFloor+backOffset, bottomFloor+backOffset+uploadHeight, depth, RGB(255, 88, 118))
	}

	if cfg.position != (Vector3D{}) {
		model.Move(cfg.position)
	}
	return model
}

// normalizedBands clamps and scales bands to 0..1.
func normalizedBands(bands []float64) []float64 {
	if len(bands) == 0 {
		return nil
	}
	max := 0.0
	for _, value := range bands {
		if value > max {
			max = value
		}
	}
	if max <= 1 {
		max = 1
	}
	out := make([]float64, len(bands))
	for i, value := range bands {
		switch {
		case value < 0:
			out[i] = 0
		case value > max:
			out[i] = 1
		default:
			out[i] = value / max
		}
	}
	return out
}

// normalizedBandPair clamps and scales two band sets against their shared peak.
func normalizedBandPair(download, upload []float64, count int) ([]float64, []float64) {
	maxValue := 1.0
	for _, value := range download {
		if value > maxValue {
			maxValue = value
		}
	}
	for _, value := range upload {
		if value > maxValue {
			maxValue = value
		}
	}
	normalize := func(values []float64) []float64 {
		out := make([]float64, count)
		offset := count - len(values)
		if offset < 0 {
			values = values[-offset:]
			offset = 0
		}
		for i, value := range values {
			switch {
			case value < 0:
				out[offset+i] = 0
			case value > maxValue:
				out[offset+i] = 1
			default:
				out[offset+i] = value / maxValue
			}
		}
		return out
	}
	return normalize(download), normalize(upload)
}

// addSpectrumBar adds one stationary extruded analyzer bar.
func addSpectrumBar(model *Model, x0, x1, y0, y1, backY, backTopY, depth float64, color Color) {
	frontBottom := y0 - depth*0.12
	model.AddFace(Face{
		Vertices: []Vector3D{
			{X: x0, Y: frontBottom, Z: 0},
			{X: x0, Y: y1, Z: 0},
			{X: x1, Y: y1, Z: 0},
			{X: x1, Y: frontBottom, Z: 0},
		},
		Char:       '█',
		RenderMode: FaceRenderFill,
		Color:      color,
		HasColor:   true,
	})
	model.AddFace(Face{
		Vertices: []Vector3D{
			{X: x1, Y: frontBottom, Z: 0},
			{X: x1, Y: y1, Z: 0},
			{X: x1 + depth, Y: backTopY, Z: depth},
			{X: x1 + depth, Y: backY, Z: depth},
		},
		Char:       '▓',
		RenderMode: FaceRenderFill,
		Color:      color.Multiply(0.62),
		HasColor:   true,
	})
	model.AddFace(Face{
		Vertices: []Vector3D{
			{X: x0, Y: y1, Z: 0},
			{X: x1, Y: y1, Z: 0},
			{X: x1 + depth, Y: backTopY, Z: depth},
			{X: x0 + depth, Y: backTopY, Z: depth},
		},
		Char:       '▒',
		RenderMode: FaceRenderFill,
		Color:      color.Multiply(0.78),
		HasColor:   true,
	})
	// Bottom face: connects the front-bottom edge to the back-bottom edge.
	// Without this face the floor of each bar is missing, which causes the
	// bottom row of the raster to be empty except at the outermost edges.
	model.AddFace(Face{
		Vertices: []Vector3D{
			{X: x0, Y: frontBottom, Z: 0},
			{X: x1, Y: frontBottom, Z: 0},
			{X: x1 + depth, Y: backY, Z: depth},
			{X: x0 + depth, Y: backY, Z: depth},
		},
		Char:       '▓',
		RenderMode: FaceRenderFill,
		Color:      color.Multiply(0.45),
		HasColor:   true,
	})
}

// addSpectrumRail adds one floor guide rail under the analyzer.
func addSpectrumRail(model *Model, width, y, z float64, color Color) {
	model.AddFace(Face{
		Vertices: []Vector3D{
			{X: -width / 2, Y: y, Z: z},
			{X: width / 2, Y: y, Z: z},
		},
		Char:     '─',
		Color:    color,
		HasColor: true,
	})
}

// spectrumBandColor maps band energy to a Winamp-like cyan-green-amber ramp.
func spectrumBandColor(value float64) Color {
	switch {
	case value >= 0.82:
		return RGB(255, 210, 80)
	case value >= 0.62:
		return RGB(150, 255, 96)
	case value >= 0.38:
		return RGB(72, 230, 210)
	default:
		return RGB(42, 120, 230)
	}
}

// spectrumRailColor returns the low-emphasis analyzer rail color.
func spectrumRailColor() Color {
	return RGB(60, 78, 96)
}
