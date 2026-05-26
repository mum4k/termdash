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

// Binary tabdemo shows the tab widget in a unified multi-widget dashboard.
package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/borderfx"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/checkbox"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/dropdown"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/heatmap"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/pie"
	"github.com/mum4k/termdash/widgets/radio"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/mum4k/termdash/widgets/slider"
	"github.com/mum4k/termdash/widgets/spectrum"
	"github.com/mum4k/termdash/widgets/tab"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
	"github.com/mum4k/termdash/widgets/threed"
)

const (
	// redrawInterval keeps the unified demo responsive without burning CPU.
	redrawInterval = 50 * time.Millisecond
	// dataInterval controls how quickly the animated widgets update.
	dataInterval = 180 * time.Millisecond
	// maxHistory bounds rolling chart series so redraw cost stays stable.
	maxHistory = 96
	// heatmapGridSize keeps the paired thermal matrices at a high-resolution data grid.
	heatmapGridSize = 15
	// alertTabIndex is the zero-based tab index used by the notification-follow demo.
	alertTabIndex = 4
)

const (
	idPageShell       = "tabContent"
	idOverviewGauge   = "tabdemo-overview-gauge"
	idOverviewMemory  = "tabdemo-overview-memory"
	idOverviewDisk    = "tabdemo-overview-disk"
	idOverviewDetails = "tabdemo-overview-details"
	idOverviewSummary = "tabdemo-overview-summary"
	idOverviewProcs   = "tabdemo-overview-processes"
	idControlsInput   = "tabdemo-controls-input"
	idControlsDisplay = "tabdemo-controls-display"
	idControlsWarp    = "tabdemo-controls-warp"
	idControlsProfile = "tabdemo-controls-profile"
	idControlsShield  = "tabdemo-controls-shield"
	idControlsAlarm   = "tabdemo-controls-alarm"
	idControlsActions = "tabdemo-controls-actions"
	idControlsStatus  = "tabdemo-controls-status"
	idSignalsLine     = "tabdemo-signals-line"
	idSignalsHeatmap  = "tabdemo-signals-heatmap"
	idSignalsHeatmapB = "tabdemo-signals-heatmap-b"
	idSignalsSpectrum = "tabdemo-signals-spectrum"
	idSignalsStatus   = "tabdemo-signals-status"
	idAlertsPie       = "tabdemo-alerts-pie"
	idAlertsDonut     = "tabdemo-alerts-donut"
	idAlertsStatus    = "tabdemo-alerts-status"
	idThreeDStage   = "tabdemo-threed-stage"
	idThreeDStatus  = "tabdemo-threed-status"
	idThreeDPyramid = "tabdemo-threed-pyramid"
)

// animatedPaneIDs lists the pane borders that participate in focus sweeps.
var animatedPaneIDs = []string{
	idOverviewGauge,
	idOverviewMemory,
	idOverviewDisk,
	idOverviewDetails,
	idOverviewSummary,
	idOverviewProcs,
	idControlsInput,
	idControlsDisplay,
	idControlsWarp,
	idControlsProfile,
	idControlsShield,
	idControlsAlarm,
	idControlsActions,
	idControlsStatus,
	idSignalsLine,
	idSignalsHeatmap,
	idSignalsHeatmapB,
	idSignalsSpectrum,
	idSignalsStatus,
	idAlertsPie,
	idAlertsDonut,
	idAlertsStatus,
	idThreeDStage,
	idThreeDStatus,
	idThreeDPyramid,
}

// controlPanelState stores the selections shown on the controls tab.
type controlPanelState struct {
	mu      sync.Mutex
	warp    bool
	shields int
	alarmY  string
	profile string
	command string
	display string
	actions int
}

// overviewWidgets groups the widgets used by the first tab.
type overviewWidgets struct {
	cpuGauge  *gauge.Gauge
	memGauge  *gauge.Gauge
	diskGauge *gauge.Gauge
	cpuText   *text.Text
	pidText   *text.Text
	summary   *text.Text
}

// controlWidgets groups the widgets used by the controls tab.
type controlWidgets struct {
	warpSwitch *checkbox.Checkbox
	shields    *slider.Slider
	alarm      *dropdown.Dropdown
	profile    *radio.Radio
	input      *textinput.TextInput
	display    *segmentdisplay.SegmentDisplay
	route      *button.Button
	prime      *button.Button
	stabilize  *button.Button
	status     *text.Text
}

// telemetryWidgets groups the widgets used by the signals tab.
type telemetryWidgets struct {
	spectrum *spectrum.Spectrum
	line     *linechart.LineChart
	heatmap  *heatmap.HeatMap
	heatmapB *heatmap.HeatMap
	pie      *pie.Pie
	donut    *donut.Donut
	status   *text.Text
	alerts   *text.Text
}

// threeDWidgets groups the widgets used by the 3D showcase tab.
type threeDWidgets struct {
	stage   *threed.ThreeD
	pyramid *threed.ThreeD
	status  *text.Text
}

// telemetryModel stores the rolling telemetry series shown in the signals tab.
type telemetryModel struct {
	mu    sync.Mutex
	phase float64
	line  []float64
	left  []int
	right []int
}

// main boots the tab demo.
func main() {
	term, err := tcell.New()
	if err != nil {
		log.Fatalf("failed to initialize terminal: %v", err)
	}
	defer term.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instructions, err := newInstructions()
	if err != nil {
		log.Fatalf("failed to create instructions widget: %v", err)
	}

	overview, overviewTab, err := newOverviewTab()
	if err != nil {
		log.Fatalf("failed to build overview tab: %v", err)
	}

	controlState := &controlPanelState{
		warp:    true,
		shields: 78,
		alarmY:  "450",
		profile: "FOCUS",
		command: "LINK IDLE",
		display: "READY",
	}
	_, controlsTab, err := newControlsTab(controlState)
	if err != nil {
		log.Fatalf("failed to build controls tab: %v", err)
	}

	telemetry, signalsTab, thermalTab, alertsTab, err := newTelemetryTabs()
	if err != nil {
		log.Fatalf("failed to build telemetry tabs: %v", err)
	}
	threeD, threeDTab, err := newThreeDTab()
	if err != nil {
		log.Fatalf("failed to build 3D tab: %v", err)
	}

	tabManager := tab.NewManager(overviewTab, controlsTab, signalsTab, thermalTab, alertsTab, threeDTab)
	opts := tab.NewOptions(
		tab.AnimatedActiveTab(false),
		tab.ActiveTextColor(cell.ColorNumber(159)),
		tab.InactiveTextColor(cell.ColorNumber(245)),
		tab.FollowNotifications(false),
		tab.SweepTextColor(cell.ColorNumber(242)),
		tab.SweepAccentColor(cell.ColorNumber(75)),
	)

	tabHeader, err := tab.NewHeader(tabManager, opts)
	if err != nil {
		log.Fatalf("failed to create tab header: %v", err)
	}

	initialContent, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("failed to create initial content widget: %v", err)
	}
	if err := initialContent.Write("Loading tab content..."); err != nil {
		log.Fatalf("failed to write initial content: %v", err)
	}

	root, err := container.New(
		term,
		container.Border(linestyle.Round),
		container.BorderColor(cell.ColorNumber(236)),
		container.FocusedColor(cell.ColorNumber(236)),
		container.BorderTitle(" termdash "),
		container.BorderTitleAlignCenter(),
		container.TitleColor(cell.ColorNumber(75)),
		container.TitleFocusedColor(cell.ColorNumber(75)),
		container.SplitHorizontal(
			container.Top(
				container.PlaceWidget(tabHeader.Widget()),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Round),
						container.BorderColor(cell.ColorNumber(236)),
						container.FocusedColor(cell.ColorNumber(236)),
						container.BorderTitle("Overview"),
						container.BorderTitleAlignCenter(),
						container.TitleColor(cell.ColorNumber(242)),
						container.TitleFocusedColor(cell.ColorNumber(242)),
						container.PaddingLeft(1),
						container.PaddingTop(1),
						container.PlaceWidget(initialContent),
						container.ID(idPageShell),
					),
					container.Bottom(
						container.PlaceWidget(instructions),
					),
					container.SplitPercent(94),
				),
			),
			container.SplitPercent(8),
		),
	)
	if err != nil {
		log.Fatalf("failed to create root container: %v", err)
	}

	tabContent := tab.NewContent(tabManager)
	if err := tabContent.Update(root); err != nil {
		log.Fatalf("failed to seed initial tab content: %v", err)
	}

	eventHandler := tab.NewEventHandler(ctx, term, tabManager, tabHeader, tabContent, root, cancel, opts)
	fx := configureBorderChrome(root)
	go func() {
		_ = fx.Run(ctx)
	}()

	go animateOverview(ctx, overview)
	go animateTelemetry(ctx, telemetry)
	go animateThreeD(ctx, threeD)
	go runAlertDrill(ctx, tabManager, eventHandler, telemetry.alerts)

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' || k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
	}

	if err := termdash.Run(
		ctx,
		term,
		root,
		termdash.KeyboardSubscriber(func(k *terminalapi.Keyboard) {
			quitter(k)
			eventHandler.HandleKeyboard(k)
		}),
		termdash.MouseSubscriber(eventHandler.HandleMouse),
		termdash.RedrawInterval(redrawInterval),
	); err != nil && err != context.Canceled {
		log.Fatalf("termdash encountered an error: %v", err)
	}
}

// newInstructions creates the footer help text for the unified demo.
func newInstructions() (*text.Text, error) {
	w, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, err
	}
	if err := w.Write(" ←/→ ", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(75)), cell.Bold())); err != nil {
		return nil, err
	}
	if err := w.Write("switch tabs   ", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(245)))); err != nil {
		return nil, err
	}
	if err := w.Write("Tab ", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(75)), cell.Bold())); err != nil {
		return nil, err
	}
	if err := w.Write("next   ", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(245)))); err != nil {
		return nil, err
	}
	if err := w.Write("q/Esc ", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(75)), cell.Bold())); err != nil {
		return nil, err
	}
	if err := w.Write("quit", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(245)))); err != nil {
		return nil, err
	}
	return w, nil
}

// demoPaneOptions returns the shared window styling used across the unified demo.
func demoPaneOptions(title string, focused bool) []container.Option {
	opts := []container.Option{
		container.Border(linestyle.Round),
		container.BorderTitle(" " + title + " "),
		container.BorderTitleAlignCenter(),
		container.BorderColor(cell.ColorNumber(236)),
		container.FocusedColor(cell.ColorNumber(75)),
		container.TitleColor(cell.ColorNumber(242)),
		container.TitleFocusedColor(cell.ColorNumber(159)),
	}
	if focused {
		opts = append(opts, container.Focused())
	}
	return opts
}

// paneOptions combines IDs, shared pane chrome, and extra options in one slice.
func paneOptions(id, title string, focused bool, extras ...container.Option) []container.Option {
	opts := []container.Option{container.ID(id)}
	opts = append(opts, demoPaneOptions(title, focused)...)
	opts = append(opts, extras...)
	return opts
}

// overviewPaneOptions keeps the overview page visually aligned with the other tabs.
func overviewPaneOptions(title string, focused bool) []container.Option {
	return demoPaneOptions(title, focused)
}

// newOverviewTab builds the first tab and keeps its original monitor flavor.
func newOverviewTab() (*overviewWidgets, *tab.Tab, error) {
	cpuGauge, err := gauge.New(gauge.Color(cell.ColorNumber(75)))
	if err != nil {
		return nil, nil, err
	}
	memGauge, err := gauge.New(gauge.Color(cell.ColorNumber(108)))
	if err != nil {
		return nil, nil, err
	}
	diskGauge, err := gauge.New(gauge.Color(cell.ColorNumber(137)))
	if err != nil {
		return nil, nil, err
	}
	cpuText, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, nil, err
	}
	summary, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, nil, err
	}
	pidText, err := text.New(text.WrapAtWords(), text.RollContent())
	if err != nil {
		return nil, nil, err
	}

	overview := &overviewWidgets{
		cpuGauge:  cpuGauge,
		memGauge:  memGauge,
		diskGauge: diskGauge,
		cpuText:   cpuText,
		summary:   summary,
		pidText:   pidText,
	}

	if err := memGauge.Percent(62, gauge.TextLabel("MEM 62%")); err != nil {
		return nil, nil, err
	}
	if err := diskGauge.Percent(47, gauge.TextLabel("DISK 47%")); err != nil {
		return nil, nil, err
	}
	if err := summary.Write("Thermals: nominal\nScheduler: balanced\nGPU queue: stable"); err != nil {
		return nil, nil, err
	}

	content := container.SplitVertical(
		container.Left(
			container.SplitHorizontal(
				container.Top(
					paneOptions(idOverviewGauge, "CPU", true, container.PlaceWidget(cpuGauge))...,
				),
				container.Bottom(
					container.SplitHorizontal(
						container.Top(
							paneOptions(idOverviewMemory, "Memory", false, container.PlaceWidget(memGauge))...,
						),
						container.Bottom(
							paneOptions(idOverviewDisk, "Disk", false, container.PlaceWidget(diskGauge))...,
						),
						container.SplitPercent(50),
					),
				),
				container.SplitPercent(34),
			),
		),
		container.Right(
			container.SplitHorizontal(
				container.Top(
					container.SplitHorizontal(
						container.Top(
							paneOptions(idOverviewDetails, "Live Metrics", false,
								container.PaddingLeft(1),
								container.PaddingTop(1),
								container.PlaceWidget(cpuText),
							)...,
						),
						container.Bottom(
							paneOptions(idOverviewSummary, "Host Summary", false,
								container.PaddingLeft(1),
								container.PaddingTop(1),
								container.PlaceWidget(summary),
							)...,
						),
						container.SplitPercent(56),
					),
				),
				container.Bottom(
					paneOptions(idOverviewProcs, "Process Queue", false,
						container.PaddingLeft(1),
						container.PaddingTop(1),
						container.PlaceWidget(pidText),
					)...,
				),
				container.SplitPercent(40),
			),
		),
		container.SplitPercent(22),
	)

	return overview, &tab.Tab{Name: "Overview", Content: content}, nil
}

// newControlsTab builds the interactive controls showcase.
func newControlsTab(state *controlPanelState) (*controlWidgets, *tab.Tab, error) {
	status, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, nil, err
	}

	display, err := segmentdisplay.New(
		segmentdisplay.AlignHorizontal(align.HorizontalCenter),
		segmentdisplay.AlignVertical(align.VerticalMiddle),
		segmentdisplay.MaximizeSegmentHeight(),
		segmentdisplay.GapPercent(8),
	)
	if err != nil {
		return nil, nil, err
	}

	widgets := &controlWidgets{
		status:  status,
		display: display,
	}

	writeStatus := func() error {
		return renderControlStatus(status, state)
	}
	writeDisplay := func() error {
		return renderSegmentDisplay(display, state)
	}

	warpSwitch, err := checkbox.New("Enable Warp Assist", checkbox.Checked(state.warp), checkbox.OnChange(func(checked bool) error {
		state.mu.Lock()
		state.warp = checked
		state.mu.Unlock()
		return writeStatus()
	}))
	if err != nil {
		return nil, nil, err
	}

	shields, err := slider.New(
		slider.Min(0),
		slider.Max(100),
		slider.Value(state.shields),
		slider.Width(28),
		slider.OnChange(func(v int) error {
			state.mu.Lock()
			state.shields = v
			state.mu.Unlock()
			return writeStatus()
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	alarm, err := dropdown.New(
		dropdown.IntRange(200, 600, 50, "%03d"),
		dropdown.Selected(5),
		dropdown.OnSelect(func(_ int, label string) error {
			state.mu.Lock()
			state.alarmY = label
			state.mu.Unlock()
			return writeStatus()
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	profile, err := radio.New(
		[]radio.Item{
			{Label: "FOCUS", SelectedCellOpts: []cell.Option{cell.FgColor(cell.ColorCyan)}},
			{Label: "POWER", SelectedCellOpts: []cell.Option{cell.FgColor(cell.ColorYellow)}},
			{Label: "MATRIX", SelectedCellOpts: []cell.Option{cell.FgColor(cell.ColorGreen)}},
		},
		radio.OnChange(func(_ int, label string) error {
			state.mu.Lock()
			state.profile = label
			state.mu.Unlock()
			return writeStatus()
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	input, err := textinput.New(
		textinput.Border(linestyle.Round),
		textinput.BorderColor(cell.ColorNumber(240)),
		textinput.PlaceHolder("enter command..."),
		textinput.PlaceHolderColor(cell.ColorNumber(240)),
		textinput.MaxWidthCells(52),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.ClearOnSubmit(),
		textinput.OnSubmit(func(command string) error {
			return submitControlCommand(widgets, state, command)
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	newActionButton := func(label string, fill cell.Color, fn func() error) (*button.Button, error) {
		return button.New(
			label,
			fn,
			button.WidthFor("Stabilize"),
			button.FillColor(fill),
			button.FocusedFillColor(cell.ColorNumber(111)),
			button.PressedFillColor(cell.ColorNumber(159)),
			button.ShadowColor(cell.ColorNumber(235)),
		)
	}

	routeButton, err := newActionButton("Route", cell.ColorNumber(60), func() error {
		return applyControlPreset(widgets, state, true, 86, "450", "FOCUS", "ROUTE")
	})
	if err != nil {
		return nil, nil, err
	}
	primeButton, err := newActionButton("Prime", cell.ColorNumber(67), func() error {
		return applyControlPreset(widgets, state, true, 94, "500", "POWER", "PRIME")
	})
	if err != nil {
		return nil, nil, err
	}
	stabilizeButton, err := newActionButton("Stabilize", cell.ColorNumber(66), func() error {
		return applyControlPreset(widgets, state, false, 34, "350", "MATRIX", "STABLE")
	})
	if err != nil {
		return nil, nil, err
	}

	widgets.warpSwitch = warpSwitch
	widgets.shields = shields
	widgets.alarm = alarm
	widgets.profile = profile
	widgets.input = input
	widgets.route = routeButton
	widgets.prime = primeButton
	widgets.stabilize = stabilizeButton

	if err := writeStatus(); err != nil {
		return nil, nil, err
	}
	if err := writeDisplay(); err != nil {
		return nil, nil, err
	}

	content := container.SplitHorizontal(
		container.Top(
			container.SplitVertical(
				container.Left(
					paneOptions(idControlsDisplay, "Segment Display", true,
						container.PaddingLeft(1),
						container.PaddingTop(0),
						container.PlaceWidget(display),
					)...,
				),
				container.Right(
					paneOptions(idControlsInput, "Command Uplink", false,
						container.PaddingLeft(1),
						container.PaddingTop(1),
						container.PlaceWidget(input),
					)...,
				),
				container.SplitPercent(42),
			),
		),
		container.Bottom(
			container.SplitVertical(
				container.Left(
					container.SplitHorizontal(
						container.Top(
							container.SplitVertical(
								container.Left(
									container.SplitHorizontal(
										container.Top(
											paneOptions(idControlsShield, "Shield Bias", false,
												container.PaddingLeft(1),
												container.PaddingTop(1),
												container.PlaceWidget(shields),
											)...,
										),
										container.Bottom(
											paneOptions(idControlsWarp, "Warp Assist", false,
												container.PaddingLeft(1),
												container.PaddingTop(1),
												container.PlaceWidget(warpSwitch),
											)...,
										),
										container.SplitPercent(52),
									),
								),
								container.Right(
									container.SplitHorizontal(
										container.Top(
											paneOptions(idControlsAlarm, "Alarm Y", false,
												container.PaddingLeft(1),
												container.PaddingTop(1),
												container.PlaceWidget(alarm),
											)...,
										),
										container.Bottom(
											paneOptions(idControlsProfile, "Profile", false,
												container.PaddingLeft(1),
												container.PaddingTop(1),
												container.PlaceWidget(profile),
											)...,
										),
										container.SplitPercent(52),
									),
								),
								container.SplitPercent(54),
							),
						),
						container.Bottom(
							paneOptions(idControlsStatus, "Status", false,
								container.PaddingLeft(1),
								container.PaddingTop(1),
								container.PlaceWidget(status),
							)...,
						),
						container.SplitPercent(54),
					),
				),
				container.Right(
					paneOptions(idControlsActions, "Action Deck", false,
						container.PaddingLeft(1),
						container.PaddingTop(1),
						container.SplitVertical(
							container.Left(
								container.PlaceWidget(routeButton),
							),
							container.Right(
								container.SplitVertical(
									container.Left(
										container.PlaceWidget(primeButton),
									),
									container.Right(
										container.PlaceWidget(stabilizeButton),
									),
									container.SplitPercent(50),
								),
							),
							container.SplitPercent(34),
						),
					)...,
				),
				container.SplitPercent(58),
			),
		),
		container.SplitPercent(28),
	)

	return widgets, &tab.Tab{Name: "Controls", Content: content}, nil
}

// newTelemetryTabs builds the signal, thermal, and alert showcase tabs.
func newTelemetryTabs() (*telemetryWidgets, *tab.Tab, *tab.Tab, *tab.Tab, error) {
	sig, err := spectrum.New(
		spectrum.ChannelLabels("LEFT", "RIGHT"),
		spectrum.MaxValue(600),
		spectrum.Gradient(cell.ColorNumber(24), cell.ColorNumber(31), cell.ColorNumber(75), cell.ColorNumber(117), cell.ColorNumber(159)),
		spectrum.PrimaryRunes('⠂', '⠆', '⠇', '⠧', '⠷', '⠿', '⣿'),
		spectrum.SecondaryRunes('⠂', '⠒', '⠓', '⠛', '⠟', '⠿', '⣿'),
		spectrum.HorizontalRunes('⠂', '⠒', '⠲', '⠶', '⠾', '⠿', '⣿'),
		spectrum.PeakRunes('⣾', '⣷'),
		spectrum.HalfDuplexRune('⣿'),
		spectrum.AxisCellOpts(cell.FgColor(cell.ColorNumber(240))),
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	chart, err := linechart.New(
		linechart.BrailleOnly(),
		linechart.DownsampleLTTB(),
		linechart.ThresholdLine(520, cell.FgColor(cell.ColorNumber(167))),
		linechart.XAxisUnscaled(),
		linechart.YAxisCustomScale(120, 720),
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	heat, err := heatmap.New(
		heatmap.AxisCellOpts(cell.FgColor(cell.ColorNumber(242))),
		heatmap.XLabelCellOpts(cell.FgColor(cell.ColorNumber(244))),
		heatmap.YLabelCellOpts(cell.FgColor(cell.ColorNumber(117))),
		heatmap.Palette(
			cell.ColorNumber(236),
			cell.ColorNumber(239),
			cell.ColorNumber(24),
			cell.ColorNumber(31),
			cell.ColorNumber(38),
			cell.ColorNumber(45),
			cell.ColorNumber(81),
		),
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	heatB, err := heatmap.New(
		heatmap.AxisCellOpts(cell.FgColor(cell.ColorNumber(242))),
		heatmap.XLabelCellOpts(cell.FgColor(cell.ColorNumber(244))),
		heatmap.YLabelCellOpts(cell.FgColor(cell.ColorNumber(153))),
		heatmap.Palette(
			cell.ColorNumber(236),
			cell.ColorNumber(238),
			cell.ColorNumber(17),
			cell.ColorNumber(24),
			cell.ColorNumber(30),
			cell.ColorNumber(37),
			cell.ColorNumber(45),
		),
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	pieWidget, err := pie.New(
		pie.ColorOption([]cell.Color{
			cell.ColorNumber(75),
			cell.ColorNumber(111),
			cell.ColorNumber(147),
			cell.ColorNumber(183),
		}),
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	load, err := donut.New(
		donut.HolePercent(28),
		donut.CellOpts(cell.FgColor(cell.ColorNumber(75))),
		donut.TextCellOpts(cell.Bold(), cell.FgColor(cell.ColorWhite)),
	)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	status, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, nil, nil, nil, err
	}
	alerts, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, nil, nil, nil, err
	}

	widgets := &telemetryWidgets{
		spectrum: sig,
		line:     chart,
		heatmap:  heat,
		heatmapB: heatB,
		pie:      pieWidget,
		donut:    load,
		status:   status,
		alerts:   alerts,
	}

	xLabels, yLabels, heatValues := telemetryHeatmapFrame(0)
	if err := heat.Values(xLabels, yLabels, heatValues); err != nil {
		return nil, nil, nil, nil, err
	}
	heat.ClearXLabels()
	xLabelsB, yLabelsB, heatValuesB := telemetryHeatmapFrameB(0)
	if err := heatB.Values(xLabelsB, yLabelsB, heatValuesB); err != nil {
		return nil, nil, nil, nil, err
	}
	heatB.ClearXLabels()
	if err := pieWidget.Values([]int{28, 24, 18, 14}); err != nil {
		return nil, nil, nil, nil, err
	}
	if err := renderTelemetryStatus(widgets.status, 0, 0, 0); err != nil {
		return nil, nil, nil, nil, err
	}
	if err := renderAlertStatus(widgets.alerts, 0, "Alert routing armed", "Awaiting next threshold drill"); err != nil {
		return nil, nil, nil, nil, err
	}

	signals := container.SplitVertical(
		container.Left(
			paneOptions(idSignalsSpectrum, "Spectrum Analyzer", true,
				container.PlaceWidget(sig),
			)...,
		),
		container.Right(
			container.SplitHorizontal(
				container.Top(
					paneOptions(idSignalsLine, "Subspace Telemetry", false,
						container.PlaceWidget(chart),
					)...,
				),
				container.Bottom(
					paneOptions(idSignalsStatus, "Signal Details", false,
						container.PaddingLeft(1),
						container.PaddingTop(1),
						container.PlaceWidget(status),
					)...,
				),
				container.SplitPercent(74),
			),
		),
		container.SplitPercent(72),
	)

	thermal := container.SplitVertical(
		container.Left(
			paneOptions(idSignalsHeatmap, "Thermal Matrix A", true,
				container.PlaceWidget(heat),
			)...,
		),
		container.Right(
			paneOptions(idSignalsHeatmapB, "Thermal Matrix B", false,
				container.PlaceWidget(heatB),
			)...,
		),
		container.SplitPercent(50),
	)

	alertsContent := container.SplitHorizontal(
		container.Top(
			container.SplitVertical(
				container.Left(
					paneOptions(idAlertsPie, "Band Mix", true,
						container.PlaceWidget(pieWidget),
					)...,
				),
				container.Right(
					paneOptions(idAlertsDonut, "Channel Load", false,
						container.PlaceWidget(load),
					)...,
				),
				container.SplitPercent(55),
			),
		),
		container.Bottom(
			paneOptions(idAlertsStatus, "Alert Route", false,
				container.PaddingLeft(1),
				container.PaddingTop(1),
				container.PlaceWidget(alerts),
			)...,
		),
		container.SplitPercent(66),
	)

	return widgets,
		&tab.Tab{Name: "Signals", Content: signals},
		&tab.Tab{Name: "Thermal", Content: thermal},
		&tab.Tab{Name: "Alerts", Content: alertsContent},
		nil
}

// newThreeDTab builds a dedicated tab for the threed widget.
func newThreeDTab() (*threeDWidgets, *tab.Tab, error) {
	stage, err := threed.New(
		threed.ShowAxes(false),
		threed.BackfaceCulling(false),
		threed.RotationStep(0.08),
		threed.UprightOnly(true),
		threed.ZoomScale(38.0),
		threed.AmbientColor(threed.Color{R: 0.6, G: 0.6, B: 0.6}),
		threed.DiffuseColor(threed.Color{R: 0.8, G: 0.8, B: 0.8}),
		threed.SpecularColor(threed.Color{R: 0.4, G: 0.4, B: 0.4}),
		threed.Shininess(24),
	)
	if err != nil {
		return nil, nil, err
	}
	stage.SetModel(threed.CreateCube(threed.Vector3D{}, 1.0, '█'))

	pyramid, err := threed.New(
		threed.ShowAxes(false),
		threed.BackfaceCulling(false),
		threed.RotationStep(0.06),
		threed.UprightOnly(false),
		threed.ZoomScale(34.0),
		threed.AmbientColor(threed.Color{R: 0.5, G: 0.5, B: 0.5}),
		threed.DiffuseColor(threed.Color{R: 0.9, G: 0.9, B: 0.9}),
		threed.SpecularColor(threed.Color{R: 0.6, G: 0.6, B: 0.6}),
		threed.Shininess(32),
	)
	if err != nil {
		return nil, nil, err
	}
	pyramid.SetModel(threed.CreatePyramid(threed.Vector3D{}, 1.0, '▲'))

	status, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, nil, err
	}
	if err := renderThreeDStatus(status, 0); err != nil {
		return nil, nil, err
	}

	widgets := &threeDWidgets{
		stage:   stage,
		pyramid: pyramid,
		status:  status,
	}

	// Layout: left side = main stage (75%) over render notes (25%),
	//         right side = pyramid demo.
	content := container.SplitVertical(
		container.Left(
			container.SplitHorizontal(
				container.Top(
					paneOptions(idThreeDStage, "ThreeD Object Stage", true,
						container.PlaceWidget(stage),
					)...,
				),
				container.Bottom(
					paneOptions(idThreeDStatus, "Render Notes", false,
						container.PaddingLeft(1),
						container.PaddingTop(1),
						container.PlaceWidget(status),
					)...,
				),
				container.SplitPercent(75),
			),
		),
		container.Right(
			paneOptions(idThreeDPyramid, "Pyramid Demo", false,
				container.PlaceWidget(pyramid),
			)...,
		),
		container.SplitPercent(60),
	)

	return widgets, &tab.Tab{Name: "ThreeD", Content: content}, nil
}

// submitControlCommand routes free-form text into the segment display and status panel.
func submitControlCommand(widgets *controlWidgets, state *controlPanelState, command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil
	}

	state.mu.Lock()
	state.command = command
	state.display = compactDisplayText(command)
	state.actions++
	state.mu.Unlock()

	if err := renderSegmentDisplay(widgets.display, state); err != nil {
		return err
	}
	return renderControlStatus(widgets.status, state)
}

// applyControlPreset pushes one canned action through the control widgets.
func applyControlPreset(widgets *controlWidgets, state *controlPanelState, warp bool, shields int, alarmY, profile, command string) error {
	state.mu.Lock()
	state.warp = warp
	state.shields = shields
	state.alarmY = alarmY
	state.profile = profile
	state.command = command
	state.display = compactDisplayText(command)
	state.actions++
	state.mu.Unlock()

	widgets.warpSwitch.SetChecked(warp)
	widgets.shields.SetValue(shields)
	if err := widgets.alarm.SetSelected(alarmIndexForLabel(alarmY)); err != nil {
		return err
	}
	if err := widgets.profile.SetSelected(profileIndexForLabel(profile)); err != nil {
		return err
	}
	if err := renderSegmentDisplay(widgets.display, state); err != nil {
		return err
	}
	return renderControlStatus(widgets.status, state)
}

// renderSegmentDisplay updates the controls tab segment display summary.
func renderSegmentDisplay(w *segmentdisplay.SegmentDisplay, state *controlPanelState) error {
	state.mu.Lock()
	display := state.display
	state.mu.Unlock()

	return w.Write([]*segmentdisplay.TextChunk{
		segmentdisplay.NewChunk(compactDisplayText(display)),
	})
}

// compactDisplayText converts free-form text into a short segment-display label.
func compactDisplayText(text string) string {
	text = strings.ToUpper(strings.TrimSpace(text))
	if text == "" {
		return "READY"
	}

	var b strings.Builder
	for _, r := range text {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune(r)
		}
		if b.Len() >= 8 {
			break
		}
	}

	compacted := strings.TrimSpace(b.String())
	if compacted == "" {
		return "READY"
	}
	return compacted
}

// alarmIndexForLabel maps the dropdown label back to its fixed range index.
func alarmIndexForLabel(label string) int {
	var value int
	if _, err := fmt.Sscanf(label, "%d", &value); err != nil {
		return 0
	}
	if value < 200 {
		value = 200
	}
	if value > 600 {
		value = 600
	}
	return (value - 200) / 50
}

// profileIndexForLabel maps the current profile label to the radio index.
func profileIndexForLabel(label string) int {
	switch strings.ToUpper(strings.TrimSpace(label)) {
	case "POWER":
		return 1
	case "MATRIX":
		return 2
	default:
		return 0
	}
}

// renderControlStatus rewrites the controls tab status summary.
func renderControlStatus(w *text.Text, state *controlPanelState) error {
	state.mu.Lock()
	warp := state.warp
	shields := state.shields
	alarmY := state.alarmY
	profile := state.profile
	command := state.command
	display := state.display
	actions := state.actions
	state.mu.Unlock()

	w.Reset()
	labelColor := cell.FgColor(cell.ColorNumber(245))
	valueColor := cell.FgColor(cell.ColorNumber(252))
	if err := w.Write("Warp Assist: ", text.WriteReplace(), text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write(onOff(warp)+"\n", text.WriteCellOpts(cell.FgColor(boolColor(warp)))); err != nil {
		return err
	}
	if err := w.Write("Shield Bias: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write(fmt.Sprintf("%d%%\n", shields), text.WriteCellOpts(valueColor)); err != nil {
		return err
	}
	if err := w.Write("Alarm Y: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write(alarmY+"\n", text.WriteCellOpts(valueColor)); err != nil {
		return err
	}
	if err := w.Write("Profile: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write(profile+"\n", text.WriteCellOpts(valueColor)); err != nil {
		return err
	}
	if err := w.Write("Display: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write(compactDisplayText(display)+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(75)))); err != nil {
		return err
	}
	if err := w.Write("Command: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write(strings.ToUpper(command)+"\n", text.WriteCellOpts(valueColor)); err != nil {
		return err
	}
	if err := w.Write("Actions: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	return w.Write(fmt.Sprintf("%d", actions), text.WriteCellOpts(valueColor))
}

// renderTelemetryStatus rewrites the telemetry sidecar text.
func renderTelemetryStatus(w *text.Text, latest, peak, load float64) error {
	w.Reset()
	labelColor := cell.FgColor(cell.ColorNumber(245))
	valueColor := cell.FgColor(cell.ColorNumber(252))
	if err := w.Write("Load: ", text.WriteReplace(), text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write(fmt.Sprintf("%.0f%%\n", load), text.WriteCellOpts(valueColor)); err != nil {
		return err
	}
	if err := w.Write("Latest: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write(fmt.Sprintf("%.0f\n", latest), text.WriteCellOpts(valueColor)); err != nil {
		return err
	}
	if err := w.Write("Peak: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	return w.Write(fmt.Sprintf("%.0f", peak), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(167))))
}

// renderAlertStatus rewrites the alert-route panel used by the notification demo.
func renderAlertStatus(w *text.Text, count int, headline, detail string) error {
	w.Reset()
	labelColor := cell.FgColor(cell.ColorNumber(245))
	valueColor := cell.FgColor(cell.ColorNumber(252))
	if err := w.Write("Follow: ", text.WriteReplace(), text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write("ACTIVE\n", text.WriteCellOpts(cell.Bold(), cell.FgColor(cell.ColorNumber(114)))); err != nil {
		return err
	}
	if err := w.Write("Event: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write(headline+"\n", text.WriteCellOpts(valueColor)); err != nil {
		return err
	}
	if err := w.Write("Detail: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write(detail+"\n", text.WriteCellOpts(valueColor)); err != nil {
		return err
	}
	if err := w.Write("Count: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	return w.Write(fmt.Sprintf("%02d", count), text.WriteCellOpts(valueColor))
}

// renderThreeDStatus rewrites the threed tab sidecar copy.
func renderThreeDStatus(w *text.Text, step int) error {
	w.Reset()

	labelColor := cell.FgColor(cell.ColorNumber(245))
	valueColor := cell.FgColor(cell.ColorNumber(252))
	if err := w.Write("Stage: ", text.WriteReplace(), text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write("cube  Y-axis orbit\n", text.WriteCellOpts(valueColor)); err != nil {
		return err
	}
	if err := w.Write("Pyramid: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write("square base  XY orbit\n", text.WriteCellOpts(valueColor)); err != nil {
		return err
	}
	if err := w.Write("Zoom: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	if err := w.Write("mouse wheel\n", text.WriteCellOpts(valueColor)); err != nil {
		return err
	}
	if err := w.Write("Frame: ", text.WriteCellOpts(labelColor)); err != nil {
		return err
	}
	return w.Write(fmt.Sprintf("%04d\n", step), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(75))))
}

// telemetryHeatmapFrame generates the animated heatmap frame for the signals tab.
func telemetryHeatmapFrame(phase float64) ([]string, []string, [][]float64) {
	xLabels, yLabels := telemetryHeatmapLabels()
	values := make([][]float64, len(yLabels))

	for row := range values {
		values[row] = make([]float64, len(xLabels))
		for col := range values[row] {
			band := float64(col) / math.Max(float64(len(xLabels)-1), 1)
			stack := float64(row) / math.Max(float64(len(yLabels)-1), 1)
			wave := 0.52 + 0.48*math.Sin(phase*0.38+stack*2.4+band*5.8)
			ripple := 0.38 + 0.34*math.Cos(phase*0.76-band*3.6+stack*1.8)
			hotspot := 0.0
			if math.Abs(band-(0.5+0.22*math.Sin(phase*0.21))) < 0.11 {
				hotspot = 68 * (1 - math.Min(1, math.Abs(stack-(0.52+0.26*math.Cos(phase*0.17+band*2.3)))/0.24))
			}
			values[row][col] = 24 + (wave+ripple)*148 + hotspot
		}
	}

	return xLabels, yLabels, values
}

// telemetryHeatmapFrameB generates a companion heatmap with a different motion field.
func telemetryHeatmapFrameB(phase float64) ([]string, []string, [][]float64) {
	xLabels, yLabels := telemetryHeatmapLabels()
	values := make([][]float64, len(yLabels))

	for row := range values {
		values[row] = make([]float64, len(xLabels))
		for col := range values[row] {
			band := float64(col) / math.Max(float64(len(xLabels)-1), 1)
			stack := float64(row) / math.Max(float64(len(yLabels)-1), 1)
			sweep := 0.48 + 0.42*math.Sin(phase*0.54+band*7.2-stack*3.1)
			pulse := 0.30 + 0.28*math.Cos(phase*1.06+stack*6.0+band*2.2)
			ring := math.Exp(-math.Pow(band-(0.22+0.48*math.Sin(phase*0.13+stack*1.7)), 2) / 0.010)
			values[row][col] = 18 + (sweep+pulse)*126 + ring*92
		}
	}

	return xLabels, yLabels, values
}

// telemetryHeatmapLabels returns the shared thermal-matrix labels with a reduced
// visible label density so the heatmaps stay readable at cell width 1.
func telemetryHeatmapLabels() ([]string, []string) {
	xLabels := make([]string, heatmapGridSize)
	for i := range xLabels {
		if i%4 == 0 {
			idx := i / 4
			if idx < 26 {
				xLabels[i] = string(rune('A' + idx))
			} else {
				xLabels[i] = fmt.Sprintf("%c%c", rune('A'+(idx-26)/26), rune('A'+(idx-26)%26))
			}
		}
	}
	yLabels := make([]string, heatmapGridSize)
	for i := range yLabels {
		if i%4 == 0 || i == len(yLabels)-1 {
			yLabels[i] = fmt.Sprintf("%02d", i+1)
		}
	}
	return xLabels, yLabels
}

// telemetryPieValues derives a stable band mix from the current stereo channels.
func telemetryPieValues(left, right []int) []int {
	bands := make([]int, 4)
	if len(left) == 0 || len(right) == 0 {
		return []int{1, 1, 1, 1}
	}

	for i := range left {
		band := i * len(bands) / len(left)
		if band >= len(bands) {
			band = len(bands) - 1
		}
		bands[band] += left[i] + right[i]
	}
	for i := range bands {
		bands[i] = 1 + bands[i]/maxInt(len(left)*12, 1)
	}
	return bands
}

// animateOverview feeds the first tab with rolling CPU data.
func animateOverview(ctx context.Context, widgets *overviewWidgets) {
	ticker := time.NewTicker(dataInterval)
	defer ticker.Stop()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	appNames := []string{"Chrome", "VSCode", "Terminal", "Slack", "Docker", "Mail", "Notes", "Zoom"}
	var pidLines []string

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cpuUsage := 35 + rnd.Float64()*55
			if err := widgets.cpuGauge.Percent(int(cpuUsage), gauge.TextLabel(fmt.Sprintf("CPU %.1f%%", cpuUsage))); err != nil {
				log.Printf("failed to update CPU gauge: %v", err)
			}

			memUsage := 52 + rnd.Intn(20)
			if err := widgets.memGauge.Percent(memUsage, gauge.TextLabel(fmt.Sprintf("MEM %d%%", memUsage))); err != nil {
				log.Printf("failed to update memory gauge: %v", err)
			}

			diskUsage := 44 + rnd.Intn(6)
			if err := widgets.diskGauge.Percent(diskUsage, gauge.TextLabel(fmt.Sprintf("DISK %d%%", diskUsage))); err != nil {
				log.Printf("failed to update disk gauge: %v", err)
			}

			widgets.cpuText.Reset()
			threadCount := 120 + rnd.Intn(55)
			queueDepth := 8 + rnd.Intn(14)
			if err := widgets.cpuText.Write(fmt.Sprintf("Usage:    %.1f%%\nThreads:  %d\nQueue:    %d\nCtx/s:    %d", cpuUsage, threadCount, queueDepth, 2600+rnd.Intn(700))); err != nil {
				log.Printf("failed to update CPU details: %v", err)
			}
			widgets.summary.Reset()
			if err := widgets.summary.Write(fmt.Sprintf("Thermals:  nominal\nScheduler: balanced\nMem free:  %d%%", 100-memUsage)); err != nil {
				log.Printf("failed to update host summary: %v", err)
			}

			pidLines = append(pidLines, fmt.Sprintf("PID %d: %s", 1000+rnd.Intn(9000), appNames[rnd.Intn(len(appNames))]))
			if len(pidLines) > 18 {
				pidLines = pidLines[1:]
			}
			widgets.pidText.Reset()
			if err := widgets.pidText.Write(strings.Join(pidLines, "\n")); err != nil {
				log.Printf("failed to update process list: %v", err)
			}
		}
	}
}

// animateTelemetry feeds the third tab with rolling spectrum and line chart data.
func animateTelemetry(ctx context.Context, widgets *telemetryWidgets) {
	ticker := time.NewTicker(dataInterval)
	defer ticker.Stop()

	model := &telemetryModel{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			latest, peak, load := model.step()
			if err := widgets.line.Series("latency", append([]float64(nil), model.line...), linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(75)))); err != nil {
				log.Printf("failed to update line chart: %v", err)
			}
			xLabels, yLabels, heatValues := telemetryHeatmapFrame(model.phase)
			if err := widgets.heatmap.Values(xLabels, yLabels, heatValues); err != nil {
				log.Printf("failed to update heatmap: %v", err)
			}
			widgets.heatmap.ClearXLabels()
			xLabelsB, yLabelsB, heatValuesB := telemetryHeatmapFrameB(model.phase)
			if err := widgets.heatmapB.Values(xLabelsB, yLabelsB, heatValuesB); err != nil {
				log.Printf("failed to update companion heatmap: %v", err)
			}
			widgets.heatmapB.ClearXLabels()
			if err := widgets.spectrum.SetStereo(append([]int(nil), model.left...), append([]int(nil), model.right...)); err != nil {
				log.Printf("failed to update spectrum: %v", err)
			}
			if err := widgets.pie.Values(telemetryPieValues(model.left, model.right)); err != nil {
				log.Printf("failed to update pie: %v", err)
			}
			if err := widgets.donut.Percent(int(load), donut.Label(fmt.Sprintf("%.0f%%", load))); err != nil {
				log.Printf("failed to update channel load: %v", err)
			}
			if err := renderTelemetryStatus(widgets.status, latest, peak, load); err != nil {
				log.Printf("failed to update telemetry status: %v", err)
			}
		}
	}
}

// animateThreeD rotates the main stage cube and the pyramid demo each tick.
func animateThreeD(ctx context.Context, widgets *threeDWidgets) {
	ticker := time.NewTicker(redrawInterval)
	defer ticker.Stop()

	step := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			widgets.stage.Rotate(threed.Vector3D{Y: 0.015})
			widgets.pyramid.Rotate(threed.Vector3D{X: 0.010, Y: 0.018})
			if err := renderThreeDStatus(widgets.status, step); err != nil {
				log.Printf("failed to update 3D status: %v", err)
			}
			step++
		}
	}
}

// step advances the rolling telemetry model one frame.
func (m *telemetryModel) step() (latest, peak, load float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.phase += 0.35

	latest = 360 + 145*math.Sin(m.phase) + 95*math.Sin(m.phase*0.37+0.5) + 40*math.Cos(m.phase*1.9)
	if latest < 120 {
		latest = 120
	}
	if latest > 700 {
		latest = 700
	}
	m.line = appendRollingFloat(m.line, latest, maxHistory)

	m.left = m.left[:0]
	m.right = m.right[:0]
	for i := 0; i < 40; i++ {
		phase := m.phase + float64(i)*0.18
		left := 220 + int(150*math.Sin(phase)+120*math.Cos(phase*0.33))
		right := 210 + int(145*math.Cos(phase*0.92)+105*math.Sin(phase*0.41))
		m.left = append(m.left, clampInt(left, 30, 580))
		m.right = append(m.right, clampInt(right, 30, 580))
	}

	for _, v := range m.line {
		if v > peak {
			peak = v
		}
	}
	load = (latest / 700) * 100
	return latest, peak, load
}

// runAlertDrill periodically raises a real tab notification so the demo can
// show FollowNotifications without making every tab feel noisy.
func runAlertDrill(ctx context.Context, tm *tab.Manager, refresher interface{ Refresh() }, status *text.Text) {
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	count := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			count++
			headline, detail := alertDrillMessage(count)
			if status != nil {
				if err := renderAlertStatus(status, count, headline, detail); err != nil {
					log.Printf("failed to update alert status: %v", err)
				}
			}
			if tm.SetNotification(alertTabIndex, true, 6*time.Second) {
				if refresher != nil {
					refresher.Refresh()
				}
			}
			timer.Reset(30 * time.Second)
		}
	}
}

// mustTextWidget creates a small empty text widget for fixed-size spacer panes.
func mustTextWidget(content string) *text.Text {
	w, err := text.New()
	if err != nil {
		panic(err)
	}
	if content != "" {
		if err := w.Write(content); err != nil {
			panic(err)
		}
	}
	return w
}

// alertDrillMessage returns restrained alarm text for the notification demo.
func alertDrillMessage(count int) (headline, detail string) {
	switch count % 3 {
	case 1:
		return "Thermal matrix drift", "Alerts tab selected by notification-follow hook"
	case 2:
		return "Channel load crest", "Band mix and load widgets promoted for inspection"
	default:
		return "Subspace threshold", "Signal telemetry exceeded the configured watch band"
	}
}

// configureBorderChrome wires focused pane borders through borderfx.
func configureBorderChrome(c *container.Container) *borderfx.Animator {
	fx := borderfx.NewAnimator(c)
	fx.SetTickRate(64 * time.Millisecond)
	fx.SetInactiveStyle(tabInactiveBorderStyle)
	for _, id := range animatedPaneIDs {
		fx.RegisterMacro(id, borderfx.Presets.Interlace, borderPaletteForPane(id))
	}
	// ThreeD tab panes get a title sweep effect for a polished active look.
	fx.Register(idThreeDStage, borderfx.TextSweep(cell.ColorWhite, cell.ColorNumber(236)))
	fx.Register(idThreeDStatus, borderfx.TextSweep(cell.ColorWhite, cell.ColorNumber(236)))
	fx.Register(idThreeDPyramid, borderfx.TextSweep(cell.ColorWhite, cell.ColorNumber(236)))
	return fx
}

// tabInactiveBorderStyle greys idle panes for a clean, receded look.
func tabInactiveBorderStyle(id string, bc container.BorderCell) container.BorderCellStyle {
	_ = id
	color := cell.ColorNumber(236)
	if bc.Title {
		color = cell.ColorNumber(242)
	}
	return container.BorderCellStyle{
		Rune:     bc.Rune,
		CellOpts: []cell.Option{cell.FgColor(color)},
	}
}

// borderPaletteForPane assigns a subtle accent palette per demo pane.
func borderPaletteForPane(id string) borderfx.Palette {
	_ = id
	return borderfx.Colors(
		cell.ColorNumber(75),
		cell.ColorNumber(243),
		cell.ColorNumber(236),
	)
}

// onOff renders a boolean state in demo-friendly language.
func onOff(v bool) string {
	if v {
		return "ONLINE"
	}
	return "OFFLINE"
}

// boolColor returns the status color for an on/off value.
func boolColor(v bool) cell.Color {
	if v {
		return cell.ColorGreen
	}
	return cell.ColorRed
}

// appendRollingFloat appends one value while keeping the slice length bounded.
func appendRollingFloat(values []float64, next float64, max int) []float64 {
	values = append(values, next)
	if len(values) > max {
		values = values[len(values)-max:]
	}
	return values
}

// clampInt constrains v to the provided inclusive range.
func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// maxInt returns the larger of the two provided integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
