// Copyright 2019 Google Inc.
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

// Binary termdashdemo showcases every termdash widget across four themed tabs.
// Navigate with ←/→ arrow keys or Tab. Press q or Esc to quit.
package main

import (
	"context"
	"flag"
	"fmt"
	"image"
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
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/borderfx"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/checkbox"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/dropdown"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/heatmap"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/modal"
	"github.com/mum4k/termdash/widgets/pie"
	"github.com/mum4k/termdash/widgets/radar"
	"github.com/mum4k/termdash/widgets/radio"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/mum4k/termdash/widgets/slider"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/spectrum"
	"github.com/mum4k/termdash/widgets/spinner"
	"github.com/mum4k/termdash/widgets/tab"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
	"github.com/mum4k/termdash/widgets/threed"
	"github.com/mum4k/termdash/widgets/timeline"
	"github.com/mum4k/termdash/widgets/treeview"
)

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

// Container IDs used by the tab system and borderfx.
const (
	idPageShell = "tabContent"
	idRoot      = "root"

	// Tab 1 – Dashboard
	idDashSeg   = "demo-dash-seg"
	idDashRoll  = "demo-dash-roll"
	idDashSpark = "demo-dash-spark"
	idDashGauge = "demo-dash-gauge"
	idDashLC    = "demo-dash-lc"
	idDashBar   = "demo-dash-bar"
	idDashDonut = "demo-dash-donut"
	idDashSine  = "demo-dash-sine"

	// Tab 2 – Controls
	idCtrlSeg     = "demo-ctrl-seg"
	idCtrlInput   = "demo-ctrl-input"
	idCtrlSlider  = "demo-ctrl-slider"
	idCtrlCheck   = "demo-ctrl-check"
	idCtrlDrop    = "demo-ctrl-drop"
	idCtrlRadio   = "demo-ctrl-radio"
	idCtrlActions = "demo-ctrl-actions"
	idCtrlStatus  = "demo-ctrl-status"

	// Tab 3 – Visualize
	idVizModal   = "demo-viz-modal"
	idVizHeatmap = "demo-viz-heatmap"
	idVizPie     = "demo-viz-pie"
	idVizStatus  = "demo-viz-status"

	// Tab 4 – Explorer
	idExplTree = "demo-expl-tree"
	idExplTime = "demo-expl-timeline"
	idExplPick = "demo-expl-picker"
	idExplPrev = "demo-expl-preview"
	idExplSpin = "demo-expl-spin"

	// Tab 5 – ThreeD
	idThreeDStage = "demo-threed-stage"
	idThreeDInfo  = "demo-threed-info"
)

// animatedPaneIDs lists every pane border registered with borderfx.
var animatedPaneIDs = []string{
	idDashSeg, idDashRoll, idDashSpark, idDashGauge,
	idDashLC, idDashBar, idDashDonut, idDashSine,
	idCtrlSeg, idCtrlInput, idCtrlSlider, idCtrlCheck,
	idCtrlDrop, idCtrlRadio, idCtrlActions, idCtrlStatus,
	idVizModal, idVizHeatmap, idVizPie, idVizStatus,
	idExplTree, idExplTime, idExplPick, idExplPrev, idExplSpin,
	idThreeDStage, idThreeDInfo,
}

// ─────────────────────────────────────────────────────────────────────────────
// Shared helpers
// ─────────────────────────────────────────────────────────────────────────────

// paneOpts returns the shared container styling for a named pane.
func paneOpts(id, title string, extras ...container.Option) []container.Option {
	opts := []container.Option{
		container.ID(id),
		container.Border(linestyle.Round),
		container.BorderTitle(" " + title + " "),
		container.BorderTitleAlignCenter(),
		container.BorderColor(cell.ColorNumber(236)),
		container.FocusedColor(cell.ColorNumber(75)),
		container.TitleColor(cell.ColorNumber(242)),
		container.TitleFocusedColor(cell.ColorNumber(159)),
	}
	return append(opts, extras...)
}

// periodic executes fn on every interval until ctx is cancelled.
func periodic(ctx context.Context, interval time.Duration, fn func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// rotateFloats rotates s left by step.
func rotateFloats(s []float64, step int) []float64 {
	return append(s[step:], s[:step]...)
}

// rotateRunes rotates s left by step.
func rotateRunes(s []rune, step int) []rune {
	return append(s[step:], s[:step]...)
}

// textState builds a rotating character-scroll state for the segment display.
func textState(txt string, capacity, step int) []rune {
	if capacity == 0 {
		return nil
	}
	var state []rune
	for i := 0; i < capacity; i++ {
		state = append(state, ' ')
	}
	state = append(state, []rune(txt)...)
	step = step % len(state)
	return rotateRunes(state, step)
}

// appendRollingFloat appends v and trims the slice to max length.
func appendRollingFloat(vals []float64, v float64, max int) []float64 {
	vals = append(vals, v)
	if len(vals) > max {
		vals = vals[len(vals)-max:]
	}
	return vals
}

// clampInt constrains v to [lo, hi].
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ─────────────────────────────────────────────────────────────────────────────
// Tab 1 – Dashboard
// ─────────────────────────────────────────────────────────────────────────────

// dashWidgets holds every widget shown on the Dashboard tab.
type dashWidgets struct {
	segDist  *segmentdisplay.SegmentDisplay
	rollT    *text.Text
	spGreen  *sparkline.SparkLine
	spRed    *sparkline.SparkLine
	gauge    *gauge.Gauge
	heartLC  *linechart.LineChart
	barChart *barchart.BarChart
	donut    *donut.Donut
	leftB    *button.Button
	rightB   *button.Button
	sineLC   *linechart.LineChart
}

// distance is a thread-safe integer shared between sine buttons and the chart.
type distance struct {
	v  int
	mu sync.Mutex
}

func (d *distance) add(v int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.v += v
}
func (d *distance) get() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.v
}

// newDashboardTab creates all Dashboard widgets and returns the tab.
func newDashboardTab(ctx context.Context, t terminalapi.Terminal) (*dashWidgets, *tab.Tab, error) {
	// ── Segment display ──────────────────────────────────────────────────────
	sd, err := newSegmentDisplay(ctx, t)
	if err != nil {
		return nil, nil, err
	}

	// ── Rolling text ─────────────────────────────────────────────────────────
	rollT, err := text.New(text.RollContent())
	if err != nil {
		return nil, nil, err
	}
	lineNum := 0
	go periodic(ctx, 1*time.Second, func() error {
		err := rollT.Write(fmt.Sprintf("line %d\n", lineNum), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(142))))
		lineNum++
		return err
	})

	// ── SparkLines ───────────────────────────────────────────────────────────
	spGreen, err := sparkline.New(sparkline.Color(cell.ColorGreen))
	if err != nil {
		return nil, nil, err
	}
	go periodic(ctx, 250*time.Millisecond, func() error {
		return spGreen.Add([]int{rand.Intn(101)})
	})

	spRed, err := sparkline.New(sparkline.Color(cell.ColorRed))
	if err != nil {
		return nil, nil, err
	}
	go periodic(ctx, 500*time.Millisecond, func() error {
		return spRed.Add([]int{rand.Intn(101)})
	})

	// ── Gauge ─────────────────────────────────────────────────────────────────
	g, err := gauge.New(gauge.Color(cell.ColorNumber(75)))
	if err != nil {
		return nil, nil, err
	}
	progress := 35
	go periodic(ctx, 2*time.Second, func() error {
		if err := g.Percent(progress, gauge.TextLabel(fmt.Sprintf("%d%%", progress))); err != nil {
			return err
		}
		progress++
		if progress > 100 {
			progress = 35
		}
		return nil
	})

	// ── Heartbeat line chart ──────────────────────────────────────────────────
	var hbInputs []float64
	for i := 0; i < 100; i++ {
		v := math.Pow(math.Sin(float64(i)), 63) * math.Sin(float64(i)+1.5) * 8
		hbInputs = append(hbInputs, v)
	}
	heartLC, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
	if err != nil {
		return nil, nil, err
	}
	hbStep := 0
	go periodic(ctx, redrawInterval/3, func() error {
		hbStep = (hbStep + 1) % len(hbInputs)
		return heartLC.Series("heartbeat", rotateFloats(hbInputs, hbStep),
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
			linechart.SeriesXLabels(map[int]string{0: "zero"}),
		)
	})

	// ── Bar chart ────────────────────────────────────────────────────────────
	bc, err := barchart.New(
		barchart.BarColors([]cell.Color{
			cell.ColorNumber(33), cell.ColorNumber(39), cell.ColorNumber(45),
			cell.ColorNumber(51), cell.ColorNumber(81), cell.ColorNumber(87),
		}),
		barchart.ValueColors([]cell.Color{
			cell.ColorBlack, cell.ColorBlack, cell.ColorBlack,
			cell.ColorBlack, cell.ColorBlack, cell.ColorBlack,
		}),
		barchart.ShowValues(),
	)
	if err != nil {
		return nil, nil, err
	}
	bcVals := make([]int, 6)
	go periodic(ctx, 1*time.Second, func() error {
		for j := range bcVals {
			bcVals[j] = rand.Intn(101)
		}
		return bc.Values(bcVals, 100)
	})

	// ── Donut ────────────────────────────────────────────────────────────────
	don, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorNumber(33))),
		donut.HolePercent(40),
		donut.TextCellOpts(cell.Bold(), cell.FgColor(cell.ColorWhite)),
	)
	if err != nil {
		return nil, nil, err
	}
	donPct := 35
	go periodic(ctx, 500*time.Millisecond, func() error {
		if err := don.Percent(donPct, donut.Label(fmt.Sprintf("%d%%", donPct))); err != nil {
			return err
		}
		donPct++
		if donPct > 100 {
			donPct = 35
		}
		return nil
	})

	// ── Sine line chart + buttons ─────────────────────────────────────────────
	var sineInputs []float64
	for i := 0; i < 200; i++ {
		sineInputs = append(sineInputs, math.Sin(float64(i)/100*math.Pi))
	}
	sineLC, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
	if err != nil {
		return nil, nil, err
	}
	sineStep := 0
	sineDist := &distance{v: 100}
	go periodic(ctx, redrawInterval/3, func() error {
		sineStep = (sineStep + 1) % len(sineInputs)
		if err := sineLC.Series("first", rotateFloats(sineInputs, sineStep),
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(33))),
		); err != nil {
			return err
		}
		step2 := (sineStep + sineDist.get()) % len(sineInputs)
		return sineLC.Series("second", rotateFloats(sineInputs, step2),
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorWhite)),
		)
	})

	const diff = 20
	leftB, err := button.New("(l)eft", func() error { sineDist.add(diff); return nil },
		button.GlobalKey('l'),
		button.WidthFor("(r)ight"),
		button.FillColor(cell.ColorNumber(220)),
		button.FocusedFillColor(cell.ColorNumber(111)),
	)
	if err != nil {
		return nil, nil, err
	}
	rightB, err := button.New("(r)ight", func() error { sineDist.add(-diff); return nil },
		button.GlobalKey('r'),
		button.FillColor(cell.ColorNumber(196)),
		button.FocusedFillColor(cell.ColorNumber(111)),
	)
	if err != nil {
		return nil, nil, err
	}

	w := &dashWidgets{
		segDist: sd, rollT: rollT,
		spGreen: spGreen, spRed: spRed,
		gauge: g, heartLC: heartLC, barChart: bc, donut: don,
		leftB: leftB, rightB: rightB, sineLC: sineLC,
	}

	// ── Layout ───────────────────────────────────────────────────────────────
	left := container.SplitHorizontal(
		container.Top(
			paneOpts(idDashSeg, "Segment Display",
				container.PlaceWidget(sd))...,
		),
		container.Bottom(
			container.SplitHorizontal(
				container.Top(
					container.SplitVertical(
						container.Left(
							paneOpts(idDashRoll, "Rolling Log",
								container.PlaceWidget(rollT))...,
						),
						container.Right(
							container.SplitHorizontal(
								container.Top(
									paneOpts(idDashSpark, "SparkLine",
										container.PlaceWidget(spGreen))...,
								),
								container.Bottom(
									container.Border(linestyle.Round),
									container.BorderColor(cell.ColorNumber(236)),
									container.FocusedColor(cell.ColorNumber(75)),
									container.PlaceWidget(spRed),
								),
							),
						),
					),
				),
				container.Bottom(
					container.SplitHorizontal(
						container.Top(
							paneOpts(idDashGauge, "Gauge",
								container.PlaceWidget(g))...,
						),
						container.Bottom(
							paneOpts(idDashLC, "Heartbeat",
								container.PlaceWidget(heartLC))...,
						),
						container.SplitPercent(28),
					),
				),
				container.SplitPercent(68),
			),
		),
		container.SplitPercent(22),
	)

	right := container.SplitHorizontal(
		container.Top(
			container.SplitVertical(
				container.Left(
					paneOpts(idDashBar, "Bar Chart",
						container.PlaceWidget(bc))...,
				),
				container.Right(
					paneOpts(idDashDonut, "Donut",
						container.PlaceWidget(don))...,
				),
				container.SplitPercent(55),
			),
		),
		container.Bottom(
			container.SplitHorizontal(
				container.Top(
					paneOpts(idDashSine, "Sine Waves",
						container.PlaceWidget(sineLC))...,
				),
				container.Bottom(
					container.SplitVertical(
						container.Left(
							container.PlaceWidget(leftB),
							container.AlignHorizontal(align.HorizontalRight),
							container.PaddingRight(1),
						),
						container.Right(
							container.PlaceWidget(rightB),
							container.AlignHorizontal(align.HorizontalLeft),
							container.PaddingLeft(1),
						),
					),
				),
				container.SplitPercent(86),
			),
		),
		container.SplitPercent(46),
	)

	content := container.SplitVertical(
		container.Left(left),
		container.Right(right),
		container.SplitPercent(68),
	)

	return w, &tab.Tab{Name: "Dashboard", Content: content}, nil
}

// newSegmentDisplay creates the animated SegmentDisplay for the Dashboard.
func newSegmentDisplay(ctx context.Context, t terminalapi.Terminal) (*segmentdisplay.SegmentDisplay, error) {
	sd, err := segmentdisplay.New()
	if err != nil {
		return nil, err
	}
	colors := []cell.Color{
		cell.ColorNumber(33), cell.ColorRed, cell.ColorYellow,
		cell.ColorNumber(33), cell.ColorGreen, cell.ColorRed,
		cell.ColorGreen, cell.ColorRed,
	}
	txt := "Termdash"
	step := 0
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		capacity := 0
		termSize := t.Size()
		for {
			select {
			case <-ticker.C:
				if capacity == 0 {
					capacity = sd.Capacity()
				}
				if t.Size().Eq(image.ZP) || !t.Size().Eq(termSize) {
					termSize = t.Size()
					capacity = sd.Capacity()
				}
				state := textState(txt, capacity, step)
				var chunks []*segmentdisplay.TextChunk
				for i := 0; i < capacity; i++ {
					if i >= len(state) {
						break
					}
					chunks = append(chunks, segmentdisplay.NewChunk(
						string(state[i]),
						segmentdisplay.WriteCellOpts(cell.FgColor(colors[i%len(colors)])),
					))
				}
				if len(chunks) == 0 {
					continue
				}
				if err := sd.Write(chunks); err != nil {
					panic(err)
				}
				step++
			case <-ctx.Done():
				return
			}
		}
	}()
	return sd, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Tab 2 – Controls
// ─────────────────────────────────────────────────────────────────────────────

// ctrlState holds the live values from all form controls.
type ctrlState struct {
	mu      sync.Mutex
	warp    bool
	shields int
	mode    string
	alarm   string
	actions int
	display string
}

// ctrlWidgets groups Tab 2's interactive widgets.
type ctrlWidgets struct {
	segDisplay *segmentdisplay.SegmentDisplay
	input      *textinput.TextInput
	warpCheck  *checkbox.Checkbox
	shields    *slider.Slider
	modeRadio  *radio.Radio
	alarmDrop  *dropdown.Dropdown
	applyB     *button.Button
	resetB     *button.Button
	clearB     *button.Button
	status     *text.Text
	state      *ctrlState
}

// newControlsTab creates all Controls widgets and returns the tab.
func newControlsTab() (*ctrlWidgets, *tab.Tab, error) {
	state := &ctrlState{
		warp:    true,
		shields: 78,
		mode:    "SCAN",
		alarm:   "400",
		display: "READY",
	}

	// Status text (updated by every control)
	statusW, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, nil, err
	}

	// Segment display
	display, err := segmentdisplay.New(
		segmentdisplay.AlignHorizontal(align.HorizontalCenter),
		segmentdisplay.AlignVertical(align.VerticalMiddle),
		segmentdisplay.MaximizeSegmentHeight(),
	)
	if err != nil {
		return nil, nil, err
	}

	w := &ctrlWidgets{status: statusW, segDisplay: display, state: state}

	writeStatus := func() error { return renderCtrlStatus(statusW, state) }
	writeDisplay := func() error { return renderCtrlDisplay(display, state) }

	// Checkbox
	warpCheck, err := checkbox.New("Enable Warp Assist",
		checkbox.Checked(state.warp),
		checkbox.UseIndicatorSet(checkbox.IndicatorSets.Rounded),
		checkbox.OnChange(func(checked bool) error {
			state.mu.Lock()
			state.warp = checked
			state.mu.Unlock()
			_ = writeDisplay()
			return writeStatus()
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	// Slider – SegmentedStyle blocks
	shieldsSlider, err := slider.New(
		slider.Min(0),
		slider.Max(100),
		slider.Value(state.shields),
		slider.Width(28),
		slider.SegmentedStyle(),
		slider.FillCellOpts(cell.FgColor(cell.ColorNumber(75))),
		slider.OnChange(func(v int) error {
			state.mu.Lock()
			state.shields = v
			state.mu.Unlock()
			_ = writeDisplay()
			return writeStatus()
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	// Radio
	modeRadio, err := radio.New(
		[]radio.Item{
			{Label: "SCAN", SelectedCellOpts: []cell.Option{cell.FgColor(cell.ColorCyan)}},
			{Label: "BOOST", SelectedCellOpts: []cell.Option{cell.FgColor(cell.ColorYellow)}},
			{Label: "STEALTH", SelectedCellOpts: []cell.Option{cell.FgColor(cell.ColorGreen)}},
		},
		radio.UseIndicatorSet(radio.IndicatorSets.Circle),
		radio.Gap(1),
		radio.OnChange(func(_ int, label string) error {
			state.mu.Lock()
			state.mode = label
			state.mu.Unlock()
			_ = writeDisplay()
			return writeStatus()
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	// Dropdown – IntRange returns []string; pass it as first arg, not as an option.
	alarmChoices := dropdown.IntRange(200, 600, 100, "%d")
	alarmDrop, err := dropdown.New(alarmChoices,
		dropdown.Selected(2),
		dropdown.Width(18),
		dropdown.GlyphSet(dropdown.GlyphProfiles.Minimal),
		dropdown.OnSelect(func(_ int, label string) error {
			state.mu.Lock()
			state.alarm = label
			state.mu.Unlock()
			_ = writeDisplay()
			return writeStatus()
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	// TextInput
	inputW, err := textinput.New(
		textinput.Border(linestyle.Round),
		textinput.BorderColor(cell.ColorNumber(240)),
		textinput.PlaceHolder("enter command..."),
		textinput.PlaceHolderColor(cell.ColorNumber(240)),
		textinput.MaxWidthCells(40),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.ClearOnSubmit(),
		textinput.OnSubmit(func(cmd string) error {
			cmd = strings.TrimSpace(strings.ToUpper(cmd))
			if cmd == "" {
				return nil
			}
			state.mu.Lock()
			state.display = cmd
			state.actions++
			state.mu.Unlock()
			_ = writeDisplay()
			return writeStatus()
		}),
	)
	if err != nil {
		return nil, nil, err
	}

	// Action buttons
	btnOpts := func(fill cell.Color) []button.Option {
		return []button.Option{
			button.WidthFor("Stabilize"),
			button.FillColor(fill),
			button.FocusedFillColor(cell.ColorNumber(111)),
			button.PressedFillColor(cell.ColorNumber(159)),
			button.ShadowColor(cell.ColorNumber(235)),
		}
	}
	applyB, err := button.New("Route", func() error {
		return applyCtrlPreset(w, state, true, 86, "400", "SCAN", "ROUTE")
	}, btnOpts(cell.ColorNumber(60))...)
	if err != nil {
		return nil, nil, err
	}
	resetB, err := button.New("Prime", func() error {
		return applyCtrlPreset(w, state, true, 94, "500", "BOOST", "PRIME")
	}, btnOpts(cell.ColorNumber(67))...)
	if err != nil {
		return nil, nil, err
	}
	clearB, err := button.New("Stabilize", func() error {
		return applyCtrlPreset(w, state, false, 34, "350", "STEALTH", "STABLE")
	}, btnOpts(cell.ColorNumber(66))...)
	if err != nil {
		return nil, nil, err
	}

	w.warpCheck = warpCheck
	w.shields = shieldsSlider
	w.modeRadio = modeRadio
	w.alarmDrop = alarmDrop
	w.input = inputW
	w.applyB = applyB
	w.resetB = resetB
	w.clearB = clearB

	if err := writeStatus(); err != nil {
		return nil, nil, err
	}
	if err := writeDisplay(); err != nil {
		return nil, nil, err
	}

	// ── Layout ───────────────────────────────────────────────────────────────
	// Three-column design so the dropdown gets a dedicated tall pane.
	// Column widths: left 28% | center 42% | right 30%
	content := container.SplitHorizontal(
		container.Top(
			container.SplitVertical(
				container.Left(
					paneOpts(idCtrlSeg, "Status Display",
						container.PaddingLeft(1),
						container.PlaceWidget(display))...,
				),
				container.Right(
					paneOpts(idCtrlInput, "Command Uplink",
						container.PaddingLeft(1),
						container.PaddingTop(1),
						container.PlaceWidget(inputW))...,
				),
				container.SplitPercent(48),
			),
		),
		container.Bottom(
			container.SplitVertical(
				// ── Left column: slider + checkbox + status ──────────────────
				container.Left(
					container.SplitHorizontal(
						container.Top(
							container.SplitHorizontal(
								container.Top(
									paneOpts(idCtrlSlider, "Shield Bias",
										container.PaddingLeft(1),
										container.PaddingTop(1),
										container.PlaceWidget(shieldsSlider))...,
								),
								container.Bottom(
									paneOpts(idCtrlCheck, "Warp Assist",
										container.PaddingLeft(1),
										container.PaddingTop(1),
										container.PlaceWidget(warpCheck))...,
								),
								container.SplitPercent(58),
							),
						),
						container.Bottom(
							paneOpts(idCtrlStatus, "Status",
								container.PaddingLeft(1),
								container.PaddingTop(1),
								container.PlaceWidget(statusW))...,
						),
						container.SplitPercent(58),
					),
				),
				container.Right(
					container.SplitVertical(
						// ── Center column: dropdown (tall) + radio ──────────
						container.Left(
							container.SplitHorizontal(
								container.Top(
									paneOpts(idCtrlDrop, "Alarm Y",
										container.PaddingLeft(1),
										container.PaddingTop(1),
										container.PlaceWidget(alarmDrop))...,
								),
								container.Bottom(
									paneOpts(idCtrlRadio, "Profile",
										container.PaddingLeft(1),
										container.PaddingTop(1),
										container.PlaceWidget(modeRadio))...,
								),
								container.SplitPercent(62),
							),
						),
						// ── Right column: action deck ────────────────────────
						container.Right(
							paneOpts(idCtrlActions, "Action Deck",
								container.PaddingLeft(1),
								container.PaddingTop(1),
								container.SplitVertical(
									container.Left(container.PlaceWidget(applyB)),
									container.Right(
										container.SplitVertical(
											container.Left(container.PlaceWidget(resetB)),
											container.Right(container.PlaceWidget(clearB)),
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
			),
		),
		container.SplitPercent(28),
	)

	return w, &tab.Tab{Name: "Controls", Content: content}, nil
}

// renderCtrlStatus rewrites the Controls tab status panel.
func renderCtrlStatus(w *text.Text, s *ctrlState) error {
	s.mu.Lock()
	warp, shields, mode, alarm, actions, display := s.warp, s.shields, s.mode, s.alarm, s.actions, s.display
	s.mu.Unlock()

	w.Reset()
	label := cell.FgColor(cell.ColorNumber(245))
	warpStr, warpColor := "OFFLINE", cell.ColorRed
	if warp {
		warpStr, warpColor = "ONLINE", cell.ColorGreen
	}
	lines := []struct {
		k string
		v string
		c cell.Color
	}{
		{"Warp Assist: ", warpStr, warpColor},
		{"Shield Bias: ", fmt.Sprintf("%d%%", shields), cell.ColorNumber(252)},
		{"Profile:     ", mode, cell.ColorNumber(75)},
		{"Alarm Y:     ", alarm, cell.ColorNumber(252)},
		{"Display:     ", display, cell.ColorNumber(159)},
		{"Actions:     ", fmt.Sprintf("%d", actions), cell.ColorNumber(252)},
	}
	for i, ln := range lines {
		if err := w.Write(ln.k, text.WriteCellOpts(label)); err != nil {
			return err
		}
		nl := "\n"
		if i == len(lines)-1 {
			nl = ""
		}
		if err := w.Write(ln.v+nl, text.WriteCellOpts(cell.FgColor(ln.c))); err != nil {
			return err
		}
	}
	return nil
}

// renderCtrlDisplay writes a compact label to the segment display.
func renderCtrlDisplay(w *segmentdisplay.SegmentDisplay, s *ctrlState) error {
	s.mu.Lock()
	d := s.display
	s.mu.Unlock()
	label := strings.ToUpper(strings.TrimSpace(d))
	if label == "" {
		label = "READY"
	}
	// Trim to 8 chars for the segment display
	runes := []rune(label)
	if len(runes) > 8 {
		runes = runes[:8]
	}
	return w.Write([]*segmentdisplay.TextChunk{
		segmentdisplay.NewChunk(string(runes)),
	})
}

// applyCtrlPreset pushes a canned preset through all form controls.
func applyCtrlPreset(w *ctrlWidgets, s *ctrlState, warp bool, shields int, alarm, mode, display string) error {
	s.mu.Lock()
	s.warp = warp
	s.shields = shields
	s.alarm = alarm
	s.mode = mode
	s.display = display
	s.actions++
	s.mu.Unlock()

	w.warpCheck.SetChecked(warp)
	w.shields.SetValue(shields)

	// Map alarm label to dropdown index: range 200..600 step 100 → (val-200)/100
	var alarmVal int
	fmt.Sscanf(alarm, "%d", &alarmVal)
	_ = w.alarmDrop.SetSelected(clampInt((alarmVal-200)/100, 0, 4))

	modeIdx := map[string]int{"SCAN": 0, "BOOST": 1, "STEALTH": 2}[mode]
	_ = w.modeRadio.SetSelected(modeIdx)

	_ = renderCtrlDisplay(w.segDisplay, s)
	return renderCtrlStatus(w.status, s)
}

// ─────────────────────────────────────────────────────────────────────────────
// Tab 3 – Visualize  (Modal · Heatmap · Pie · BorderFX)
// ─────────────────────────────────────────────────────────────────────────────

// vizWidgets groups Tab 3's data-viz widgets.
type vizWidgets struct {
	radarW  *radar.Radar
	donut2  *donut.Donut
	lineW   *linechart.LineChart
	heatW   *heatmap.HeatMap
	pieW    *pie.Pie
	statusW *text.Text
	modalW  *modal.Modal
	phase   float64
	mu      sync.Mutex
}

// newVisualizeTab creates all Visualize widgets and returns the tab.
func newVisualizeTab(ctx context.Context) (*vizWidgets, *tab.Tab, error) {
	// ── Radar ────────────────────────────────────────────────────────────────
	rdr, err := radar.New(
		radar.SweepSpeed(50.0),
		radar.BeamWidth(28.0),
		radar.RangeRings(3),
		radar.BeamColor(0, 200, 64),
		radar.ContactColor(255, 140, 0),
		radar.ContactChar('◆'),
		radar.Border(linestyle.Round),
		radar.BorderTitle(" RADAR SWEEP "),
	)
	if err != nil {
		return nil, nil, err
	}

	// Seed initial contacts.
	initContacts := []*radar.Contact{
		{Angle: 42.0, Distance: 0.35, Label: "A1"},
		{Angle: 158.0, Distance: 0.62, Label: "B2"},
		{Angle: 270.0, Distance: 0.80, Label: "C3"},
	}
	if err := rdr.SetContacts(initContacts); err != nil {
		return nil, nil, err
	}

	// ── Donut (modal) ────────────────────────────────────────────────────────
	don2, err := donut.New(
		donut.HolePercent(35),
		donut.CellOpts(cell.FgColor(cell.ColorNumber(75))),
		donut.TextCellOpts(cell.Bold(), cell.FgColor(cell.ColorWhite)),
		donut.Label("System Load"),
		donut.LabelAlign(align.HorizontalCenter),
	)
	if err != nil {
		return nil, nil, err
	}
	if err := don2.Percent(62, donut.Label("62%")); err != nil {
		return nil, nil, err
	}

	// ── LineChart with threshold (modal) ─────────────────────────────────────
	lineW, err := linechart.New(
		linechart.BrailleOnly(),
		linechart.ThresholdLine(0.75, cell.FgColor(cell.ColorNumber(167))),
		linechart.YAxisCustomScale(-1.1, 1.1),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorNumber(240))),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorNumber(244))),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorNumber(244))),
	)
	if err != nil {
		return nil, nil, err
	}

	// ── Build Modal with three draggable sub-windows ──────────────────────────
	radarDW := modal.NewDraggableWidget("modal-radar", rdr, 0, 0, 42, 24, nil)
	radarDW.Title = "Radar Sweep"
	radarDW.Border = true
	radarDW.Minimizable = true

	donutDW := modal.NewDraggableWidget("modal-donut", don2, 44, 0, 26, 12, nil)
	donutDW.Title = "System Load"
	donutDW.Border = true
	donutDW.Minimizable = true

	lcDW := modal.NewDraggableWidget("modal-lc", lineW, 44, 13, 44, 13, nil)
	lcDW.Title = "Signal · threshold 0.75"
	lcDW.Border = true
	lcDW.Minimizable = true

	modalW := modal.NewModal("viz-modal", []*modal.DraggableWidget{radarDW, donutDW, lcDW}, nil)

	// ── Heatmap ──────────────────────────────────────────────────────────────
	heatW, err := heatmap.New(
		heatmap.AxisCellOpts(cell.FgColor(cell.ColorNumber(242))),
		heatmap.XLabelCellOpts(cell.FgColor(cell.ColorNumber(244))),
		heatmap.YLabelCellOpts(cell.FgColor(cell.ColorNumber(117))),
		heatmap.Palette(
			cell.ColorNumber(236), cell.ColorNumber(239),
			cell.ColorNumber(24), cell.ColorNumber(31),
			cell.ColorNumber(38), cell.ColorNumber(45),
			cell.ColorNumber(81),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	// ── Pie ──────────────────────────────────────────────────────────────────
	pieW, err := pie.New(
		pie.ColorOption([]cell.Color{
			cell.ColorNumber(75), cell.ColorNumber(111),
			cell.ColorNumber(147), cell.ColorNumber(183),
			cell.ColorNumber(219),
		}),
	)
	if err != nil {
		return nil, nil, err
	}
	if err := pieW.Values([]int{32, 24, 18, 14, 12}); err != nil {
		return nil, nil, err
	}

	// ── Status text ───────────────────────────────────────────────────────────
	statusW, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, nil, err
	}
	if err := renderVizStatus(statusW, 0, 62.0); err != nil {
		return nil, nil, err
	}

	// Seed heatmap
	xl, yl, hv := vizHeatmapFrame(0)
	if err := heatW.Values(xl, yl, hv); err != nil {
		return nil, nil, err
	}

	w := &vizWidgets{
		radarW: rdr, donut2: don2, lineW: lineW,
		heatW: heatW, pieW: pieW, statusW: statusW, modalW: modalW,
	}

	// ── Layout ───────────────────────────────────────────────────────────────
	content := container.SplitHorizontal(
		container.Top(
			container.SplitVertical(
				container.Left(
					paneOpts(idVizModal, "Interactive Panels",
						container.PlaceWidget(modalW))...,
				),
				container.Right(
					container.SplitHorizontal(
						container.Top(
							paneOpts(idVizHeatmap, "Thermal Heatmap",
								container.PlaceWidget(heatW))...,
						),
						container.Bottom(
							paneOpts(idVizPie, "Band Distribution",
								container.PlaceWidget(pieW))...,
						),
						container.SplitPercent(60),
					),
				),
				container.SplitPercent(60),
			),
		),
		container.Bottom(
			paneOpts(idVizStatus, "Telemetry Status",
				container.PaddingLeft(2),
				container.PaddingTop(1),
				container.PlaceWidget(statusW))...,
		),
		container.SplitPercent(88),
	)

	return w, &tab.Tab{Name: "Visualize", Content: content}, nil
}

// vizHeatmapFrame generates an animated heatmap frame.
func vizHeatmapFrame(phase float64) ([]string, []string, [][]float64) {
	const cols, rows = 14, 10
	xLabels := make([]string, cols)
	for i := range xLabels {
		if i%3 == 0 {
			xLabels[i] = fmt.Sprintf("%c", 'A'+i/3)
		}
	}
	yLabels := make([]string, rows)
	for i := range yLabels {
		if i%2 == 0 {
			yLabels[i] = fmt.Sprintf("%02d", i+1)
		}
	}
	values := make([][]float64, rows)
	for r := range values {
		values[r] = make([]float64, cols)
		for c := range values[r] {
			band := float64(c) / float64(cols-1)
			stack := float64(r) / float64(rows-1)
			wave := 0.5 + 0.5*math.Sin(phase*0.4+stack*2.5+band*5.0)
			ripple := 0.4 + 0.35*math.Cos(phase*0.8-band*3.2+stack*1.9)
			values[r][c] = 20 + (wave+ripple)*130
		}
	}
	return xLabels, yLabels, values
}

// renderVizStatus updates the Visualize tab status panel.
func renderVizStatus(w *text.Text, contacts int, load float64) error {
	w.Reset()
	label := cell.FgColor(cell.ColorNumber(245))
	value := cell.FgColor(cell.ColorNumber(252))
	lines := []struct{ k, v string }{
		{"Radar contacts: ", fmt.Sprintf("%d active", contacts)},
		{"System load:    ", fmt.Sprintf("%.0f%%", load)},
		{"Threshold line: ", "0.75  (red dashed)"},
		{"Heatmap palette:", "7-step thermal blue"},
		{"Pie segments:   ", "5 bands"},
		{"Drag windows:   ", "click & drag modal panels"},
	}
	for i, ln := range lines {
		if err := w.Write(ln.k, text.WriteCellOpts(label)); err != nil {
			return err
		}
		nl := "\n"
		if i == len(lines)-1 {
			nl = ""
		}
		if err := w.Write(ln.v+nl, text.WriteCellOpts(value)); err != nil {
			return err
		}
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Tab 4 – Explorer  (catalog · live previews · timeline scrubber)
// ─────────────────────────────────────────────────────────────────────────────

// explWidgets groups Tab 4's widgets.
type explWidgets struct {
	mu       sync.Mutex
	root     *container.Container
	tree     *treeview.TreeView
	line     *linechart.LineChart
	bar      *barchart.BarChart
	sparkA   *sparkline.SparkLine
	donut    *donut.Donut
	pie      *pie.Pie
	heat     *heatmap.HeatMap
	gauge    *gauge.Gauge
	radar    *radar.Radar
	spectrum *spectrum.Spectrum
	timeLine *timeline.Timeline
	picker   *timeline.TimeRangePicker
	prevTime *timeline.Timeline
	prevPick *timeline.TimeRangePicker
	spinW    *text.Text
	previews map[string]func() []container.Option
}

// newExplorerTab creates all Explorer widgets and returns the tab.
func newExplorerTab(ctx context.Context) (*explWidgets, *tab.Tab, error) {
	w, err := newExplorerWidgets()
	if err != nil {
		return nil, nil, err
	}

	// ── TreeView – termdash widget catalog ────────────────────────────────────
	mkLeaf := func(label string) *treeview.TreeNode {
		n := &treeview.TreeNode{Label: label}
		if _, ok := w.previews[label]; ok {
			n.OnClick = func() error {
				return w.selectCatalogItem(label)
			}
		}
		return n
	}
	mkGroup := func(label string, children ...*treeview.TreeNode) *treeview.TreeNode {
		return &treeview.TreeNode{
			Label:         label,
			ExpandedState: true,
			Children:      children,
		}
	}
	catalog := []*treeview.TreeNode{
		mkGroup("Visualization",
			mkLeaf("LineChart"), mkLeaf("BarChart"), mkLeaf("SparkLine"),
			mkLeaf("Donut"), mkLeaf("Pie"), mkLeaf("HeatMap"),
			mkLeaf("Gauge"), mkLeaf("Radar"), mkLeaf("Spectrum"),
			mkLeaf("Timeline"), mkLeaf("TimeRangePicker"),
		),
		mkGroup("Input Controls",
			mkLeaf("Button"), mkLeaf("TextInput"), mkLeaf("Checkbox"),
			mkLeaf("Radio"), mkLeaf("Slider"), mkLeaf("Dropdown"),
		),
		mkGroup("Display",
			mkLeaf("Text"), mkLeaf("SegmentDisplay"), mkLeaf("ThreeD"),
		),
		mkGroup("Layout & FX",
			mkLeaf("Modal"), mkLeaf("TreeView ← you are here"),
			mkLeaf("BorderFX"), mkLeaf("Tab"),
		),
	}
	tv, err := treeview.New(
		treeview.Nodes(catalog...),
		treeview.LabelColor(cell.ColorNumber(252)),
		treeview.ExpandedIcon("▼"),
		treeview.CollapsedIcon("▶"),
		treeview.LeafIcon("·"),
		treeview.IndentationPerLevel(2),
	)
	if err != nil {
		return nil, nil, err
	}
	w.tree = tv

	// ── Layout ───────────────────────────────────────────────────────────────
	content := container.SplitVertical(
		container.Left(
			paneOpts(idExplTree, "Widget Catalog",
				container.PlaceWidget(tv))...,
		),
		container.Right(
			container.SplitHorizontal(
				container.Top(
					paneOpts(idExplPrev, "LineChart Preview",
						container.PlaceWidget(w.line))...,
				),
				container.Bottom(
					container.SplitHorizontal(
						container.Top(
							paneOpts(idExplPick, "Timeline Range Picker",
								container.PlaceWidget(w.picker))...,
						),
						container.Bottom(
							container.SplitVertical(
								container.Left(
									paneOpts(idExplTime, "Live Event Stream",
										container.PlaceWidget(w.timeLine))...,
								),
								container.Right(
									paneOpts(idExplSpin, "System Status",
										container.PaddingLeft(1),
										container.PaddingTop(1),
										container.PlaceWidget(w.spinW))...,
								),
								container.SplitPercent(64),
							),
						),
						container.SplitPercent(36),
					),
				),
				container.SplitPercent(62),
			),
		),
		container.SplitPercent(32),
	)

	return w, &tab.Tab{Name: "Explorer", Content: content}, nil
}

// setRoot allows catalog clicks to swap the preview pane after the root exists.
func (w *explWidgets) setRoot(root *container.Container) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.root = root
}

// selectCatalogItem swaps the right-hand preview pane to the selected widget.
func (w *explWidgets) selectCatalogItem(name string) error {
	w.mu.Lock()
	root := w.root
	buildPreview := w.previews[name]
	w.mu.Unlock()
	if root == nil || buildPreview == nil {
		return nil
	}
	return root.Update(idExplPrev, buildPreview()...)
}

// newExplorerWidgets creates every live visualization used by the Explorer catalog.
func newExplorerWidgets() (*explWidgets, error) {
	lineW, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorNumber(238))),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorNumber(245))),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorNumber(245))),
		linechart.ThresholdLine(0.82, cell.FgColor(cell.ColorNumber(203))),
	)
	if err != nil {
		return nil, err
	}
	barW, err := barchart.New(
		barchart.BarWidth(6),
		barchart.BarGap(2),
		barchart.ShowValues(),
	)
	if err != nil {
		return nil, err
	}
	sparkW, err := sparkline.New(
		sparkline.Color(cell.ColorNumber(75)),
		sparkline.AlertColor(cell.ColorNumber(203)),
		sparkline.Threshold(78),
		sparkline.Label("Activity", cell.FgColor(cell.ColorNumber(245))),
	)
	if err != nil {
		return nil, err
	}
	donutW, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorNumber(75))),
		donut.TextCellOpts(cell.FgColor(cell.ColorNumber(231)), cell.Bold()),
		donut.Label("COMPUTE", cell.FgColor(cell.ColorNumber(245))),
	)
	if err != nil {
		return nil, err
	}
	pieW, err := pie.New(pie.ColorOption([]cell.Color{
		cell.ColorNumber(75), cell.ColorNumber(118), cell.ColorNumber(220),
		cell.ColorNumber(203), cell.ColorNumber(159),
	}))
	if err != nil {
		return nil, err
	}
	heatW, err := heatmap.New(
		heatmap.AxisCellOpts(cell.FgColor(cell.ColorNumber(240))),
		heatmap.XLabelCellOpts(cell.FgColor(cell.ColorNumber(245))),
		heatmap.YLabelCellOpts(cell.FgColor(cell.ColorNumber(245))),
		heatmap.Palette(cell.ColorNumber(235), cell.ColorNumber(24), cell.ColorNumber(31), cell.ColorNumber(75), cell.ColorNumber(159)),
	)
	if err != nil {
		return nil, err
	}
	gaugeW, err := gauge.New(
		gauge.Height(3),
		gauge.Color(cell.ColorNumber(75)),
		gauge.TextLabel(" throughput"),
		gauge.Threshold(82, linestyle.Light, cell.FgColor(cell.ColorNumber(203))),
	)
	if err != nil {
		return nil, err
	}
	radarW, err := radar.New(
		radar.SweepSpeed(95),
		radar.BeamColor(90, 255, 190),
		radar.ContactColor(255, 210, 90),
		radar.RangeRings(4),
	)
	if err != nil {
		return nil, err
	}
	spectrumW, err := spectrum.New(
		spectrum.Stereo(),
		spectrum.ChannelLabels("LEFT", "RIGHT"),
		spectrum.MaxValue(100),
		spectrum.Gradient(cell.ColorNumber(24), cell.ColorNumber(75), cell.ColorNumber(159), cell.ColorNumber(220), cell.ColorNumber(203)),
		spectrum.AxisCellOpts(cell.FgColor(cell.ColorNumber(238))),
		spectrum.Threshold(78),
		spectrum.ThresholdLineColor(cell.ColorNumber(203)),
		spectrum.AlertColor(cell.ColorNumber(203)),
	)
	if err != nil {
		return nil, err
	}
	timelineW, err := timeline.New(timeline.FollowTail(), timeline.MaxEvents(200))
	if err != nil {
		return nil, err
	}
	pickerW, err := timeline.NewTimeRangePicker(func(start, end time.Time, hasRange bool) {
		if hasRange {
			timelineW.SetTimeFilter(start, end)
			return
		}
		timelineW.ClearTimeFilter()
	})
	if err != nil {
		return nil, err
	}
	timelinePreview, err := timeline.New(timeline.FollowTail(), timeline.MaxEvents(200))
	if err != nil {
		return nil, err
	}
	pickerPreview, err := timeline.NewTimeRangePicker(func(start, end time.Time, hasRange bool) {
		if hasRange {
			timelinePreview.SetTimeFilter(start, end)
			return
		}
		timelinePreview.ClearTimeFilter()
	})
	if err != nil {
		return nil, err
	}
	spinW, err := text.New()
	if err != nil {
		return nil, err
	}
	buttonW, err := button.New("Run Action", func() error {
		return writeExplorerStatus(spinW, "Button press received")
	},
		button.Width(18),
		button.Height(3),
		button.FillColor(cell.ColorNumber(75)),
		button.FocusedFillColor(cell.ColorNumber(159)),
		button.PressedFillColor(cell.ColorNumber(118)),
		button.TextColor(cell.ColorBlack),
	)
	if err != nil {
		return nil, err
	}
	inputW, err := textinput.New(
		textinput.Label("Query", cell.FgColor(cell.ColorNumber(245))),
		textinput.DefaultText("termdash widgets"),
		textinput.PlaceHolder("type here"),
		textinput.FillColor(cell.ColorNumber(234)),
		textinput.TextColor(cell.ColorNumber(231)),
		textinput.CursorColor(cell.ColorNumber(159)),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.ColorNumber(75)),
		textinput.MaxWidthCells(36),
	)
	if err != nil {
		return nil, err
	}
	checkW, err := checkbox.New("Enable alerts",
		checkbox.Checked(true),
		checkbox.UseIndicatorSet(checkbox.IndicatorSets.Heavy),
		checkbox.CellOpts(cell.FgColor(cell.ColorNumber(245))),
		checkbox.FocusedCellOpts(cell.FgColor(cell.ColorNumber(159))),
		checkbox.CheckedCellOpts(cell.FgColor(cell.ColorNumber(118))),
	)
	if err != nil {
		return nil, err
	}
	radioW, err := radio.New([]radio.Item{
		{Label: "Low"},
		{Label: "Balanced"},
		{Label: "High"},
	},
		radio.Selected(1),
		radio.UseIndicatorSet(radio.IndicatorSets.Diamond),
	)
	if err != nil {
		return nil, err
	}
	sliderW, err := slider.New(
		slider.Width(34),
		slider.Value(64),
		slider.Step(5),
		slider.SegmentedBlocksStyle(),
		slider.FillCellOpts(cell.FgColor(cell.ColorNumber(75))),
		slider.TrackCellOpts(cell.FgColor(cell.ColorNumber(238))),
		slider.KnobCellOpts(cell.FgColor(cell.ColorNumber(231))),
		slider.FocusedKnobCellOpts(cell.FgColor(cell.ColorNumber(159))),
	)
	if err != nil {
		return nil, err
	}
	dropdownW, err := dropdown.New([]string{"Telemetry", "Operations", "Charts", "Controls"},
		dropdown.Selected(1),
		dropdown.Width(24),
		dropdown.GlyphSet(dropdown.GlyphProfiles.Minimal),
		dropdown.CellOpts(cell.FgColor(cell.ColorNumber(245)), cell.BgColor(cell.ColorNumber(234))),
		dropdown.FocusedCellOpts(cell.FgColor(cell.ColorNumber(231)), cell.BgColor(cell.ColorNumber(236))),
		dropdown.SelectedCellOpts(cell.FgColor(cell.ColorNumber(159)), cell.BgColor(cell.ColorNumber(236))),
		dropdown.BorderCellOpts(cell.FgColor(cell.ColorNumber(75))),
	)
	if err != nil {
		return nil, err
	}
	textW, err := explorerText("Text widget\n\nRich terminal copy with word wrapping, color, and scroll handling.", cell.ColorNumber(245))
	if err != nil {
		return nil, err
	}
	segmentW, err := explorerSegmentDisplay()
	if err != nil {
		return nil, err
	}
	threedW, err := explorerThreeD()
	if err != nil {
		return nil, err
	}
	treePreview, err := explorerTreeView()
	if err != nil {
		return nil, err
	}
	modalPreview, err := explorerModalPreview()
	if err != nil {
		return nil, err
	}
	borderPreview, err := explorerText("BorderFX preview\n\nFocus this pane and watch the registered border animation render around the selected preview.", cell.ColorNumber(159))
	if err != nil {
		return nil, err
	}
	tabPreview := newExplorerTabPreview()
	w := &explWidgets{
		line: lineW, bar: barW, sparkA: sparkW, donut: donutW, pie: pieW,
		heat: heatW, gauge: gaugeW, radar: radarW, spectrum: spectrumW,
		timeLine: timelineW, picker: pickerW, prevTime: timelinePreview, prevPick: pickerPreview, spinW: spinW,
	}
	w.previews = map[string]func() []container.Option{
		"LineChart":               explorerWidgetPreview("LineChart", lineW),
		"BarChart":                explorerWidgetPreview("BarChart", barW),
		"SparkLine":               explorerWidgetPreview("SparkLine", sparkW),
		"Donut":                   explorerWidgetPreview("Donut", donutW),
		"Pie":                     explorerWidgetPreview("Pie", pieW),
		"HeatMap":                 explorerWidgetPreview("HeatMap", heatW),
		"Gauge":                   explorerWidgetPreview("Gauge", gaugeW),
		"Radar":                   explorerWidgetPreview("Radar", radarW),
		"Spectrum":                explorerWidgetPreview("Spectrum", spectrumW),
		"Timeline":                explorerWidgetPreview("Timeline", timelinePreview),
		"TimeRangePicker":         explorerWidgetPreview("TimeRangePicker", pickerPreview),
		"Button":                  explorerControlPreview("Button", buttonW),
		"TextInput":               explorerControlPreview("TextInput", inputW),
		"Checkbox":                explorerControlPreview("Checkbox", checkW),
		"Radio":                   explorerControlPreview("Radio", radioW),
		"Slider":                  explorerControlPreview("Slider", sliderW),
		"Dropdown":                explorerControlPreview("Dropdown", dropdownW),
		"Text":                    explorerWidgetPreview("Text", textW),
		"SegmentDisplay":          explorerWidgetPreview("SegmentDisplay", segmentW),
		"ThreeD":                  explorerWidgetPreview("ThreeD", threedW),
		"Modal":                   explorerWidgetPreview("Modal", modalPreview),
		"TreeView ← you are here": explorerWidgetPreview("TreeView", treePreview),
		"BorderFX":                explorerWidgetPreview("BorderFX", borderPreview),
		"Tab":                     explorerWidgetPreview("Tab", tabPreview),
	}
	seedExplorerTimeline(timelineW, pickerW)
	seedExplorerTimeline(timelinePreview, pickerPreview)
	if err := writeExplorerStatus(spinW, "Catalog ready"); err != nil {
		return nil, err
	}
	return w, nil
}

// explorerWidgetPreview returns a standard preview pane for one catalog item.
func explorerWidgetPreview(title string, widget widgetapi.Widget) func() []container.Option {
	return func() []container.Option {
		return paneOpts(idExplPrev, title+" Preview", container.PlaceWidget(widget))
	}
}

// explorerControlPreview adds padding around compact input widgets.
func explorerControlPreview(title string, widget widgetapi.Widget) func() []container.Option {
	return func() []container.Option {
		return paneOpts(idExplPrev, title+" Preview",
			container.PaddingLeft(2),
			container.PaddingTop(2),
			container.PlaceWidget(widget),
		)
	}
}

// explorerText creates a one-off text widget for catalog previews.
func explorerText(copy string, color cell.Color) (*text.Text, error) {
	w, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, err
	}
	if err := w.Write(copy, text.WriteCellOpts(cell.FgColor(color))); err != nil {
		return nil, err
	}
	return w, nil
}

// explorerSegmentDisplay creates a real SegmentDisplay preview.
func explorerSegmentDisplay() (*segmentdisplay.SegmentDisplay, error) {
	sd, err := segmentdisplay.New(segmentdisplay.MaximizeDisplayedText())
	if err != nil {
		return nil, err
	}
	chunks := []*segmentdisplay.TextChunk{
		segmentdisplay.NewChunk("TERM", segmentdisplay.WriteCellOpts(cell.FgColor(cell.ColorNumber(75)))),
		segmentdisplay.NewChunk("DASH", segmentdisplay.WriteCellOpts(cell.FgColor(cell.ColorNumber(118)))),
	}
	if err := sd.Write(chunks); err != nil {
		return nil, err
	}
	return sd, nil
}

// explorerThreeD creates a compact live ThreeD widget preview.
func explorerThreeD() (*threed.ThreeD, error) {
	stage, err := threed.New(
		threed.ShowAxes(false),
		threed.EnableLogging(false),
		threed.BackfaceCulling(false),
		threed.ZoomScale(20.0),
		threed.UprightOnly(false),
		threed.AmbientColor(threed.Color{R: 0.38, G: 0.38, B: 0.38}),
		threed.DiffuseColor(threed.Color{R: 1.0, G: 1.0, B: 1.0}),
	)
	if err != nil {
		return nil, err
	}
	model := threed.Cube(
		threed.ModelSize(1.4),
		threed.ModelRune('█'),
		threed.ModelColor(threed.NeonCyan),
	)
	stage.SetModel(model)
	stage.Rotate(threed.Vector3D{X: 0.55, Y: 0.45, Z: 0.12})
	return stage, nil
}

// explorerTreeView creates a nested TreeView preview.
func explorerTreeView() (*treeview.TreeView, error) {
	return treeview.New(
		treeview.Nodes(
			&treeview.TreeNode{
				Label:         "Catalog",
				ExpandedState: true,
				Children: []*treeview.TreeNode{
					{Label: "Visualization"},
					{Label: "Input Controls"},
					{Label: "Display"},
					{Label: "Layout & FX"},
				},
			},
		),
		treeview.LabelColor(cell.ColorNumber(252)),
		treeview.ExpandedIcon("▼"),
		treeview.CollapsedIcon("▶"),
		treeview.LeafIcon("·"),
		treeview.IndentationPerLevel(2),
	)
}

// explorerModalPreview creates a draggable modal widget preview.
func explorerModalPreview() (*modal.Modal, error) {
	body, err := explorerText("Drag me by the title bar.\nClick - to minimize and restore.", cell.ColorNumber(245))
	if err != nil {
		return nil, err
	}
	logW, err := explorerText("Modal child window\nsame border/minimize behavior", cell.ColorNumber(159))
	if err != nil {
		return nil, err
	}
	main := modal.NewDraggableWidget("explorer-modal-main", body, 4, 2, 34, 8, nil)
	main.Title = "Modal Window"
	aux := modal.NewDraggableWidget("explorer-modal-log", logW, 26, 10, 32, 7, nil)
	aux.Title = "Compact Log"
	return modal.NewModal("explorer-modal-preview", []*modal.DraggableWidget{main, aux}, nil), nil
}

type explorerTabPreview struct {
	mu     sync.Mutex
	active int
	frame  int
	rects  []image.Rectangle
}

func newExplorerTabPreview() *explorerTabPreview {
	return &explorerTabPreview{}
}

func (p *explorerTabPreview) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	_ = meta
	p.mu.Lock()
	defer p.mu.Unlock()

	ar := cvs.Area()
	p.frame++
	p.rects = p.rects[:0]

	x := ar.Min.X
	for i, label := range []string{"Metrics", "Events", "Graph"} {
		tabText := fmt.Sprintf(" %s %s ", explorerTabIcon(i == p.active), label)
		width := runewidth.StringWidth(tabText)
		p.rects = append(p.rects, image.Rect(x-ar.Min.X, 0, x-ar.Min.X+width, 2))

		fg := cell.ColorNumber(245)
		bg := cell.ColorNumber(236)
		if i == p.active {
			fg = cell.ColorNumber(159)
			bg = cell.ColorNumber(24)
		}
		if err := draw.Text(cvs, tabText, image.Point{X: x, Y: ar.Min.Y},
			draw.TextCellOpts(cell.FgColor(fg), cell.BgColor(bg), cell.Bold()),
			draw.TextMaxX(ar.Max.X),
			draw.TextOverrunMode(draw.OverrunModeTrim),
		); err != nil {
			return err
		}

		marker := strings.Repeat("─", width)
		if i == p.active {
			marker = explorerTabSweep(width, p.frame)
		}
		if err := draw.Text(cvs, marker, image.Point{X: x, Y: ar.Min.Y + 1},
			draw.TextCellOpts(cell.FgColor(explorerTabUnderlineColor(i == p.active))),
			draw.TextMaxX(ar.Max.X),
			draw.TextOverrunMode(draw.OverrunModeTrim),
		); err != nil {
			return err
		}
		x += width + 1
	}

	return p.drawActiveBody(cvs, ar)
}

func (p *explorerTabPreview) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	_ = meta
	p.mu.Lock()
	defer p.mu.Unlock()
	switch k.Key {
	case keyboard.KeyArrowLeft:
		p.active = (p.active + 2) % 3
	case keyboard.KeyArrowRight, keyboard.KeyTab:
		p.active = (p.active + 1) % 3
	}
	return nil
}

func (p *explorerTabPreview) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	_ = meta
	if m.Button != mouse.ButtonLeft {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, r := range p.rects {
		if m.Position.In(r) {
			p.active = i
			return nil
		}
	}
	return nil
}

func (p *explorerTabPreview) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize:  image.Point{X: 36, Y: 8},
		WantKeyboard: widgetapi.KeyScopeFocused,
		WantMouse:    widgetapi.MouseScopeWidget,
	}
}

func (p *explorerTabPreview) drawActiveBody(cvs *canvas.Canvas, ar image.Rectangle) error {
	content := [][]string{
		{"METRICS", "load 64%", "p95 42ms", "queue 07"},
		{"EVENTS", "12:08 deploy started", "12:09 cache warm", "12:10 checks green"},
		{"GRAPH", "▁▃▅▇▆▅▇█▆▃", "series: live", "mode: preview"},
	}
	rows := content[p.active]
	for i, row := range rows {
		color := cell.ColorNumber(245)
		if i == 0 {
			color = cell.ColorNumber(159)
		}
		if err := draw.Text(cvs, row, image.Point{X: ar.Min.X + 2, Y: ar.Min.Y + 3 + i},
			draw.TextCellOpts(cell.FgColor(color)),
			draw.TextMaxX(ar.Max.X),
			draw.TextOverrunMode(draw.OverrunModeTrim),
		); err != nil {
			return err
		}
	}
	return nil
}

func explorerTabIcon(active bool) string {
	if active {
		return "◆"
	}
	return "○"
}

func explorerTabUnderlineColor(active bool) cell.Color {
	if active {
		return cell.ColorNumber(75)
	}
	return cell.ColorNumber(238)
}

func explorerTabSweep(width, frame int) string {
	if width <= 0 {
		return ""
	}
	head := frame % width
	runes := []rune(strings.Repeat("─", width))
	runes[head] = '●'
	if head > 0 {
		runes[head-1] = '•'
	}
	return string(runes)
}

// writeExplorerStatus refreshes the small status pane.
func writeExplorerStatus(w *text.Text, message string) error {
	w.Reset()
	rows := []struct {
		k string
		v string
	}{
		{"Status:   ", message + "\n"},
		{"Catalog:  ", "all leaves selectable\n"},
		{"Preview:  ", "right pane swaps live widgets\n"},
		{"Tip:      ", "click tree items to inspect"},
	}
	for _, row := range rows {
		if err := w.Write(row.k, text.WriteCellOpts(cell.FgColor(cell.ColorNumber(245)))); err != nil {
			return err
		}
		if err := w.Write(row.v, text.WriteCellOpts(cell.FgColor(cell.ColorNumber(252)))); err != nil {
			return err
		}
	}
	return nil
}

// seedExplorerTimeline gives the Timeline and TimeRangePicker meaningful initial data.
func seedExplorerTimeline(tl *timeline.Timeline, picker *timeline.TimeRangePicker) {
	seeds := []struct {
		sev   timeline.Severity
		title string
	}{
		{timeline.SeverityInfo, "System Boot"},
		{timeline.SeverityDebug, "Config Loaded"},
		{timeline.SeverityInfo, "Tab System Ready"},
		{timeline.SeverityWarn, "Memory Pressure"},
		{timeline.SeverityInfo, "BorderFX Active"},
		{timeline.SeverityDebug, "Radar Armed"},
		{timeline.SeverityInfo, "Spectrum Calibrated"},
		{timeline.SeverityWarn, "Timeline Range Armed"},
	}
	now := time.Now()
	events := make([]timeline.Event, 0, len(seeds))
	for i, s := range seeds {
		ts := now.Add(-time.Duration(len(seeds)-i) * 8 * time.Second)
		e := timeline.Event{
			Time:      ts.Format("15:04:05"),
			Title:     s.title,
			Timestamp: ts,
			Severity:  s.sev,
		}
		tl.AddEvent(e)
		events = append(events, e)
	}
	picker.SetPickerEvents(events)
}

// ─────────────────────────────────────────────────────────────────────────────
// Animation goroutines
// ─────────────────────────────────────────────────────────────────────────────

// animateVisualize feeds Tab 3 with rolling data.
func animateVisualize(ctx context.Context, w *vizWidgets) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var lineVals []float64
	radarAngle := 0.0
	loadPct := 62.0
	lineStep := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.mu.Lock()
			w.phase += 0.3
			phase := w.phase
			w.mu.Unlock()

			// Radar contacts – rotate slowly around the sweep
			radarAngle = math.Mod(radarAngle+3.0, 360.0)
			contacts := []*radar.Contact{
				{Angle: math.Mod(radarAngle+42, 360), Distance: 0.35 + 0.08*math.Sin(phase*0.7), Label: "A1"},
				{Angle: math.Mod(radarAngle+158, 360), Distance: 0.62 + 0.06*math.Cos(phase*0.5), Label: "B2"},
				{Angle: math.Mod(radarAngle+270, 360), Distance: 0.80, Label: "C3"},
			}
			if err := w.radarW.SetContacts(contacts); err != nil {
				log.Printf("radar: %v", err)
			}

			// Donut – system load
			loadPct = 45 + 30*math.Sin(phase*0.18)
			if loadPct < 10 {
				loadPct = 10
			}
			if err := w.donut2.Percent(int(loadPct), donut.Label(fmt.Sprintf("%.0f%%", loadPct))); err != nil {
				log.Printf("donut2: %v", err)
			}

			// LineChart with threshold at 0.75
			lineStep = (lineStep + 1) % 360
			v := 0.5 + 0.55*math.Sin(float64(lineStep)*math.Pi/60) + 0.2*math.Cos(float64(lineStep)*math.Pi/24)
			lineVals = appendRollingFloat(lineVals, v, 96)
			if err := w.lineW.Series("signal", append([]float64(nil), lineVals...),
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(75))),
			); err != nil {
				log.Printf("linechart: %v", err)
			}

			// Heatmap
			xl, yl, hv := vizHeatmapFrame(phase)
			if err := w.heatW.Values(xl, yl, hv); err != nil {
				log.Printf("heatmap: %v", err)
			}

			// Pie – 5 segments derived from sine phase
			pv := []int{
				clampInt(int(28+12*math.Sin(phase*0.31)), 4, 50),
				clampInt(int(22+10*math.Cos(phase*0.47)), 4, 50),
				clampInt(int(18+8*math.Sin(phase*0.61+1.0)), 4, 50),
				clampInt(int(14+6*math.Cos(phase*0.73+2.0)), 4, 50),
				clampInt(int(12+5*math.Sin(phase*0.89+3.0)), 4, 50),
			}
			if err := w.pieW.Values(pv); err != nil {
				log.Printf("pie: %v", err)
			}

			// Status
			if err := renderVizStatus(w.statusW, len(contacts), loadPct); err != nil {
				log.Printf("viz status: %v", err)
			}
		}
	}
}

// animateExplorer feeds Tab 4 with live events and spinner frames.
func animateExplorer(ctx context.Context, w *explWidgets) {
	eventTicker := time.NewTicker(4 * time.Second)
	activityTicker := time.NewTicker(300 * time.Millisecond)
	spinTicker := time.NewTicker(120 * time.Millisecond)
	defer eventTicker.Stop()
	defer activityTicker.Stop()
	defer spinTicker.Stop()

	sp := spinner.Must("dots")
	spinStep := 0
	var lineVals []float64
	radarAngle := 0.0

	eventNames := []struct {
		sev   timeline.Severity
		title string
	}{
		{timeline.SeverityInfo, "Radar contact updated"},
		{timeline.SeverityDebug, "Heatmap frame rendered"},
		{timeline.SeverityWarn, "Signal threshold crossed"},
		{timeline.SeverityInfo, "Tab navigation event"},
		{timeline.SeverityError, "Transient network blip"},
		{timeline.SeverityInfo, "BorderFX tick complete"},
		{timeline.SeverityCritical, "Alert: load spike detected"},
		{timeline.SeverityDebug, "Donut animation step"},
		{timeline.SeverityInfo, "Pie chart recalculated"},
		{timeline.SeverityWarn, "Memory GC pause: 2ms"},
	}
	eventIdx := 0

	spinnerLabels := []string{
		"Scanning subsystems",
		"Monitoring telemetry",
		"Aggregating metrics",
		"Rendering widgets",
		"Processing events",
	}
	spinLabelIdx := 0
	spinLabelTimer := 0

	for {
		select {
		case <-ctx.Done():
			return

		case <-eventTicker.C:
			e := eventNames[eventIdx%len(eventNames)]
			eventIdx++
			event := timeline.Event{
				Time:      time.Now().Format("15:04:05"),
				Title:     e.title,
				Timestamp: time.Now(),
				Severity:  e.sev,
			}
			w.timeLine.AddEvent(event)
			w.picker.AddPickerEvent(event)
			w.prevTime.AddEvent(event)
			w.prevPick.AddPickerEvent(event)

		case <-activityTicker.C:
			phase := float64(spinStep) * 0.18
			load := clampInt(int(58+32*math.Sin(phase)), 4, 98)
			lineVals = appendRollingFloat(lineVals, 0.55+0.34*math.Sin(phase)+0.18*math.Cos(phase*1.7), 90)
			if err := w.line.Series("latency", append([]float64(nil), lineVals...),
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(75))),
			); err != nil {
				log.Printf("explorer linechart: %v", err)
			}
			if err := w.bar.Values([]int{
				clampInt(int(46+28*math.Sin(phase*0.8)), 2, 100),
				clampInt(int(63+24*math.Cos(phase*1.1)), 2, 100),
				clampInt(int(38+34*math.Sin(phase*1.4+1.5)), 2, 100),
				clampInt(int(74+18*math.Cos(phase*0.7+0.6)), 2, 100),
			}, 100,
				barchart.Labels([]string{"cpu", "mem", "io", "net"}),
				barchart.BarColors([]cell.Color{cell.ColorNumber(75), cell.ColorNumber(118), cell.ColorNumber(220), cell.ColorNumber(203)}),
				barchart.ValueColors([]cell.Color{cell.ColorNumber(231), cell.ColorNumber(231), cell.ColorNumber(231), cell.ColorNumber(231)}),
			); err != nil {
				log.Printf("explorer barchart: %v", err)
			}
			if err := w.sparkA.Add([]int{load}); err != nil {
				log.Printf("explorer sparkline: %v", err)
			}
			if err := w.donut.Percent(load, donut.Label(fmt.Sprintf("LOAD %02d%%", load), cell.FgColor(cell.ColorNumber(245)))); err != nil {
				log.Printf("explorer donut: %v", err)
			}
			if err := w.pie.Values([]int{
				clampInt(int(24+14*math.Sin(phase)), 4, 50),
				clampInt(int(20+12*math.Cos(phase*0.9)), 4, 50),
				clampInt(int(18+9*math.Sin(phase*1.3+1.7)), 4, 50),
				clampInt(int(15+7*math.Cos(phase*1.6+0.8)), 4, 50),
				clampInt(int(12+5*math.Sin(phase*2.1)), 4, 50),
			}); err != nil {
				log.Printf("explorer pie: %v", err)
			}
			xl, yl, hv := vizHeatmapFrame(phase)
			if err := w.heat.Values(xl, yl, hv); err != nil {
				log.Printf("explorer heatmap: %v", err)
			}
			if err := w.gauge.Percent(load); err != nil {
				log.Printf("explorer gauge: %v", err)
			}
			radarAngle = math.Mod(radarAngle+5.0, 360.0)
			if err := w.radar.SetContacts([]*radar.Contact{
				{Angle: math.Mod(radarAngle+35, 360), Distance: 0.34, Label: "A1"},
				{Angle: math.Mod(radarAngle+165, 360), Distance: 0.62, Label: "B2"},
				{Angle: math.Mod(radarAngle+290, 360), Distance: 0.78, Label: "C3"},
			}); err != nil {
				log.Printf("explorer radar: %v", err)
			}
			primary, secondary := explorerSpectrumFrame(phase, 64)
			if err := w.spectrum.SetStereo(primary, secondary); err != nil {
				log.Printf("explorer spectrum: %v", err)
			}

		case <-spinTicker.C:
			spinLabelTimer++
			if spinLabelTimer%40 == 0 {
				spinLabelIdx = (spinLabelIdx + 1) % len(spinnerLabels)
			}
			label := spinnerLabels[spinLabelIdx]
			frame := sp.DecorateLeft(label, spinStep)
			spinStep++

			w.spinW.Reset()
			statusLines := []struct{ k, v string }{
				{"", frame + "\n"},
				{"Uptime:   ", fmt.Sprintf("%s\n", fmtDuration(time.Since(time.Now().Add(-time.Duration(spinStep)*120*time.Millisecond))))},
				{"Events:   ", fmt.Sprintf("%d logged\n", eventIdx)},
				{"Catalog:  ", "11 visualizations\n"},
				{"Tabs:     ", "5 loaded"},
			}
			for _, ln := range statusLines {
				kColor := cell.FgColor(cell.ColorNumber(245))
				vColor := cell.FgColor(cell.ColorNumber(252))
				if ln.k == "" {
					vColor = cell.FgColor(cell.ColorNumber(75))
				}
				if ln.k != "" {
					_ = w.spinW.Write(ln.k, text.WriteCellOpts(kColor))
				}
				_ = w.spinW.Write(ln.v, text.WriteCellOpts(vColor))
			}
		}
	}
}

// fmtDuration formats a duration as HH:MM:SS.
func fmtDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// explorerSpectrumFrame returns a compact stereo signal for the Explorer preview.
func explorerSpectrumFrame(phase float64, n int) ([]int, []int) {
	left := make([]int, n)
	right := make([]int, n)
	for i := 0; i < n; i++ {
		x := float64(i) / float64(n)
		left[i] = clampInt(int(48+38*math.Sin(phase+x*math.Pi*4)+12*math.Sin(phase*1.7+x*math.Pi*11)), 0, 100)
		right[i] = clampInt(int(44+34*math.Cos(phase*0.9+x*math.Pi*5)+14*math.Sin(phase*1.2+x*math.Pi*9)), 0, 100)
	}
	return left, right
}

// ─────────────────────────────────────────────────────────────────────────────
// BorderFX
// ─────────────────────────────────────────────────────────────────────────────

// configureBorderFX wires all pane borders through borderfx.
func configureBorderFX(root *container.Container) *borderfx.Animator {
	fx := borderfx.NewAnimator(root)
	fx.SetTickRate(64 * time.Millisecond)
	fx.SetInactiveStyle(inactiveBorderStyle)
	for _, id := range animatedPaneIDs {
		fx.RegisterMacro(id, borderfx.Presets.Interlace, accentPalette())
	}
	return fx
}

// inactiveBorderStyle dims idle pane borders.
func inactiveBorderStyle(_ string, bc container.BorderCell) container.BorderCellStyle {
	color := cell.ColorNumber(236)
	if bc.Title {
		color = cell.ColorNumber(242)
	}
	return container.BorderCellStyle{
		Rune:     bc.Rune,
		CellOpts: []cell.Option{cell.FgColor(color)},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Tab 5 – ThreeD  (Möbius Riders)
// ─────────────────────────────────────────────────────────────────────────────

// threedWidgets groups Tab 5's 3D scene widgets.
type threedWidgets struct {
	stage *threed.ThreeD
	info  *text.Text
	mu    sync.Mutex
	step  int
}

// newThreeDTab creates the Möbius Riders 3D tab.
func newThreeDTab() (*threedWidgets, *tab.Tab, error) {
	stage, err := threed.New(
		threed.ShowAxes(false),
		threed.EnableLogging(false),
		threed.BackfaceCulling(false), // strip needs both sides visible
		threed.RotationStep(0.08),
		threed.ZoomScale(18.0),
		threed.UprightOnly(false),
		threed.AmbientColor(threed.Color{R: 0.38, G: 0.38, B: 0.38}),
		threed.DiffuseColor(threed.Color{R: 1.00, G: 1.00, B: 1.00}),
		threed.SpecularColor(threed.Color{R: 0.52, G: 0.52, B: 0.52}),
		threed.Shininess(42),
	)
	if err != nil {
		return nil, nil, err
	}

	infoW, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, nil, err
	}

	stage.SetModel(buildMobiusScene(0))
	// Set an initial tilt so the Möbius twist and all three riders are visible.
	stage.Rotate(threed.Vector3D{X: 0.55, Y: 0.40, Z: 0.00})
	if err := renderThreeDInfo(infoW, 0); err != nil {
		return nil, nil, err
	}

	w := &threedWidgets{stage: stage, info: infoW}

	content := container.SplitVertical(
		container.Left(
			paneOpts(idThreeDStage, "Möbius Riders · 3D",
				container.PlaceWidget(stage))...,
		),
		container.Right(
			paneOpts(idThreeDInfo, "Scene Info",
				container.PaddingLeft(1),
				container.PaddingTop(1),
				container.PlaceWidget(infoW))...,
		),
		container.SplitPercent(74),
	)

	return w, &tab.Tab{Name: "ThreeD", Content: content}, nil
}

// renderThreeDInfo writes the scene legend into the info panel.
func renderThreeDInfo(w *text.Text, step int) error {
	w.Reset()
	lbl := cell.FgColor(cell.ColorNumber(245))
	type row struct{ k, v string }
	rows := []row{
		{"Scene    ", "Möbius Riders"},
		{"Strip    ", "blue · light-blue twist"},
		{"Sphere   ", "RGB latitude bands"},
		{"Cube     ", "6 per-face colors"},
		{"Pyramid  ", "4 per-face colors"},
		{"Frame    ", fmt.Sprintf("%04d", step)},
		{"", ""},
		{"↑↓←→    ", "orbit camera"},
		{"scroll   ", "zoom in / out"},
		{"Tab / →  ", "next tab"},
	}
	for i, r := range rows {
		if r.k == "" && r.v == "" {
			if err := w.Write("\n", text.WriteCellOpts(lbl)); err != nil {
				return err
			}
			continue
		}
		if err := w.Write(r.k, text.WriteCellOpts(lbl)); err != nil {
			return err
		}
		nl := "\n"
		if i == len(rows)-1 {
			nl = ""
		}
		if err := w.Write(r.v+nl, text.WriteCellOpts(cell.FgColor(cell.ColorNumber(252)))); err != nil {
			return err
		}
	}
	return nil
}

// animateThreeD advances the Möbius Riders scene each tick.
func animateThreeD(ctx context.Context, w *threedWidgets) {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.mu.Lock()
			w.step++
			step := w.step
			w.mu.Unlock()
			w.stage.SetModel(buildMobiusScene(step))
			w.stage.Rotate(threed.Vector3D{X: 0.006, Y: 0.014, Z: 0.000})
			_ = renderThreeDInfo(w.info, step)
		}
	}
}

// ── Möbius scene geometry ────────────────────────────────────────────────────

// mobiusPoint returns a point on the Möbius strip in 3D space.
// The center circle lies in the XZ plane; Y is up.
// u ∈ [0, 2π] runs around the loop; v ∈ [-halfWidth, halfWidth] crosses it.
func mobiusPoint(u, v, R float64) threed.Vector3D {
	return threed.Vector3D{
		X: (R + v*math.Sin(u/2)) * math.Cos(u),
		Y: v * math.Cos(u/2),
		Z: (R + v*math.Sin(u/2)) * math.Sin(u),
	}
}

// mobiusNormal returns the unit surface normal at v=0 on the strip.
// The direction alternates over the full 2π loop (Möbius property).
func mobiusNormal(u float64) threed.Vector3D {
	nx := -math.Cos(u) * math.Cos(u/2)
	ny := math.Sin(u / 2)
	nz := -math.Sin(u) * math.Cos(u/2)
	mag := math.Sqrt(nx*nx + ny*ny + nz*nz)
	if mag < 1e-9 {
		return threed.Vector3D{Y: 1}
	}
	return threed.Vector3D{X: nx / mag, Y: ny / mag, Z: nz / mag}
}

// buildMobiusStrip assembles the strip mesh with alternating blue stripes.
func buildMobiusStrip(R, halfWidth float64, segments int) *threed.Model {
	blue := threed.Color{R: 0.14, G: 0.40, B: 0.92}
	light := threed.Color{R: 0.52, G: 0.78, B: 1.00}
	model := threed.NewModel()
	for i := 0; i < segments; i++ {
		u0 := 2 * math.Pi * float64(i) / float64(segments)
		u1 := 2 * math.Pi * float64(i+1) / float64(segments)
		p00 := mobiusPoint(u0, -halfWidth, R)
		p01 := mobiusPoint(u0, +halfWidth, R)
		p10 := mobiusPoint(u1, -halfWidth, R)
		p11 := mobiusPoint(u1, +halfWidth, R)
		clr := blue
		if (i/4)%2 == 0 {
			clr = light
		}
		model.AddFace(threed.Face{
			Vertices: []threed.Vector3D{p00, p01, p11, p10},
			Char:     '█',
			Color:    clr,
			HasColor: true,
		})
	}
	return model
}

// buildColoredSphere creates a sphere with rainbow latitude-band colors.
func buildColoredSphere(center threed.Vector3D, radius float64) *threed.Model {
	latSegs, lonSegs := 8, 16
	bands := []threed.Color{
		{R: 1.00, G: 0.18, B: 0.18}, // red
		{R: 1.00, G: 0.62, B: 0.10}, // orange
		{R: 0.94, G: 0.94, B: 0.12}, // yellow
		{R: 0.18, G: 0.88, B: 0.28}, // green
		{R: 0.18, G: 0.38, B: 1.00}, // blue
		{R: 0.80, G: 0.18, B: 0.94}, // purple
		{R: 0.96, G: 0.20, B: 0.60}, // magenta
		{R: 0.18, G: 0.88, B: 0.88}, // cyan
	}
	model := threed.NewModel()
	for lat := 0; lat < latSegs; lat++ {
		phi0 := math.Pi * float64(lat) / float64(latSegs)
		phi1 := math.Pi * float64(lat+1) / float64(latSegs)
		clr := bands[lat%len(bands)]
		for lon := 0; lon < lonSegs; lon++ {
			th0 := 2 * math.Pi * float64(lon) / float64(lonSegs)
			th1 := 2 * math.Pi * float64(lon+1) / float64(lonSegs)
			v00 := threed.Vector3D{X: center.X + radius*math.Sin(phi0)*math.Cos(th0), Y: center.Y + radius*math.Cos(phi0), Z: center.Z + radius*math.Sin(phi0)*math.Sin(th0)}
			v01 := threed.Vector3D{X: center.X + radius*math.Sin(phi0)*math.Cos(th1), Y: center.Y + radius*math.Cos(phi0), Z: center.Z + radius*math.Sin(phi0)*math.Sin(th1)}
			v10 := threed.Vector3D{X: center.X + radius*math.Sin(phi1)*math.Cos(th0), Y: center.Y + radius*math.Cos(phi1), Z: center.Z + radius*math.Sin(phi1)*math.Sin(th0)}
			v11 := threed.Vector3D{X: center.X + radius*math.Sin(phi1)*math.Cos(th1), Y: center.Y + radius*math.Cos(phi1), Z: center.Z + radius*math.Sin(phi1)*math.Sin(th1)}
			model.AddFace(threed.Face{
				Vertices: []threed.Vector3D{v00, v01, v11, v10},
				Char:     '█', Color: clr, HasColor: true,
			})
		}
	}
	return model
}

// buildColoredCube creates a cube with a distinct color on each face.
func buildColoredCube(center threed.Vector3D, size float64) *threed.Model {
	h := size / 2
	cx, cy, cz := center.X, center.Y, center.Z
	vv := [8]threed.Vector3D{
		{X: cx - h, Y: cy - h, Z: cz - h}, // 0
		{X: cx + h, Y: cy - h, Z: cz - h}, // 1
		{X: cx + h, Y: cy + h, Z: cz - h}, // 2
		{X: cx - h, Y: cy + h, Z: cz - h}, // 3
		{X: cx - h, Y: cy - h, Z: cz + h}, // 4
		{X: cx + h, Y: cy - h, Z: cz + h}, // 5
		{X: cx + h, Y: cy + h, Z: cz + h}, // 6
		{X: cx - h, Y: cy + h, Z: cz + h}, // 7
	}
	type faceSpec struct {
		idx [4]int
		clr threed.Color
	}
	specs := []faceSpec{
		{[4]int{3, 2, 1, 0}, threed.Color{R: 1.00, G: 0.18, B: 0.18}}, // front  – red
		{[4]int{4, 5, 6, 7}, threed.Color{R: 0.18, G: 0.88, B: 0.28}}, // back   – green
		{[4]int{7, 3, 0, 4}, threed.Color{R: 0.18, G: 0.40, B: 1.00}}, // left   – blue
		{[4]int{1, 2, 6, 5}, threed.Color{R: 0.86, G: 0.18, B: 0.92}}, // right  – purple
		{[4]int{7, 6, 2, 3}, threed.Color{R: 1.00, G: 0.70, B: 0.10}}, // top    – orange
		{[4]int{0, 1, 5, 4}, threed.Color{R: 0.18, G: 0.90, B: 0.90}}, // bottom – cyan
	}
	model := threed.NewModel()
	for _, s := range specs {
		model.AddFace(threed.Face{
			Vertices: []threed.Vector3D{vv[s.idx[0]], vv[s.idx[1]], vv[s.idx[2]], vv[s.idx[3]]},
			Char:     '█', Color: s.clr, HasColor: true,
		})
	}
	return model
}

// buildColoredPyramid creates a square pyramid with a distinct color per face.
func buildColoredPyramid(center threed.Vector3D, size float64) *threed.Model {
	ht := size * 0.85
	s := size * 0.52
	cx, cy, cz := center.X, center.Y, center.Z
	apex := threed.Vector3D{X: cx, Y: cy + ht/2, Z: cz}
	b00 := threed.Vector3D{X: cx - s, Y: cy - ht/2, Z: cz - s}
	b10 := threed.Vector3D{X: cx + s, Y: cy - ht/2, Z: cz - s}
	b11 := threed.Vector3D{X: cx + s, Y: cy - ht/2, Z: cz + s}
	b01 := threed.Vector3D{X: cx - s, Y: cy - ht/2, Z: cz + s}
	model := threed.NewModel()
	// base
	model.AddFace(threed.Face{Vertices: []threed.Vector3D{b00, b10, b11, b01}, Char: '█', Color: threed.Color{R: 0.28, G: 0.10, B: 0.50}, HasColor: true})
	// four triangular sides
	model.AddFace(threed.Face{Vertices: []threed.Vector3D{b00, apex, b10}, Char: '█', Color: threed.Color{R: 1.00, G: 0.18, B: 0.28}, HasColor: true}) // red
	model.AddFace(threed.Face{Vertices: []threed.Vector3D{b10, apex, b11}, Char: '█', Color: threed.Color{R: 0.18, G: 0.88, B: 0.28}, HasColor: true}) // green
	model.AddFace(threed.Face{Vertices: []threed.Vector3D{b11, apex, b01}, Char: '█', Color: threed.Color{R: 0.18, G: 0.42, B: 1.00}, HasColor: true}) // blue
	model.AddFace(threed.Face{Vertices: []threed.Vector3D{b01, apex, b00}, Char: '█', Color: threed.Color{R: 0.88, G: 0.18, B: 0.90}, HasColor: true}) // purple
	return model
}

// mergeThreed merges any number of threed models into one.
func mergeThreed(models ...*threed.Model) *threed.Model {
	merged := threed.NewModel()
	for _, m := range models {
		if m == nil {
			continue
		}
		for _, face := range m.Faces {
			merged.AddFace(face)
		}
	}
	return merged
}

// buildMobiusScene composes the full animated scene for the given step.
func buildMobiusScene(step int) *threed.Model {
	const (
		R          = 2.0   // Möbius strip major radius
		halfWidth  = 0.55  // half cross-section width
		segments   = 56    // strip mesh resolution
		liftHeight = 0.70  // how far above the strip center each rider floats
		riderSpeed = 0.025 // angular advance per frame (radians)
	)

	strip := buildMobiusStrip(R, halfWidth, segments)

	baseAngle := float64(step) * riderSpeed
	riders := []struct {
		delta float64
		build func(threed.Vector3D) *threed.Model
	}{
		{0, func(c threed.Vector3D) *threed.Model { return buildColoredSphere(c, 0.38) }},
		{2 * math.Pi / 3, func(c threed.Vector3D) *threed.Model { return buildColoredCube(c, 0.52) }},
		{4 * math.Pi / 3, func(c threed.Vector3D) *threed.Model { return buildColoredPyramid(c, 0.60) }},
	}

	for _, r := range riders {
		u := baseAngle + r.delta
		center := mobiusPoint(u, 0, R)
		n := mobiusNormal(u)
		objCenter := threed.Vector3D{
			X: center.X + liftHeight*n.X,
			Y: center.Y + liftHeight*n.Y,
			Z: center.Z + liftHeight*n.Z,
		}
		strip = mergeThreed(strip, r.build(objCenter))
	}
	return strip
}

// accentPalette returns the shared accent palette for all panes.
func accentPalette() borderfx.Palette {
	return borderfx.Colors(
		cell.ColorNumber(75),
		cell.ColorNumber(243),
		cell.ColorNumber(236),
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// Main
// ─────────────────────────────────────────────────────────────────────────────

const (
	termboxTerminal = "termbox"
	tcellTerminal   = "tcell"
)

func main() {
	terminalPtr := flag.String("terminal", "tcell",
		"Terminal backend: 'termbox' or 'tcell' (default: tcell)")
	flag.Parse()

	var t terminalapi.Terminal
	var err error
	switch *terminalPtr {
	case termboxTerminal:
		t, err = termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	case tcellTerminal:
		t, err = tcell.New(tcell.ColorMode(terminalapi.ColorMode256))
	default:
		log.Fatalf("unknown terminal %q – choose termbox or tcell", *terminalPtr)
	}
	if err != nil {
		log.Fatalf("failed to open terminal: %v", err)
	}
	defer t.Close()
	// EnableMouseMotion is on the concrete tcell terminal; type-assert to reach it.
	if tcellT, ok := t.(*tcell.Terminal); ok {
		tcellT.EnableMouseMotion()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── Build tabs ───────────────────────────────────────────────────────────
	_, dashTab, err := newDashboardTab(ctx, t)
	if err != nil {
		log.Fatalf("dashboard tab: %v", err)
	}
	_, ctrlTab, err := newControlsTab()
	if err != nil {
		log.Fatalf("controls tab: %v", err)
	}
	vizW, vizTab, err := newVisualizeTab(ctx)
	if err != nil {
		log.Fatalf("visualize tab: %v", err)
	}
	explW, explTab, err := newExplorerTab(ctx)
	if err != nil {
		log.Fatalf("explorer tab: %v", err)
	}
	threedW, threedTab, err := newThreeDTab()
	if err != nil {
		log.Fatalf("threed tab: %v", err)
	}

	// ── Tab manager & styling ─────────────────────────────────────────────────
	tabManager := tab.NewManager(dashTab, ctrlTab, vizTab, explTab, threedTab)
	tabOpts := tab.NewOptions(
		tab.AnimatedActiveTab(true),
		tab.ActiveTextColor(cell.ColorNumber(159)),
		tab.InactiveTextColor(cell.ColorNumber(245)),
		tab.SweepTextColor(cell.ColorNumber(242)),
		tab.SweepAccentColor(cell.ColorNumber(75)),
	)
	tabHeader, err := tab.NewHeader(tabManager, tabOpts)
	if err != nil {
		log.Fatalf("tab header: %v", err)
	}

	// ── Instructions footer ───────────────────────────────────────────────────
	instructions, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("instructions: %v", err)
	}
	pairs := []struct{ key, desc string }{
		{"←/→", "switch tabs  "},
		{"Tab", "next tab  "},
		{"l/r", "shift sine  "},
		{"q/Esc", "quit"},
	}
	for _, p := range pairs {
		_ = instructions.Write(p.key+" ", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(75)), cell.Bold()))
		_ = instructions.Write(p.desc, text.WriteCellOpts(cell.FgColor(cell.ColorNumber(245))))
	}

	// ── Initial placeholder ───────────────────────────────────────────────────
	placeholder, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("placeholder: %v", err)
	}
	_ = placeholder.Write("Loading…")

	// ── Root container ────────────────────────────────────────────────────────
	root, err := container.New(
		t,
		container.ID(idRoot),
		container.Border(linestyle.Round),
		container.BorderColor(cell.ColorNumber(236)),
		container.FocusedColor(cell.ColorNumber(236)),
		container.BorderTitle(" termdash · widget showcase "),
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
						container.BorderTitle(" Dashboard "),
						container.BorderTitleAlignCenter(),
						container.TitleColor(cell.ColorNumber(242)),
						container.TitleFocusedColor(cell.ColorNumber(242)),
						container.PlaceWidget(placeholder),
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
		log.Fatalf("root container: %v", err)
	}
	explW.setRoot(root)

	// ── Seed first tab ────────────────────────────────────────────────────────
	tabContent := tab.NewContent(tabManager)
	if err := tabContent.Update(root); err != nil {
		log.Fatalf("seed tab content: %v", err)
	}

	// ── Event handler (tab switching) ────────────────────────────────────────
	eventHandler := tab.NewEventHandler(ctx, t, tabManager, tabHeader, tabContent, root, cancel, tabOpts)

	// ── BorderFX ──────────────────────────────────────────────────────────────
	fx := configureBorderFX(root)
	go func() { _ = fx.Run(ctx) }()

	// ── Animation goroutines ──────────────────────────────────────────────────
	go animateVisualize(ctx, vizW)
	go animateExplorer(ctx, explW)
	go animateThreeD(ctx, threedW)

	// ── Keyboard handler ──────────────────────────────────────────────────────
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' || k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
	}

	if err := termdash.Run(
		ctx, t, root,
		termdash.KeyboardSubscriber(func(k *terminalapi.Keyboard) {
			quitter(k)
			eventHandler.HandleKeyboard(k)
		}),
		termdash.MouseSubscriber(eventHandler.HandleMouse),
		termdash.RedrawInterval(redrawInterval),
	); err != nil && err != context.Canceled {
		log.Fatalf("termdash: %v", err)
	}
}
