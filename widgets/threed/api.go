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

import (
	"context"
	"image"
	"io"
	"log"
)

// ShapeKind identifies a built-in primitive shape.
type ShapeKind int

const (
	ShapeCube ShapeKind = iota
	ShapePyramid
	ShapeTetrahedron
	ShapeOctahedron
	ShapeSphere
)

// ModelOption configures shape, glyph, board, chart, and conversion helpers.
type ModelOption interface {
	setModelOption(*modelOptions)
}

type modelOptionFunc func(*modelOptions)

func (f modelOptionFunc) setModelOption(opts *modelOptions) {
	f(opts)
}

type modelOptions struct {
	position   Vector3D
	size       float64
	char       rune
	color      Color
	hasColor   bool
	cellWidth  float64
	cellHeight float64
	lat        int
	lon        int
	centered   bool
}

func newModelOptions(opts ...ModelOption) modelOptions {
	cfg := modelOptions{
		size:       1,
		char:       '█',
		cellWidth:  0.06,
		cellHeight: 0.16,
		lat:        8,
		lon:        12,
		centered:   true,
	}
	for _, opt := range opts {
		opt.setModelOption(&cfg)
	}
	return cfg
}

// ModelPosition sets the model origin or center.
func ModelPosition(position Vector3D) ModelOption {
	return modelOptionFunc(func(opts *modelOptions) {
		opts.position = position
	})
}

// ModelSize sets the size used by shapes and single glyphs.
func ModelSize(size float64) ModelOption {
	return modelOptionFunc(func(opts *modelOptions) {
		if size > 0 {
			opts.size = size
		}
	})
}

// ModelRune sets the face or glyph rune.
func ModelRune(char rune) ModelOption {
	return modelOptionFunc(func(opts *modelOptions) {
		if char != 0 {
			opts.char = char
		}
	})
}

// ModelColor applies a fixed color to all faces produced by the helper.
func ModelColor(color Color) ModelOption {
	return modelOptionFunc(func(opts *modelOptions) {
		opts.color = color
		opts.hasColor = true
	})
}

// ModelCellSize sets the cell spacing used by text, logic, and game boards.
func ModelCellSize(width, height float64) ModelOption {
	return modelOptionFunc(func(opts *modelOptions) {
		if width > 0 {
			opts.cellWidth = width
		}
		if height > 0 {
			opts.cellHeight = height
		}
	})
}

// ModelSegments sets sphere tessellation as latitude and longitude segments.
func ModelSegments(lat, lon int) ModelOption {
	return modelOptionFunc(func(opts *modelOptions) {
		if lat > 0 {
			opts.lat = lat
		}
		if lon > 0 {
			opts.lon = lon
		}
	})
}

// ModelCentered controls whether text board helpers center their rows.
func ModelCentered(centered bool) ModelOption {
	return modelOptionFunc(func(opts *modelOptions) {
		opts.centered = centered
	})
}

// Shape creates a built-in primitive model.
func Shape(kind ShapeKind, opts ...ModelOption) *Model {
	cfg := newModelOptions(opts...)
	var model *Model
	switch kind {
	case ShapePyramid:
		model = CreatePyramid(cfg.position, cfg.size, cfg.char)
	case ShapeTetrahedron:
		model = CreateTetrahedron(cfg.position, cfg.size)
	case ShapeOctahedron:
		model = CreateOctahedron(cfg.position, cfg.size)
	case ShapeSphere:
		model = CreateSphere(cfg.position, cfg.size, cfg.lat, cfg.lon, cfg.char)
	default:
		model = CreateCube(cfg.position, cfg.size, cfg.char)
	}
	if cfg.hasColor {
		model.SetColor(cfg.color)
	}
	return model
}

// Cube creates a cube model.
func Cube(opts ...ModelOption) *Model {
	return Shape(ShapeCube, opts...)
}

// Pyramid creates a square pyramid model.
func Pyramid(opts ...ModelOption) *Model {
	return Shape(ShapePyramid, opts...)
}

// Tetrahedron creates a tetrahedron model.
func Tetrahedron(opts ...ModelOption) *Model {
	return Shape(ShapeTetrahedron, opts...)
}

// Octahedron creates an octahedron model.
func Octahedron(opts ...ModelOption) *Model {
	return Shape(ShapeOctahedron, opts...)
}

// Sphere creates a low-poly sphere model.
func Sphere(opts ...ModelOption) *Model {
	return Shape(ShapeSphere, opts...)
}

// Glyph converts a UTF-8 string into a single 3D glyph billboard.
func Glyph(text string, opts ...ModelOption) *Model {
	cfg := newModelOptions(opts...)
	color := cfg.color
	if !cfg.hasColor {
		color = terminalBoardColor(RenderableRune(text, cfg.char))
	}
	model := NewModel()
	addGlyphBillboard(model, cfg.position, cfg.size, RenderableRune(text, cfg.char), color)
	return model
}

// SymbolSpinner converts a UTF-8 string or bundled emoji into an animated model.
func SymbolSpinner(text string, step int) *Model {
	return NewAnimatedSymbolSpinner(text, step)
}

// ModelFromImage converts an image into a 3D model.
func ModelFromImage(img image.Image, opts ...ModelOption) *Model {
	model := ImageToModel(img)
	return applyOutputOptions(model, newModelOptions(opts...))
}

// ModelFromImageFile loads an image file and converts it into a 3D model.
func ModelFromImageFile(path string, opts ...ModelOption) (*Model, error) {
	model, err := LoadImageModel(path)
	if err != nil {
		return nil, err
	}
	return applyOutputOptions(model, newModelOptions(opts...)), nil
}

// ModelFromKML converts parsed KML data into a 3D model.
func ModelFromKML(kml *KML, opts ...ModelOption) (*Model, error) {
	model, err := GenerateModelFromKML(kml, quietLogger())
	if err != nil {
		return nil, err
	}
	return applyOutputOptions(model, newModelOptions(opts...)), nil
}

// ModelFromKMLURL fetches a KML document and converts it into a 3D model.
func ModelFromKMLURL(ctx context.Context, url string, opts ...ModelOption) (*Model, error) {
	kml, err := FetchAndParseKML(ctx, url, quietLogger())
	if err != nil {
		return nil, err
	}
	return ModelFromKML(kml, opts...)
}

// LineChart converts data into a 3D line chart model.
func LineChart(data []float64, opts ...ModelOption) *Model {
	return applyOutputOptions(GenerateLineChartModel(data), newModelOptions(opts...))
}

// BarChart converts data into a 3D bar chart model.
func BarChart(data []float64, opts ...ModelOption) *Model {
	return applyOutputOptions(GenerateBarChartModel(data), newModelOptions(opts...))
}

func applyOutputOptions(model *Model, cfg modelOptions) *Model {
	if model == nil {
		return nil
	}
	if cfg.size != 1 {
		model.Scale(cfg.size)
	}
	if cfg.position != (Vector3D{}) {
		model.Move(cfg.position)
	}
	if cfg.hasColor {
		model.SetColor(cfg.color)
	}
	return model
}

func quietLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}
