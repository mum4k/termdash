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
	"math"
	"sync"
	"testing"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

func TestShapeConstructors(t *testing.T) {
	tests := []struct {
		name      string
		model     *Model
		wantFaces int
	}{
		{name: "cube", model: CreateCube(Vector3D{}, 1, '#'), wantFaces: 6},
		{name: "pyramid", model: CreatePyramid(Vector3D{}, 1, '^'), wantFaces: 5},
		{name: "sphere", model: CreateSphere(Vector3D{}, 1, 8, 12, 'o'), wantFaces: 96},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(tc.model.Faces); got != tc.wantFaces {
				t.Fatalf("len(Faces) = %d, want %d", got, tc.wantFaces)
			}
			for i, face := range tc.model.Faces {
				if len(face.Vertices) < 3 {
					t.Fatalf("face %d has %d vertices, want at least 3", i, len(face.Vertices))
				}
			}
		})
	}
}

func TestPyramidAndSphereNormalsPointOutward(t *testing.T) {
	tests := []struct {
		name   string
		center Vector3D
		model  *Model
	}{
		{name: "pyramid", center: Vector3D{}, model: CreatePyramid(Vector3D{}, 1, '^')},
		{name: "sphere", center: Vector3D{}, model: CreateSphere(Vector3D{}, 1, 8, 12, 'o')},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for i, face := range tc.model.Faces {
				faceCenter := centerOfVertices(face.Vertices)
				outward := faceCenter.Subtract(tc.center)
				if got := computeFaceNormal(face.Vertices).Dot(outward); got <= 0 {
					t.Fatalf("face %d normal points inward: normal dot outward = %f", i, got)
				}
			}
		})
	}
}

func TestNormalizeReturnsUnitVector(t *testing.T) {
	v := Vector3D{X: 3, Y: 4, Z: 12}.Normalize()
	got := math.Sqrt(v.Dot(v))
	if math.Abs(got-1) > 1e-12 {
		t.Fatalf("normalized length = %f, want 1", got)
	}
}

func TestModelSetColor(t *testing.T) {
	model := CreateCube(Vector3D{}, 1, '#')
	color := Color{R: 1, G: 0.25, B: 0.5}

	model.SetColor(color)

	for i, face := range model.Faces {
		if !face.HasColor {
			t.Fatalf("face %d HasColor = false, want true", i)
		}
		if face.Color != color {
			t.Fatalf("face %d Color = %+v, want %+v", i, face.Color, color)
		}
	}
}

func TestCalculatePhongShadingPreservesPreferredChar(t *testing.T) {
	opts := defaultOptions()
	_, _, got := calculatePhongShading(
		Vector3D{X: 0, Y: 0, Z: -1},
		Vector3D{X: 0, Y: 0, Z: -1},
		Color{R: 0.5, G: 0.7, B: 1.0},
		opts,
		'✈',
	)
	if got != '✈' {
		t.Fatalf("calculatePhongShading() char = %q, want %q", got, '✈')
	}
}

func TestDrawGlyphFacePlacesRuneAtCenter(t *testing.T) {
	cvs, err := canvas.New(image.Rect(0, 0, 8, 8))
	if err != nil {
		t.Fatalf("canvas.New() => unexpected error: %v", err)
	}

	drawGlyphFace(cvs, []Vector2D{
		{X: 2, Y: 2},
		{X: 4, Y: 2},
		{X: 4, Y: 4},
		{X: 2, Y: 4},
	}, Color{R: 1, G: 1, B: 1}.ToCellColor(), '✈')

	got, err := cvs.Cell(image.Point{X: 3, Y: 3})
	if err != nil {
		t.Fatalf("cvs.Cell() => unexpected error: %v", err)
	}
	if got.Rune != '✈' {
		t.Fatalf("center rune = %q, want %q", got.Rune, '✈')
	}
}

func TestDrawUsesFaceLineColor(t *testing.T) {
	widget, err := New(
		ShowAxes(false),
		BackfaceCulling(false),
		AmbientColor(Color{R: 1, G: 1, B: 1}),
		DiffuseColor(Color{}),
		SpecularColor(Color{}),
	)
	if err != nil {
		t.Fatalf("New() => unexpected error: %v", err)
	}

	model := NewModel()
	model.AddFace(Face{
		Vertices: []Vector3D{
			{X: -1, Y: 0, Z: 0},
			{X: 1, Y: 0, Z: 0},
		},
		Char:     '=',
		Color:    Color{R: 0.85, G: 0.22, B: 0.18},
		HasColor: true,
	})
	widget.SetModel(model)

	cvs, err := canvas.New(image.Rect(0, 0, 20, 10))
	if err != nil {
		t.Fatalf("canvas.New() => unexpected error: %v", err)
	}
	if err := widget.Draw(cvs, nil); err != nil {
		t.Fatalf("Draw() => unexpected error: %v", err)
	}

	var found bool
	wantColor := Color{R: 0.85, G: 0.22, B: 0.18}.ToCellColor()
	for y := 0; y < cvs.Area().Dy() && !found; y++ {
		for x := 0; x < cvs.Area().Dx(); x++ {
			got, err := cvs.Cell(image.Point{X: x, Y: y})
			if err != nil {
				t.Fatalf("cvs.Cell() => unexpected error: %v", err)
			}
			if got.Rune == '=' {
				found = true
				if got.Opts == nil || got.Opts.FgColor != wantColor {
					t.Fatalf("line fg color = %v, want %v", got.Opts.FgColor, wantColor)
				}
				break
			}
		}
	}
	if !found {
		t.Fatal("rendered line rune not found on canvas")
	}
}

func TestSubcellSceneProducesBrailleDetail(t *testing.T) {
	cvs, err := canvas.New(image.Rect(0, 0, 4, 4))
	if err != nil {
		t.Fatalf("canvas.New() => unexpected error: %v", err)
	}

	scene := newSubcellScene(cvs.Area())
	scene.FillPolygon([]Vector2D{
		{X: 0.25, Y: 0.25},
		{X: 1.75, Y: 0.25},
		{X: 1.10, Y: 0.90},
		{X: 0.25, Y: 1.40},
	}, Color{R: 1, G: 0.8, B: 0.2}, 0)
	if err := scene.CopyTo(cvs); err != nil {
		t.Fatalf("scene.CopyTo() => unexpected error: %v", err)
	}

	var sawBraille bool
	for y := 0; y < cvs.Area().Dy(); y++ {
		for x := 0; x < cvs.Area().Dx(); x++ {
			got, err := cvs.Cell(image.Point{X: x, Y: y})
			if err != nil {
				t.Fatalf("cvs.Cell() => unexpected error: %v", err)
			}
			if got.Rune >= 0x2800 && got.Rune <= 0x28FF {
				sawBraille = true
				break
			}
		}
		if sawBraille {
			break
		}
	}
	if !sawBraille {
		t.Fatal("scene rendered no braille detail, want a braille-resolved fill")
	}
}

func TestPolygonCoverageSamplesPartialSubcell(t *testing.T) {
	halfCell := []Vector2D{
		{X: 0, Y: 0},
		{X: 0.5, Y: 0},
		{X: 0.5, Y: 1},
		{X: 0, Y: 1},
	}

	got := polygonCoverage(0, 0, halfCell, 2)
	if math.Abs(got-0.5) > 1e-12 {
		t.Fatalf("polygonCoverage() = %f, want 0.5", got)
	}
}

func TestSubcellCoveredKeepsDitheredEdgesLight(t *testing.T) {
	if !subcellCovered(0.25, 0, 0) {
		t.Fatal("subcellCovered() dropped a lightly covered low-threshold edge dot")
	}
	if subcellCovered(0.25, 1, 0) {
		t.Fatal("subcellCovered() kept every lightly covered edge dot, want ordered dither")
	}
	if !subcellCovered(0.5, 1, 0) {
		t.Fatal("subcellCovered() dropped a half-covered dot, want solid coverage to render")
	}
}

func TestSubcellSceneClearOnlyDirtyBounds(t *testing.T) {
	scene := newSubcellScene(image.Rect(0, 0, 8, 8))
	scene.FillPolygon([]Vector2D{
		{X: 1, Y: 1},
		{X: 5, Y: 1},
		{X: 5, Y: 5},
		{X: 1, Y: 5},
	}, Color{R: 1, G: 1, B: 1}, 0)
	if !scene.dirty {
		t.Fatal("FillPolygon did not mark dirty bounds")
	}
	scene.Clear()
	if scene.dirty {
		t.Fatal("Clear left dirty bounds active")
	}
	for i, filled := range scene.filled {
		if filled {
			t.Fatalf("filled[%d] = true after Clear, want false", i)
		}
	}
}

func TestSpectrumRailsDoNotCrossFilledBars(t *testing.T) {
	tests := []struct {
		name  string
		model *Model
	}{
		{name: "audio", model: SpectrumAnalyzer([]float64{1, 0.8, 0.5}, ModelSize(3))},
		{name: "network", model: NetworkSpectrum([]float64{1, 0.8, 0.5}, []float64{0.2, 0.3, 0.4}, ModelSize(3))},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			minFillY := math.Inf(1)
			type yRange struct {
				min float64
				max float64
			}
			var fillRanges []yRange
			var railYs []float64
			for _, face := range tc.model.Faces {
				if len(face.Vertices) == 2 {
					railYs = append(railYs, face.Vertices[0].Y)
					continue
				}
				if face.RenderMode != FaceRenderFill {
					continue
				}
				rng := yRange{min: math.Inf(1), max: math.Inf(-1)}
				for _, vertex := range face.Vertices {
					if vertex.Y < minFillY {
						minFillY = vertex.Y
					}
					if vertex.Y < rng.min {
						rng.min = vertex.Y
					}
					if vertex.Y > rng.max {
						rng.max = vertex.Y
					}
				}
				fillRanges = append(fillRanges, rng)
			}
			if len(railYs) == 0 {
				t.Fatal("spectrum model has no rail line faces")
			}
			if len(fillRanges) == 0 || math.IsInf(minFillY, 1) {
				t.Fatal("spectrum model has no filled bar faces")
			}
			for _, y := range railYs {
				for _, rng := range fillRanges {
					if y >= rng.min && y <= rng.max {
						t.Fatalf("rail y=%f crosses filled bar range [%f,%f]", y, rng.min, rng.max)
					}
				}
			}
		})
	}
}

func TestUprightOnlyLocksPitchAndRoll(t *testing.T) {
	widget, err := New(UprightOnly(true))
	if err != nil {
		t.Fatalf("New() => unexpected error: %v", err)
	}

	widget.Rotate(Vector3D{X: 1.2, Y: 0.5, Z: 0.8})
	if widget.rotation.X != 0 {
		t.Fatalf("rotation.X = %f, want 0 in upright mode", widget.rotation.X)
	}
	if widget.rotation.Z != 0 {
		t.Fatalf("rotation.Z = %f, want 0 in upright mode", widget.rotation.Z)
	}
	if widget.rotation.Y == 0 {
		t.Fatal("rotation.Y = 0, want Y rotation to remain active")
	}

	if err := widget.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowUp}, nil); err != nil {
		t.Fatalf("Keyboard(up) => unexpected error: %v", err)
	}
	if err := widget.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowRight}, nil); err != nil {
		t.Fatalf("Keyboard(right) => unexpected error: %v", err)
	}
	if widget.rotation.X != 0 {
		t.Fatalf("rotation.X after keyboard = %f, want 0 in upright mode", widget.rotation.X)
	}
	if widget.rotation.Z != 0 {
		t.Fatalf("rotation.Z after keyboard = %f, want 0 in upright mode", widget.rotation.Z)
	}
}

// TestThreeDConcurrentDrawAndRotate exercises the widget under concurrent
// Draw, Rotate, SetModel, Keyboard, and Mouse calls. Run with -race to detect
// data races; the test will also fail on any panics caused by torn reads.
func TestThreeDConcurrentDrawAndRotate(t *testing.T) {
	widget, err := New()
	if err != nil {
		t.Fatalf("New() => unexpected error: %v", err)
	}
	widget.SetModel(CreateCube(Vector3D{}, 1, '#'))

	cvs, err := canvas.New(image.Rect(0, 0, 20, 20))
	if err != nil {
		t.Fatalf("canvas.New() => unexpected error: %v", err)
	}

	const iters = 50
	var wg sync.WaitGroup

	// Goroutine 1–2: Draw.
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				if err := widget.Draw(cvs, nil); err != nil {
					t.Errorf("Draw() error: %v", err)
					return
				}
			}
		}()
	}

	// Goroutine 3–4: Rotate.
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				widget.Rotate(Vector3D{Y: 0.1})
			}
		}()
	}

	// Goroutine 5: SetModel.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < iters; j++ {
			widget.SetModel(CreateSphere(Vector3D{}, 1, 6, 8, 'o'))
		}
	}()

	// Goroutine 6: Keyboard events.
	wg.Add(1)
	go func() {
		defer wg.Done()
		keys := []keyboard.Key{
			keyboard.KeyArrowUp, keyboard.KeyArrowDown,
			keyboard.KeyArrowLeft, keyboard.KeyArrowRight,
		}
		for j := 0; j < iters; j++ {
			k := &terminalapi.Keyboard{Key: keys[j%len(keys)]}
			if err := widget.Keyboard(k, nil); err != nil {
				t.Errorf("Keyboard() error: %v", err)
				return
			}
		}
	}()

	// Goroutine 7–8: Mouse scroll events.
	for i := 0; i < 2; i++ {
		btn := mouse.ButtonWheelUp
		if i == 1 {
			btn = mouse.ButtonWheelDown
		}
		wg.Add(1)
		go func(b mouse.Button) {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				m := &terminalapi.Mouse{Button: b}
				if err := widget.Mouse(m, nil); err != nil {
					t.Errorf("Mouse() error: %v", err)
					return
				}
			}
		}(btn)
	}

	wg.Wait()
}

func centerOfVertices(vertices []Vector3D) Vector3D {
	var center Vector3D
	for _, vertex := range vertices {
		center = center.Add(vertex)
	}
	scale := 1 / float64(len(vertices))
	return center.Multiply(scale)
}
