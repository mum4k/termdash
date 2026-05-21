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

import "testing"

func TestRenderableRune(t *testing.T) {
	tests := []struct {
		name     string
		frame    string
		fallback rune
		want     rune
	}{
		{name: "single cell", frame: "✶", fallback: '*', want: '✶'},
		{name: "variation selector frame", frame: "✈️", fallback: '*', want: '✈'},
		{name: "wide emoji survives", frame: "👌", fallback: '*', want: '👌'},
		{name: "empty falls back", frame: "", fallback: '*', want: '*'},
	}

	for _, tc := range tests {
		if got := RenderableRune(tc.frame, tc.fallback); got != tc.want {
			t.Fatalf("%s: RenderableRune(%q, %q) = %q, want %q", tc.name, tc.frame, tc.fallback, got, tc.want)
		}
	}
}

func TestNewAnimatedSpinnerStarPrism(t *testing.T) {
	model := NewAnimatedSpinnerStarPrism("✈️", 3)
	if model == nil {
		t.Fatal("NewAnimatedSpinnerStarPrism() => nil model")
	}
	if got, wantMin := len(model.Faces), 52; got < wantMin {
		t.Fatalf("len(model.Faces) = %d, want at least %d", got, wantMin)
	}
	if got, want := len(model.Faces[0].Vertices), 50; got != want {
		t.Fatalf("len(model.Faces[0].Vertices) = %d, want %d", got, want)
	}
	if got, want := model.Faces[0].Char, '✈'; got != want {
		t.Fatalf("model.Faces[0].Char = %q, want %q", got, want)
	}
	for i, face := range model.Faces {
		if len(face.Vertices) < 3 {
			t.Fatalf("face %d has %d vertices, want at least 3", i, len(face.Vertices))
		}
		if !face.HasColor {
			t.Fatalf("face %d should carry color", i)
		}
	}
}

func TestNewAnimatedSymbolSpinner(t *testing.T) {
	model := NewAnimatedSymbolSpinner("A", 2)
	if model == nil {
		t.Fatal("NewAnimatedSymbolSpinner() => nil model")
	}
	if got, want := len(model.Faces), 2; got != want {
		t.Fatalf("len(model.Faces) = %d, want %d", got, want)
	}
	for i, face := range model.Faces {
		if got, want := face.Char, 'A'; got != want {
			t.Fatalf("face %d char = %q, want %q", i, got, want)
		}
		if got, want := face.RenderMode, FaceRenderGlyph; got != want {
			t.Fatalf("face %d RenderMode = %v, want %v", i, got, want)
		}
		if len(face.Vertices) != 4 {
			t.Fatalf("face %d has %d vertices, want 4", i, len(face.Vertices))
		}
	}
	frontNormal := computeFaceNormal(model.Faces[0].Vertices)
	backNormal := computeFaceNormal(model.Faces[1].Vertices)
	if frontNormal.Dot(backNormal) > -0.99 {
		t.Fatalf("billboard normals should oppose each other, got dot=%f", frontNormal.Dot(backNormal))
	}
}

func TestNewAnimatedSymbolSpinnerEmojiAsset(t *testing.T) {
	model := NewAnimatedSymbolSpinner("👌", 2)
	if model == nil {
		t.Fatal("NewAnimatedSymbolSpinner() => nil model")
	}
	if got := len(model.Faces); got <= 300 {
		t.Fatalf("len(model.Faces) = %d, want high-resolution asset-backed extrusion", got)
	}
	if got, want := model.Faces[0].RenderMode, FaceRenderFill; got != want {
		t.Fatalf("model.Faces[0].RenderMode = %v, want %v", got, want)
	}
	if got, want := model.Faces[0].Char, '█'; got != want {
		t.Fatalf("model.Faces[0].Char = %q, want %q", got, want)
	}
	if clr := model.Faces[0].Color; !(clr.R > clr.B && clr.G > clr.B) {
		t.Fatalf("front face color = %+v, want warm source-derived color", clr)
	}
}

func TestBundledEmojiAssetName(t *testing.T) {
	tests := []struct {
		frame string
		want  string
	}{
		{frame: "👌", want: "1F44C"},
		{frame: "✈️", want: "2708"},
		{frame: "", want: ""},
	}
	for _, tc := range tests {
		if got := bundledEmojiAssetName(tc.frame); got != tc.want {
			t.Fatalf("bundledEmojiAssetName(%q) = %q, want %q", tc.frame, got, tc.want)
		}
	}
}

func TestShouldFillFaceBackground(t *testing.T) {
	if !shouldFillFaceBackground('█') {
		t.Fatal("shouldFillFaceBackground('█') = false, want true")
	}
	if shouldFillFaceBackground('👌') {
		t.Fatal("shouldFillFaceBackground('👌') = true, want false")
	}
}

func TestSpinnerStarColorsStayInRange(t *testing.T) {
	for step := 0; step < 12; step++ {
		front, side, back := spinnerStarColors(step)
		for _, clr := range []Color{front, side, back} {
			for _, component := range []float64{clr.R, clr.G, clr.B} {
				if component < 0 || component > 1 {
					t.Fatalf("spinnerStarColors(%d) produced out-of-range component %.3f", step, component)
				}
			}
		}
	}
}

func TestDensifyClosedPath(t *testing.T) {
	base := []Vector3D{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
	}
	refined := densifyClosedPath(base, 5)
	if got, want := len(refined), len(base)*5; got != want {
		t.Fatalf("len(refined) = %d, want %d", got, want)
	}
	if refined[0] != base[0] {
		t.Fatalf("refined[0] = %+v, want %+v", refined[0], base[0])
	}
}
