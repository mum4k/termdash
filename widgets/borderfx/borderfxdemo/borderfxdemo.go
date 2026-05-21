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

// Binary borderfxdemo - animated LCARS-style borders.
//
// MUST be run from the termdash repo root:
//
//	cd /path/to/termdash
//	go run ./widgets/borderfx/borderfxdemo/
//
// Press q or Esc to quit. Use odd number keys to select telemetry data
// profiles and even number keys to select telemetry style profiles.
package main

import (
	"context"
	"fmt"
	"image"
	"log"
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
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/borderfx"
	"github.com/mum4k/termdash/widgets/checkbox"
	"github.com/mum4k/termdash/widgets/dropdown"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/radio"
	"github.com/mum4k/termdash/widgets/slider"
	spin "github.com/mum4k/termdash/widgets/spinner"
	"github.com/mum4k/termdash/widgets/text"
)

const (
	idSensors = "sensors"
	idLCARS   = "lcars"
	idWarp    = "warp"
	idComms   = "comms"
	idHelp    = "help"

	transientSelectDuration = 2500 * time.Millisecond
	torpedoLoadDelay        = 900 * time.Millisecond
	cloakEnableDelay        = 1400 * time.Millisecond
	sparkWarmupSamples      = 224
	sparkSamplesPerTick     = 1
	sparkFrameInterval      = 170 * time.Millisecond
)

var (
	primaryDim = cell.ColorNumber(236)

	activeTitle  = cell.ColorNumber(195)
	inactiveEdge = cell.ColorNumber(239)
	inactiveText = cell.ColorNumber(244)
)

var demoPanelIDs = []string{idSensors, idLCARS, idWarp, idComms, idHelp}

var panelTitles = map[string]borderfx.TitleSpec{
	idSensors: standardPanelTitle(" Sensor Array ", "dots_6"),
	idLCARS:   standardPanelTitle(" LCARS Telemetry ", "dots_10"),
	idWarp:    standardPanelTitle(" Warp Core Flux ", "star"),
	idComms:   standardPanelTitle(" Comms ", "spin_2"),
	idHelp:    standardPanelTitle(" Controls ", "pulse"),
}

// standardPanelTitle returns the shared title reveal style used across the
// borderfx demo.
func standardPanelTitle(base, rightSpinner string) borderfx.TitleSpec {
	return borderfx.DecryptingTitle(base, titleCharset(base)).
		WithRightSpinner(spin.Must(rightSpinner))
}

// titleCharset returns the scrambling charset used while a focused window
// title resolves into its final readable text.
func titleCharset(base string) string {
	switch base {
	case " Warp Core Flux ":
		return borderfx.DecryptCharsets.WarpCoreFlux
	case " LCARS Telemetry ":
		return borderfx.DecryptCharsets.LCARSTelemetry
	case " Comms ":
		return borderfx.DecryptCharsets.Comms
	case " Controls ":
		return borderfx.DecryptCharsets.Controls
	default:
		return borderfx.DecryptCharsets.Default
	}
}

// forEachDemoPanel applies fn to each panel ID in the demo's fixed layout.
func forEachDemoPanel(fn func(string)) {
	for _, id := range demoPanelIDs {
		fn(id)
	}
}

type panelStyle int

const (
	styleInterlace panelStyle = iota + 1
	styleFire
	styleIce
	styleRainbow
	styleMatrix
	styleSynthwave
	styleNeon
)

type telemetryDataProfile int

const (
	dataProfileSine telemetryDataProfile = iota + 1
	dataProfileBurst
	dataProfileSaw
	dataProfilePulse
	dataProfileStorm
)

type telemetryStyleProfile int

const (
	styleProfileBraille telemetryStyleProfile = iota + 1
	styleProfileNeedle
	styleProfileBars
	styleProfileMatrix
	styleProfileMinimal
)

type telemetryProfiles struct {
	mu           sync.RWMutex
	data         telemetryDataProfile
	style        telemetryStyleProfile
	lastKey      string
	lastEventKey keyboard.Key
	lastEventAt  time.Time
}

type styleRegistry struct {
	mu     sync.RWMutex
	styles map[string]panelStyle
}

type clickTracker struct {
	count    int
	pressed  bool
	pressPos image.Point
	lastAt   time.Time
	lastPos  image.Point
}

type sparkProbe struct {
	mu     sync.Mutex
	values []int
}

type shipState struct {
	mu             sync.RWMutex
	warpOnline     bool
	shieldsPct     int
	torpedoSet     int
	torpedoLoaded  int
	torpedoLoading bool
	cloakOnline    bool
	cloakEnabling  bool
	cloakDisabling bool
	alarmThreshold int
	redAlert       bool
	version        int
}

type hoverTooltip struct {
	mu      sync.Mutex
	text    string
	anchor  image.Point
	visible bool
}

type statusControls struct {
	ship      *shipState
	warp      *radio.Radio
	shields   *slider.Slider
	torpedoes *dropdown.Dropdown
	alarm     *dropdown.Dropdown
	cloak     *checkbox.Checkbox
}

type demoTerminal struct {
	*tcell.Terminal
	tooltip         *hoverTooltip
	controls        *statusControls
	controlsLive    func() bool
	keyboardHandler func(*terminalapi.Keyboard)
	loadingMu       sync.RWMutex
	loadingStyle    borderfx.LoadingBackground
	loadingFrames   map[string]string
	loadingVisible  bool
}

// newTermdashTerminal builds the demo terminal using the standard termdash
// tcell adapter, then wraps it with the demo-specific overlay behavior.
func newTermdashTerminal() (*demoTerminal, error) {
	base, err := tcell.New()
	if err != nil {
		return nil, fmt.Errorf("tcell.New: %w", err)
	}
	return newDemoTerminal(base), nil
}

func newDemoTerminal(base *tcell.Terminal) *demoTerminal {
	return &demoTerminal{
		Terminal:      base,
		tooltip:       &hoverTooltip{},
		loadingFrames: make(map[string]string),
	}
}

func (dt *demoTerminal) Flush() error {
	if style, frames, visible := dt.loadingSnapshot(); visible {
		dt.drawLoadingOverlay(style, frames)
		return dt.Terminal.Flush()
	}
	dt.tooltip.draw(dt.Terminal)
	if dt.controls != nil {
		live := true
		if dt.controlsLive != nil {
			live = dt.controlsLive()
		}
		dt.controls.draw(dt.Terminal, dt.Size(), live)
	}
	return dt.Terminal.Flush()
}

func (dt *demoTerminal) setStatusControls(controls *statusControls, live func() bool) {
	dt.controls = controls
	dt.controlsLive = live
}

func (dt *demoTerminal) setKeyboardHandler(handler func(*terminalapi.Keyboard)) {
	dt.keyboardHandler = handler
}

// setLoadingBackground stores the reusable loading background style.
func (dt *demoTerminal) setLoadingBackground(style borderfx.LoadingBackground) {
	dt.loadingMu.Lock()
	dt.loadingStyle = style
	dt.loadingMu.Unlock()
}

// setLoadingFrame updates the loading copy rendered inside one panel.
func (dt *demoTerminal) setLoadingFrame(id, frame string) {
	dt.loadingMu.Lock()
	dt.loadingFrames[id] = frame
	dt.loadingMu.Unlock()
}

// setLoadingVisible enables or disables the boot overlay.
func (dt *demoTerminal) setLoadingVisible(visible bool) {
	dt.loadingMu.Lock()
	dt.loadingVisible = visible
	dt.loadingMu.Unlock()
}

// loadingSnapshot returns a consistent copy of the current loading overlay
// state for rendering.
func (dt *demoTerminal) loadingSnapshot() (borderfx.LoadingBackground, map[string]string, bool) {
	dt.loadingMu.RLock()
	defer dt.loadingMu.RUnlock()

	frames := make(map[string]string, len(dt.loadingFrames))
	for id, frame := range dt.loadingFrames {
		frames[id] = frame
	}
	return dt.loadingStyle, frames, dt.loadingVisible
}

// drawLoadingOverlay paints the boot overlay into each panel's content area.
func (dt *demoTerminal) drawLoadingOverlay(style borderfx.LoadingBackground, frames map[string]string) {
	forEachDemoPanel(func(id string) {
		rect := loadingPanelRect(dt.Size(), id)
		if rect.Empty() {
			return
		}
		_ = style.Draw(dt.Terminal, rect, frames[id])
	})
}

func (dt *demoTerminal) Event(ctx context.Context) terminalapi.Event {
	ev := dt.Terminal.Event(ctx)
	if k, ok := ev.(*terminalapi.Keyboard); ok && dt.keyboardHandler != nil {
		dt.keyboardHandler(k)
	}
	return ev
}

func newStyleRegistry() *styleRegistry {
	return &styleRegistry{
		styles: map[string]panelStyle{
			idSensors: defaultStyleForPanel(idSensors),
			idLCARS:   defaultStyleForPanel(idLCARS),
			idWarp:    defaultStyleForPanel(idWarp),
			idComms:   defaultStyleForPanel(idComms),
			idHelp:    defaultStyleForPanel(idHelp),
		},
	}
}

func newTelemetryProfiles() *telemetryProfiles {
	return &telemetryProfiles{
		data:  dataProfileSine,
		style: styleProfileBraille,
	}
}

func (tp *telemetryProfiles) snapshot() (telemetryDataProfile, telemetryStyleProfile, string) {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.data, tp.style, tp.lastKey
}

func (tp *telemetryProfiles) setData(profile telemetryDataProfile, key keyboard.Key, now time.Time) bool {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	if tp.isDuplicateLocked(key, now) {
		return false
	}
	tp.data = profile
	tp.lastKey = key.String()
	tp.markHandledLocked(key, now)
	return true
}

func (tp *telemetryProfiles) cycleData(key keyboard.Key, now time.Time) bool {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	if tp.isDuplicateLocked(key, now) {
		return false
	}
	if tp.data >= dataProfileStorm {
		tp.data = dataProfileSine
	} else {
		tp.data++
	}
	tp.lastKey = key.String()
	tp.markHandledLocked(key, now)
	return true
}

func (tp *telemetryProfiles) setStyle(profile telemetryStyleProfile, key keyboard.Key, now time.Time) bool {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	if tp.isDuplicateLocked(key, now) {
		return false
	}
	tp.style = profile
	tp.lastKey = key.String()
	tp.markHandledLocked(key, now)
	return true
}

func (tp *telemetryProfiles) cycleStyle(key keyboard.Key, now time.Time) bool {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	if tp.isDuplicateLocked(key, now) {
		return false
	}
	if tp.style >= styleProfileMinimal {
		tp.style = styleProfileBraille
	} else {
		tp.style++
	}
	tp.lastKey = key.String()
	tp.markHandledLocked(key, now)
	return true
}

func (tp *telemetryProfiles) isDuplicateLocked(key keyboard.Key, now time.Time) bool {
	return tp.lastEventKey == key && !tp.lastEventAt.IsZero() && now.Sub(tp.lastEventAt) < 100*time.Millisecond
}

func (tp *telemetryProfiles) markHandledLocked(key keyboard.Key, now time.Time) {
	tp.lastEventKey = key
	tp.lastEventAt = now
}

func newShipState() *shipState {
	return &shipState{
		warpOnline:     false,
		shieldsPct:     87,
		torpedoSet:     12,
		torpedoLoaded:  12,
		alarmThreshold: 500,
	}
}

func newStatusControls(ship *shipState) (*statusControls, error) {
	sc := &statusControls{ship: ship}

	warp, err := radio.New(
		[]radio.Item{
			{
				Label:            "ON",
				CellOpts:         []cell.Option{cell.FgColor(cell.ColorNumber(118))},
				SelectedCellOpts: []cell.Option{cell.FgColor(cell.ColorNumber(118))},
			},
			{
				Label:            "OFF",
				CellOpts:         []cell.Option{cell.FgColor(cell.ColorNumber(203))},
				SelectedCellOpts: []cell.Option{cell.FgColor(cell.ColorNumber(203))},
			},
		},
		radio.OnChange(func(index int, label string) error {
			_ = label
			ship.setWarpOnline(index == 0)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	shields, err := slider.New(
		slider.Min(1),
		slider.Max(100),
		slider.Value(87),
		slider.Width(18),
		slider.FillCellOpts(cell.FgColor(cell.ColorNumber(221))),
		slider.TrackCellOpts(cell.FgColor(cell.ColorNumber(239))),
		slider.KnobCellOpts(cell.FgColor(cell.ColorWhite)),
		slider.FocusedKnobCellOpts(cell.FgColor(cell.ColorWhite)),
		slider.OnChange(func(value int) error {
			ship.setShieldsPct(value)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	torpedoes, err := dropdown.New(
		dropdown.IntRange(1, 12, 1, "%02d"),
		dropdown.Selected(11),
		dropdown.CellOpts(cell.FgColor(cell.ColorNumber(81))),
		dropdown.FocusedCellOpts(cell.FgColor(cell.ColorNumber(81))),
		dropdown.SelectedCellOpts(cell.FgColor(cell.ColorWhite), cell.BgColor(cell.ColorNumber(60))),
		dropdown.BorderCellOpts(cell.FgColor(cell.ColorNumber(81))),
		dropdown.OnSelect(func(index int, label string) error {
			_ = label
			version := ship.startTorpedoLoad(index + 1)
			afterDelay(torpedoLoadDelay, func() {
				ship.finishTorpedoLoad(version)
			})
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	alarm, err := dropdown.New(
		dropdown.IntRange(200, 600, 50, "%03d"),
		dropdown.Selected(6),
		dropdown.CellOpts(cell.FgColor(cell.ColorNumber(81))),
		dropdown.FocusedCellOpts(cell.FgColor(cell.ColorNumber(81))),
		dropdown.SelectedCellOpts(cell.FgColor(cell.ColorWhite), cell.BgColor(cell.ColorNumber(60))),
		dropdown.BorderCellOpts(cell.FgColor(cell.ColorNumber(81))),
		dropdown.OnSelect(func(index int, label string) error {
			_ = label
			ship.setAlarmThreshold(200 + index*50)
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	var cloak *checkbox.Checkbox
	cloak, err = checkbox.New(
		"",
		checkbox.UseIndicatorSet(checkbox.IndicatorSets.Classic),
		checkbox.CellOpts(cell.FgColor(cell.ColorNumber(87))),
		checkbox.FocusedCellOpts(cell.FgColor(cell.ColorNumber(87))),
		checkbox.CheckedCellOpts(cell.FgColor(cell.ColorNumber(87))),
		checkbox.OnChange(func(checked bool) error {
			_ = checked
			version, targetOnline := ship.startCloakToggle()
			if version == 0 {
				_, _, _, _, _, cloakOnline, cloakEnabling, _, _, _, _ := ship.snapshot()
				cloak.SetChecked(cloakOnline || cloakEnabling)
				return nil
			}
			afterDelay(cloakEnableDelay, func() {
				ship.finishCloakToggle(version, targetOnline)
			})
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	sc.warp = warp
	sc.shields = shields
	sc.torpedoes = torpedoes
	sc.alarm = alarm
	sc.cloak = cloak
	sc.syncFromShip()
	return sc, nil
}

func (sr *styleRegistry) get(id string) panelStyle {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.styles[id]
}

func (sr *styleRegistry) set(id string, style panelStyle) {
	sr.mu.Lock()
	sr.styles[id] = style
	sr.mu.Unlock()
}

// afterDelay runs fn on a timer without blocking the current input callback.
func afterDelay(delay time.Duration, fn func()) {
	go func() {
		time.Sleep(delay)
		fn()
	}()
}

// registerPanelStyles refreshes the animated border effect for every demo pane.
func registerPanelStyles(fx *borderfx.Animator, styles *styleRegistry) {
	forEachDemoPanel(func(id string) {
		fx.Register(id, themedStyle(styles.get(id)))
	})
}

func (ss *shipState) snapshot() (bool, int, int, int, bool, bool, bool, bool, int, bool, int) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.warpOnline, ss.shieldsPct, ss.torpedoSet, ss.torpedoLoaded, ss.torpedoLoading, ss.cloakOnline, ss.cloakEnabling, ss.cloakDisabling, ss.alarmThreshold, ss.redAlert, ss.version
}

func (ss *shipState) setWarpOnline(next bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.warpOnline == next {
		return
	}
	ss.warpOnline = next
	ss.version++
}

func (ss *shipState) setShieldsPct(next int) {
	if next < 1 {
		next = 1
	}
	if next > 100 {
		next = 100
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.shieldsPct == next {
		return
	}
	ss.shieldsPct = next
	ss.version++
}

func (ss *shipState) setAlarmThreshold(next int) {
	if next < 200 {
		next = 200
	}
	if next > 600 {
		next = 600
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.alarmThreshold == next {
		return
	}
	ss.alarmThreshold = next
	ss.version++
}

func (ss *shipState) setRedAlert(active bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.redAlert == active {
		return
	}
	ss.redAlert = active
	ss.version++
}

func (ss *shipState) startTorpedoLoad(next int) int {
	if next < 1 {
		next = 1
	}
	if next > 12 {
		next = 12
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.torpedoSet = next
	ss.torpedoLoading = true
	ss.version++
	return ss.version
}

func (ss *shipState) finishTorpedoLoad(expectedVersion int) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.version != expectedVersion {
		return
	}
	ss.torpedoLoaded = ss.torpedoSet
	ss.torpedoLoading = false
	ss.version++
}

func (ss *shipState) startCloakToggle() (int, bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.cloakEnabling || ss.cloakDisabling {
		return 0, false
	}
	targetOnline := !ss.cloakOnline
	if targetOnline {
		ss.cloakEnabling = true
	} else {
		ss.cloakDisabling = true
	}
	ss.version++
	return ss.version, targetOnline
}

func (ss *shipState) finishCloakToggle(expectedVersion int, targetOnline bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if ss.version != expectedVersion {
		return
	}
	if targetOnline && !ss.cloakEnabling {
		return
	}
	if !targetOnline && !ss.cloakDisabling {
		return
	}
	ss.cloakEnabling = false
	ss.cloakDisabling = false
	ss.cloakOnline = targetOnline
	ss.version++
}

func main() {
	t, err := newTermdashTerminal()
	if err != nil {
		log.Fatal(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Widgets ---

	graphProfiles := newTelemetryProfiles()

	logW, _ := text.New(text.RollContent(), text.WrapAtWords())
	sparkW, _ := linechart.New(
		linechart.BrailleOnly(),
		linechart.DownsampleLTTB(),
		linechart.ThresholdLine(500, cell.FgColor(cell.ColorRed)),
		linechart.XAxisUnscaled(),
		linechart.YAxisCustomScale(120, 700),
		linechart.YAxisAdaptive(),
	)
	statusW, _ := text.New()
	eventsW, _ := text.New(text.RollContent())
	helpW, _ := text.New()

	bootDone := make(chan struct{})
	styles := newStyleRegistry()
	ship := newShipState()
	controls, err := newStatusControls(ship)
	if err != nil {
		log.Fatal(err)
	}
	t.setLoadingBackground(borderfx.NewLoadingBackground())
	t.setLoadingVisible(true)
	t.setLoadingFrame(idSensors, bootSensorFrame(0))
	t.setLoadingFrame(idLCARS, bootStatusFrame(0))
	t.setLoadingFrame(idWarp, bootWarpFrame(0))
	t.setLoadingFrame(idComms, bootEventsFrame(0))
	t.setLoadingFrame(idHelp, bootHelpFrame(0))
	mouseSelectMode := false
	mouseSelectVersion := 0
	var mouseStateMu sync.Mutex
	clicks := &clickTracker{}
	sparkData := &sparkProbe{}
	enableMouse := func() {
		mouseStateMu.Lock()
		defer mouseStateMu.Unlock()
		t.EnableMouseMotion()
	}
	disableMouse := func() {
		mouseStateMu.Lock()
		defer mouseStateMu.Unlock()
		t.tooltip.hide()
		if t.controls != nil {
			t.controls.torpedoes.Close()
			t.controls.alarm.Close()
		}
		t.DisableMouse()
	}
	setMouseSelectMode := func(next bool) {
		mouseStateMu.Lock()
		mouseSelectMode = next
		mouseSelectVersion++
		mouseStateMu.Unlock()
	}
	mouseSelectEnabled := func() bool {
		mouseStateMu.Lock()
		defer mouseStateMu.Unlock()
		return mouseSelectMode
	}
	mouseSelectSnapshot := func() (bool, int) {
		mouseStateMu.Lock()
		defer mouseStateMu.Unlock()
		return mouseSelectMode, mouseSelectVersion
	}

	// --- Layout: rounded borders on every panel ---

	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(55,
			grid.ColWidthPerc(55,
				grid.Widget(logW,
					container.Border(linestyle.Round),
					container.BorderTitle(panelTitles[idSensors].Plain()),
					container.BorderTitleAlignCenter(),
					container.MarginTop(1),
					container.MarginLeft(2),
					container.MarginRight(1),
					container.MarginBottom(1),
					container.PaddingLeft(1),
					container.PaddingRight(1),
					container.Focused(),
					container.ID(idSensors),
				),
			),
			grid.ColWidthPerc(45,
				grid.Widget(statusW,
					container.Border(linestyle.Round),
					container.BorderTitle(panelTitles[idLCARS].Plain()),
					container.MarginTop(1),
					container.MarginLeft(1),
					container.MarginRight(2),
					container.MarginBottom(1),
					container.PaddingLeft(1),
					container.PaddingRight(1),
					container.ID(idLCARS),
				),
			),
		),
		grid.RowHeightPerc(30,
			grid.ColWidthPerc(55,
				grid.Widget(sparkW,
					container.Border(linestyle.Round),
					container.BorderTitle(panelTitles[idWarp].Plain()),
					container.MarginTop(1),
					container.MarginLeft(2),
					container.MarginRight(1),
					container.MarginBottom(1),
					container.PaddingLeft(1),
					container.PaddingRight(1),
					container.ID(idWarp),
				),
			),
			grid.ColWidthPerc(45,
				grid.Widget(eventsW,
					container.Border(linestyle.Round),
					container.BorderTitle(panelTitles[idComms].Plain()),
					container.MarginTop(1),
					container.MarginLeft(1),
					container.MarginRight(2),
					container.MarginBottom(1),
					container.PaddingLeft(1),
					container.PaddingRight(1),
					container.ID(idComms),
				),
			),
		),
		grid.RowHeightPerc(15,
			grid.Widget(helpW,
				container.Border(linestyle.Round),
				container.BorderTitle(panelTitles[idHelp].Plain()),
				container.MarginTop(1),
				container.MarginLeft(2),
				container.MarginRight(2),
				container.MarginBottom(1),
				container.PaddingLeft(1),
				container.PaddingRight(1),
				container.ID(idHelp),
			),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		log.Fatal(err)
	}
	rootOpts := append(gridOpts,
		container.KeyFocusNext(keyboard.KeyTab),
		container.KeyFocusPrevious(keyboard.KeyBacktab),
	)
	c, err := container.New(t, rootOpts...)
	if err != nil {
		log.Fatal(err)
	}
	t.setStatusControls(controls, func() bool { return c.ActiveID() == idLCARS })

	// --- Animated borders ---

	fx := borderfx.NewAnimator(c)
	fx.SetInactiveStyle(inactivePins)
	go func() {
		_ = fx.Run(ctx)
	}()
	go runBootSequence(ctx, c, fx, styles, graphProfiles, ship, t, logW, statusW, eventsW, helpW, bootDone)
	go borderfx.NewTitleController(c, fx, panelTitles, func(id string) *borderfx.Effect {
		return themedStyle(styles.get(id))
	}).Run(ctx, bootDone)
	go trackStatusPanel(ctx, c, statusW, ship, bootDone)

	go feedLog(ctx, logW, bootDone)
	go feedSparkline(ctx, sparkW, sparkData, ship, graphProfiles, bootDone)
	go feedEvents(ctx, eventsW, bootDone)
	go refreshSparkTooltip(ctx, t, sparkW, mouseSelectEnabled, bootDone)
	t.setKeyboardHandler(func(k *terminalapi.Keyboard) {
		if handleGraphProfileKey(k.Key, graphProfiles) {
			reseedSparkline(sparkW, sparkData, ship, graphProfiles)
			writeHelp(helpW, true, mouseSelectEnabled(), graphProfileStatus(graphProfiles))
		}
	})

	// --- Keyboard ---

	quitter := func(k *terminalapi.Keyboard) {
		activeID := c.ActiveID()
		if activeID == "" {
			activeID = idSensors
		}
		if handleGraphProfileKey(k.Key, graphProfiles) {
			reseedSparkline(sparkW, sparkData, ship, graphProfiles)
			writeHelp(helpW, true, mouseSelectEnabled(), graphProfileStatus(graphProfiles))
			return
		}
		switch k.Key {
		case keyboard.KeyEsc, 'q', 'Q':
			cancel()
		case 'm', 'M':
			next := !mouseSelectEnabled()
			setMouseSelectMode(next)
			t.tooltip.hide()
			controls.torpedoes.Close()
			controls.alarm.Close()
			if next {
				disableMouse()
			} else {
				enableMouse()
			}
			writeHelp(helpW, true, next, graphProfileStatus(graphProfiles))
		case 'z', 'Z':
			styles.set(activeID, styleInterlace)
			fx.Register(activeID, themedStyle(styleInterlace))
		case 'x', 'X':
			styles.set(activeID, styleFire)
			fx.Register(activeID, themedStyle(styleFire))
		case 'c', 'C':
			styles.set(activeID, styleIce)
			fx.Register(activeID, themedStyle(styleIce))
		case 'v', 'V':
			styles.set(activeID, styleRainbow)
			fx.Register(activeID, themedStyle(styleRainbow))
		case 'b', 'B':
			styles.set(activeID, styleMatrix)
			fx.Register(activeID, themedStyle(styleMatrix))
		case 'n', 'N':
			styles.set(activeID, styleSynthwave)
			fx.Register(activeID, themedStyle(styleSynthwave))
		case 'g', 'G':
			styles.set(activeID, styleNeon)
			fx.Register(activeID, themedStyle(styleNeon))
		}
	}

	scheduleFocusRestore := func() {
		scheduleFocusModeRestore(
			ctx,
			transientSelectDuration,
			mouseSelectSnapshot,
			setMouseSelectMode,
			enableMouse,
			func(selectMode bool) { writeHelp(helpW, true, selectMode, graphProfileStatus(graphProfiles)) },
		)
	}

	mouseWatcher := func(m *terminalapi.Mouse) {
		handleDemoMouse(
			m,
			time.Now(),
			clicks,
			mouseSelectEnabled,
			setMouseSelectMode,
			enableMouse,
			disableMouse,
			c.ActiveID,
			func(pos image.Point) string { return panelAt(t.Size(), pos) },
			func(pos image.Point) bool { return focusPanelAt(c, t.Size(), pos) },
			func(pos image.Point) bool {
				if controls.handleMouse(t.Size(), pos, c.ActiveID() == idLCARS) {
					return true
				}
				return handleStatusControlClick(t.Size(), pos, ship)
			},
			func(selectMode bool) { writeHelp(helpW, true, selectMode, graphProfileStatus(graphProfiles)) },
			func(pos image.Point) {
				t.tooltip.update(pos, sparklineReadout(sparkW, t.Size(), pos))
			},
			scheduleFocusRestore,
		)
	}

	enableMouse()
	go func() {
		time.Sleep(180 * time.Millisecond)
		if !mouseSelectEnabled() {
			enableMouse()
		}
	}()

	if err := termdash.Run(ctx, t, c,
		termdash.KeyboardSubscriber(quitter),
		termdash.MouseSubscriber(mouseWatcher),
		termdash.RedrawInterval(60*time.Millisecond),
	); err != nil {
		log.Fatal(err)
	}
}

// -------------------------------------------------------------------

func isFocusMouse(m *terminalapi.Mouse) bool {
	return m.Button == mouse.ButtonLeft
}

func focusPanelAt(c *container.Container, size, pos image.Point) bool {
	if targetID := panelAt(size, pos); targetID != "" {
		_ = c.Update(targetID, container.Focused())
		return true
	}
	return false
}

// textSelectablePanel reports whether a panel primarily contains copyable text.
func textSelectablePanel(id string) bool {
	switch id {
	case idSensors, idComms, idHelp:
		return true
	default:
		return false
	}
}

func handleDemoMouse(
	m *terminalapi.Mouse,
	now time.Time,
	clicks *clickTracker,
	mouseSelectEnabled func() bool,
	setMouseSelectMode func(bool),
	enableMouse func(),
	disableMouse func(),
	activePanel func() string,
	panelIDAt func(image.Point) string,
	focusAt func(image.Point) bool,
	activateAt func(image.Point) bool,
	writeHelp func(bool),
	hoverAt func(image.Point),
	scheduleFocusRestore func(),
) {
	if !mouseSelectEnabled() {
		hoverAt(m.Position)
	}

	if m.Button != mouse.ButtonRelease {
		if mouseSelectEnabled() {
			return
		}
		if isFocusMouse(m) {
			targetID := panelIDAt(m.Position)
			if targetID != "" && activePanel() == targetID && textSelectablePanel(targetID) {
				setMouseSelectMode(true)
				disableMouse()
				writeHelp(true)
				scheduleFocusRestore()
				return
			}
			clicks.press(m.Position)
			focusAt(m.Position)
			activateAt(m.Position)
		}
		return
	}

	clickCount := clicks.releaseCount(now, m.Position)
	if clickCount >= 2 {
		setMouseSelectMode(true)
		disableMouse()
		writeHelp(true)
		scheduleFocusRestore()
		return
	}
	if mouseSelectEnabled() {
		return
	}
}

func (ht *hoverTooltip) update(anchor image.Point, text string) {
	ht.mu.Lock()
	defer ht.mu.Unlock()
	ht.anchor = anchor
	ht.text = text
	ht.visible = text != ""
}

func (ht *hoverTooltip) snapshot() (image.Point, bool) {
	ht.mu.Lock()
	defer ht.mu.Unlock()
	return ht.anchor, ht.visible
}

func (ht *hoverTooltip) hide() {
	ht.update(image.Point{}, "")
}

func (ht *hoverTooltip) draw(t terminalapi.Terminal) {
	ht.mu.Lock()
	text := ht.text
	anchor := ht.anchor
	visible := ht.visible
	ht.mu.Unlock()
	if !visible || text == "" {
		return
	}

	pos, ok := tooltipOrigin(t.Size(), anchor, text)
	if !ok {
		return
	}

	opts := []cell.Option{
		cell.FgColor(cell.ColorNumber(87)),
		cell.BgColor(cell.ColorNumber(236)),
	}
	for row, line := range strings.Split(text, "\n") {
		for col, r := range line {
			_ = t.SetCell(image.Point{X: pos.X + col, Y: pos.Y + row}, r, opts...)
		}
	}
}

func (sc *statusControls) syncFromShip() {
	warpOnline, shieldsPct, torpedoSet, _, _, cloakOnline, cloakEnabling, _, alarmThreshold, _, _ := sc.ship.snapshot()
	if warpOnline {
		_ = sc.warp.SetSelected(0)
	} else {
		_ = sc.warp.SetSelected(1)
		sc.torpedoes.Close()
		sc.alarm.Close()
	}
	sc.shields.SetValue(shieldsPct)
	_ = sc.torpedoes.SetSelected(torpedoSet - 1)
	_ = sc.alarm.SetSelected((alarmThreshold - 200) / 50)
	sc.cloak.SetChecked(cloakOnline || cloakEnabling)
}

func (sc *statusControls) draw(t terminalapi.Terminal, size image.Point, active bool) {
	if !active {
		return
	}

	sc.syncFromShip()
	warpOnline, shieldsPct, _, _, _, _, _, _, _, _, _ := sc.ship.snapshot()
	if !warpOnline {
		sc.drawWarpOffline(t, size, shieldsPct)
		return
	}

	for _, target := range []struct {
		w    widgetapi.Widget
		rect image.Rectangle
	}{
		{w: sc.warp, rect: sc.warpArea(size)},
		{w: sc.shields, rect: sc.shieldsArea(size)},
		{w: sc.torpedoes, rect: sc.torpedoArea(size)},
		{w: sc.alarm, rect: sc.alarmArea(size)},
		{w: sc.cloak, rect: sc.cloakArea(size)},
	} {
		if target.rect.Empty() {
			continue
		}
		cvs, err := canvas.New(target.rect)
		if err != nil {
			continue
		}
		if err := target.w.Draw(cvs, &widgetapi.Meta{Focused: true}); err != nil {
			continue
		}
		_ = cvs.Apply(t)
	}
}

func (sc *statusControls) handleMouse(size, pos image.Point, active bool) bool {
	if !active {
		return false
	}

	sc.syncFromShip()
	warpOnline, _, _, _, _, _, _, _, _, _, _ := sc.ship.snapshot()
	warpArea := sc.warpArea(size)
	shieldsArea := sc.shieldsArea(size)
	torpedoArea := sc.torpedoArea(size)
	alarmArea := sc.alarmArea(size)
	cloakArea := sc.cloakArea(size)

	if !warpOnline {
		if !pos.In(warpArea) {
			sc.torpedoes.Close()
			sc.alarm.Close()
			return false
		}
		rel := pos.Sub(warpArea.Min)
		_ = sc.warp.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: rel}, &widgetapi.EventMeta{Focused: true})
		return true
	}

	if pos.In(sc.torpedoTrigger(size)) {
		sc.alarm.Close()
	}
	if pos.In(sc.alarmTrigger(size)) {
		sc.torpedoes.Close()
	}

	handled := pos.In(warpArea) || pos.In(shieldsArea) || pos.In(torpedoArea) || pos.In(alarmArea) || pos.In(cloakArea)
	for _, target := range []struct {
		w    widgetapi.Widget
		rect image.Rectangle
	}{
		{w: sc.warp, rect: warpArea},
		{w: sc.shields, rect: shieldsArea},
		{w: sc.torpedoes, rect: torpedoArea},
		{w: sc.alarm, rect: alarmArea},
		{w: sc.cloak, rect: cloakArea},
	} {
		rel := pos.Sub(target.rect.Min)
		_ = target.w.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: rel}, &widgetapi.EventMeta{Focused: true})
	}
	return handled
}

func (sc *statusControls) torpedoTrigger(size image.Point) image.Rectangle {
	_, _, _, torpedoRect, _, _ := statusControlRects(size, true)
	return torpedoRect
}

func (sc *statusControls) alarmTrigger(size image.Point) image.Rectangle {
	_, _, _, _, _, alarmRect := statusControlRects(size, true)
	return alarmRect
}

func (sc *statusControls) cloakArea(size image.Point) image.Rectangle {
	_, _, _, _, cloakRect, _ := statusControlRects(size, true)
	opts := sc.cloak.Options()
	return image.Rect(cloakRect.Min.X, cloakRect.Min.Y, cloakRect.Min.X+opts.MinimumSize.X, cloakRect.Min.Y+opts.MinimumSize.Y)
}

func (sc *statusControls) warpArea(size image.Point) image.Rectangle {
	onRect, _, _, _, _, _ := statusControlRects(size, true)
	opts := sc.warp.Options()
	return image.Rect(onRect.Min.X, onRect.Min.Y, onRect.Min.X+opts.MinimumSize.X, onRect.Min.Y+opts.MinimumSize.Y)
}

func (sc *statusControls) shieldsArea(size image.Point) image.Rectangle {
	_, _, sliderRect, _, _, _ := statusControlRects(size, true)
	opts := sc.shields.Options()
	return image.Rect(sliderRect.Min.X, sliderRect.Min.Y, sliderRect.Min.X+opts.MinimumSize.X, sliderRect.Min.Y+opts.MinimumSize.Y)
}

func (sc *statusControls) torpedoArea(size image.Point) image.Rectangle {
	trigger := sc.torpedoTrigger(size)
	canvasSize := sc.torpedoes.CanvasSize(size.Y - trigger.Min.Y)
	return image.Rect(trigger.Min.X, trigger.Min.Y, trigger.Min.X+canvasSize.X, trigger.Min.Y+canvasSize.Y)
}

func (sc *statusControls) alarmArea(size image.Point) image.Rectangle {
	trigger := sc.alarmTrigger(size)
	canvasSize := sc.alarm.CanvasSize(size.Y - trigger.Min.Y)
	return image.Rect(trigger.Min.X, trigger.Min.Y, trigger.Min.X+canvasSize.X, trigger.Min.Y+canvasSize.Y)
}

// drawWarpOffline renders dimmed LCARS controls while the warp core is offline.
// The warp radio row (ON/OFF) is intentionally omitted here: it is already
// rendered by the text widget at the correct column position (which shifts by
// one when "OFFLINE" replaces "ONLINE" in the status label). The remaining
// controls (shields, torpedoes, cloak, alarm) use fixed column positions and
// are safe to paint via direct terminal writes.
func (sc *statusControls) drawWarpOffline(t terminalapi.Terminal, size image.Point, shieldsPct int) {
	muted := cell.ColorNumber(241)
	sc.drawControlText(t, sc.shieldsArea(size).Min, shieldsSlider(shieldsPct), muted)
	sc.drawControlText(t, sc.torpedoTrigger(size).Min, sc.torpedoes.TriggerText(), muted)
	sc.drawControlText(t, sc.cloakArea(size).Min, sc.cloak.Text(), muted)
	sc.drawControlText(t, sc.alarmTrigger(size).Min, sc.alarm.TriggerText(), muted)
}

// drawControlText writes a one-line control overlay directly to the terminal.
func (sc *statusControls) drawControlText(t terminalapi.Terminal, pos image.Point, text string, fg cell.Color) {
	for i, r := range []rune(text) {
		_ = t.SetCell(image.Point{X: pos.X + i, Y: pos.Y}, r, cell.FgColor(fg))
	}
}

func tooltipOrigin(size, anchor image.Point, text string) (image.Point, bool) {
	lines := strings.Split(text, "\n")
	width := 0
	for _, line := range lines {
		lineWidth := len([]rune(line))
		if lineWidth > width {
			width = lineWidth
		}
	}
	height := len(lines)
	if width == 0 || height == 0 || size.X <= 0 || size.Y <= 0 || width > size.X || height > size.Y {
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

func refreshSparkTooltip(ctx context.Context, t *demoTerminal, chart *linechart.LineChart, mouseSelectEnabled func() bool, bootDone <-chan struct{}) {
	select {
	case <-ctx.Done():
		return
	case <-bootDone:
	}

	ticker := time.NewTicker(180 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if mouseSelectEnabled() {
				t.tooltip.hide()
				continue
			}
			refreshHoverTooltip(t.tooltip, chart, t.Size())
		}
	}
}

func refreshHoverTooltip(tooltip *hoverTooltip, chart *linechart.LineChart, size image.Point) {
	anchor, visible := tooltip.snapshot()
	if !visible {
		return
	}
	tooltip.update(anchor, sparklineReadout(chart, size, anchor))
}

func scheduleFocusModeRestore(
	ctx context.Context,
	delay time.Duration,
	mouseSelectSnapshot func() (bool, int),
	setMouseSelectMode func(bool),
	enableMouse func(),
	writeHelp func(bool),
) {
	_, version := mouseSelectSnapshot()
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		selected, currentVersion := mouseSelectSnapshot()
		if !selected || currentVersion != version {
			return
		}
		setMouseSelectMode(false)
		enableMouse()
		writeHelp(false)
	}()
}

func runBootSequence(ctx context.Context, c *container.Container, fx *borderfx.Animator, styles *styleRegistry, graphProfiles *telemetryProfiles, ship *shipState, t *demoTerminal, logW, statusW, eventsW, helpW *text.Text, bootDone chan struct{}) {
	forEachDemoPanel(func(id string) {
		fx.Register(id, borderfx.Cycle([]cell.Color{inactiveEdge}))
		_ = c.Update(id,
			container.BorderColor(inactiveEdge),
			container.FocusedColor(primaryDim),
			container.TitleColor(inactiveText),
			container.TitleFocusedColor(activeTitle),
		)
	})

	for i := 0; i < 4; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}
		t.setLoadingFrame(idSensors, bootSensorFrame(i))
		t.setLoadingFrame(idLCARS, bootStatusFrame(i))
		t.setLoadingFrame(idWarp, bootWarpFrame(i))
		t.setLoadingFrame(idComms, bootEventsFrame(i))
		t.setLoadingFrame(idHelp, bootHelpFrame(i))
		time.Sleep(260 * time.Millisecond)
	}

	time.Sleep(650 * time.Millisecond)

	logW.Reset()
	statusW.Reset()
	eventsW.Reset()
	helpW.Reset()
	warpOnline, shieldsPct, torpedoSet, torpedoLoaded, torpedoLoading, cloakOnline, cloakEnabling, cloakDisabling, alarmThreshold, redAlert, _ := ship.snapshot()
	writeStatus(statusW, true, warpOnline, shieldsPct, torpedoSet, torpedoLoaded, torpedoLoading, cloakOnline, cloakEnabling, cloakDisabling, alarmThreshold, redAlert, true)
	writeHelp(helpW, true, false, graphProfileStatus(graphProfiles))

	registerPanelStyles(fx, styles)

	_ = c.Update(idSensors,
		container.BorderColor(primaryDim),
		container.FocusedColor(primaryDim),
		container.TitleColor(activeTitle),
		container.TitleFocusedColor(activeTitle),
	)
	for _, id := range []string{idLCARS, idWarp, idComms, idHelp} {
		_ = c.Update(id,
			container.BorderColor(inactiveEdge),
			container.FocusedColor(primaryDim),
			container.TitleColor(inactiveText),
			container.TitleFocusedColor(activeTitle),
		)
	}

	t.setLoadingVisible(false)
	close(bootDone)
}

func trackStatusPanel(ctx context.Context, c *container.Container, statusW *text.Text, ship *shipState, bootDone <-chan struct{}) {
	select {
	case <-ctx.Done():
		return
	case <-bootDone:
	}

	lastActive := ""
	lastVersion := -1
	lastBlink := -1
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		activeID := c.ActiveID()
		warpOnline, shieldsPct, torpedoSet, torpedoLoaded, torpedoLoading, cloakOnline, cloakEnabling, cloakDisabling, alarmThreshold, redAlert, version := ship.snapshot()
		blinkPhase := 0
		if redAlert {
			blinkPhase = int(time.Now().UnixNano() / int64(350*time.Millisecond) % 2)
		}
		if activeID != lastActive || version != lastVersion || blinkPhase != lastBlink {
			writeStatus(statusW, activeID != idLCARS, warpOnline, shieldsPct, torpedoSet, torpedoLoaded, torpedoLoading, cloakOnline, cloakEnabling, cloakDisabling, alarmThreshold, redAlert, blinkPhase == 0)
			lastActive = activeID
			lastVersion = version
			lastBlink = blinkPhase
		}
		time.Sleep(90 * time.Millisecond)
	}
}

func closeTo(a, b image.Point) bool {
	dx := a.X - b.X
	if dx < 0 {
		dx = -dx
	}
	dy := a.Y - b.Y
	if dy < 0 {
		dy = -dy
	}
	return dx <= 2 && dy <= 1
}

func panelAt(size, pos image.Point) string {
	for id, rect := range panelRects(size) {
		if pos.In(rect) {
			return id
		}
	}
	return ""
}

func panelRects(size image.Point) map[string]image.Rectangle {
	width, height := size.X, size.Y
	if width <= 0 || height <= 0 {
		return nil
	}

	topH := (height * 55) / 100
	midH := (height * 30) / 100
	if topH < 1 {
		topH = 1
	}
	if midH < 1 {
		midH = 1
	}
	if topH+midH >= height {
		midH = height - topH - 1
		if midH < 1 {
			midH = 1
		}
	}
	botY := topH + midH

	leftW := (width * 55) / 100
	if leftW < 1 {
		leftW = 1
	}
	if leftW >= width {
		leftW = width - 1
	}

	return map[string]image.Rectangle{
		idSensors: image.Rect(0, 0, leftW, topH),
		idLCARS:   image.Rect(leftW, 0, width, topH),
		idWarp:    image.Rect(0, topH, leftW, botY),
		idComms:   image.Rect(leftW, topH, width, botY),
		idHelp:    image.Rect(0, botY, width, height),
	}
}

// loadingPanelRect returns the drawable content area used by the loading
// overlay inside a bordered panel.
func loadingPanelRect(size image.Point, id string) image.Rectangle {
	rect := panelRects(size)[id]
	if rect.Empty() {
		return image.Rectangle{}
	}
	return image.Rect(rect.Min.X+1, rect.Min.Y+1, rect.Max.X-1, rect.Max.Y-1)
}

func sparklineGraphRect(size image.Point) image.Rectangle {
	rect := panelRects(size)[idWarp]
	if rect.Empty() {
		return image.Rectangle{}
	}
	return image.Rect(
		rect.Min.X+2,
		rect.Min.Y+2,
		rect.Max.X-2,
		rect.Max.Y-2,
	)
}

func statusContentOrigin(size image.Point) image.Point {
	rect := panelRects(size)[idLCARS]
	return image.Point{X: rect.Min.X + 3, Y: rect.Min.Y + 2}
}

func radioChoiceText(selected bool, label string) string {
	indicator := radio.IndicatorSets.Circle.Unselected
	if selected {
		indicator = radio.IndicatorSets.Circle.Selected
	}
	return indicator + " " + label
}

func dropdownPreview(label string) string {
	return "[" + label + " " + string(dropdown.GlyphProfiles.Classic.ClosedArrow) + "]"
}

func checkboxIndicator(checked bool) string {
	if checked {
		return checkbox.IndicatorSets.Classic.Checked
	}
	return checkbox.IndicatorSets.Classic.Unchecked
}

func statusControlRects(size image.Point, warpOnline bool) (image.Rectangle, image.Rectangle, image.Rectangle, image.Rectangle, image.Rectangle, image.Rectangle) {
	origin := statusContentOrigin(size)
	lineY := origin.Y + 1
	stateText := "ONLINE"
	if !warpOnline {
		stateText = "OFFLINE"
	}
	const prefix = "  WARP CORE       "
	const spacer = "   "
	onLabel := radioChoiceText(true, "ON")
	offLabel := radioChoiceText(false, "OFF")
	const shieldPrefix = "  SHIELDS         000%  ["
	const shieldWidth = 18
	const torpedoPrefix = "  TORPEDOES       "
	torpedoLabel := dropdownPreview("12")
	const cloakPrefix = "  CLOAK           "
	cloakLabel := checkboxIndicator(false)
	const alarmPrefix = "  ALARM           "
	alarmLabel := dropdownPreview("500")

	onX := origin.X + len([]rune(prefix)) + len([]rune(stateText)) + len([]rune(spacer))
	offX := onX + len([]rune(onLabel)) + len([]rune(spacer))
	sliderX := origin.X + len([]rune(shieldPrefix))
	sliderY := origin.Y + 2
	torpedoX := origin.X + len([]rune(torpedoPrefix))
	torpedoY := origin.Y + 6
	cloakX := origin.X + len([]rune(cloakPrefix))
	cloakY := origin.Y + 7
	alarmX := origin.X + len([]rune(alarmPrefix))
	alarmY := origin.Y + 8
	return image.Rect(onX, lineY, onX+len([]rune(onLabel)), lineY+1),
		image.Rect(offX, lineY, offX+len([]rune(offLabel)), lineY+1),
		image.Rect(sliderX, sliderY, sliderX+shieldWidth, sliderY+1),
		image.Rect(torpedoX, torpedoY, torpedoX+len([]rune(torpedoLabel)), torpedoY+1),
		image.Rect(cloakX, cloakY, cloakX+len([]rune(cloakLabel)), cloakY+1),
		image.Rect(alarmX, alarmY, alarmX+len([]rune(alarmLabel)), alarmY+1)
}

func handleStatusControlClick(size, pos image.Point, ship *shipState) bool {
	warpOnline, _, _, _, _, _, _, _, _, _, _ := ship.snapshot()
	onRect, offRect, _, _, _, _ := statusControlRects(size, warpOnline)
	switch {
	case pos.In(onRect):
		ship.setWarpOnline(true)
		return true
	case pos.In(offRect):
		ship.setWarpOnline(false)
		return true
	default:
		return false
	}
}

func (sp *sparkProbe) add(v int) {
	sp.addBatch([]int{v})
}

func (sp *sparkProbe) addBatch(values []int) {
	if len(values) == 0 {
		return
	}
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.values = append(sp.values, values...)
	if len(sp.values) > 512 {
		copy(sp.values, sp.values[len(sp.values)-512:])
		sp.values = sp.values[:512]
	}
}

func (sp *sparkProbe) clear() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.values = nil
}

// floatValues returns a copy of the probe history as float64 samples.
func (sp *sparkProbe) floatValues() []float64 {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	values := make([]float64, len(sp.values))
	for i, value := range sp.values {
		values[i] = float64(value)
	}
	return values
}

// sparklineReadout returns an X/Y hover readout only when the cursor is over a
// rendered line segment inside the warp graph.
func sparklineReadout(chart *linechart.LineChart, size, pos image.Point) string {
	graph := sparklineGraphRect(size)
	if !pos.In(graph) || graph.Dx() <= 0 || graph.Dy() <= 0 || chart == nil {
		return ""
	}

	sample, ok := chart.ValueAt(
		image.Point{X: graph.Dx(), Y: graph.Dy()},
		image.Point{X: pos.X - graph.Min.X, Y: pos.Y - graph.Min.Y},
	)
	if !ok {
		return ""
	}
	return fmt.Sprintf(" X: %03d \n Y: %03d ", sample.X, int(math.Round(sample.Y)))
}

func (ct *clickTracker) press(pos image.Point) {
	ct.pressed = true
	ct.pressPos = pos
}

// releaseCount returns the number of consecutive release clicks near the same
// location within the selection cadence window.
func (ct *clickTracker) releaseCount(now time.Time, pos image.Point) int {
	if !ct.pressed || !closeTo(ct.pressPos, pos) {
		ct.pressed = false
		return 0
	}
	ct.pressed = false

	if ct.lastAt.IsZero() || now.Sub(ct.lastAt) > 475*time.Millisecond || !closeTo(ct.lastPos, pos) {
		ct.count = 1
		ct.lastAt = now
		ct.lastPos = pos
		return ct.count
	}

	ct.count++
	ct.lastAt = now
	ct.lastPos = pos
	return ct.count
}

type sparkProfileStyle struct {
	color cell.Color
}

func sparkStyleForProfile(profile telemetryStyleProfile) sparkProfileStyle {
	switch profile {
	case styleProfileNeedle:
		return sparkProfileStyle{
			color: cell.ColorNumber(116),
		}
	case styleProfileBars:
		return sparkProfileStyle{
			color: cell.ColorNumber(117),
		}
	case styleProfileMatrix:
		return sparkProfileStyle{
			color: cell.ColorNumber(118),
		}
	case styleProfileMinimal:
		return sparkProfileStyle{
			color: cell.ColorNumber(250),
		}
	default:
		return sparkProfileStyle{
			color: cell.ColorNumber(117),
		}
	}
}

func isDataProfileKey(key keyboard.Key) bool {
	_, ok := dataProfileForKey(key)
	return ok
}

func isStyleProfileKey(key keyboard.Key) bool {
	_, ok := styleProfileForKey(key)
	return ok
}

func handleGraphProfileKey(key keyboard.Key, profiles *telemetryProfiles) bool {
	now := time.Now()
	if profile, ok := dataProfileForKey(key); ok {
		return profiles.setData(profile, key, now)
	}
	if profile, ok := styleProfileForKey(key); ok {
		return profiles.setStyle(profile, key, now)
	}
	switch key {
	case 'd', 'D':
		return profiles.cycleData(key, now)
	case 's', 'S':
		return profiles.cycleStyle(key, now)
	default:
		return false
	}
}

func dataProfileForKey(key keyboard.Key) (telemetryDataProfile, bool) {
	switch key {
	case '1', '!', keyboard.KeyEnd:
		return dataProfileBurst, true
	case '3', '#', '"', keyboard.KeyPgDn:
		return dataProfileSaw, true
	case '5', '%':
		return dataProfilePulse, true
	case '7', '&', keyboard.KeyHome:
		return dataProfileStorm, true
	case '9', '(', keyboard.KeyPgUp:
		return dataProfileSine, true
	default:
		return dataProfileSine, false
	}
}

func styleProfileForKey(key keyboard.Key) (telemetryStyleProfile, bool) {
	switch key {
	case '2', '@', 'é', keyboard.KeyArrowDown:
		return styleProfileNeedle, true
	case '4', '$', '\'', keyboard.KeyArrowLeft:
		return styleProfileBars, true
	case '6', '^', '-', keyboard.KeyArrowRight:
		return styleProfileMatrix, true
	case '8', '*', '_', keyboard.KeyArrowUp:
		return styleProfileMinimal, true
	case '0', ')', 'à', keyboard.KeyInsert:
		return styleProfileBraille, true
	default:
		return styleProfileBraille, false
	}
}

func dataProfileName(profile telemetryDataProfile) string {
	switch profile {
	case dataProfileBurst:
		return "Burst"
	case dataProfileSaw:
		return "Saw"
	case dataProfilePulse:
		return "Pulse"
	case dataProfileStorm:
		return "Storm"
	default:
		return "Sine"
	}
}

func styleProfileName(profile telemetryStyleProfile) string {
	switch profile {
	case styleProfileNeedle:
		return "Needle"
	case styleProfileBars:
		return "Bars"
	case styleProfileMatrix:
		return "Matrix"
	case styleProfileMinimal:
		return "Minimal"
	default:
		return "Braille"
	}
}

func graphProfileStatus(tp *telemetryProfiles) string {
	data, style, lastKey := tp.snapshot()
	if lastKey == "" {
		return fmt.Sprintf("Data=%s  Style=%s", dataProfileName(data), styleProfileName(style))
	}
	return fmt.Sprintf("Data=%s  Style=%s  Key=%s", dataProfileName(data), styleProfileName(style), lastKey)
}

func inactivePins(_ string, bc container.BorderCell) container.BorderCellStyle {
	if (bc.Point.X == bc.Border.Min.X || bc.Point.X == bc.Border.Max.X-1) &&
		(bc.Point.Y == bc.Border.Min.Y || bc.Point.Y == bc.Border.Max.Y-1) {
		return container.BorderCellStyle{
			Rune:     '◉',
			CellOpts: []cell.Option{cell.FgColor(inactiveText)},
		}
	}
	return container.BorderCellStyle{
		Rune:     bc.Rune,
		CellOpts: []cell.Option{cell.FgColor(inactiveEdge)},
	}
}

// bootSensorFrame returns the staged loading copy for the sensor window.
func bootSensorFrame(stage int) string {
	lines := []string{
		" :: carrier lock ::\n\n  preparing window .............\n\n  signal lattice ...............\n\n  telemetry uplink .............\n",
		" :: carrier lock ::\n  ::::::::::::::::::::::::::::::\n  preparing window .............\n  ::::::::::::::::::::::::::::::\n  telemetry uplink .............\n",
		" :: carrier lock ::\n  preparing window .... staged\n  signal lattice ...... acquired\n  phase bus ........... stable\n  focus plane ......... warming\n",
		" :: carrier lock ::\n  preparing window .... aligned\n  telemetry uplink .... synced\n  render plane ........ ready\n  interface gate ...... opening\n",
	}
	if stage < 0 {
		stage = 0
	}
	if stage >= len(lines) {
		stage = len(lines) - 1
	}
	return lines[stage]
}

// bootStatusFrame returns the staged loading copy for the LCARS window.
func bootStatusFrame(stage int) string {
	lines := [][]string{
		{
			"  preparing telemetry window",
			"  warp core .......... offline",
			"  shield grid ....... cold start",
			"  threat board ...... asleep",
		},
		{
			"  preparing telemetry window",
			"  warp core .......... offline",
			"  shield grid ....... parking",
			"  threat board ...... indexing",
		},
		{
			"  preparing telemetry window",
			"  warp core .......... offline",
			"  shield grid ....... standby",
			"  threat board ...... queued",
		},
		{
			"  telemetry window ready",
			"  warp core .......... offline",
			"  shield grid ....... standby",
			"  threat board ...... standing by",
		},
	}
	if stage < 0 {
		stage = 0
	}
	if stage >= len(lines) {
		stage = len(lines) - 1
	}
	return strings.Join(lines[stage], "\n\n") + "\n"
}

// bootWarpFrame returns the staged loading copy for the warp graph window.
func bootWarpFrame(stage int) string {
	lines := [][]string{
		{
			"  preparing flux graph",
			"  subspace trace .... dark",
			"  sampling lattice .. interlaced",
			"  threshold rail .... parked",
		},
		{
			"  preparing flux graph",
			"  subspace trace .... buffering",
			"  sampling lattice .. aligning",
			"  threshold rail .... parked",
		},
		{
			"  flux graph warming",
			"  subspace trace .... staged",
			"  sampling lattice .. synchronized",
			"  threshold rail .... watching",
		},
		{
			"  flux graph ready",
			"  subspace trace .... offline",
			"  sampling lattice .. synchronized",
			"  threshold rail .... watching",
		},
	}
	if stage < 0 {
		stage = 0
	}
	if stage >= len(lines) {
		stage = len(lines) - 1
	}
	return strings.Join(lines[stage], "\n\n") + "\n"
}

func bootEventsFrame(stage int) string {
	lines := [][]string{
		{
			"  preparing comms window",
			"  IN  15:04  carrier trace buffering",
			"  OPS 15:04  subspace archive mounting",
			"  NAV 15:04  relay paths reserving",
		},
		{
			"  preparing comms window",
			"  IN  15:04  carrier trace buffering",
			"  OPS 15:04  priority channels mapping",
			"  NAV 15:04  relay paths reserving",
		},
		{
			"  preparing comms window",
			"  IN  15:04  archive mounted",
			"  OPS 15:04  priority channels locked",
			"  NAV 15:04  relay paths green",
		},
		{
			"  comms window ready",
			"  IN  15:04  archive mounted",
			"  OPS 15:04  priority channels locked",
			"  NAV 15:04  relay paths green",
		},
	}
	if stage < 0 {
		stage = 0
	}
	if stage >= len(lines) {
		stage = len(lines) - 1
	}
	return strings.Join(lines[stage], "\n\n") + "\n"
}

func bootHelpFrame(stage int) string {
	lines := []string{
		"  preparing control window\n\n  scanlines settling\n\n  awaiting bridge sync\n",
		"  preparing control window\n\n  scanlines settling\n\n  routing input lattice\n",
		"  preparing control window\n\n  focus map warming\n\n  routing input lattice\n",
		"  control window ready\n\n  focus map warmed\n\n  operator handoff pending\n",
	}
	if stage < 0 {
		stage = 0
	}
	if stage >= len(lines) {
		stage = len(lines) - 1
	}
	return lines[stage]
}

func writeStatus(w *text.Text, inactive bool, warpOnline bool, shieldsPct, torpedoSet, torpedoLoaded int, torpedoLoading, cloakOnline, cloakEnabling, cloakDisabling bool, alarmThreshold int, redAlert bool, blinkOn bool) {
	muted := cell.ColorNumber(246)
	green := cell.ColorGreen
	yellow := cell.ColorYellow
	cyan := cell.ColorCyan
	red := cell.ColorRed
	onColor := cell.ColorNumber(118)
	offColor := cell.ColorNumber(203)
	controlMuted := cell.ColorNumber(241)
	sliderFill := cell.ColorNumber(221)
	sliderTrack := cell.ColorNumber(239)
	sliderKnob := cell.ColorWhite
	loadedColor := cell.ColorNumber(118)
	dropdownColor := cell.ColorNumber(81)
	cloakColor := cell.ColorNumber(87)
	alertColor := cell.ColorRed
	poweredDown := !warpOnline
	if inactive || poweredDown {
		green, yellow, cyan, red = muted, muted, muted, muted
		onColor, offColor, controlMuted = muted, muted, muted
		sliderFill, sliderTrack, sliderKnob = muted, muted, muted
		loadedColor, dropdownColor, cloakColor, alertColor = muted, muted, muted, muted
	}
	statusLabel := "ONLINE"
	statusColor := green
	onLabel := radioChoiceText(true, "ON")
	offLabel := radioChoiceText(false, "OFF")
	if !warpOnline {
		statusLabel = "OFFLINE"
		statusColor = muted
		onLabel = radioChoiceText(false, "ON")
		offLabel = radioChoiceText(true, "OFF")
	}
	if !inactive && !poweredDown {
		onLabel = strings.Repeat(" ", len([]rune(onLabel)))
		offLabel = strings.Repeat(" ", len([]rune(offLabel)))
	}
	slider := []rune(shieldsSlider(shieldsPct))
	w.Reset()
	_ = w.Write("\n", text.WriteReplace())
	_ = w.Write("  WARP CORE       ", text.WriteCellOpts(cell.FgColor(statusColor)))
	_ = w.Write(statusLabel, text.WriteCellOpts(cell.FgColor(statusColor)))
	_ = w.Write("   ", text.WriteCellOpts(cell.FgColor(controlMuted)))
	_ = w.Write(onLabel, text.WriteCellOpts(cell.FgColor(onColor)))
	_ = w.Write("   ", text.WriteCellOpts(cell.FgColor(controlMuted)))
	_ = w.Write(offLabel+"\n", text.WriteCellOpts(cell.FgColor(offColor)))
	_ = w.Write(fmt.Sprintf("  SHIELDS         %3d%%  [", shieldsPct), text.WriteCellOpts(cell.FgColor(yellow)))
	if !inactive && !poweredDown {
		slider = []rune(strings.Repeat(" ", len(slider)))
	}
	for _, r := range slider[:len(slider)-1] {
		color := sliderTrack
		if r == '█' {
			color = sliderFill
		}
		_ = w.Write(string(r), text.WriteCellOpts(cell.FgColor(color)))
	}
	_ = w.Write(string(slider[len(slider)-1]), text.WriteCellOpts(cell.FgColor(sliderKnob)))
	_ = w.Write("]\n", text.WriteCellOpts(cell.FgColor(yellow)))
	_ = w.Write("  LIFE SUPPORT    NOMINAL\n", text.WriteCellOpts(cell.FgColor(green)))
	_ = w.Write("  HULL            96%%\n", text.WriteCellOpts(cell.FgColor(green)))
	_ = w.Write("  PHASERS         STANDBY\n", text.WriteCellOpts(cell.FgColor(cyan)))
	_ = w.Write("  TORPEDOES       ", text.WriteCellOpts(cell.FgColor(cyan)))
	torpedoControl := dropdownPreview(fmt.Sprintf("%02d", torpedoSet))
	if !inactive && !poweredDown {
		torpedoControl = strings.Repeat(" ", len([]rune(torpedoControl)))
	}
	_ = w.Write(torpedoControl, text.WriteCellOpts(cell.FgColor(dropdownColor)))
	_ = w.Write("  ", text.WriteCellOpts(cell.FgColor(controlMuted)))
	if torpedoLoading {
		_ = w.Write("Loading...\n", text.WriteCellOpts(cell.FgColor(yellow)))
	} else {
		_ = w.Write(fmt.Sprintf("%02d LOADED\n", torpedoLoaded), text.WriteCellOpts(cell.FgColor(loadedColor)))
	}
	cloakBox := checkboxIndicator(false)
	cloakStatus := "OFFLINE"
	cloakStatusColor := red
	if cloakEnabling {
		cloakBox = checkboxIndicator(true)
		cloakStatus = "ENABLING CLOAKING..."
		cloakStatusColor = yellow
	} else if cloakDisabling {
		cloakBox = checkboxIndicator(false)
		cloakStatus = "DISABLING CLOAKING..."
		cloakStatusColor = yellow
	} else if cloakOnline {
		cloakBox = checkboxIndicator(true)
		cloakStatus = "ONLINE"
		cloakStatusColor = green
	}
	_ = w.Write("  CLOAK           ", text.WriteCellOpts(cell.FgColor(red)))
	if !inactive && !poweredDown {
		cloakBox = strings.Repeat(" ", len([]rune(cloakBox)))
	}
	_ = w.Write(cloakBox, text.WriteCellOpts(cell.FgColor(cloakColor)))
	_ = w.Write("  ", text.WriteCellOpts(cell.FgColor(controlMuted)))
	_ = w.Write(cloakStatus+"\n", text.WriteCellOpts(cell.FgColor(cloakStatusColor)))
	_ = w.Write("  ALARM           ", text.WriteCellOpts(cell.FgColor(yellow)))
	alarmControl := dropdownPreview(fmt.Sprintf("%03d", alarmThreshold))
	if !inactive && !poweredDown {
		alarmControl = strings.Repeat(" ", len([]rune(alarmControl)))
	}
	_ = w.Write(alarmControl+"\n", text.WriteCellOpts(cell.FgColor(dropdownColor)))
	_ = w.Write("\n")
	alertText := "INACTIVE"
	alertOpts := []cell.Option{cell.FgColor(cell.ColorNumber(240))}
	alertIcon := "   "
	if redAlert {
		alertText = "ACTIVE"
		alertOpts = []cell.Option{cell.FgColor(alertColor)}
		if blinkOn {
			alertIcon = "🚨 "
		}
	}
	_ = w.Write("  RED ALERT       ", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(240))))
	_ = w.Write(alertText, text.WriteCellOpts(alertOpts...))
	_ = w.Write("  ", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(240))))
	_ = w.Write(alertIcon+"\n", text.WriteCellOpts(cell.FgColor(alertColor)))
}

func shieldsSlider(pct int) string {
	const width = 18
	if pct < 1 {
		pct = 1
	}
	if pct > 100 {
		pct = 100
	}
	knob := (pct - 1) * (width - 1) / 99
	runes := make([]rune, width)
	for i := range runes {
		switch {
		case i < knob:
			runes[i] = '█'
		case i == knob:
			runes[i] = '●'
		default:
			runes[i] = '░'
		}
	}
	return string(runes)
}

func writeHelp(w *text.Text, inactive bool, mouseSelectMode bool, sparkReadout string) {
	graphStatus := sparkReadout
	if graphStatus == "" {
		graphStatus = "Data=Sine  Style=Braille"
	}
	primary := cell.ColorNumber(250)
	secondary := cell.ColorNumber(220)
	accent := cell.ColorRed
	if inactive {
		primary, secondary, accent = inactiveText, inactiveText, inactiveText
	}
	mouseLine := "Mouse=Focus"
	if mouseSelectMode {
		mouseLine = "Mouse=Select"
	}
	_ = w.Write("  Graph controls: odd keys cycle DATA, even keys cycle STYLE.  ",
		text.WriteReplace(),
		text.WriteCellOpts(cell.FgColor(primary)))
	_ = w.Write("1=Burst 3=Saw 5=Pulse 7=Storm 9=Sine  2/4/6/8/0=Styles  d/s=Cycle  "+graphStatus+"  ",
		text.WriteCellOpts(cell.FgColor(secondary)))
	_ = w.Write("Tab=Focus  focused text click=Select  m="+mouseLine+"  q=Quit  z..g=border presets  panes showcase decrypt charsets + preset profiles",
		text.WriteCellOpts(cell.FgColor(accent)))
}

func defaultStyle() panelStyle {
	return styleInterlace
}

func defaultStyleForPanel(id string) panelStyle {
	switch id {
	case idLCARS:
		return styleNeon
	case idWarp:
		return styleIce
	case idComms:
		return styleMatrix
	case idHelp:
		return styleRainbow
	default:
		return defaultStyle()
	}
}

func themedStyle(style panelStyle) *borderfx.Effect {
	if style == styleNeon {
		return borderfx.Presets.Noise.With(borderfx.Colors(
			cell.ColorNumber(159),
			cell.ColorNumber(117),
			cell.ColorNumber(39),
		))
	}
	if style == styleRainbow {
		return borderfx.Presets.Noise.With(borderfx.Colors(
			cell.ColorNumber(195),
			cell.ColorNumber(251),
			cell.ColorNumber(240),
		))
	}
	return themedMacro(style).With(panelPalette())
}

func themedMacro(style panelStyle) borderfx.Macro {
	switch style {
	case styleInterlace:
		return borderfx.Presets.Interlace
	case styleFire:
		return borderfx.Presets.Shard
	case styleIce:
		return borderfx.Presets.Pulse
	case styleRainbow:
		return borderfx.Presets.Ribbon
	case styleMatrix:
		return borderfx.Presets.Power
	case styleSynthwave:
		return borderfx.Presets.Emoji
	case styleNeon:
		return borderfx.Presets.Noise
	default:
		return borderfx.Presets.Focus
	}
}

func panelPalette() borderfx.Palette {
	return borderfx.Colors(
		cell.ColorNumber(117),
		cell.ColorNumber(251),
		cell.ColorNumber(233),
	)
}

// -------------------------------------------------------------------

func feedLog(ctx context.Context, w *text.Text, bootDone <-chan struct{}) {
	select {
	case <-ctx.Done():
		return
	case <-bootDone:
	}
	msgs := []struct {
		icon  string
		msg   string
		color cell.Color
	}{
		{"SCAN", "sensor sweep: sector 31427 clear", cell.ColorCyan},
		{"DFLT", "deflector array: nominal output", cell.ColorWhite},
		{"COMM", "subspace signal detected - bearing 247 mark 3", cell.ColorYellow},
		{"SHLD", "shield harmonics: realigned to 47.3 kHz", cell.ColorGreen},
		{"TEMP", "tachyon grid: no temporal anomalies", cell.ColorWhite},
		{"WARP", "warp field: stable at factor 6.2", cell.ColorGreen},
		{"HAIL", "hailing frequencies open - all bands", cell.ColorCyan},
		{"TRNS", "transporter: pattern buffers at 99.8%%", cell.ColorWhite},
		{"WARN", "ion storm detected at 2.4 parsecs", cell.ColorYellow},
		{"PHSR", "phaser bank 3: fully recharged", cell.ColorGreen},
		{"BIO", "2 non-humanoid life signs - bearing 118", cell.ColorYellow},
		{"CORE", "antimatter containment: 99.97%%", cell.ColorGreen},
		{"ORBT", "entering orbit: class M planet", cell.ColorCyan},
		{"LAB", "anomalous readings in cargo bay 4", cell.ColorYellow},
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m := msgs[i%len(msgs)]
			ts := time.Now().Format("15:04:05")
			_ = w.Write(fmt.Sprintf(" %-4s %s  %s\n", m.icon, ts, m.msg),
				text.WriteCellOpts(cell.FgColor(m.color)))
			i++
		}
	}
}

func feedSparkline(ctx context.Context, s *linechart.LineChart, probe *sparkProbe, ship *shipState, profiles *telemetryProfiles, bootDone <-chan struct{}) {
	select {
	case <-ctx.Done():
		return
	case <-bootDone:
	}

	reseedSparkline(s, probe, ship, profiles)
	phase := 0.0

	ticker := time.NewTicker(sparkFrameInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dataProfile, styleProfile, _ := profiles.snapshot()
			frame := make([]int, 0, sparkSamplesPerTick)
			for i := 0; i < sparkSamplesPerTick; i++ {
				phase += profilePhaseStep(dataProfile)
				frame = append(frame, fabricatedPingSample(dataProfile, phase))
			}
			applySparklineBatch(s, probe, ship, frame, sparkStyleForProfile(styleProfile))
		}
	}
}

func reseedSparkline(s *linechart.LineChart, probe *sparkProbe, ship *shipState, profiles *telemetryProfiles) {
	dataProfile, styleProfile, _ := profiles.snapshot()
	phase := 0.0
	warmup := make([]int, 0, sparkWarmupSamples)
	for i := 0; i < sparkWarmupSamples; i++ {
		phase += profilePhaseStep(dataProfile)
		warmup = append(warmup, fabricatedPingSample(dataProfile, phase))
	}
	probe.clear()
	applySparklineBatch(s, probe, ship, warmup, sparkStyleForProfile(styleProfile))
}

// fabricatedPingSample produces one telemetry sample for the selected data
// profile. Each profile intentionally has distinct motion so switching profiles
// is obvious to operators.
func fabricatedPingSample(profile telemetryDataProfile, phase float64) int {
	switch profile {
	case dataProfileBurst:
		return clampGraphSample(profileBurstSample(phase))
	case dataProfileSaw:
		return clampGraphSample(profileSawSample(phase))
	case dataProfilePulse:
		return clampGraphSample(profilePulseSample(phase))
	case dataProfileStorm:
		return clampGraphSample(profileStormSample(phase))
	default:
		return clampGraphSample(profileSineSample(phase))
	}
}

func profilePhaseStep(profile telemetryDataProfile) float64 {
	switch profile {
	case dataProfileBurst:
		return 0.18
	case dataProfileSaw:
		return 0.11
	case dataProfilePulse:
		return 0.16
	case dataProfileStorm:
		return 0.21
	default:
		return 0.12
	}
}

func profileSineSample(phase float64) int {
	base := 302.0 + 64.0*math.Sin(phase*0.34) + 28.0*math.Sin(phase*0.93+1.4)
	jitter := 10.0*math.Sin(phase*5.1) + 6.0*math.Sin(phase*8.4+0.7)
	v := int(base + jitter)
	return (v / 4) * 4
}

func profileBurstSample(phase float64) int {
	baseline := 294.0 + 36.0*math.Sin(phase*0.26+0.3)
	jitter := 14.0*math.Sin(phase*5.3) + 10.0*math.Sin(phase*10.9+0.8)
	spikeGate := math.Pow(0.5+0.5*math.Sin(phase*0.49+0.2), 11)
	spike := spikeGate * (96.0 + 34.0*(0.5+0.5*math.Sin(phase*3.1+1.5)))
	v := int(baseline + jitter + spike)
	return (v / 6) * 6
}

func profileSawSample(phase float64) int {
	saw := (frac(phase*0.29)*2.0 - 1.0) * 78.0
	envelope := 300.0 + 26.0*math.Sin(phase*0.57+1.1)
	edge := 12.0 * math.Sin(phase*7.2)
	v := int(envelope + saw + edge)
	return (v / 5) * 5
}

func profilePulseSample(phase float64) int {
	base := 288.0 + 24.0*math.Sin(phase*0.37+0.9)
	pulseGate := math.Pow(0.5+0.5*math.Sin(phase*0.74), 10)
	pulse := pulseGate * (134.0 + 24.0*(0.5+0.5*math.Sin(phase*2.8+1.8)))
	jitter := 10.0*math.Sin(phase*9.2+0.1) + 6.0*math.Sin(phase*13.5)
	v := int(base + pulse + jitter)
	return (v / 6) * 6
}

func profileStormSample(phase float64) int {
	base := 290.0 + 48.0*math.Sin(phase*0.25+0.4) + 24.0*math.Sin(phase*0.71+1.9)
	jitter := 18.0*math.Sin(phase*4.8) + 12.0*math.Sin(phase*9.7+0.8) + 6.0*math.Sin(phase*17.3)
	ramp := (frac(phase*0.31)*2.0 - 1.0) * 20.0
	spikeGate := math.Pow(0.5+0.5*math.Sin(phase*0.47+0.2), 12)
	dropGate := math.Pow(0.5+0.5*math.Sin(phase*0.67+2.4), 16)
	spike := spikeGate * (96.0 + 42.0*(0.5+0.5*math.Sin(phase*3.1+1.5)))
	drop := dropGate * (72.0 + 24.0*(0.5+0.5*math.Sin(phase*2.7+0.9)))
	v := int(base + jitter + ramp + spike - drop)
	return (v / 4) * 4
}

func clampGraphSample(v int) int {
	switch {
	case v < 120:
		return 120
	case v > 700:
		return 700
	default:
		return v
	}
}

func frac(v float64) float64 {
	return v - math.Floor(v)
}

// applySparklineBatch appends values to the graph, updates tooltip history, and
// synchronizes red-alert state against the current alarm threshold.
func applySparklineBatch(s *linechart.LineChart, probe *sparkProbe, ship *shipState, values []int, style sparkProfileStyle) {
	if len(values) == 0 {
		return
	}
	probe.addBatch(values)
	_, _, _, _, _, _, _, _, alarmThreshold, _, _ := ship.snapshot()
	s.SetThresholdLine(float64(alarmThreshold), cell.FgColor(cell.ColorRed))
	red := false
	for _, v := range values {
		if v >= alarmThreshold {
			red = true
			break
		}
	}
	ship.setRedAlert(red)
	history := probe.floatValues()
	_ = alarmThreshold
	_ = s.Series("subspace", history, linechart.SeriesCellOpts(cell.FgColor(style.color)))
}

func feedEvents(ctx context.Context, w *text.Text, bootDone <-chan struct{}) {
	select {
	case <-ctx.Done():
		return
	case <-bootDone:
	}
	evts := []struct {
		icon  string
		msg   string
		color cell.Color
	}{
		{"IN", "incoming: starbase 12 priority one", cell.ColorCyan},
		{"NAV", "course correction: +0.3 mark 2", cell.ColorWhite},
		{"DIP", "diplomatic envoy requesting docking", cell.ColorYellow},
		{"AWY", "away team: check-in confirmed", cell.ColorGreen},
		{"RED", "vessel at warp 9 - closing fast", cell.ColorRed},
		{"SCI", "science: anomaly catalogued #7741", cell.ColorMagenta},
		{"OPS", "duty shift rotation complete", cell.ColorWhite},
		{"MED", "all crew cleared for duty", cell.ColorGreen},
		{"UNK", "unidentified craft decloaking - port bow", cell.ColorRed},
		{"CRG", "cargo transfer from station complete", cell.ColorCyan},
	}
	ticker := time.NewTicker(700 * time.Millisecond)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e := evts[i%len(evts)]
			ts := time.Now().Format("15:04")
			_ = w.Write(fmt.Sprintf("  %-3s %s  %s\n", e.icon, ts, e.msg),
				text.WriteCellOpts(cell.FgColor(e.color)))
			i++
		}
	}
}
