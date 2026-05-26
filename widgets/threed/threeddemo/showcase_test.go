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

package main

import (
	"testing"

	"github.com/mum4k/termdash/widgets/threed"
)

// TestBuildShowcaseScenes verifies that the showcase exposes the expected scene set.
func TestBuildShowcaseScenes(t *testing.T) {
	scenes := buildShowcaseScenes(false)
	if got, want := len(scenes), showcaseSceneCount; got != want {
		t.Fatalf("len(buildShowcaseScenes(false)) = %d, want %d", got, want)
	}
	if got, want := scenes[0].Name, "Terminal Forms"; got != want {
		t.Fatalf("scenes[0].Name = %q, want %q", got, want)
	}
	for i, want := range []string{"Cube Core", "Pyramid Spire", "Sphere Shell", "Octa Prism", "Tetra Field"} {
		if got := scenes[i+4].Name; got != want {
			t.Fatalf("scenes[%d].Name = %q, want %q", i+4, got, want)
		}
	}

	withAsset := buildShowcaseScenes(true)
	if got, want := withAsset[3].Name, "Image Relief"; got != want {
		t.Fatalf("asset scene name = %q, want %q", got, want)
	}
}

// TestShapeSelectionsAre3DShapes verifies demo selections 5-9 are shape scenes.
func TestShapeSelectionsAre3DShapes(t *testing.T) {
	for i, scene := range buildShowcaseScenes(false)[4:9] {
		model := scene.Build(7, nil)
		var polygons, glyphs, lines int
		for _, face := range model.Faces {
			if len(face.Vertices) >= 3 {
				polygons++
			}
			if face.RenderMode == threed.FaceRenderGlyph {
				glyphs++
			}
			if len(face.Vertices) == 2 {
				lines++
			}
		}
		if polygons < 4 {
			t.Fatalf("selection %d (%s) polygon faces = %d, want 3D shape geometry", i+5, scene.Name, polygons)
		}
		if glyphs != 0 {
			t.Fatalf("selection %d (%s) glyph faces = %d, want shape only", i+5, scene.Name, glyphs)
		}
		if lines != 0 {
			t.Fatalf("selection %d (%s) line faces = %d, want no rails/shadow lines", i+5, scene.Name, lines)
		}
		if scene.Orbit == (threed.Vector3D{}) {
			t.Fatalf("selection %d (%s) has no orbit", i+5, scene.Name)
		}
	}
}

// TestCircuitScenesIncludeGlyphForms verifies the reference-inspired scenes use
// glyph-rendered shapes in addition to filled meshes.
func TestCircuitScenesIncludeGlyphForms(t *testing.T) {
	for _, scene := range buildShowcaseScenes(false)[:2] {
		model := scene.Build(4, nil)
		var glyphFaces int
		for _, face := range model.Faces {
			if face.RenderMode == threed.FaceRenderGlyph {
				glyphFaces++
			}
		}
		if glyphFaces < 10 {
			t.Fatalf("%s scene glyph faces = %d, want a dense glyph layer", scene.Name, glyphFaces)
		}
	}
}

// TestShowcaseSceneBuildersReturnFaces verifies that every curated scene produces geometry.
func TestShowcaseSceneBuildersReturnFaces(t *testing.T) {
	for _, scene := range buildShowcaseScenes(false) {
		model := scene.Build(3, nil)
		if model == nil {
			t.Fatalf("%s scene returned nil model", scene.Name)
		}
		if len(model.Faces) == 0 {
			t.Fatalf("%s scene returned no faces", scene.Name)
		}
	}
}

// TestAssetSceneUsesProvidedModel verifies that the asset scene accepts an image-backed model.
func TestAssetSceneUsesProvidedModel(t *testing.T) {
	asset := threed.CreateCube(threed.Vector3D{}, 1, '█')
	asset.SetColor(threed.Color{R: 0.9, G: 0.4, B: 0.2})

	model := buildAssetReliefScene(2, asset)
	if model == nil || len(model.Faces) == 0 {
		t.Fatal("buildAssetReliefScene() => empty model")
	}

	var foundAssetColor bool
	for _, face := range model.Faces {
		if face.HasColor && face.Color.R > 0.8 && face.Color.G < 0.5 {
			foundAssetColor = true
			break
		}
	}
	if !foundAssetColor {
		t.Fatal("asset scene did not preserve any source-derived asset color")
	}
}

// TestGraphSceneRowsAreRectangular keeps the terminal boards crisp.
func TestGraphSceneRowsAreRectangular(t *testing.T) {
	for name, rows := range map[string][]string{
		"Signal Lattice": signalRigRows(3),
		"Prism Field":    prismFieldRows(4),
	} {
		if len(rows) == 0 {
			t.Fatalf("%s rows are empty", name)
		}
		want := len([]rune(rows[0]))
		for i, row := range rows {
			if got := len([]rune(row)); got != want {
				t.Errorf("%s row %d width = %d, want %d: %q", name, i, got, want, row)
			}
		}
	}
}

// TestCenteredGlyphBoardKeepsBoardCentered verifies scene boards aren't clipped
// by stale hand-tuned left origins when their text changes.
func TestCenteredGlyphBoardKeepsBoardCentered(t *testing.T) {
	rows := []string{
		"╭────╮",
		"│ ABC│",
		"╰────╯",
	}
	model := threed.CenteredGlyphBoard(threed.Vector3D{X: 0, Y: 0, Z: 0}, rows, 0.1, 0.2, graphColor)
	if model == nil || len(model.Faces) == 0 {
		t.Fatal("CenteredGlyphBoard() returned empty model")
	}
	minX, maxX := 1.0, -1.0
	for _, face := range model.Faces {
		for _, vertex := range face.Vertices {
			if vertex.X < minX {
				minX = vertex.X
			}
			if vertex.X > maxX {
				maxX = vertex.X
			}
		}
	}
	center := (minX + maxX) / 2
	if center < -0.02 || center > 0.02 {
		t.Fatalf("model center X = %f, want near 0", center)
	}
}

// TestNormalizeShowcaseAssetModelCentersGeometry verifies the helper recenters and scales models.
func TestNormalizeShowcaseAssetModelCentersGeometry(t *testing.T) {
	model := threed.CreateCube(threed.Vector3D{X: 10, Y: 5, Z: -3}, 8, '█')
	normalizeShowcaseAssetModel(model)

	center := model.Center()
	if center.X < -0.6 || center.X > 0.6 {
		t.Fatalf("center.X = %f, want near 0", center.X)
	}
	if center.Z < -0.6 || center.Z > 0.6 {
		t.Fatalf("center.Z = %f, want near 0", center.Z)
	}
}

// TestMergeModelsCountsFaces verifies model composition preserves all faces.
func TestMergeModelsCountsFaces(t *testing.T) {
	left := threed.CreateCube(threed.Vector3D{}, 1, '█')
	right := threed.CreateCube(threed.Vector3D{X: 2}, 1, '█')
	merged := mergeModels(left, nil, right)
	if got, want := len(merged.Faces), len(left.Faces)+len(right.Faces); got != want {
		t.Fatalf("len(merged.Faces) = %d, want %d", got, want)
	}
}
