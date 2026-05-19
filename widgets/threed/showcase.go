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
	"math"
	"strings"
)

// ShowcaseSceneCount is the number of curated showcase scenes.
const ShowcaseSceneCount = 9

// ShowcaseScene describes one curated ThreeD showcase scene.
type ShowcaseScene struct {
	Name     string
	Summary  string
	Features []string
	Orbit    Vector3D
	Build    func(step int, asset *Model) *Model
}

// BuildShowcaseScenes returns the curated 3D showcase scenes.
func BuildShowcaseScenes(assetLoaded bool) []ShowcaseScene {
	assetName := "Prism Field"
	assetSummary := "A full-width prism-board fallback with stacked forms, clean graph rails, and labeled terminal modules."
	assetFeatures := []string{
		"large readable prism forms",
		"clean waveform and rail graphs",
		"sharp terminal-native glyph layout",
	}
	if assetLoaded {
		assetName = "Image Relief"
		assetSummary = "Asset-backed extrusion on a clean display pedestal."
		assetFeatures = []string{
			"image extrusion via LoadImageModel",
			"source-derived face colors",
			"same lighting stack as procedural geometry",
		}
	}

	return []ShowcaseScene{
		{
			Name:    "Terminal Forms",
			Summary: "Sharp box-drawing forms, chips, nodes, and traces arranged as terminal-native geometry.",
			Features: []string{
				"crisp box-drawing geometry",
				"multiple high-contrast shape families",
				"explicit title-like glyph placement",
			},
			Orbit: Vector3D{},
			Build: buildCircuitBloomScene,
		},
		{
			Name:    "Shape Board",
			Summary: "A compact board of cubes, pyramids, diamonds, rings, bars, and node networks.",
			Features: []string{
				"large readable terminal forms",
				"wireframe and solid glyph shapes",
				"tight cyan, green, amber, and rose accents",
			},
			Orbit: Vector3D{},
			Build: buildFormCatalogScene,
		},
		{
			Name:    "Signal Lattice",
			Summary: "Clean terminal charts: bars, traces, packets, rails, and a compact signal matrix.",
			Features: []string{
				"readable bar and trace charts",
				"clean packet rails",
				"compact signal matrix",
			},
			Orbit: Vector3D{},
			Build: buildSignalRigScene,
		},
		{
			Name:     assetName,
			Summary:  assetSummary,
			Features: assetFeatures,
			Orbit:    Vector3D{},
			Build:    buildAssetReliefScene,
		},
		{
			Name:    "Cube Core",
			Summary: "A shaded cube with a clean technical plinth and orbit rails.",
			Features: []string{
				"large solid cube primitive",
				"cyan face lighting",
				"simple reusable Cube API path",
			},
			Orbit: Vector3D{Y: 0.050},
			Build: buildCubeCoreScene,
		},
		{
			Name:    "Pyramid Spire",
			Summary: "A crisp pyramid primitive staged over a minimal base grid.",
			Features: []string{
				"square pyramid geometry",
				"warm amber face palette",
				"clean pedestal composition",
			},
			Orbit: Vector3D{Y: 0.045},
			Build: buildPyramidSpireScene,
		},
		{
			Name:    "Sphere Shell",
			Summary: "A low-poly sphere showing smooth shaded terminal fill.",
			Features: []string{
				"segmented sphere primitive",
				"soft white surface shading",
				"animated inspection orbit",
			},
			Orbit: Vector3D{Y: 0.040},
			Build: buildSphereShellScene,
		},
		{
			Name:    "Octa Prism",
			Summary: "An octahedron primitive with a tight ring accent and facet contrast.",
			Features: []string{
				"diamond-like octahedron geometry",
				"rose highlight facets",
				"compact ring-band accent",
			},
			Orbit: Vector3D{Y: 0.052},
			Build: buildOctaPrismScene,
		},
		{
			Name:    "Tetra Field",
			Summary: "A tetrahedron field showing multiple primitive instances in one model.",
			Features: []string{
				"multi-shape composition",
				"reusable model append path",
				"balanced shape spacing",
			},
			Orbit: Vector3D{Y: 0.048},
			Build: buildTetraFieldScene,
		},
	}
}

// buildCircuitBloomScene assembles the main reference-inspired circuit scene.
func buildCircuitBloomScene(step int, _ *Model) *Model {
	rows := []string{
		"╭─────────────────────────╮     ◆   ◆   ◆       ┌───────┐",
		"│  ┌──────────┐   ╱╲      │   ◆───◆───◆───◆     │ █ █ █ │",
		"│  │  CUBE    │  ╱  ╲     │   │   │   │   │     │ █ █ █ │",
		"│  │  ┌────┐  │ ╱────╲    │   ◆───◆───◆───◆     └───┬───┘",
		"│  │  │    │  │ ╲    ╱    │       │   │               │",
		"│  │  └────┘  │  ╲  ╱     │   ┌───┘   └───────┐   ┌──┴──┐",
		"│  └──────────┘   ╲╱      │   │  ┌─────┐      │   │ NODE │",
		"╰───────────────┬─────────╯   │  │ ▣ ▣ │  ◇   │   └─────┘",
		"                │             │  └──┬──┘ ◇◇◇  │",
		"  ○────○────○───┼───○────○    └─────┼────◇────┘",
		"  │    │    │   │   │    │          │",
		"  ○────○────○   ●   ○────○     ┌────┴────┐   ▲",
		"       │        │        │     │ BARS    │  ▲▲▲",
		"  ┌────┴────┐   │   ┌────┴──┐  │ █ █ █ █ │ ▲▲▲▲▲",
		"  │ TRACE   ├───┘   │ RING  │  │ █ █ █ █ │   │",
		"  └─────────┘       │ ○ ○ ○ │  └─────────┘   │",
		"                    └───────┘        ────────┘",
	}
	halo := []string{
		"    ·       ·          ·              ·          ·",
		"       ·          ·          ·             ·",
		"  ·          ·              ·        ·           ·",
	}
	return mergeModels(
		createGlyphRows(Vector3D{X: -1.72, Y: 1.65, Z: -0.22}, rows, 0.038, 0.20, terminalFormColor),
		createGlyphRows(Vector3D{X: -1.70, Y: 1.83, Z: -0.40}, halo, 0.038, 0.20, mutedCircuitColor),
		createPulseGlyphs(step),
	)
}

// buildFormCatalogScene assembles the mixed primitive and glyph form lineup.
func buildFormCatalogScene(step int, _ *Model) *Model {
	rows := []string{
		"┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐",
		"│ WIREFRAME    │  │ PYRAMID      │  │ DIAMOND      │  │ RING ARRAY   │",
		"│    ┌────┐    │  │      ▲       │  │      ◆       │  │   ○ ○ ○      │",
		"│ ┌──┼────┼──┐ │  │     ▲▲▲      │  │     ◆ ◆      │  │ ○       ○    │",
		"│ │  └────┘  │ │  │    ▲▲▲▲▲     │  │    ◆   ◆     │  │ ○       ○    │",
		"│ └──────────┘ │  │   ▲▲▲▲▲▲▲    │  │     ◆ ◆      │  │   ○ ○ ○      │",
		"└──────────────┘  └──────┬───────┘  └──────◆───────┘  └──────┬───────┘",
		"                         │                       │            │",
		"┌──────────────┐  ┌──────┴───────┐  ┌────────────┴─┐  ┌──────┴───────┐",
		"│ BAR BLOCKS   │  │ NODE MESH    │  │ CHIP STRIP   │  │ TRACE BUS    │",
		"│ █ █ █ █ █ █  │  │ ◆──◆──◆──◆   │  │ ▣ ▣ ▣ ▣ ▣    │  │ ═══╦═══╦═══  │",
		"│ █ █ █ █ █ █  │  │ │  │  │  │   │  │ ├─┬─┬─┬─┤    │  │    ║   ║     │",
		"│ █ █ █ █ █ █  │  │ ◆──◆──◆──◆   │  │ ▣ ▣ ▣ ▣ ▣    │  │ ═══╩═══╩═══  │",
		"└──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘",
	}
	return mergeModels(
		createGlyphRows(Vector3D{X: -1.78, Y: 1.48, Z: -0.22}, rows, 0.036, 0.21, terminalFormColor),
		createPulseGlyphs(step+8),
	)
}

// buildSignalRigScene assembles the telemetry-inspired model stack.
func buildSignalRigScene(step int, _ *Model) *Model {
	rows := signalRigRows(step)
	return createCenteredGlyphRows(Vector3D{X: 0.02, Y: 1.54, Z: -0.22}, rows, 0.043, 0.22, graphColor)
}

func signalRigRows(step int) []string {
	frame := step % 6
	pulse := []string{"▁", "▂", "▃", "▄", "▅", "▆"}[frame]
	return normalizeTerminalBoardRows([]string{
		"╭──────────────────────────── SIGNAL LATTICE ───────────────────────────╮",
		"│ AMPLITUDE                         PACKET RAIL                         │",
		"│ ████▇▇▆▆▅▅▄▄▃▃▂▂▁▁                ═══╦══════╦══════╦══════╦═══       │",
		"│ ███▇▇▆▆▅▅▄▄▃▃▂▂▁                  ║      ║      ║      ║           │",
		"│ ██▆▆▅▅▄▄▃▃▂▂▁                    ═╩══════╩══════╩══════╩═══       │",
		"│                                                                      │",
		"│ TRACE A  ─────╮   ╭────────╮      ╭──────╮       ╭────────────       │",
		"│               ╰───╯        ╰──────╯      ╰───────╯                   │",
		"│ TRACE B  ────────────╮      ╭──────────────╮       ╭──────           │",
		"│                      ╰──────╯              ╰───────╯                 │",
		"│                                                                      │",
		"│ MATRIX   ◆──◆──◆──◆──◆        BINS  " + pulse + "  ███ ███  ██  ████  ██  ███ │",
		"│          │  │  │  │  │              ▁▂▃▄▅▆▇█▇▆▅▄▃▂▁                 │",
		"│          ◆──◆──◆──◆──◆              ─────┬─────┬─────┬─────          │",
		"╰──────────────────────────────────────────────────────────────────────╯",
	})
}

// buildAssetReliefScene assembles the asset-backed or procedural fallback scene.
func buildAssetReliefScene(step int, asset *Model) *Model {
	if asset != nil {
		return buildImageReliefScene(step, asset)
	}
	return buildPrismFieldScene(step)
}

// buildImageReliefScene places the optional image model on a display pedestal.
func buildImageReliefScene(step int, asset *Model) *Model {
	base := createBoxModel(Vector3D{X: 0, Y: -1.20, Z: 0}, 8.2, 0.18, 4.6, Color{R: 0.14, G: 0.18, B: 0.25})
	plinth := createBoxModel(Vector3D{X: 0, Y: -0.92, Z: 0}, 4.6, 0.22, 2.6, Color{R: 0.26, G: 0.34, B: 0.46})
	grid := createGridModel(-1.08, 7.8, 4.0, 14, 6)
	grid.SetColor(Color{R: 0.22, G: 0.32, B: 0.48})
	model := cloneModel(asset)
	normalizeShowcaseAssetModel(model)
	spinAccent := createRingBand(Vector3D{X: 0, Y: -0.55, Z: 0}, 1.55, 1.78, 0.10, 24, Color{R: 0.95, G: 0.76, B: 0.25})
	shiftDynamicZ(model, 0.10*math.Sin(float64(step)*0.08))
	return mergeModels(base, plinth, grid, spinAccent, model)
}

// buildPrismFieldScene creates a procedural fallback scene when no asset is present.
func buildPrismFieldScene(step int) *Model {
	rows := prismFieldRows(step)
	return createCenteredGlyphRows(Vector3D{X: 0.02, Y: 1.58, Z: -0.22}, rows, 0.042, 0.22, graphColor)
}

func buildCubeCoreScene(step int, _ *Model) *Model {
	cube := Cube(ModelSize(1.55), ModelRune('█'), ModelColor(NeonCyan))
	accent := Cube(
		ModelSize(0.38+0.04*math.Sin(float64(step)*0.08)),
		ModelPosition(Vector3D{X: 1.25, Y: 0.80, Z: -0.18}),
		ModelRune('█'),
		ModelColor(SoftWhite),
	)
	return mergeModels(cube, accent)
}

func buildPyramidSpireScene(step int, _ *Model) *Model {
	pyramid := Pyramid(ModelSize(1.85), ModelRune('▲'), ModelColor(Amber))
	capstone := Tetrahedron(
		ModelSize(0.44+0.03*math.Sin(float64(step)*0.10)),
		ModelPosition(Vector3D{X: 0, Y: 1.18, Z: 0}),
		ModelColor(SoftWhite),
	)
	return mergeModels(pyramid, capstone)
}

func buildSphereShellScene(step int, _ *Model) *Model {
	sphere := Sphere(ModelSize(1.18), ModelSegments(10, 16), ModelRune('█'), ModelColor(SoftWhite))
	inner := Sphere(
		ModelSize(0.46+0.04*math.Sin(float64(step)*0.07)),
		ModelSegments(6, 10),
		ModelRune('▓'),
		ModelColor(NeonCyan),
	)
	return mergeModels(sphere, inner)
}

func buildOctaPrismScene(step int, _ *Model) *Model {
	octa := Octahedron(ModelSize(1.72), ModelColor(Rose))
	core := Cube(
		ModelSize(0.36+0.04*math.Sin(float64(step)*0.09)),
		ModelRune('█'),
		ModelColor(SoftWhite),
	)
	return mergeModels(octa, core)
}

func buildTetraFieldScene(step int, _ *Model) *Model {
	left := Tetrahedron(ModelSize(1.05), ModelPosition(Vector3D{X: -1.05, Y: 0.02, Z: 0}), ModelColor(NeonGreen))
	center := Tetrahedron(ModelSize(1.28), ModelPosition(Vector3D{X: 0, Y: 0.26, Z: -0.10}), ModelColor(NeonCyan))
	right := Tetrahedron(ModelSize(0.95), ModelPosition(Vector3D{X: 1.10, Y: -0.04, Z: 0.04}), ModelColor(Amber))
	node := Sphere(
		ModelSize(0.22+0.03*math.Sin(float64(step)*0.12)),
		ModelSegments(4, 8),
		ModelRune('●'),
		ModelColor(SoftWhite),
	)
	return mergeModels(left, center, right, node)
}

func prismFieldRows(step int) []string {
	frame := step % 6
	bar := []string{"▁", "▂", "▄", "▆", "█", "▇"}[frame]
	rise := []string{"▁▂▄▆█▇▅▃▂▁", "▂▄▆█▇▅▃▂▁▂", "▄▆█▇▅▃▂▁▂▄", "▆█▇▅▃▂▁▂▄▆", "█▇▅▃▂▁▂▄▆█", "▇▅▃▂▁▂▄▆█▇"}[frame]
	return normalizeTerminalBoardRows([]string{
		"╭────────────────────────────── PRISM FIELD ──────────────────────────────╮",
		"│ PRISM STACKS                         PROFILE / WAVEFORM                 │",
		"│   ╭────╮   ╭────╮   ╭────╮            ╭────────────────────────────╮     │",
		"│   │ ██ │   │ ██ │   │ █  │            │ " + rise + "  " + bar + "  " + rise + " │     │",
		"│   │ ██ │   │ ██ │   │ ██ │            ╰──────────────┬─────────────╯     │",
		"│   ╰─┬──╯   ╰─┬──╯   ╰─┬──╯                           │                   │",
		"│     │        │        │                ◆────◆────◆────◆────◆              │",
		"│ ╭───┴────────┴────────┴───╮            │    │    │    │    │              │",
		"│ │  " + bar + "  █  ▇  ▆  ▄  ▂  ▁  █  " + bar + "  │            ◆────◆────◆────◆────◆              │",
		"│ ╰─────────────────────────╯                                                  │",
		"│                                                                              │",
		"│ RAIL A  ═════╦═════╦═════╦═════╦═════╦═════╦═════       NODE BUS             │",
		"│ RAIL B  ─────┴─────┴─────┴─────┴─────┴─────┴─────   ○──○──○──○──○           │",
		"│ DEPTH   ▁▁▂▂▄▄▆▆████▆▆▄▄▂▂▁▁        CROSS  ◇──◇──◇──◇                       │",
		"╰──────────────────────────────────────────────────────────────────────────────╯",
	})
}

func normalizeTerminalBoardRows(rows []string) []string {
	maxWidth := 0
	for _, row := range rows {
		if width := len([]rune(row)); width > maxWidth {
			maxWidth = width
		}
	}
	if maxWidth == 0 {
		return rows
	}
	out := make([]string, len(rows))
	for i, row := range rows {
		runes := []rune(row)
		pad := maxWidth - len(runes)
		if pad <= 0 {
			out[i] = row
			continue
		}
		fill := strings.Repeat(" ", pad)
		if len(runes) > 0 && strings.ContainsRune("╮╯┐┘", runes[len(runes)-1]) {
			fill = strings.Repeat("─", pad)
		}
		if len(runes) > 1 && strings.ContainsRune("│╮╯┐┘", runes[len(runes)-1]) {
			out[i] = string(runes[:len(runes)-1]) + fill + string(runes[len(runes)-1])
			continue
		}
		out[i] = row + fill
	}
	return out
}

// createBoxModel creates a rectangular prism centered at the provided point.
func createBoxModel(center Vector3D, sx, sy, sz float64, color Color) *Model {
	hx, hy, hz := sx/2, sy/2, sz/2
	vertices := []Vector3D{
		{X: center.X - hx, Y: center.Y - hy, Z: center.Z - hz},
		{X: center.X + hx, Y: center.Y - hy, Z: center.Z - hz},
		{X: center.X + hx, Y: center.Y + hy, Z: center.Z - hz},
		{X: center.X - hx, Y: center.Y + hy, Z: center.Z - hz},
		{X: center.X - hx, Y: center.Y - hy, Z: center.Z + hz},
		{X: center.X + hx, Y: center.Y - hy, Z: center.Z + hz},
		{X: center.X + hx, Y: center.Y + hy, Z: center.Z + hz},
		{X: center.X - hx, Y: center.Y + hy, Z: center.Z + hz},
	}

	model := NewModel()
	faces := [][]int{
		{3, 2, 1, 0},
		{4, 5, 6, 7},
		{7, 3, 0, 4},
		{1, 2, 6, 5},
		{7, 6, 2, 3},
		{0, 1, 5, 4},
	}
	for _, idx := range faces {
		model.AddFace(Face{
			Vertices: []Vector3D{vertices[idx[0]], vertices[idx[1]], vertices[idx[2]], vertices[idx[3]]},
			Char:     '█',
			Color:    color,
			HasColor: true,
		})
	}
	return model
}

// createGridModel creates a wire grid in the XZ plane at the provided Y level.
func createGridModel(y, width, depth float64, cols, rows int) *Model {
	model := NewModel()
	left := -width / 2
	right := width / 2
	front := -depth / 2
	back := depth / 2

	for i := 0; i <= cols; i++ {
		t := float64(i) / math.Max(float64(cols), 1)
		x := left + t*width
		model.AddFace(Face{
			Vertices: []Vector3D{
				{X: x, Y: y, Z: front},
				{X: x, Y: y, Z: back},
			},
			Char: '─',
		})
	}
	for i := 0; i <= rows; i++ {
		t := float64(i) / math.Max(float64(rows), 1)
		z := front + t*depth
		model.AddFace(Face{
			Vertices: []Vector3D{
				{X: left, Y: y, Z: z},
				{X: right, Y: y, Z: z},
			},
			Char: '─',
		})
	}
	return model
}

func terminalFormColor(r rune) Color {
	switch {
	case strings.ContainsRune("┌┐└┘─│├┤┬┴┼╭╮╰╯╦╩═║", r):
		return Color{R: 0.54, G: 0.92, B: 0.96}
	case strings.ContainsRune("▲△╱╲", r):
		return Color{R: 0.96, G: 0.74, B: 0.22}
	case strings.ContainsRune("◆◇", r):
		return Color{R: 0.46, G: 0.94, B: 0.32}
	case strings.ContainsRune("○●", r):
		return Color{R: 0.66, G: 0.88, B: 1.00}
	case strings.ContainsRune("█▣", r):
		return Color{R: 0.96, G: 0.28, B: 0.46}
	default:
		return Color{R: 0.82, G: 0.86, B: 0.88}
	}
}

func mutedCircuitColor(r rune) Color {
	if r == '·' {
		return Color{R: 0.18, G: 0.28, B: 0.26}
	}
	return Color{R: 0.26, G: 0.34, B: 0.34}
}

func graphColor(r rune) Color {
	switch {
	case strings.ContainsRune("┌┐└┘─│├┤┬┴┼╭╮╰╯╦╩═║╔╗╚╝", r):
		return Color{R: 0.50, G: 0.88, B: 0.92}
	case strings.ContainsRune("█▇▆▅▄▃▂▁", r):
		return Color{R: 0.70, G: 0.92, B: 1.00}
	case strings.ContainsRune("◆◇", r):
		return Color{R: 0.48, G: 0.96, B: 0.34}
	case strings.ContainsRune("○●", r):
		return Color{R: 0.95, G: 0.78, B: 0.24}
	default:
		return Color{R: 0.84, G: 0.86, B: 0.88}
	}
}

func createPulseGlyphs(step int) *Model {
	model := NewModel()
	points := []Vector3D{
		{X: -1.05, Y: 1.42, Z: -0.34},
		{X: -0.72, Y: -0.30, Z: -0.34},
		{X: 0.58, Y: 1.18, Z: -0.34},
		{X: 1.12, Y: -1.18, Z: -0.34},
		{X: 0.06, Y: -1.48, Z: -0.34},
	}
	glyphs := []rune{'✦', '◆', '●', '✧', '◇'}
	for i, p := range points {
		if (step+i)%3 == 0 {
			addGlyphBillboard(model, p, 0.08, glyphs[i%len(glyphs)], Color{R: 0.98, G: 0.96, B: 0.42})
		}
	}
	return model
}

// createFrontGridModel creates a faint etched grid in the XY plane.
func createFrontGridModel(z, width, height float64, cols, rows int) *Model {
	model := NewModel()
	left := -width / 2
	right := width / 2
	bottom := -height / 2
	top := height / 2
	color := Color{R: 0.12, G: 0.20, B: 0.22}

	for i := 0; i <= cols; i++ {
		t := float64(i) / math.Max(float64(cols), 1)
		x := left + t*width
		addColoredLine(model, Vector3D{X: x, Y: bottom, Z: z}, Vector3D{X: x, Y: top, Z: z}, '│', color)
	}
	for i := 0; i <= rows; i++ {
		t := float64(i) / math.Max(float64(rows), 1)
		y := bottom + t*height
		addColoredLine(model, Vector3D{X: left, Y: y, Z: z}, Vector3D{X: right, Y: y, Z: z}, '─', color)
	}
	return model
}

// createCircuitCluster creates a compact field of terminal glyphs and traces.
func createCircuitCluster(center Vector3D, rows, cols int, spacing float64, primary, accent Color, step int) *Model {
	model := NewModel()
	glyphs := []rune{'┌', '┐', '└', '┘', '┼', '╷', '╵', '╶', '╴', '•', '◆', '▪', '█', '╋', '╂', '┬', '┴'}
	width := float64(cols-1) * spacing
	height := float64(rows-1) * spacing

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			index := (row*cols + col + step) % len(glyphs)
			x := center.X - width/2 + float64(col)*spacing
			y := center.Y - height/2 + float64(row)*spacing
			z := center.Z
			color := primary
			if (row+col+step)%4 == 0 {
				color = accent
			}
			if (row*2+col+step)%9 == 0 {
				addCircuitChip(model, Vector3D{X: x, Y: y, Z: z - 0.02}, spacing*0.54, spacing*0.36, color)
				continue
			}
			addGlyphBillboard(model, Vector3D{X: x, Y: y, Z: z - 0.04}, spacing*1.20, glyphs[index], color)
		}
	}

	for row := 0; row < rows; row++ {
		y := center.Y - height/2 + float64(row)*spacing
		if row%2 == 0 {
			addColoredLine(model, Vector3D{X: center.X - width/2, Y: y, Z: center.Z - 0.16}, Vector3D{X: center.X + width/2, Y: y, Z: center.Z - 0.16}, '─', primary)
		}
	}
	for col := 0; col < cols; col++ {
		x := center.X - width/2 + float64(col)*spacing
		if col%3 == 0 {
			addColoredLine(model, Vector3D{X: x, Y: center.Y - height/2, Z: center.Z - 0.18}, Vector3D{X: x, Y: center.Y + height/2, Z: center.Z - 0.18}, '│', accent)
		}
	}

	return model
}

// createChipBank creates a row of small solid modules like the blocks in the reference.
func createChipBank(center Vector3D, count int, spacing float64, color Color) *Model {
	model := NewModel()
	start := -float64(count-1) * spacing / 2
	for i := 0; i < count; i++ {
		x := center.X + start + float64(i)*spacing
		model = mergeModels(model, createBoxModel(Vector3D{X: x, Y: center.Y, Z: center.Z}, spacing*0.44, spacing*0.76, 0.18, color))
	}
	return model
}

func createTraceHeader(center Vector3D, width float64, color Color) *Model {
	model := NewModel()
	addColoredLine(model, Vector3D{X: center.X - width/2, Y: center.Y, Z: center.Z}, Vector3D{X: center.X + width/2, Y: center.Y, Z: center.Z}, '─', color)
	for i := 0; i < 7; i++ {
		x := center.X - width/2 + float64(i)*width/6
		addGlyphBillboard(model, Vector3D{X: x, Y: center.Y + 0.18*math.Sin(float64(i)), Z: center.Z - 0.03}, 0.20, []rune{'■', '┼', '◆', '●', '╋', '┤', '├'}[i], color)
	}
	return model
}

func addCircuitChip(model *Model, center Vector3D, width, height float64, color Color) {
	chip := createBoxModel(center, width, height, 0.12, color)
	for _, face := range chip.Faces {
		model.AddFace(face)
	}
}

func addGlyphBillboard(model *Model, center Vector3D, size float64, glyph rune, color Color) {
	half := size / 2
	model.AddFace(Face{
		Vertices: []Vector3D{
			{X: center.X - half, Y: center.Y - half, Z: center.Z},
			{X: center.X - half, Y: center.Y + half, Z: center.Z},
			{X: center.X + half, Y: center.Y + half, Z: center.Z},
			{X: center.X + half, Y: center.Y - half, Z: center.Z},
		},
		Char:       glyph,
		RenderMode: FaceRenderGlyph,
		Color:      color,
		HasColor:   true,
	})
}

// createRingBand creates a flat extruded ring in the XZ plane.
func createRingBand(center Vector3D, innerRadius, outerRadius, height float64, segments int, color Color) *Model {
	if segments < 8 {
		segments = 8
	}
	topY := center.Y + height/2
	bottomY := center.Y - height/2
	outerTop := make([]Vector3D, segments)
	outerBottom := make([]Vector3D, segments)
	innerTop := make([]Vector3D, segments)
	innerBottom := make([]Vector3D, segments)

	for i := 0; i < segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		cosA, sinA := math.Cos(angle), math.Sin(angle)
		outerTop[i] = Vector3D{X: center.X + cosA*outerRadius, Y: topY, Z: center.Z + sinA*outerRadius}
		outerBottom[i] = Vector3D{X: center.X + cosA*outerRadius, Y: bottomY, Z: center.Z + sinA*outerRadius}
		innerTop[i] = Vector3D{X: center.X + cosA*innerRadius, Y: topY, Z: center.Z + sinA*innerRadius}
		innerBottom[i] = Vector3D{X: center.X + cosA*innerRadius, Y: bottomY, Z: center.Z + sinA*innerRadius}
	}

	model := NewModel()
	for i := 0; i < segments; i++ {
		next := (i + 1) % segments
		model.AddFace(Face{
			Vertices: []Vector3D{outerTop[i], outerTop[next], innerTop[next], innerTop[i]},
			Char:     '█',
			Color:    color,
			HasColor: true,
		})
		model.AddFace(Face{
			Vertices: []Vector3D{outerBottom[next], outerBottom[i], innerBottom[i], innerBottom[next]},
			Char:     '█',
			Color:    color,
			HasColor: true,
		})
		model.AddFace(Face{
			Vertices: []Vector3D{outerTop[i], outerBottom[i], outerBottom[next], outerTop[next]},
			Char:     '█',
			Color:    color,
			HasColor: true,
		})
		model.AddFace(Face{
			Vertices: []Vector3D{innerTop[next], innerBottom[next], innerBottom[i], innerTop[i]},
			Char:     '█',
			Color:    color,
			HasColor: true,
		})
	}
	return model
}

// addLine adds a two-vertex face (line segment) to a model using the given character.
func addLine(model *Model, a, b Vector3D, char rune) {
	model.AddFace(Face{
		Vertices: []Vector3D{a, b},
		Char:     char,
	})
}

func addColoredLine(model *Model, a, b Vector3D, char rune, color Color) {
	model.AddFace(Face{
		Vertices: []Vector3D{a, b},
		Char:     char,
		Color:    color,
		HasColor: true,
	})
}

// mergeModels concatenates multiple models into one combined model.
func mergeModels(models ...*Model) *Model {
	merged := NewModel()
	for _, model := range models {
		if model == nil {
			continue
		}
		for _, face := range model.Faces {
			merged.AddFace(face)
		}
	}
	return merged
}

// cloneModel makes a deep copy of a model so scene builders can modify it safely.
func cloneModel(model *Model) *Model {
	if model == nil {
		return nil
	}
	cloned := NewModel()
	for _, face := range model.Faces {
		verts := make([]Vector3D, len(face.Vertices))
		copy(verts, face.Vertices)
		cloned.AddFace(Face{
			Vertices:   verts,
			Char:       face.Char,
			RenderMode: face.RenderMode,
			Color:      face.Color,
			HasColor:   face.HasColor,
			Normal:     face.Normal,
		})
	}
	return cloned
}

// normalizeShowcaseAssetModel centers and scales an asset model for the showcase stage.
func normalizeShowcaseAssetModel(model *Model) {
	if model == nil || len(model.Faces) == 0 {
		return
	}
	minX, minY, minZ := math.Inf(1), math.Inf(1), math.Inf(1)
	maxX, maxY, maxZ := math.Inf(-1), math.Inf(-1), math.Inf(-1)
	for _, face := range model.Faces {
		for _, vertex := range face.Vertices {
			if vertex.X < minX {
				minX = vertex.X
			}
			if vertex.X > maxX {
				maxX = vertex.X
			}
			if vertex.Y < minY {
				minY = vertex.Y
			}
			if vertex.Y > maxY {
				maxY = vertex.Y
			}
			if vertex.Z < minZ {
				minZ = vertex.Z
			}
			if vertex.Z > maxZ {
				maxZ = vertex.Z
			}
		}
	}
	center := Vector3D{
		X: (minX + maxX) / 2,
		Y: (minY + maxY) / 2,
		Z: (minZ + maxZ) / 2,
	}
	span := math.Max(maxX-minX, math.Max(maxY-minY, maxZ-minZ))
	if span == 0 {
		span = 1
	}
	scale := 3.0 / span
	for fi := range model.Faces {
		for vi := range model.Faces[fi].Vertices {
			vertex := model.Faces[fi].Vertices[vi]
			model.Faces[fi].Vertices[vi] = Vector3D{
				X: (vertex.X - center.X) * scale,
				Y: (vertex.Y-center.Y)*scale + 0.15,
				Z: (vertex.Z - center.Z) * scale,
			}
		}
	}
}

// shiftDynamicZ adds a small animated Z offset to a cloned asset model.
func shiftDynamicZ(model *Model, delta float64) {
	if model == nil {
		return
	}
	for fi := range model.Faces {
		for vi := range model.Faces[fi].Vertices {
			model.Faces[fi].Vertices[vi].Z += delta
		}
	}
}

// shiftLineModel repositions a generated line model in world space.
func shiftLineModel(model *Model, dx, dy, dz float64) {
	if model == nil {
		return
	}
	for fi := range model.Faces {
		for vi := range model.Faces[fi].Vertices {
			model.Faces[fi].Vertices[vi].X += dx
			model.Faces[fi].Vertices[vi].Y += dy
			model.Faces[fi].Vertices[vi].Z += dz
		}
	}
}
