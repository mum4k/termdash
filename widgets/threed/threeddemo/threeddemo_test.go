package main

import (
	"strings"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/threed"
)

// TestBuildDemoScenes verifies that the showcase exposes the expected scene set.
func TestBuildDemoScenes(t *testing.T) {
	scenes := buildDemoScenes(false)
	if got, want := len(scenes), sceneCount; got != want {
		t.Fatalf("len(buildDemoScenes(false)) = %d, want %d", got, want)
	}
	if got, want := scenes[0].Name, "Terminal Forms"; got != want {
		t.Fatalf("scenes[0].Name = %q, want %q", got, want)
	}

	withAsset := buildDemoScenes(true)
	if got, want := withAsset[3].Name, "Image Relief"; got != want {
		t.Fatalf("asset scene name = %q, want %q", got, want)
	}
}

// TestCircuitScenesIncludeGlyphForms verifies the reference-inspired scenes use
// glyph-rendered shapes in addition to filled meshes.
func TestCircuitScenesIncludeGlyphForms(t *testing.T) {
	for _, scene := range buildDemoScenes(false)[:2] {
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

// TestSceneBuildersReturnFaces verifies that every curated scene produces geometry.
func TestSceneBuildersReturnFaces(t *testing.T) {
	scenes := buildDemoScenes(false)
	for _, scene := range scenes {
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

// TestSceneCatalogTextHighlightsActive verifies the scene list emphasizes the active scene.
func TestSceneCatalogTextHighlightsActive(t *testing.T) {
	lines := sceneCatalogText(buildDemoScenes(false), 1, false)
	if got := lines[1].line; got != "▶ 2. Shape Board" {
		t.Fatalf("active line = %q, want highlighted Shape Board", got)
	}
	if got := lines[1].color; got != cell.ColorNumber(159) {
		t.Fatalf("active line color = %v, want %v", got, cell.ColorNumber(159))
	}
}

// TestSceneDetailsTextIncludesFeatures verifies the details panel copy exposes metadata.
func TestSceneDetailsTextIncludesFeatures(t *testing.T) {
	scene := buildDemoScenes(true)[3]
	details := sceneDetailsText(scene, 12, true, "/tmp/example.png")
	if !strings.Contains(details, "Scene: Image Relief") {
		t.Fatalf("details = %q, want scene heading", details)
	}
	if !strings.Contains(details, "image extrusion via threed.LoadImageModel") {
		t.Fatalf("details = %q, want feature list", details)
	}
	if !strings.Contains(details, "Asset file: loaded") {
		t.Fatalf("details = %q, want asset status", details)
	}
}

// TestControlSummaryLinesIncludesAssetState verifies the helper reflects asset availability.
func TestControlSummaryLinesIncludesAssetState(t *testing.T) {
	lines := controlSummaryLines(true)
	last := lines[len(lines)-1]
	if got, want := last.label, "Asset"; got != want {
		t.Fatalf("last.label = %q, want %q", got, want)
	}
	if got, want := last.value, "image relief online"; got != want {
		t.Fatalf("last.value = %q, want %q", got, want)
	}
}

// TestNormalizeAssetModelCentersGeometry verifies the helper recenters and scales models.
func TestNormalizeAssetModelCentersGeometry(t *testing.T) {
	model := threed.CreateCube(threed.Vector3D{X: 10, Y: 5, Z: -3}, 8, '█')
	normalizeAssetModel(model)

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
