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

// Binary spectrumdemo shows the functionality of a spectrum widget.
package main

import (
	"context"
	"fmt"
	"image"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/borderfx"
	"github.com/mum4k/termdash/widgets/spectrum"
)

const (
	// frameInterval is tuned to feel responsive without creating terminal flicker.
	frameInterval = 58 * time.Millisecond
	// spectrumMax fixes the demo scale so profile changes remain comparable.
	spectrumMax = 600
	// stereoBins controls how many bands feed the mirrored analyzer panes.
	stereoBins = 44
	// networkBins controls how many bands feed the half-duplex telemetry pane.
	networkBins = 56
	// defaultThreshold starts the telemetry alarm line near the previous alert
	// behavior used by the demos.
	defaultThreshold = 500
	// thresholdMin keeps the dropdown compact while covering the useful range.
	thresholdMin = 200
	// thresholdStep controls the selectable MinY-to-MaxY threshold increments.
	thresholdStep = 50
)

// Stable container IDs let the demo update titles and focus effects at runtime.
const (
	idHarmonic = "spectrum-harmonic"
	idVectors  = "spectrum-vectors"
	idSubspace = "spectrum-subspace"
)

// dataProfile identifies the synthetic signal shape currently feeding the demo.
type dataProfile int

const (
	dataProfileLattice dataProfile = iota
	dataProfileComb
	dataProfilePulse
	dataProfileCarrier
	dataProfileStorm
)

// styleProfile identifies the glyph and color treatment applied to spectra.
type styleProfile int

const (
	styleProfileAnalyzer styleProfile = iota
	styleProfileNeedle
	styleProfileLCARS
	styleProfileMatrix
	styleProfileWire
)

// demoProfiles stores the live data and style selections.
type demoProfiles struct {
	mu    sync.RWMutex
	data  dataProfile
	style styleProfile
}

// spectrumTooltip displays the current graph sample next to the mouse.
type spectrumTooltip struct {
	mu      sync.Mutex
	text    string
	anchor  image.Point
	visible bool
}

// spectrumTerminal overlays demo controls after termdash draws the containers.
type spectrumTerminal struct {
	terminalapi.Terminal
	threshold *spectrum.AlertControl
	tooltip   *spectrumTooltip
	profiles  *demoProfiles
	activeID  func() string
}

// profileSynth owns reusable sample buffers for the animated data feed.
type profileSynth struct {
	phase float64
	max   int

	left        []int
	right       []int
	half        []int
	leftTarget  []int
	rightTarget []int
	halfTarget  []int
}

// paneFX binds a container ID to the borderfx palette used for focus sweeps.
type paneFX struct {
	id      string
	palette borderfx.Palette
}

// sweepPanes lists the windows that participate in focus-sweep animation.
var sweepPanes = []paneFX{
	{id: idHarmonic, palette: borderfx.Palettes.Cyan},
	{id: idVectors, palette: borderfx.Palettes.Synthwave},
	{id: idSubspace, palette: borderfx.Palettes.Matrix},
}

// main boots the interactive spectrum showcase.
func main() {
	baseTerm, err := tcell.New()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stereo, err := spectrum.New(
		spectrum.ChannelLabels("LEFT PHASE", "RIGHT PHASE"),
		spectrum.MaxValue(spectrumMax),
		spectrum.Gradient(cell.ColorGreen, cell.ColorNumber(118), cell.ColorYellow, cell.ColorNumber(214), cell.ColorRed),
		spectrum.PeakRunes('◆', '◇'),
	)
	if err != nil {
		panic(err)
	}
	horizontal, err := spectrum.New(
		spectrum.Horizontal(),
		spectrum.ChannelLabels("FORWARD", "AFT"),
		spectrum.MaxValue(spectrumMax),
		spectrum.Gradient(cell.ColorGreen, cell.ColorNumber(118), cell.ColorYellow, cell.ColorNumber(214), cell.ColorRed),
		spectrum.PeakRunes('◀', '▶'),
	)
	if err != nil {
		panic(err)
	}
	network, err := spectrum.New(
		spectrum.HalfDuplex(),
		spectrum.ChannelLabels("PING", "JITTER"),
		spectrum.MaxValue(spectrumMax),
		spectrum.Gradient(cell.ColorGreen, cell.ColorNumber(118), cell.ColorYellow, cell.ColorNumber(214), cell.ColorRed),
		spectrum.PeakRunes('▲', '▼'),
		spectrum.HalfDuplexRune('⣿'),
	)
	if err != nil {
		panic(err)
	}
	profiles := newDemoProfiles()
	thresholds, err := spectrum.NewAlertControl(thresholdMin, spectrumMax, thresholdStep, defaultThreshold, func(v int) error {
		return applyNetworkThreshold(network, v)
	})
	if err != nil {
		panic(err)
	}
	t := newSpectrumTerminal(baseTerm, thresholds, profiles)
	defer t.Close()
	baseTerm.EnableMouseMotion()

	if err := applySpectrumProfiles(stereo, network, horizontal, profiles); err != nil {
		panic(err)
	}
	if err := applyNetworkThreshold(network, thresholds.Threshold()); err != nil {
		panic(err)
	}

	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(50,
			grid.ColWidthPerc(52,
				grid.Widget(stereo, paneOptions(idHarmonic, dataInstructionTitle(profiles), false)...),
			),
			grid.ColWidthPerc(48,
				grid.Widget(horizontal, paneOptions(idVectors, styleInstructionTitle(profiles), false)...),
			),
		),
		grid.RowHeightPerc(50,
			grid.Widget(network, paneOptions(idSubspace, focusInstructionTitle(profiles), true)...),
		),
	)
	gridOpts, err := builder.Build()
	if err != nil {
		panic(err)
	}

	c, err := container.New(t, append(gridOpts,
		container.KeyFocusNext(keyboard.KeyTab),
		container.KeyFocusPrevious(keyboard.KeyBacktab),
	)...)
	if err != nil {
		panic(err)
	}
	t.setActiveIDProvider(c.ActiveID)

	if err := applySpectrumTitles(c, profiles); err != nil {
		panic(err)
	}
	fx := configureFocusSweep(c)
	go func() {
		_ = fx.Run(ctx)
	}()
	go animate(ctx, stereo, network, horizontal, profiles, thresholds)
	go refreshSpectrumTooltip(ctx, t, network)

	quitter := func(k *terminalapi.Keyboard) {
		if handleSpectrumProfileKey(k.Key, profiles) {
			if err := applySpectrumProfiles(stereo, network, horizontal, profiles); err != nil {
				panic(err)
			}
			if err := applySpectrumTitles(c, profiles); err != nil {
				panic(err)
			}
			return
		}
		if k.Key == 'q' || k.Key == 'Q' || k.Key == keyboard.KeyEsc {
			cancel()
		}
	}
	mouseWatcher := func(m *terminalapi.Mouse) {
		if m.Button == mouse.ButtonLeft {
			if thresholds.HandleMouse(m.Position, networkWidgetArea(t.Size()), networkPrimaryLabel(profiles)) {
				_ = c.Update(idSubspace, container.Focused())
				t.tooltip.hide()
				return
			}
			activeID := c.ActiveID()
			targetID := spectrumPaneAt(t.Size(), m.Position)
			if targetID != "" && targetID != activeID {
				_ = c.Update(targetID, container.Focused())
				t.tooltip.hide()
				return
			}
			if targetID == idSubspace && activeID == idSubspace {
				showNetworkTooltip(t.tooltip, network, t.Size(), m.Position)
				return
			}
			t.tooltip.hide()
			return
		}
	}

	if err := termdash.Run(ctx, t, c,
		termdash.KeyboardSubscriber(quitter),
		termdash.MouseSubscriber(mouseWatcher),
		termdash.RedrawInterval(frameInterval),
	); err != nil {
		panic(err)
	}
}

// paneOptions returns the shared container styling for spectrum demo windows.
func paneOptions(id, title string, focused bool) []container.Option {
	opts := []container.Option{
		container.ID(id),
		container.Border(linestyle.Round),
		container.BorderTitle(title),
		container.BorderTitleAlignCenter(),
		container.BorderColor(cell.ColorNumber(239)),
		container.FocusedColor(cell.ColorNumber(51)),
		container.TitleColor(cell.ColorNumber(245)),
		container.TitleFocusedColor(cell.ColorNumber(195)),
		container.PaddingLeft(1),
		container.PaddingTop(1),
	}
	if focused {
		opts = append(opts, container.Focused())
	}
	return opts
}

// configureFocusSweep wires borderfx so only the focused pane animates.
func configureFocusSweep(c *container.Container) *borderfx.Animator {
	fx := borderfx.NewAnimator(c)
	fx.SetTickRate(frameInterval)
	fx.SetInactiveStyle(inactiveBorderStyle)
	for _, pane := range sweepPanes {
		fx.RegisterMacro(pane.id, borderfx.Presets.Rail, pane.palette)
	}
	return fx
}

// newSpectrumTerminal wraps tcell so demo overlays can draw after containers.
func newSpectrumTerminal(base terminalapi.Terminal, thresholds *spectrum.AlertControl, profiles *demoProfiles) *spectrumTerminal {
	return &spectrumTerminal{
		Terminal:  base,
		threshold: thresholds,
		tooltip:   &spectrumTooltip{},
		profiles:  profiles,
	}
}

// Flush overlays threshold controls and hover readouts before screen flush.
func (st *spectrumTerminal) Flush() error {
	if st.threshold != nil {
		graphArea := networkWidgetArea(st.Size())
		_ = st.threshold.Draw(st.Terminal, graphArea, networkPrimaryLabel(st.profiles))
		_ = st.threshold.DrawAlert(st.Terminal, spectrumPaneRects(st.Size())[idSubspace], st.activePaneID() == idSubspace)
	}
	if st.tooltip != nil {
		st.tooltip.draw(st.Terminal)
	}
	return st.Terminal.Flush()
}

// setActiveIDProvider stores a callback that reports the currently focused pane.
func (st *spectrumTerminal) setActiveIDProvider(fn func() string) {
	st.activeID = fn
}

// activePaneID returns the ID of the focused pane, if known.
func (st *spectrumTerminal) activePaneID() string {
	if st.activeID == nil {
		return ""
	}
	return st.activeID()
}

// inactiveBorderStyle greys inactive window borders while keeping titles legible.
func inactiveBorderStyle(id string, bc container.BorderCell) container.BorderCellStyle {
	_ = id
	color := cell.ColorNumber(238)
	if bc.Title {
		color = cell.ColorNumber(244)
	}
	return container.BorderCellStyle{
		Rune:     softenedDemoBorderRune(bc.Rune),
		CellOpts: []cell.Option{cell.FgColor(color)},
	}
}

// softenedDemoBorderRune replaces sharp rounded corners with softer caps.
func softenedDemoBorderRune(r rune) rune {
	switch r {
	case '╭':
		return '○'
	case '╮':
		return '○'
	case '╰':
		return '○'
	case '╯':
		return '○'
	default:
		return r
	}
}

// applyNetworkThreshold updates the graph alarm line and alert color.
func applyNetworkThreshold(network *spectrum.Spectrum, threshold int) error {
	return network.Configure(
		spectrum.Threshold(threshold),
		spectrum.ThresholdLineColor(cell.ColorRed),
		spectrum.AlertColor(cell.ColorRed),
	)
}

// focusSpectrumPaneAt moves keyboard focus to the pane under the mouse.
func focusSpectrumPaneAt(c *container.Container, size, pos image.Point) bool {
	if targetID := spectrumPaneAt(size, pos); targetID != "" {
		_ = c.Update(targetID, container.Focused())
		return true
	}
	return false
}

// spectrumPaneAt returns the pane ID containing pos.
func spectrumPaneAt(size, pos image.Point) string {
	for id, rect := range spectrumPaneRects(size) {
		if pos.In(rect) {
			return id
		}
	}
	return ""
}

// spectrumPaneRects approximates the grid layout for mouse hit testing.
func spectrumPaneRects(size image.Point) map[string]image.Rectangle {
	if size.X <= 0 || size.Y <= 0 {
		return nil
	}
	topH := (size.Y * 50) / 100
	if topH < 1 {
		topH = 1
	}
	if topH >= size.Y {
		topH = size.Y - 1
	}
	leftW := (size.X * 52) / 100
	if leftW < 1 {
		leftW = 1
	}
	if leftW >= size.X {
		leftW = size.X - 1
	}
	return map[string]image.Rectangle{
		idHarmonic: image.Rect(0, 0, leftW, topH),
		idVectors:  image.Rect(leftW, 0, size.X, topH),
		idSubspace: image.Rect(0, topH, size.X, size.Y),
	}
}

// networkWidgetArea returns the bottom graph canvas area in terminal cells.
func networkWidgetArea(size image.Point) image.Rectangle {
	rect := spectrumPaneRects(size)[idSubspace]
	if rect.Empty() || rect.Dx() < 4 || rect.Dy() < 4 {
		return image.Rectangle{}
	}
	return image.Rect(rect.Min.X+2, rect.Min.Y+2, rect.Max.X-1, rect.Max.Y-1)
}

// networkPrimaryLabel returns the bottom graph's leading label.
func networkPrimaryLabel(profiles *demoProfiles) string {
	data, _ := profiles.snapshot()
	return "PING " + dataProfileName(data)
}

// showNetworkTooltip reveals the bottom-graph readout on demand.
func showNetworkTooltip(tooltip *spectrumTooltip, network *spectrum.Spectrum, size, pos image.Point) {
	text := networkHoverReadout(network, size, pos)
	if text == "" {
		tooltip.hide()
		return
	}
	tooltip.update(pos, text)
}

// networkHoverReadout formats the X/Y sample under the mouse.
func networkHoverReadout(network *spectrum.Spectrum, size, pos image.Point) string {
	graph := networkWidgetArea(size)
	if graph.Empty() || !pos.In(graph) {
		return ""
	}
	sample, ok := network.ValueAt(graph.Size(), pos.Sub(graph.Min))
	if !ok {
		return ""
	}
	return fmt.Sprintf(" X: %03d \n Y: %03d ", sample.X, sample.Y)
}

// refreshSpectrumTooltip keeps the readout current while the mouse is still.
func refreshSpectrumTooltip(ctx context.Context, t *spectrumTerminal, network *spectrum.Spectrum) {
	ticker := time.NewTicker(160 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			anchor, visible := t.tooltip.snapshot()
			if !visible {
				continue
			}
			showNetworkTooltip(t.tooltip, network, t.Size(), anchor)
		}
	}
}

// update makes the tooltip visible at anchor with text.
func (st *spectrumTooltip) update(anchor image.Point, text string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.anchor = anchor
	st.text = text
	st.visible = text != ""
}

// snapshot returns the current tooltip anchor and visibility.
func (st *spectrumTooltip) snapshot() (image.Point, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	return st.anchor, st.visible
}

// hide clears the tooltip.
func (st *spectrumTooltip) hide() {
	st.update(image.Point{}, "")
}

// draw overlays the tooltip onto the terminal.
func (st *spectrumTooltip) draw(t terminalapi.Terminal) {
	st.mu.Lock()
	text := st.text
	anchor := st.anchor
	visible := st.visible
	st.mu.Unlock()
	if !visible || text == "" {
		return
	}
	pos, ok := tooltipOrigin(t.Size(), anchor, text)
	if !ok {
		return
	}
	opts := []cell.Option{cell.FgColor(cell.ColorNumber(87)), cell.BgColor(cell.ColorNumber(236))}
	for row, line := range strings.Split(text, "\n") {
		drawTerminalText(t, image.Point{X: pos.X, Y: pos.Y + row}, line, opts...)
	}
}

// tooltipOrigin keeps a tooltip on screen near its anchor.
func tooltipOrigin(size, anchor image.Point, text string) (image.Point, bool) {
	lines := strings.Split(text, "\n")
	width := 0
	for _, line := range lines {
		if lineWidth := len([]rune(line)); lineWidth > width {
			width = lineWidth
		}
	}
	height := len(lines)
	if width == 0 || height == 0 || width > size.X || height > size.Y {
		return image.Point{}, false
	}
	x := anchor.X - width/2
	if x < 0 {
		x = 0
	}
	if x+width > size.X {
		x = size.X - width
	}
	y := anchor.Y + 1
	if y+height > size.Y {
		y = anchor.Y - height
	}
	if y < 0 || y+height > size.Y {
		return image.Point{}, false
	}
	return image.Point{X: x, Y: y}, true
}

// drawTerminalText writes text directly to the terminal.
func drawTerminalText(t terminalapi.Terminal, pos image.Point, text string, opts ...cell.Option) {
	cur := pos
	for _, r := range text {
		_ = t.SetCell(cur, r, opts...)
		cur.X++
	}
}

// newDemoProfiles returns the default live demo profile state.
func newDemoProfiles() *demoProfiles {
	return &demoProfiles{
		data:  dataProfileLattice,
		style: styleProfileAnalyzer,
	}
}

// snapshot returns the currently selected data and style profiles.
func (dp *demoProfiles) snapshot() (dataProfile, styleProfile) {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.data, dp.style
}

// setData selects the active synthetic data profile.
func (dp *demoProfiles) setData(profile dataProfile) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.data = profile
}

// setStyle selects the active rendering profile.
func (dp *demoProfiles) setStyle(profile styleProfile) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.style = profile
}

// handleSpectrumProfileKey applies number-key profile shortcuts.
func handleSpectrumProfileKey(key keyboard.Key, profiles *demoProfiles) bool {
	if profile, ok := dataProfileForKey(key); ok {
		profiles.setData(profile)
		return true
	}
	if profile, ok := styleProfileForKey(key); ok {
		profiles.setStyle(profile)
		return true
	}
	return false
}

// dataProfileForKey maps odd number keys to synthetic data profiles.
func dataProfileForKey(key keyboard.Key) (dataProfile, bool) {
	switch key {
	case '1', '!', keyboard.KeyEnd:
		return dataProfileLattice, true
	case '3', '#', '"', keyboard.KeyPgDn:
		return dataProfileComb, true
	case '5', '%':
		return dataProfilePulse, true
	case '7', '&', keyboard.KeyHome:
		return dataProfileCarrier, true
	case '9', '(', keyboard.KeyPgUp:
		return dataProfileStorm, true
	default:
		return dataProfileLattice, false
	}
}

// styleProfileForKey maps even number keys to rendering profiles.
func styleProfileForKey(key keyboard.Key) (styleProfile, bool) {
	switch key {
	case '2', '@', keyboard.KeyArrowDown:
		return styleProfileNeedle, true
	case '4', '$', '\'', keyboard.KeyArrowLeft:
		return styleProfileLCARS, true
	case '6', '^', '-', keyboard.KeyArrowRight:
		return styleProfileMatrix, true
	case '8', '*', '_', keyboard.KeyArrowUp:
		return styleProfileWire, true
	case '0', ')', keyboard.KeyInsert:
		return styleProfileAnalyzer, true
	default:
		return styleProfileAnalyzer, false
	}
}

// applySpectrumProfiles updates live widget labels, colors, and glyphs.
func applySpectrumProfiles(stereo, network, horizontal *spectrum.Spectrum, profiles *demoProfiles) error {
	data, style := profiles.snapshot()
	styleOpts := spectrumStyleOptions(style)
	stereoOpts := append([]spectrum.Option{
		spectrum.Vertical(),
		spectrum.Stereo(),
		spectrum.ChannelLabels("LEFT "+dataProfileName(data), "RIGHT "+styleProfileName(style)),
	}, styleOpts...)
	if err := stereo.Configure(stereoOpts...); err != nil {
		return err
	}

	horizontalOpts := append([]spectrum.Option{
		spectrum.Horizontal(),
		spectrum.Stereo(),
		spectrum.ChannelLabels("FORWARD "+styleProfileName(style), "AFT "+dataProfileName(data)),
	}, styleOpts...)
	if err := horizontal.Configure(horizontalOpts...); err != nil {
		return err
	}

	networkOpts := append([]spectrum.Option{
		spectrum.Vertical(),
		spectrum.HalfDuplex(),
		spectrum.ChannelLabels("PING "+dataProfileName(data), "LATENCY "+styleProfileName(style)),
	}, styleOpts...)
	return network.Configure(networkOpts...)
}

// applySpectrumTitles refreshes the instructional border titles.
func applySpectrumTitles(c *container.Container, profiles *demoProfiles) error {
	if err := c.Update(idHarmonic, container.BorderTitle(dataInstructionTitle(profiles))); err != nil {
		return err
	}
	if err := c.Update(idVectors, container.BorderTitle(styleInstructionTitle(profiles))); err != nil {
		return err
	}
	return c.Update(idSubspace, container.BorderTitle(focusInstructionTitle(profiles)))
}

// dataInstructionTitle returns the title for data-profile shortcuts.
func dataInstructionTitle(profiles *demoProfiles) string {
	_ = profiles
	return "DATA SHAPE | Press 1, 3, 5, 7, 9 to change data shape"
}

// styleInstructionTitle returns the title for style-profile shortcuts.
func styleInstructionTitle(profiles *demoProfiles) string {
	_ = profiles
	return "STYLE PROFILE | Press 2, 4, 6, 8, 0 to change style"
}

// focusInstructionTitle returns the title for focus navigation.
func focusInstructionTitle(profiles *demoProfiles) string {
	data, style := profiles.snapshot()
	return "ACTIVE SIGNAL | Tab focus | q quit | " + dataProfileName(data) + " / " + styleProfileName(style)
}

// spectrumStyleOptions returns the drawing options for a visual profile.
func spectrumStyleOptions(profile styleProfile) []spectrum.Option {
	switch profile {
	case styleProfileNeedle:
		return []spectrum.Option{
			spectrum.Gradient(cell.ColorNumber(35), cell.ColorNumber(82), cell.ColorNumber(154), cell.ColorNumber(220), cell.ColorNumber(196)),
			spectrum.PeakRunes('▲', '▼'),
			spectrum.HalfDuplexRune('│'),
			spectrum.AxisCellOpts(cell.FgColor(cell.ColorNumber(245))),
		}
	case styleProfileLCARS:
		return []spectrum.Option{
			spectrum.Gradient(cell.ColorNumber(39), cell.ColorNumber(51), cell.ColorNumber(87), cell.ColorNumber(141), cell.ColorNumber(201), cell.ColorNumber(226)),
			spectrum.PrimaryRunes('·', '✧', '✦'),
			spectrum.SecondaryRunes('·', '✶', '✧'),
			spectrum.HorizontalRunes('·', '✦', '✸'),
			spectrum.PeakRunes('✸', '✧'),
			spectrum.HalfDuplexRune('✶'),
			spectrum.AxisCellOpts(cell.FgColor(cell.ColorNumber(63))),
		}
	case styleProfileMatrix:
		return []spectrum.Option{
			spectrum.Gradient(cell.ColorNumber(24), cell.ColorNumber(45), cell.ColorNumber(87), cell.ColorNumber(123), cell.ColorNumber(159), cell.ColorNumber(201)),
			spectrum.PrimaryRunes('·', '○', '●'),
			spectrum.SecondaryRunes('·', '◦', '◎'),
			spectrum.HorizontalRunes('·', '○', '◉'),
			spectrum.PeakRunes('◉', '◎'),
			spectrum.HalfDuplexRune('●'),
			spectrum.AxisCellOpts(cell.FgColor(cell.ColorNumber(61))),
		}
	case styleProfileWire:
		return []spectrum.Option{
			spectrum.Gradient(cell.ColorNumber(39), cell.ColorNumber(87), cell.ColorNumber(159), cell.ColorNumber(231)),
			spectrum.PeakRunes('╿', '╽'),
			spectrum.HalfDuplexRune('┆'),
			spectrum.AxisCellOpts(cell.FgColor(cell.ColorNumber(66))),
		}
	default:
		return []spectrum.Option{
			spectrum.Gradient(cell.ColorGreen, cell.ColorNumber(118), cell.ColorYellow, cell.ColorNumber(214), cell.ColorRed),
			spectrum.PrimaryRunes('⠂', '⠆', '⠇', '⠧', '⠷', '⠿', '⣿'),
			spectrum.SecondaryRunes('⠂', '⠒', '⠓', '⠛', '⠟', '⠿', '⣿'),
			spectrum.HorizontalRunes('⠂', '⠒', '⠲', '⠶', '⠾', '⠿', '⣿'),
			spectrum.PeakRunes('⣾', '⣷'),
			spectrum.HalfDuplexRune('⣿'),
			spectrum.AxisCellOpts(cell.FgColor(cell.ColorNumber(240))),
		}
	}
}

// dataProfileName returns a compact label for a synthetic data profile.
func dataProfileName(profile dataProfile) string {
	switch profile {
	case dataProfileComb:
		return "COMB"
	case dataProfilePulse:
		return "PULSE"
	case dataProfileCarrier:
		return "CARRIER"
	case dataProfileStorm:
		return "STORM"
	default:
		return "LATTICE"
	}
}

// styleProfileName returns a compact label for a rendering profile.
func styleProfileName(profile styleProfile) string {
	switch profile {
	case styleProfileNeedle:
		return "NEEDLE"
	case styleProfileLCARS:
		return "STELLAR"
	case styleProfileMatrix:
		return "ORBIT"
	case styleProfileWire:
		return "WIRE"
	default:
		return "ANALYZER"
	}
}

// animate runs the frame loop and pushes synthesized samples into each widget.
func animate(ctx context.Context, stereo, network, horizontal *spectrum.Spectrum, profiles *demoProfiles, thresholds *spectrum.AlertControl) {
	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()

	synth := newProfileSynth(stereoBins, networkBins, spectrumMax)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			data, _ := profiles.snapshot()
			left, right, half := synth.step(data)
			_ = stereo.Update(left, right)
			_ = horizontal.Update(left, right)
			_ = network.Update(half, nil)
			if thresholds != nil {
				thresholds.UpdateSamples(half)
			}
		}
	}
}

// newProfileSynth allocates reusable buffers for the profiled demo feed.
func newProfileSynth(stereoBins, halfBins, max int) *profileSynth {
	if stereoBins < 1 {
		stereoBins = 1
	}
	if halfBins < 1 {
		halfBins = 1
	}
	if max < 1 {
		max = 1
	}
	return &profileSynth{
		max:         max,
		left:        make([]int, stereoBins),
		right:       make([]int, stereoBins),
		half:        make([]int, halfBins),
		leftTarget:  make([]int, stereoBins),
		rightTarget: make([]int, stereoBins),
		halfTarget:  make([]int, halfBins),
	}
}

// step advances the active data profile and returns reusable sample slices.
func (ps *profileSynth) step(profile dataProfile) (left, right, half []int) {
	ps.phase += dataProfileStep(profile)
	fillStereoTargets(ps.leftTarget, profile, ps.phase, 0.0, ps.max)
	fillStereoTargets(ps.rightTarget, profile, ps.phase, 0.47, ps.max)
	fillHalfTargets(ps.halfTarget, profile, ps.phase, ps.max)
	smoothSamples(ps.left, ps.leftTarget, 0.66, 0.18, ps.max)
	smoothSamples(ps.right, ps.rightTarget, 0.62, 0.16, ps.max)
	smoothSamples(ps.half, ps.halfTarget, 0.70, 0.20, ps.max)
	return ps.left, ps.right, ps.half
}

// dataProfileStep returns the phase speed for a data profile.
func dataProfileStep(profile dataProfile) float64 {
	switch profile {
	case dataProfileComb:
		return 0.115
	case dataProfilePulse:
		return 0.132
	case dataProfileCarrier:
		return 0.102
	case dataProfileStorm:
		return 0.124
	default:
		return 0.095
	}
}

// fillStereoTargets generates the two mirrored analyzer channels.
func fillStereoTargets(dst []int, profile dataProfile, phase, offset float64, max int) {
	last := maxFloat(float64(len(dst)-1), 1)
	for i := range dst {
		band := float64(i) / last
		dst[i] = clampSample(int(stereoEnergy(profile, phase, band, offset)*float64(max)), max)
	}
}

// fillHalfTargets generates the lower half-duplex telemetry channel.
func fillHalfTargets(dst []int, profile dataProfile, phase float64, max int) {
	last := maxFloat(float64(len(dst)-1), 1)
	for i := range dst {
		band := float64(i) / last
		base := stereoEnergy(profile, phase*0.92, band, 0.31) * 0.66
		ping := math.Pow(0.5+0.5*math.Sin(phase*2.4+band*18.0), 12) * 0.32
		dst[i] = clampSample(int((base+ping)*float64(max)), max)
	}
}

// stereoEnergy returns normalized signal energy for one band.
func stereoEnergy(profile dataProfile, phase, band, offset float64) float64 {
	switch profile {
	case dataProfileComb:
		return quantizedEnergy(0.22 + 0.28*sine01(phase*0.8+band*math.Pi*4.4+offset) + 0.34*math.Pow(sine01(phase*1.9+band*32.0), 5) + 0.16*sine01(phase*0.35-band*11.0))
	case dataProfilePulse:
		return quantizedEnergy(0.18 + 0.24*sine01(phase*0.7+band*math.Pi*2.2) + 0.46*math.Pow(sine01(phase*1.35+band*9.2+offset), 9) + 0.12*sine01(phase*5.3+band*17.0))
	case dataProfileCarrier:
		center := 0.50 + 0.34*math.Sin(phase*0.64+offset)
		sweep := math.Exp(-math.Pow(band-center, 2) / 0.018)
		return quantizedEnergy(0.20 + 0.46*sweep + 0.22*sine01(phase*1.1+band*math.Pi*5.8) + 0.12*sine01(phase*3.9-band*24.0))
	case dataProfileStorm:
		return quantizedEnergy(0.18 + 0.22*sine01(phase*1.4+band*math.Pi*7.4) + 0.20*sine01(phase*3.8-band*29.0) + 0.34*math.Pow(sine01(phase*2.6+band*43.0+offset), 6))
	default:
		return quantizedEnergy(0.18 + 0.42*sine01(phase+band*math.Pi*2.7+offset) + 0.26*sine01(phase*0.52-band*math.Pi*6.0) + 0.14*sine01(phase*2.1+band*15.0))
	}
}

// smoothSamples applies fast attack and slower release to reduce flicker.
func smoothSamples(dst, target []int, attack, release float64, max int) {
	for i := range dst {
		diff := target[i] - dst[i]
		rate := release
		if diff > 0 {
			rate = attack
		}
		dst[i] = clampSample(int(float64(dst[i])+float64(diff)*rate+0.5), max)
	}
}

// quantizedEnergy snaps normalized energy into small digital steps.
func quantizedEnergy(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		v = 1
	}
	return math.Round(v*36) / 36
}

// sine01 returns a sine wave normalized to the range [0, 1].
func sine01(v float64) float64 {
	return 0.5 + 0.5*math.Sin(v)
}

// clampSample keeps generated values inside the configured widget range.
func clampSample(v, max int) int {
	switch {
	case v < 0:
		return 0
	case v > max:
		return max
	default:
		return v
	}
}

// maxFloat returns the larger of two float64 values.
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
