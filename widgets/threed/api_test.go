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
	"image"
	"image/color"
	"testing"
)

func TestFriendlyShapeAPI(t *testing.T) {
	model := Cube(ModelSize(2), ModelRune('#'), ModelColor(NeonCyan), ModelPosition(Vector3D{X: 1}))
	if model == nil || len(model.Faces) != 6 {
		t.Fatalf("Cube() faces = %d, want 6", len(model.Faces))
	}
	for _, face := range model.Faces {
		if face.Char != '#' {
			t.Fatalf("face char = %q, want #", face.Char)
		}
		if !face.HasColor || face.Color != NeonCyan {
			t.Fatalf("face color = %+v, want NeonCyan", face.Color)
		}
	}
	if center := model.Center(); center.X < 0.9 || center.X > 1.1 {
		t.Fatalf("center.X = %f, want near 1", center.X)
	}
}

func TestBoardAPIs(t *testing.T) {
	for name, model := range map[string]*Model{
		"text":  TextBoard([]string{"ABC", "DEF"}),
		"logic": LogicBoard([]string{"◆──◆", "│ █│"}),
		"game":  GameBoard([]string{"#####", "#@*!#"}),
	} {
		if model == nil || len(model.Faces) == 0 {
			t.Fatalf("%s board returned empty model", name)
		}
		var glyphs int
		for _, face := range model.Faces {
			if face.RenderMode == FaceRenderGlyph {
				glyphs++
			}
		}
		if glyphs == 0 {
			t.Fatalf("%s board produced no glyph faces", name)
		}
	}
}

func TestGlyphAndImageAPIs(t *testing.T) {
	glyph := Glyph("✦", ModelSize(1.2), ModelColor(Amber))
	if glyph == nil || len(glyph.Faces) != 1 {
		t.Fatalf("Glyph() faces = %d, want 1", len(glyph.Faces))
	}
	if got := glyph.Faces[0].Char; got != '✦' {
		t.Fatalf("Glyph() char = %q, want ✦", got)
	}

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.Black)
		}
	}
	model := ModelFromImage(img, ModelSize(0.5), ModelPosition(Vector3D{X: 1}))
	if model == nil || len(model.Faces) == 0 {
		t.Fatal("ModelFromImage() returned empty model")
	}
}

func TestKMLAPI(t *testing.T) {
	model, err := ModelFromKML(&KML{
		Document: Document{
			Placemarks: []Placemark{
				{
					Name: "point",
					Point: &Point{
						Coordinates: "0,0,0",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("ModelFromKML() => unexpected error: %v", err)
	}
	if model == nil || len(model.Faces) == 0 {
		t.Fatal("ModelFromKML() returned empty model")
	}
}

func TestModelCompositionHelpers(t *testing.T) {
	left := Cube(ModelPosition(Vector3D{X: -1}))
	right := Sphere(ModelPosition(Vector3D{X: 1}), ModelSegments(4, 6))
	combined := NewModel()
	combined.Append(left, right)
	if got, wantMin := len(combined.Faces), len(left.Faces)+len(right.Faces); got != wantMin {
		t.Fatalf("Append() faces = %d, want %d", got, wantMin)
	}

	cloned := combined.Clone()
	cloned.Move(Vector3D{X: 2})
	if combined.Center().X == cloned.Center().X {
		t.Fatal("Clone()/Move() mutated the original model")
	}
}

// TestSpectrumAnalyzerBuildsBars verifies the reusable analyzer helper creates geometry.
func TestSpectrumAnalyzerBuildsBars(t *testing.T) {
	model := SpectrumAnalyzer([]float64{0.1, 0.5, 0.9}, ModelSize(2))
	if model == nil || len(model.Faces) == 0 {
		t.Fatalf("SpectrumAnalyzer() returned no geometry")
	}
	if got, wantMin := len(model.Faces), 3+2; got < wantMin {
		t.Fatalf("SpectrumAnalyzer() faces = %d, want at least %d", got, wantMin)
	}
}
