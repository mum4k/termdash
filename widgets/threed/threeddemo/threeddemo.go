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
)

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
	if index >= threed.ShowcaseSceneCount {
		index = threed.ShowcaseSceneCount - 1
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
		threed.ZoomScale(30.0),
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
	scenes := threed.BuildShowcaseScenes(assetLoaded)
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
					container.SplitPercent(66),
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
		"1-9 switch scenes   Space pause/resume   Arrow keys orbit   Mouse wheel zoom   q / Esc quit",
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
	case k.Key >= keyboard.Key('1') && k.Key <= keyboard.Key('9'):
		state.setScene(int(k.Key - keyboard.Key('1')))
	}
}

// animateScenes keeps the stage and side panels synchronized with the current scene.
func animateScenes(
	ctx context.Context,
	stage *threed.ThreeD,
	catalog *text.Text,
	details *text.Text,
	scenes []threed.ShowcaseScene,
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
func sceneCatalogText(scenes []threed.ShowcaseScene, active int, assetLoaded bool) []struct {
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
func renderSceneCatalog(w *text.Text, scenes []threed.ShowcaseScene, active int, assetLoaded bool) error {
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
func sceneDetailsText(scene threed.ShowcaseScene, step int, assetLoaded bool, assetPath string) string {
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
func renderSceneDetails(w *text.Text, scene threed.ShowcaseScene, step int, assetLoaded bool, assetPath string) error {
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
		{"Keys", "1-9 scene select", cell.ColorNumber(252)},
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
