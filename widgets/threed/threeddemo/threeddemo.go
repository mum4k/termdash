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

// Binary threeddemo shows the threed widget as a polished multi-scene showcase.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/threed"
)

const (
	// sceneTick controls how frequently the animated showcase refreshes.
	sceneTick = 40 * time.Millisecond
	// sceneCount is the number of curated scenes shown in the demo.
	sceneCount = 4
)

// sceneSpec describes one curated threed showcase scene.
type sceneSpec struct {
	Name     string
	Summary  string
	Features []string
	Orbit    threed.Vector3D
	Build    func(step int, asset *threed.Model) *threed.Model
}

// demoState stores the interactive scene state for the showcase.
type demoState struct {
	mu        sync.RWMutex
	scene     int
	paused    bool
	assetPath string
}

// sceneIndex returns the active zero-based scene index.
func (s *demoState) sceneIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scene
}

// setScene activates a specific zero-based scene index.
func (s *demoState) setScene(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 {
		index = 0
	}
	if index >= sceneCount {
		index = sceneCount - 1
	}
	s.scene = index
}

// togglePause flips the scene animation pause state.
func (s *demoState) togglePause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paused = !s.paused
}

// pauseState reports whether the scene animation is paused.
func (s *demoState) pauseState() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.paused
}

// main boots the showcase demo.
func main() {
	imagePath := flag.String("image", "", "optional image file to extrude in scene 4")
	flag.Parse()

	term, err := tcell.New()
	if err != nil {
		log.Fatalf("failed to initialize terminal: %v", err)
	}
	defer term.Close()
	term.EnableMouse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stage, err := threed.New(
		threed.ShowAxes(false),
		threed.EnableLogging(false),
		threed.BackfaceCulling(false),
		threed.RotationStep(0.08),
		threed.ZoomScale(28.0),
		threed.UprightOnly(true),
		threed.AmbientColor(threed.Color{R: 0.52, G: 0.52, B: 0.52}),
		threed.DiffuseColor(threed.Color{R: 1.00, G: 1.00, B: 1.00}),
		threed.SpecularColor(threed.Color{R: 0.44, G: 0.44, B: 0.44}),
		threed.Shininess(48),
	)
	if err != nil {
		log.Fatalf("failed to create threed widget: %v", err)
	}

	catalog, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("failed to create scene catalog: %v", err)
	}
	details, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("failed to create scene details: %v", err)
	}
	controls, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("failed to create controls panel: %v", err)
	}

	assetModel, assetLoaded := loadOptionalAsset(*imagePath)
	scenes := buildDemoScenes(assetLoaded)
	state := &demoState{assetPath: *imagePath}
	stage.SetModel(scenes[0].Build(0, assetModel))

	if err := renderSceneCatalog(catalog, scenes, 0, assetLoaded); err != nil {
		log.Fatalf("failed to render scene catalog: %v", err)
	}
	if err := renderSceneDetails(details, scenes[0], 0, assetLoaded, state.assetPath); err != nil {
		log.Fatalf("failed to render scene details: %v", err)
	}
	if err := renderControls(controls, assetLoaded); err != nil {
		log.Fatalf("failed to render controls panel: %v", err)
	}

	root, err := container.New(
		term,
		container.ID("root"),
		container.Border(linestyle.Round),
		container.BorderTitle(" ThreeD Showcase "),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						pane("Render Stage",
							container.PlaceWidget(stage),
						)...,
					),
					container.Right(
						container.SplitHorizontal(
							container.Top(
								pane("Scene Catalog",
									container.PaddingLeft(1),
									container.PaddingTop(1),
									container.PlaceWidget(catalog),
								)...,
							),
							container.Bottom(
								container.SplitHorizontal(
									container.Top(
										pane("Render Notes",
											container.PaddingLeft(1),
											container.PaddingTop(1),
											container.PlaceWidget(details),
										)...,
									),
									container.Bottom(
										pane("Controls",
											container.PaddingLeft(1),
											container.PaddingTop(1),
											container.PlaceWidget(controls),
										)...,
									),
									container.SplitPercent(70),
								),
							),
							container.SplitPercent(30),
						),
					),
					container.SplitPercent(62),
				),
			),
			container.Bottom(
				container.PlaceWidget(mustFooter()),
			),
			container.SplitPercent(93),
		),
	)
	if err != nil {
		log.Fatalf("failed to create root container: %v", err)
	}

	go animateScenes(ctx, stage, catalog, details, scenes, assetModel, assetLoaded, state)

	if err := termdash.Run(
		ctx,
		term,
		root,
		termdash.RedrawInterval(sceneTick),
		termdash.KeyboardSubscriber(func(k *terminalapi.Keyboard) {
			handleKeyboard(k, cancel, state)
		}),
	); err != nil {
		log.Fatalf("termdash terminated with error: %v", err)
	}
}

// pane returns a standard demo pane configuration.
func pane(title string, opts ...container.Option) []container.Option {
	base := []container.Option{
		container.Border(linestyle.Round),
		container.BorderTitle(" " + title + " "),
		container.BorderColor(cell.ColorNumber(240)),
	}
	return append(base, opts...)
}

// mustFooter creates the footer widget used for the operator help strip.
func mustFooter() *text.Text {
	footer, err := text.New()
	if err != nil {
		log.Fatalf("failed to create footer widget: %v", err)
	}
	if err := footer.Write(
		"1-4 switch scenes   Space pause/resume   Arrow keys orbit   Mouse wheel zoom   q / Esc quit",
		text.WriteCellOpts(cell.FgColor(cell.ColorNumber(114))),
	); err != nil {
		log.Fatalf("failed to write footer: %v", err)
	}
	return footer
}

// loadOptionalAsset tries to extrude an optional image file for scene four.
func loadOptionalAsset(path string) (*threed.Model, bool) {
	if strings.TrimSpace(path) == "" {
		return nil, false
	}
	model, err := threed.LoadImageModel(path)
	if err != nil {
		log.Printf("warning: failed to load image model %q: %v", path, err)
		return nil, false
	}
	return model, true
}

// handleKeyboard applies demo-level keyboard actions that sit beside the widget.
func handleKeyboard(k *terminalapi.Keyboard, cancel context.CancelFunc, state *demoState) {
	switch {
	case k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC:
		cancel()
	case k.Key == keyboard.Key('q') || k.Key == keyboard.Key('Q'):
		cancel()
	case k.Key == keyboard.Key(' '):
		state.togglePause()
	case k.Key >= keyboard.Key('1') && k.Key <= keyboard.Key('4'):
		state.setScene(int(k.Key - keyboard.Key('1')))
	}
}

// animateScenes keeps the stage and side panels synchronized with the current scene.
func animateScenes(
	ctx context.Context,
	stage *threed.ThreeD,
	catalog *text.Text,
	details *text.Text,
	scenes []sceneSpec,
	assetModel *threed.Model,
	assetLoaded bool,
	state *demoState,
) {
	ticker := time.NewTicker(sceneTick)
	defer ticker.Stop()

	step := 0
	lastScene := -1

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sceneIndex := state.sceneIndex()
			paused := state.pauseState()
			scene := scenes[sceneIndex]

			if !paused {
				stage.SetModel(scene.Build(step, assetModel))
				stage.Rotate(scene.Orbit)
				step++
			}

			if sceneIndex != lastScene || step%4 == 0 {
				if err := renderSceneCatalog(catalog, scenes, sceneIndex, assetLoaded); err != nil {
					log.Printf("failed to update scene catalog: %v", err)
				}
				if err := renderSceneDetails(details, scene, step, assetLoaded, state.assetPath); err != nil {
					log.Printf("failed to update scene details: %v", err)
				}
				lastScene = sceneIndex
			}
		}
	}
}

// sceneCatalogText returns the catalog copy for the current scene list.
func sceneCatalogText(scenes []sceneSpec, active int, assetLoaded bool) []struct {
	line  string
	color cell.Color
	bold  bool
} {
	lines := make([]struct {
		line  string
		color cell.Color
		bold  bool
	}, 0, len(scenes)+2)
	for i, scene := range scenes {
		line := fmt.Sprintf("  %d. %s", i+1, scene.Name)
		color := cell.ColorNumber(245)
		bold := false
		if i == active {
			line = fmt.Sprintf("▶ %d. %s", i+1, scene.Name)
			color = cell.ColorNumber(159)
			bold = true
		}
		lines = append(lines, struct {
			line  string
			color cell.Color
			bold  bool
		}{line: line, color: color, bold: bold})
	}
	lines = append(lines, struct {
		line  string
		color cell.Color
		bold  bool
	}{line: "", color: cell.ColorNumber(245)})
	assetLine := "Asset scene: procedural prism field"
	if assetLoaded {
		assetLine = "Asset scene: image relief armed"
	}
	lines = append(lines, struct {
		line  string
		color cell.Color
		bold  bool
	}{line: assetLine, color: cell.ColorNumber(114)})
	return lines
}

// renderSceneCatalog writes the available scene list and highlights the active scene.
func renderSceneCatalog(w *text.Text, scenes []sceneSpec, active int, assetLoaded bool) error {
	w.Reset()
	for _, line := range sceneCatalogText(scenes, active, assetLoaded) {
		opts := []text.WriteOption{text.WriteCellOpts(cell.FgColor(line.color))}
		if line.bold {
			opts = []text.WriteOption{text.WriteCellOpts(cell.Bold(), cell.FgColor(line.color))}
		}
		if err := w.Write(line.line+"\n", opts...); err != nil {
			return err
		}
	}
	return nil
}

// sceneDetailsText builds the detail panel copy for the active scene.
func sceneDetailsText(scene sceneSpec, step int, assetLoaded bool, assetPath string) string {
	var b strings.Builder
	b.WriteString("Scene: ")
	b.WriteString(scene.Name)
	b.WriteString("\n")
	b.WriteString(scene.Summary)
	b.WriteString("\n\nFeatures\n")
	for _, feature := range scene.Features {
		b.WriteString("• ")
		b.WriteString(feature)
		b.WriteString("\n")
	}
	b.WriteString("\nFrame: ")
	b.WriteString(fmt.Sprintf("%04d", step))
	if strings.TrimSpace(assetPath) != "" {
		b.WriteString("\nAsset file: ")
		if assetLoaded {
			b.WriteString("loaded")
		} else {
			b.WriteString("unavailable")
		}
	}
	return b.String()
}

// renderSceneDetails writes the active scene summary and feature highlights.
func renderSceneDetails(w *text.Text, scene sceneSpec, step int, assetLoaded bool, assetPath string) error {
	w.Reset()
	if err := w.Write("Scene: ", text.WriteReplace(), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(245)))); err != nil {
		return err
	}
	if err := w.Write(scene.Name+"\n", text.WriteCellOpts(cell.Bold(), cell.FgColor(cell.ColorNumber(159)))); err != nil {
		return err
	}
	if err := w.Write(scene.Summary+"\n\n", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(252)))); err != nil {
		return err
	}
	if err := w.Write("Features\n", text.WriteCellOpts(cell.Bold(), cell.FgColor(cell.ColorNumber(81)))); err != nil {
		return err
	}
	for _, feature := range scene.Features {
		if err := w.Write("• "+feature+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(250)))); err != nil {
			return err
		}
	}
	if err := w.Write("\nFrame: ", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(245)))); err != nil {
		return err
	}
	if err := w.Write(fmt.Sprintf("%04d\n", step), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(252)))); err != nil {
		return err
	}
	if strings.TrimSpace(assetPath) == "" {
		return nil
	}
	label := "Asset file: unavailable\n"
	color := cell.ColorNumber(167)
	if assetLoaded {
		label = "Asset file: loaded\n"
		color = cell.ColorNumber(114)
	}
	return w.Write(label, text.WriteCellOpts(cell.FgColor(color)))
}

// controlSummaryLines returns the static operator-help rows.
func controlSummaryLines(assetLoaded bool) []struct {
	label string
	value string
	color cell.Color
} {
	return []struct {
		label string
		value string
		color cell.Color
	}{
		{"Keys", "1-4 scene select", cell.ColorNumber(252)},
		{"Orbit", "Arrow keys or slow auto orbit", cell.ColorNumber(252)},
		{"Zoom", "Mouse wheel", cell.ColorNumber(252)},
		{"Pause", "Space toggles animation", cell.ColorNumber(252)},
		{"Asset", ternary(assetLoaded, "image relief online", "procedural fallback"), ternaryColor(assetLoaded, cell.ColorNumber(114), cell.ColorNumber(245))},
	}
}

// renderControls writes the interaction summary for the operator panel.
func renderControls(w *text.Text, assetLoaded bool) error {
	w.Reset()
	for _, line := range controlSummaryLines(assetLoaded) {
		if err := w.Write(line.label+": ", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(245)))); err != nil {
			return err
		}
		if err := w.Write(line.value+"\n", text.WriteCellOpts(cell.FgColor(line.color))); err != nil {
			return err
		}
	}
	return nil
}

// buildDemoScenes returns the curated 3D showcase scenes.
func buildDemoScenes(assetLoaded bool) []sceneSpec {
	assetName := "Prism Field"
	assetSummary := "A clean prism-board fallback with stacked bars, rails, and labeled blocks."
	assetFeatures := []string{
		"filled face shading",
		"procedural bar-field composition",
		"high-density braille fill pass",
	}
	if assetLoaded {
		assetName = "Image Relief"
		assetSummary = "Asset-backed extrusion on a clean display pedestal."
		assetFeatures = []string{
			"image extrusion via threed.LoadImageModel",
			"source-derived face colors",
			"same lighting stack as procedural geometry",
		}
	}

	return []sceneSpec{
		{
			Name:    "Terminal Forms",
			Summary: "Sharp box-drawing forms, chips, nodes, and traces arranged as terminal-native geometry.",
			Features: []string{
				"crisp box-drawing geometry",
				"multiple high-contrast shape families",
				"explicit title-like glyph placement",
			},
			Orbit: threed.Vector3D{},
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
			Orbit: threed.Vector3D{},
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
			Orbit: threed.Vector3D{},
			Build: buildSignalRigScene,
		},
		{
			Name:     assetName,
			Summary:  assetSummary,
			Features: assetFeatures,
			Orbit:    threed.Vector3D{},
			Build:    buildAssetReliefScene,
		},
	}
}

// buildCircuitBloomScene assembles the main reference-inspired circuit scene.
func buildCircuitBloomScene(step int, _ *threed.Model) *threed.Model {
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
		createGlyphRows(threed.Vector3D{X: -1.72, Y: 1.65, Z: -0.22}, rows, 0.038, 0.20, terminalFormColor),
		createGlyphRows(threed.Vector3D{X: -1.70, Y: 1.83, Z: -0.40}, halo, 0.038, 0.20, mutedCircuitColor),
		createPulseGlyphs(step),
	)
}

// buildFormCatalogScene assembles the mixed primitive and glyph form lineup.
func buildFormCatalogScene(step int, _ *threed.Model) *threed.Model {
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
		createGlyphRows(threed.Vector3D{X: -1.78, Y: 1.48, Z: -0.22}, rows, 0.036, 0.21, terminalFormColor),
		createPulseGlyphs(step+8),
	)
}

// buildSignalRigScene assembles the telemetry-inspired model stack.
func buildSignalRigScene(step int, _ *threed.Model) *threed.Model {
	frame := step % 6
	pulse := []string{"▁", "▂", "▃", "▄", "▅", "▆"}[frame]
	rows := []string{
		"┌──────────────────── SIGNAL LATTICE ────────────────────┐",
		"│  AMPLITUDE       PACKET RAIL              MATRIX       │",
		"│  ████▇▇▆▅▄▃▂     ═══╦══════╦══════╦═══    ◆──◆──◆──◆   │",
		"│  ███▇▆▅▄▃▂▁       ║      ║      ║       │  │  │  │   │",
		"│  ██▆▅▄▃▂▁         ═╩══════╩══════╩═══    ◆──◆──◆──◆   │",
		"│                                                       │",
		"│  TRACE A     ┌───┐       ┌────┐       ┌──┐            │",
		"│  ────────╮   │   ╰───────╯    ╰───────╯  ╰────        │",
		"│          ╰───╯                                         │",
		"│  TRACE B     ╭────╮          ╭──────╮                 │",
		"│  ────────────╯    ╰──────────╯      ╰────────         │",
		"│                                                       │",
		"│  BINS       " + pulse + " ███ ███  ██  ████  ██  ███        │",
		"└───────────────────────────────────────────────────────┘",
	}
	return createGlyphRows(threed.Vector3D{X: -1.55, Y: 1.34, Z: -0.22}, rows, 0.041, 0.22, graphColor)
}

// buildAssetReliefScene assembles the asset-backed or procedural fallback scene.
func buildAssetReliefScene(step int, asset *threed.Model) *threed.Model {
	if asset != nil {
		return buildImageReliefScene(step, asset)
	}
	return buildPrismFieldScene(step)
}

// buildImageReliefScene places the optional image model on a display pedestal.
func buildImageReliefScene(step int, asset *threed.Model) *threed.Model {
	base := createBoxModel(threed.Vector3D{X: 0, Y: -1.20, Z: 0}, 8.2, 0.18, 4.6, threed.Color{R: 0.14, G: 0.18, B: 0.25})
	plinth := createBoxModel(threed.Vector3D{X: 0, Y: -0.92, Z: 0}, 4.6, 0.22, 2.6, threed.Color{R: 0.26, G: 0.34, B: 0.46})
	grid := createGridModel(-1.08, 7.8, 4.0, 14, 6)
	grid.SetColor(threed.Color{R: 0.22, G: 0.32, B: 0.48})
	model := cloneModel(asset)
	normalizeAssetModel(model)
	spinAccent := createRingBand(threed.Vector3D{X: 0, Y: -0.55, Z: 0}, 1.55, 1.78, 0.10, 24, threed.Color{R: 0.95, G: 0.76, B: 0.25})
	shiftDynamicZ(model, 0.10*math.Sin(float64(step)*0.08))
	return mergeModels(base, plinth, grid, spinAccent, model)
}

// buildPrismFieldScene creates a procedural fallback scene when no asset is present.
func buildPrismFieldScene(step int) *threed.Model {
	frame := step % 4
	bar := []string{"▂", "▄", "▆", "█"}[frame]
	rows := []string{
		"┌──────────────────── PRISM FIELD ───────────────────────┐",
		"│  STACKS                         PROFILE                │",
		"│  ┌────┐  ┌────┐  ┌────┐         ┌─────────────────┐    │",
		"│  │ ██ │  │ █  │  │ ██ │         │ ▁▂▄▆█▆▄▂▁▂▄▆█   │    │",
		"│  │ ██ │  │ ██ │  │ █  │         └────────┬────────┘    │",
		"│  └─┬──┘  └─┬──┘  └─┬──┘                  │             │",
		"│    │       │       │        ◆────◆────◆───┘             │",
		"│  ┌─┴───────┴───────┴─┐     │    │    │                 │",
		"│  │  " + bar + "  █  ▆  ▄  ▂  █  │     ◆────◆────◆                 │",
		"│  └───────────────────┘                                      │",
		"│                                                             │",
		"│  RAIL A  ═════╦═════╦═════╦═════╦═════                    │",
		"│  RAIL B  ─────┴─────┴─────┴─────┴─────   ○ ○ ○ ○ ○        │",
		"└─────────────────────────────────────────────────────────────┘",
	}
	return createGlyphRows(threed.Vector3D{X: -1.60, Y: 1.34, Z: -0.22}, rows, 0.040, 0.22, graphColor)
}

// createBoxModel creates a rectangular prism centered at the provided point.
func createBoxModel(center threed.Vector3D, sx, sy, sz float64, color threed.Color) *threed.Model {
	hx, hy, hz := sx/2, sy/2, sz/2
	vertices := []threed.Vector3D{
		{X: center.X - hx, Y: center.Y - hy, Z: center.Z - hz},
		{X: center.X + hx, Y: center.Y - hy, Z: center.Z - hz},
		{X: center.X + hx, Y: center.Y + hy, Z: center.Z - hz},
		{X: center.X - hx, Y: center.Y + hy, Z: center.Z - hz},
		{X: center.X - hx, Y: center.Y - hy, Z: center.Z + hz},
		{X: center.X + hx, Y: center.Y - hy, Z: center.Z + hz},
		{X: center.X + hx, Y: center.Y + hy, Z: center.Z + hz},
		{X: center.X - hx, Y: center.Y + hy, Z: center.Z + hz},
	}

	model := threed.NewModel()
	faces := [][]int{
		{3, 2, 1, 0},
		{4, 5, 6, 7},
		{7, 3, 0, 4},
		{1, 2, 6, 5},
		{7, 6, 2, 3},
		{0, 1, 5, 4},
	}
	for _, idx := range faces {
		model.AddFace(threed.Face{
			Vertices: []threed.Vector3D{vertices[idx[0]], vertices[idx[1]], vertices[idx[2]], vertices[idx[3]]},
			Char:     '█',
			Color:    color,
			HasColor: true,
		})
	}
	return model
}

// createGridModel creates a wire grid in the XZ plane at the provided Y level.
func createGridModel(y, width, depth float64, cols, rows int) *threed.Model {
	model := threed.NewModel()
	left := -width / 2
	right := width / 2
	front := -depth / 2
	back := depth / 2

	for i := 0; i <= cols; i++ {
		t := float64(i) / math.Max(float64(cols), 1)
		x := left + t*width
		model.AddFace(threed.Face{
			Vertices: []threed.Vector3D{
				{X: x, Y: y, Z: front},
				{X: x, Y: y, Z: back},
			},
			Char: '─',
		})
	}
	for i := 0; i <= rows; i++ {
		t := float64(i) / math.Max(float64(rows), 1)
		z := front + t*depth
		model.AddFace(threed.Face{
			Vertices: []threed.Vector3D{
				{X: left, Y: y, Z: z},
				{X: right, Y: y, Z: z},
			},
			Char: '─',
		})
	}
	return model
}

type glyphColorFunc func(rune) threed.Color

func createGlyphRows(origin threed.Vector3D, rows []string, cellWidth, cellHeight float64, colorFor glyphColorFunc) *threed.Model {
	model := threed.NewModel()
	for row, line := range rows {
		col := 0
		for _, r := range line {
			if r != ' ' {
				addGlyphBillboard(model, threed.Vector3D{
					X: origin.X + float64(col)*cellWidth,
					Y: origin.Y - float64(row)*cellHeight,
					Z: origin.Z,
				}, cellWidth*1.65, r, colorFor(r))
			}
			col++
		}
	}
	return model
}

func terminalFormColor(r rune) threed.Color {
	switch {
	case strings.ContainsRune("┌┐└┘─│├┤┬┴┼╭╮╰╯╦╩═║", r):
		return threed.Color{R: 0.54, G: 0.92, B: 0.96}
	case strings.ContainsRune("▲△╱╲", r):
		return threed.Color{R: 0.96, G: 0.74, B: 0.22}
	case strings.ContainsRune("◆◇", r):
		return threed.Color{R: 0.46, G: 0.94, B: 0.32}
	case strings.ContainsRune("○●", r):
		return threed.Color{R: 0.66, G: 0.88, B: 1.00}
	case strings.ContainsRune("█▣", r):
		return threed.Color{R: 0.96, G: 0.28, B: 0.46}
	default:
		return threed.Color{R: 0.82, G: 0.86, B: 0.88}
	}
}

func mutedCircuitColor(r rune) threed.Color {
	if r == '·' {
		return threed.Color{R: 0.18, G: 0.28, B: 0.26}
	}
	return threed.Color{R: 0.26, G: 0.34, B: 0.34}
}

func graphColor(r rune) threed.Color {
	switch {
	case strings.ContainsRune("┌┐└┘─│├┤┬┴┼╭╮╰╯╦╩═║╔╗╚╝", r):
		return threed.Color{R: 0.50, G: 0.88, B: 0.92}
	case strings.ContainsRune("█▇▆▅▄▃▂▁", r):
		return threed.Color{R: 0.70, G: 0.92, B: 1.00}
	case strings.ContainsRune("◆◇", r):
		return threed.Color{R: 0.48, G: 0.96, B: 0.34}
	case strings.ContainsRune("○●", r):
		return threed.Color{R: 0.95, G: 0.78, B: 0.24}
	default:
		return threed.Color{R: 0.84, G: 0.86, B: 0.88}
	}
}

func createPulseGlyphs(step int) *threed.Model {
	model := threed.NewModel()
	points := []threed.Vector3D{
		{X: -1.05, Y: 1.42, Z: -0.34},
		{X: -0.72, Y: -0.30, Z: -0.34},
		{X: 0.58, Y: 1.18, Z: -0.34},
		{X: 1.12, Y: -1.18, Z: -0.34},
		{X: 0.06, Y: -1.48, Z: -0.34},
	}
	glyphs := []rune{'✦', '◆', '●', '✧', '◇'}
	for i, p := range points {
		if (step+i)%3 == 0 {
			addGlyphBillboard(model, p, 0.08, glyphs[i%len(glyphs)], threed.Color{R: 0.98, G: 0.96, B: 0.42})
		}
	}
	return model
}

// createFrontGridModel creates a faint etched grid in the XY plane.
func createFrontGridModel(z, width, height float64, cols, rows int) *threed.Model {
	model := threed.NewModel()
	left := -width / 2
	right := width / 2
	bottom := -height / 2
	top := height / 2
	color := threed.Color{R: 0.12, G: 0.20, B: 0.22}

	for i := 0; i <= cols; i++ {
		t := float64(i) / math.Max(float64(cols), 1)
		x := left + t*width
		addColoredLine(model, threed.Vector3D{X: x, Y: bottom, Z: z}, threed.Vector3D{X: x, Y: top, Z: z}, '│', color)
	}
	for i := 0; i <= rows; i++ {
		t := float64(i) / math.Max(float64(rows), 1)
		y := bottom + t*height
		addColoredLine(model, threed.Vector3D{X: left, Y: y, Z: z}, threed.Vector3D{X: right, Y: y, Z: z}, '─', color)
	}
	return model
}

// createCircuitCluster creates a compact field of terminal glyphs and traces.
func createCircuitCluster(center threed.Vector3D, rows, cols int, spacing float64, primary, accent threed.Color, step int) *threed.Model {
	model := threed.NewModel()
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
				addCircuitChip(model, threed.Vector3D{X: x, Y: y, Z: z - 0.02}, spacing*0.54, spacing*0.36, color)
				continue
			}
			addGlyphBillboard(model, threed.Vector3D{X: x, Y: y, Z: z - 0.04}, spacing*1.20, glyphs[index], color)
		}
	}

	for row := 0; row < rows; row++ {
		y := center.Y - height/2 + float64(row)*spacing
		if row%2 == 0 {
			addColoredLine(model, threed.Vector3D{X: center.X - width/2, Y: y, Z: center.Z - 0.16}, threed.Vector3D{X: center.X + width/2, Y: y, Z: center.Z - 0.16}, '─', primary)
		}
	}
	for col := 0; col < cols; col++ {
		x := center.X - width/2 + float64(col)*spacing
		if col%3 == 0 {
			addColoredLine(model, threed.Vector3D{X: x, Y: center.Y - height/2, Z: center.Z - 0.18}, threed.Vector3D{X: x, Y: center.Y + height/2, Z: center.Z - 0.18}, '│', accent)
		}
	}

	return model
}

// createChipBank creates a row of small solid modules like the blocks in the reference.
func createChipBank(center threed.Vector3D, count int, spacing float64, color threed.Color) *threed.Model {
	model := threed.NewModel()
	start := -float64(count-1) * spacing / 2
	for i := 0; i < count; i++ {
		x := center.X + start + float64(i)*spacing
		model = mergeModels(model, createBoxModel(threed.Vector3D{X: x, Y: center.Y, Z: center.Z}, spacing*0.44, spacing*0.76, 0.18, color))
	}
	return model
}

func createTraceHeader(center threed.Vector3D, width float64, color threed.Color) *threed.Model {
	model := threed.NewModel()
	addColoredLine(model, threed.Vector3D{X: center.X - width/2, Y: center.Y, Z: center.Z}, threed.Vector3D{X: center.X + width/2, Y: center.Y, Z: center.Z}, '─', color)
	for i := 0; i < 7; i++ {
		x := center.X - width/2 + float64(i)*width/6
		addGlyphBillboard(model, threed.Vector3D{X: x, Y: center.Y + 0.18*math.Sin(float64(i)), Z: center.Z - 0.03}, 0.20, []rune{'■', '┼', '◆', '●', '╋', '┤', '├'}[i], color)
	}
	return model
}

func addCircuitChip(model *threed.Model, center threed.Vector3D, width, height float64, color threed.Color) {
	chip := createBoxModel(center, width, height, 0.12, color)
	for _, face := range chip.Faces {
		model.AddFace(face)
	}
}

func addGlyphBillboard(model *threed.Model, center threed.Vector3D, size float64, glyph rune, color threed.Color) {
	half := size / 2
	model.AddFace(threed.Face{
		Vertices: []threed.Vector3D{
			{X: center.X - half, Y: center.Y - half, Z: center.Z},
			{X: center.X - half, Y: center.Y + half, Z: center.Z},
			{X: center.X + half, Y: center.Y + half, Z: center.Z},
			{X: center.X + half, Y: center.Y - half, Z: center.Z},
		},
		Char:       glyph,
		RenderMode: threed.FaceRenderGlyph,
		Color:      color,
		HasColor:   true,
	})
}

// createRingBand creates a flat extruded ring in the XZ plane.
func createRingBand(center threed.Vector3D, innerRadius, outerRadius, height float64, segments int, color threed.Color) *threed.Model {
	if segments < 8 {
		segments = 8
	}
	topY := center.Y + height/2
	bottomY := center.Y - height/2
	outerTop := make([]threed.Vector3D, segments)
	outerBottom := make([]threed.Vector3D, segments)
	innerTop := make([]threed.Vector3D, segments)
	innerBottom := make([]threed.Vector3D, segments)

	for i := 0; i < segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		cosA, sinA := math.Cos(angle), math.Sin(angle)
		outerTop[i] = threed.Vector3D{X: center.X + cosA*outerRadius, Y: topY, Z: center.Z + sinA*outerRadius}
		outerBottom[i] = threed.Vector3D{X: center.X + cosA*outerRadius, Y: bottomY, Z: center.Z + sinA*outerRadius}
		innerTop[i] = threed.Vector3D{X: center.X + cosA*innerRadius, Y: topY, Z: center.Z + sinA*innerRadius}
		innerBottom[i] = threed.Vector3D{X: center.X + cosA*innerRadius, Y: bottomY, Z: center.Z + sinA*innerRadius}
	}

	model := threed.NewModel()
	for i := 0; i < segments; i++ {
		next := (i + 1) % segments
		model.AddFace(threed.Face{
			Vertices: []threed.Vector3D{outerTop[i], outerTop[next], innerTop[next], innerTop[i]},
			Char:     '█',
			Color:    color,
			HasColor: true,
		})
		model.AddFace(threed.Face{
			Vertices: []threed.Vector3D{outerBottom[next], outerBottom[i], innerBottom[i], innerBottom[next]},
			Char:     '█',
			Color:    color,
			HasColor: true,
		})
		model.AddFace(threed.Face{
			Vertices: []threed.Vector3D{outerTop[i], outerBottom[i], outerBottom[next], outerTop[next]},
			Char:     '█',
			Color:    color,
			HasColor: true,
		})
		model.AddFace(threed.Face{
			Vertices: []threed.Vector3D{innerTop[next], innerBottom[next], innerBottom[i], innerTop[i]},
			Char:     '█',
			Color:    color,
			HasColor: true,
		})
	}
	return model
}

// addLine adds a two-vertex face (line segment) to a model using the given character.
func addLine(model *threed.Model, a, b threed.Vector3D, char rune) {
	model.AddFace(threed.Face{
		Vertices: []threed.Vector3D{a, b},
		Char:     char,
	})
}

func addColoredLine(model *threed.Model, a, b threed.Vector3D, char rune, color threed.Color) {
	model.AddFace(threed.Face{
		Vertices: []threed.Vector3D{a, b},
		Char:     char,
		Color:    color,
		HasColor: true,
	})
}

// mergeModels concatenates multiple models into one combined model.
func mergeModels(models ...*threed.Model) *threed.Model {
	merged := threed.NewModel()
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
func cloneModel(model *threed.Model) *threed.Model {
	if model == nil {
		return nil
	}
	cloned := threed.NewModel()
	for _, face := range model.Faces {
		verts := make([]threed.Vector3D, len(face.Vertices))
		copy(verts, face.Vertices)
		cloned.AddFace(threed.Face{
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

// normalizeAssetModel centers and scales an asset model for the showcase stage.
func normalizeAssetModel(model *threed.Model) {
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
	center := threed.Vector3D{
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
			model.Faces[fi].Vertices[vi] = threed.Vector3D{
				X: (vertex.X - center.X) * scale,
				Y: (vertex.Y-center.Y)*scale + 0.15,
				Z: (vertex.Z - center.Z) * scale,
			}
		}
	}
}

// shiftDynamicZ adds a small animated Z offset to a cloned asset model.
func shiftDynamicZ(model *threed.Model, delta float64) {
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
func shiftLineModel(model *threed.Model, dx, dy, dz float64) {
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

// ternary returns one of two strings based on the condition.
func ternary(condition bool, yes, no string) string {
	if condition {
		return yes
	}
	return no
}

// ternaryColor returns one of two colors based on the condition.
func ternaryColor(condition bool, yes, no cell.Color) cell.Color {
	if condition {
		return yes
	}
	return no
}
