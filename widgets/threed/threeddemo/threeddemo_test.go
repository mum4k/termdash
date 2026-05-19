package main

import (
	"strings"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/threed"
)

// TestSceneCatalogTextHighlightsActive verifies the scene list emphasizes the active scene.
func TestSceneCatalogTextHighlightsActive(t *testing.T) {
	lines := sceneCatalogText(threed.BuildShowcaseScenes(false), 1, false)
	if got := lines[1].line; got != "▶ 2. Shape Board" {
		t.Fatalf("active line = %q, want highlighted Shape Board", got)
	}
	if got := lines[1].color; got != cell.ColorNumber(159) {
		t.Fatalf("active line color = %v, want %v", got, cell.ColorNumber(159))
	}
}

// TestSceneDetailsTextIncludesFeatures verifies the details panel copy exposes metadata.
func TestSceneDetailsTextIncludesFeatures(t *testing.T) {
	scene := threed.BuildShowcaseScenes(true)[3]
	details := sceneDetailsText(scene, 12, true, "/tmp/example.png")
	if !strings.Contains(details, "Scene: Image Relief") {
		t.Fatalf("details = %q, want scene heading", details)
	}
	if !strings.Contains(details, "image extrusion via LoadImageModel") {
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

func TestControlSummaryLinesIncludesNineSelections(t *testing.T) {
	lines := controlSummaryLines(false)
	if got, want := lines[0].value, "1-9 scene select"; got != want {
		t.Fatalf("keys summary = %q, want %q", got, want)
	}
}
