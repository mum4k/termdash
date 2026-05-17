package main

import (
	"context"
	"image"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
)

func TestClickTrackerReleaseCount(t *testing.T) {
	ct := &clickTracker{}
	now := time.Now()
	pos := image.Point{X: 10, Y: 4}

	ct.press(pos)
	if got := ct.releaseCount(now, pos); got != 1 {
		t.Fatalf("first release count = %d, want 1", got)
	}
	ct.press(pos)
	if got := ct.releaseCount(now.Add(200*time.Millisecond), pos); got != 2 {
		t.Fatalf("second release count = %d, want 2", got)
	}
	ct.press(pos)
	if got := ct.releaseCount(now.Add(350*time.Millisecond), pos); got != 3 {
		t.Fatalf("third release count = %d, want 3", got)
	}
}

func TestNewShipStateStartsWarpOffline(t *testing.T) {
	ship := newShipState()
	warpOnline, _, _, _, _, _, _, _, _, _, _ := ship.snapshot()
	if warpOnline {
		t.Fatal("newShipState() starts with warp online, want offline")
	}
}

func TestNewStyleRegistryUsesShowcaseDefaults(t *testing.T) {
	styles := newStyleRegistry()
	want := map[string]panelStyle{
		idSensors: styleInterlace,
		idLCARS:   styleNeon,
		idWarp:    styleIce,
		idComms:   styleMatrix,
		idHelp:    styleRainbow,
	}
	for _, id := range []string{idSensors, idLCARS, idWarp, idComms, idHelp} {
		if got := styles.get(id); got != want[id] {
			t.Fatalf("styleRegistry[%q] = %v, want %v", id, got, want[id])
		}
	}
}

func TestPanelTitlesUseConfiguredDecryptCharsets(t *testing.T) {
	wantSpinners := map[string]string{
		idSensors: "dots_6",
		idLCARS:   "dots_10",
		idWarp:    "star",
		idComms:   "spin_2",
		idHelp:    "pulse",
	}
	wantCharsets := map[string]string{
		idSensors: titleCharset(" Sensor Array "),
		idLCARS:   titleCharset(" LCARS Telemetry "),
		idWarp:    titleCharset(" Warp Core Flux "),
		idComms:   titleCharset(" Comms "),
		idHelp:    titleCharset(" Controls "),
	}
	for _, id := range []string{idSensors, idLCARS, idWarp, idComms, idHelp} {
		spec, ok := panelTitles[id]
		if !ok {
			t.Fatalf("panelTitles missing %q", id)
		}
		if spec.Charset != wantCharsets[id] {
			t.Fatalf("panelTitles[%q].Charset = %q, want %q", id, spec.Charset, wantCharsets[id])
		}
		if len(spec.LeftSpin.Frames()) != 0 || len(spec.Spinner.Frames()) != 0 {
			t.Fatalf("panelTitles[%q] unexpectedly configured left or legacy spinner ornament", id)
		}
		if got, want := spec.RightSpin.Name(), wantSpinners[id]; got != want {
			t.Fatalf("panelTitles[%q].RightSpin = %q, want %q", id, got, want)
		}
	}
}

func TestBootFramesKeepWarpOffline(t *testing.T) {
	for stage := 0; stage < 4; stage++ {
		if got := bootStatusFrame(stage); !strings.Contains(got, "warp core .......... offline") {
			t.Fatalf("bootStatusFrame(%d) = %q, want offline warp core", stage, got)
		}
	}
	if got := bootWarpFrame(3); !strings.Contains(got, "subspace trace .... offline") {
		t.Fatalf("bootWarpFrame(3) = %q, want offline trace banner", got)
	}
}

func TestClickTrackerResetsOnDistanceAndTimeout(t *testing.T) {
	ct := &clickTracker{}
	now := time.Now()

	ct.press(image.Point{X: 1, Y: 1})
	if got := ct.releaseCount(now, image.Point{X: 1, Y: 1}); got != 1 {
		t.Fatalf("first click count = %d, want 1", got)
	}
	ct.press(image.Point{X: 8, Y: 1})
	if got := ct.releaseCount(now.Add(100*time.Millisecond), image.Point{X: 8, Y: 1}); got != 1 {
		t.Fatalf("distant click count = %d, want 1", got)
	}
	ct.press(image.Point{X: 8, Y: 1})
	if got := ct.releaseCount(now.Add(700*time.Millisecond), image.Point{X: 8, Y: 1}); got != 1 {
		t.Fatalf("timed out click count = %d, want 1", got)
	}
}

func TestClickTrackerIgnoresHoverRelease(t *testing.T) {
	ct := &clickTracker{}
	if got := ct.releaseCount(time.Now(), image.Point{X: 10, Y: 4}); got != 0 {
		t.Fatalf("hover release count = %d, want 0", got)
	}
}

func TestPanelAt(t *testing.T) {
	size := image.Point{X: 100, Y: 40}
	cases := []struct {
		pos  image.Point
		want string
	}{
		{pos: image.Point{X: 10, Y: 5}, want: idSensors},
		{pos: image.Point{X: 80, Y: 5}, want: idLCARS},
		{pos: image.Point{X: 10, Y: 28}, want: idWarp},
		{pos: image.Point{X: 80, Y: 28}, want: idComms},
		{pos: image.Point{X: 50, Y: 37}, want: idHelp},
	}
	for _, tc := range cases {
		if got := panelAt(size, tc.pos); got != tc.want {
			t.Fatalf("panelAt(%v) = %q, want %q", tc.pos, got, tc.want)
		}
	}
}

func TestIsFocusMouse(t *testing.T) {
	if !isFocusMouse(&terminalapi.Mouse{Button: mouse.ButtonLeft}) {
		t.Fatal("left click did not focus")
	}
	for _, button := range []mouse.Button{mouse.ButtonRelease, mouse.ButtonRight, mouse.ButtonMiddle} {
		if isFocusMouse(&terminalapi.Mouse{Button: button}) {
			t.Fatalf("button %v unexpectedly focuses", button)
		}
	}
}

func TestDemoMouseLeftClickSwitchesFocusedWindow(t *testing.T) {
	size := image.Point{X: 100, Y: 40}
	ft := faketerm.MustNew(size)
	c := mustDemoTestContainer(t, ft)

	if got := c.ActiveID(); got != idSensors {
		t.Fatalf("initial active ID = %q, want %q", got, idSensors)
	}

	handleDemoMouse(
		&terminalapi.Mouse{Position: image.Point{X: 80, Y: 5}, Button: mouse.ButtonLeft},
		time.Now(),
		&clickTracker{},
		func() bool { return false },
		func(bool) {},
		func() {},
		func() {},
		c.ActiveID,
		func(pos image.Point) string { return panelAt(ft.Size(), pos) },
		func(pos image.Point) bool { return focusPanelAt(c, ft.Size(), pos) },
		func(image.Point) bool { return false },
		func(bool) {},
		func(image.Point) {},
		func() {},
	)

	if got := c.ActiveID(); got != idLCARS {
		t.Fatalf("active ID after left click = %q, want %q", got, idLCARS)
	}
}

func TestDemoMouseHoverUpdatesSparklineReadout(t *testing.T) {
	hovered := false
	pos := image.Point{X: 12, Y: 8}

	handleDemoMouse(
		&terminalapi.Mouse{Position: pos, Button: mouse.ButtonRelease},
		time.Now(),
		&clickTracker{},
		func() bool { return false },
		func(bool) {},
		func() {},
		func() {},
		func() string { return "" },
		func(image.Point) string { return "" },
		func(image.Point) bool { return false },
		func(image.Point) bool { return false },
		func(bool) {},
		func(got image.Point) {
			if got != pos {
				t.Fatalf("hover position = %v, want %v", got, pos)
			}
			hovered = true
		},
		func() {},
	)

	if !hovered {
		t.Fatal("hover callback was not called")
	}
}

func TestDemoMouseWheelHoverDoesNotInterfereWithScroll(t *testing.T) {
	hovered := false
	focused := false
	activated := false
	selectMode := false

	handleDemoMouse(
		&terminalapi.Mouse{Position: image.Point{X: 12, Y: 8}, Button: mouse.ButtonWheelUp},
		time.Now(),
		&clickTracker{},
		func() bool { return selectMode },
		func(next bool) { selectMode = next },
		func() {},
		func() {},
		func() string { return idWarp },
		func(image.Point) string { return idWarp },
		func(image.Point) bool {
			focused = true
			return true
		},
		func(image.Point) bool {
			activated = true
			return true
		},
		func(bool) {},
		func(image.Point) { hovered = true },
		func() {},
	)

	if !hovered {
		t.Fatal("wheel hover did not run hover callback")
	}
	if focused {
		t.Fatal("wheel hover unexpectedly focused a panel")
	}
	if activated {
		t.Fatal("wheel hover unexpectedly activated a control")
	}
	if selectMode {
		t.Fatal("wheel hover unexpectedly enabled select mode")
	}
}

func TestDemoMouseDoubleClickEntersTemporarySelectMode(t *testing.T) {
	size := image.Point{X: 100, Y: 40}
	ft := faketerm.MustNew(size)
	c := mustDemoTestContainer(t, ft)
	clicks := &clickTracker{}
	selectMode := false
	enableCalls := 0
	disableCalls := 0
	restoreCalls := 0

	handle := func(button mouse.Button, now time.Time) {
		handleDemoMouse(
			&terminalapi.Mouse{Position: image.Point{X: 80, Y: 5}, Button: button},
			now,
			clicks,
			func() bool { return selectMode },
			func(next bool) { selectMode = next },
			func() { enableCalls++ },
			func() { disableCalls++ },
			c.ActiveID,
			func(pos image.Point) string { return panelAt(ft.Size(), pos) },
			func(pos image.Point) bool { return focusPanelAt(c, ft.Size(), pos) },
			func(image.Point) bool { return false },
			func(bool) {},
			func(image.Point) {},
			func() { restoreCalls++ },
		)
	}

	now := time.Now()
	handle(mouse.ButtonLeft, now)
	handle(mouse.ButtonRelease, now.Add(50*time.Millisecond))
	handle(mouse.ButtonLeft, now.Add(150*time.Millisecond))
	handle(mouse.ButtonRelease, now.Add(200*time.Millisecond))

	if !selectMode {
		t.Fatal("double click left mouse select mode disabled")
	}
	if enableCalls != 0 {
		t.Fatalf("enable mouse calls = %d, want 0", enableCalls)
	}
	if disableCalls != 1 {
		t.Fatalf("disable mouse calls = %d, want 1", disableCalls)
	}
	if restoreCalls != 1 {
		t.Fatalf("restore calls = %d, want 1", restoreCalls)
	}
	if got := c.ActiveID(); got != idLCARS {
		t.Fatalf("active ID after double click = %q, want %q", got, idLCARS)
	}
}

func TestDemoMouseFocusedTextPanelEntersSelectMode(t *testing.T) {
	size := image.Point{X: 100, Y: 40}
	ft := faketerm.MustNew(size)
	c := mustDemoTestContainer(t, ft)
	selectMode := false
	disableCalls := 0
	restoreCalls := 0

	handleDemoMouse(
		&terminalapi.Mouse{Position: image.Point{X: 10, Y: 5}, Button: mouse.ButtonLeft},
		time.Now(),
		&clickTracker{},
		func() bool { return selectMode },
		func(next bool) { selectMode = next },
		func() {},
		func() { disableCalls++ },
		c.ActiveID,
		func(pos image.Point) string { return panelAt(ft.Size(), pos) },
		func(pos image.Point) bool { return focusPanelAt(c, ft.Size(), pos) },
		func(image.Point) bool { return false },
		func(bool) {},
		func(image.Point) {},
		func() { restoreCalls++ },
	)

	if !selectMode {
		t.Fatal("focused text click did not enter select mode")
	}
	if disableCalls != 1 {
		t.Fatalf("disable mouse calls = %d, want 1", disableCalls)
	}
	if restoreCalls != 1 {
		t.Fatalf("restore calls = %d, want 1", restoreCalls)
	}
}

func TestStatusControlRectsAndClick(t *testing.T) {
	size := image.Point{X: 140, Y: 40}
	ship := newShipState()
	warpOnline, _, _, _, _, _, _, _, _, _, _ := ship.snapshot()
	onRect, offRect, sliderRect, torpedoRect, cloakRect, alarmRect := statusControlRects(size, warpOnline)
	if onRect.Empty() || offRect.Empty() || sliderRect.Empty() || torpedoRect.Empty() || cloakRect.Empty() || alarmRect.Empty() {
		t.Fatal("status control rects should not be empty")
	}

	onPoint := image.Point{X: onRect.Min.X, Y: onRect.Min.Y}
	offPoint := image.Point{X: offRect.Min.X, Y: offRect.Min.Y}

	if !handleStatusControlClick(size, offPoint, ship) {
		t.Fatal("clicking OFF control should be handled")
	}
	if online, _, _, _, _, _, _, _, _, _, _ := ship.snapshot(); online {
		t.Fatal("warp core should be offline after clicking OFF")
	}

	onRect, _, sliderRect, _, _, _ = statusControlRects(size, false)
	onPoint = image.Point{X: onRect.Min.X, Y: onRect.Min.Y}
	if !handleStatusControlClick(size, onPoint, ship) {
		t.Fatal("clicking ON control should be handled")
	}
	if online, _, _, _, _, _, _, _, _, _, _ := ship.snapshot(); !online {
		t.Fatal("warp core should be online after clicking ON")
	}
}

func TestStatusControlsHandleMouseAndDraw(t *testing.T) {
	size := image.Point{X: 140, Y: 40}
	ship := newShipState()
	controls, err := newStatusControls(ship)
	if err != nil {
		t.Fatalf("newStatusControls => unexpected error: %v", err)
	}

	warpArea := controls.warpArea(size)
	sliderArea := controls.shieldsArea(size)
	torpedoRect := controls.torpedoTrigger(size)
	alarmRect := controls.alarmTrigger(size)
	cloakArea := controls.cloakArea(size)
	if !controls.handleMouse(size, image.Point{X: warpArea.Max.X - 1, Y: warpArea.Min.Y}, true) {
		t.Fatal("clicking warp OFF radio should be handled")
	}
	if online, _, _, _, _, _, _, _, _, _, _ := ship.snapshot(); online {
		t.Fatal("warp core should be offline after clicking OFF radio")
	}
	if controls.handleMouse(size, image.Point{X: sliderArea.Max.X - 1, Y: sliderArea.Min.Y}, true) {
		t.Fatal("shields slider should be disabled while warp is offline")
	}
	ft := faketerm.MustNew(size)
	controls.draw(ft, size, true)
	buffer := ft.BackBuffer()
	sample := func(rect image.Rectangle) image.Point {
		dx := rect.Dx() / 2
		if dx < 0 {
			dx = 0
		}
		dy := rect.Dy() / 2
		if dy < 0 {
			dy = 0
		}
		return image.Point{
			X: rect.Min.X + dx,
			Y: rect.Min.Y + dy,
		}
	}
	for _, point := range []image.Point{
		sample(sliderArea),
		sample(torpedoRect),
		sample(alarmRect),
		sample(cloakArea),
	} {
		if got, want := buffer[point.X][point.Y].Opts.FgColor, cell.ColorNumber(241); got != want {
			t.Fatalf("offline control color at %v = %v, want %v", point, got, want)
		}
	}
	if !controls.handleMouse(size, image.Point{X: warpArea.Min.X, Y: warpArea.Min.Y}, true) {
		t.Fatal("clicking warp ON radio should be handled")
	}
	if online, _, _, _, _, _, _, _, _, _, _ := ship.snapshot(); !online {
		t.Fatal("warp core should be online after clicking ON radio")
	}

	if !controls.handleMouse(size, image.Point{X: sliderArea.Max.X - 1, Y: sliderArea.Min.Y}, true) {
		t.Fatal("clicking shields slider should be handled")
	}
	if _, shieldsPct, _, _, _, _, _, _, _, _, _ := ship.snapshot(); shieldsPct != 100 {
		t.Fatalf("shields percentage = %d, want 100", shieldsPct)
	}

	if !controls.handleMouse(size, image.Point{X: torpedoRect.Min.X + 1, Y: torpedoRect.Min.Y}, true) {
		t.Fatal("clicking torpedo trigger should be handled")
	}
	torpedoArea := controls.torpedoArea(size)
	torpedoOption := image.Point{X: torpedoArea.Min.X + 2, Y: torpedoArea.Min.Y + 7}
	if !controls.handleMouse(size, torpedoOption, true) {
		t.Fatal("clicking torpedo option should be handled")
	}
	_, _, torpedoSet, _, torpedoLoading, _, _, _, _, _, version := ship.snapshot()
	if torpedoSet != 6 {
		t.Fatalf("torpedo set = %d, want 6", torpedoSet)
	}
	if !torpedoLoading {
		t.Fatal("torpedo load should enter loading state")
	}
	ship.finishTorpedoLoad(version)
	_, _, _, torpedoLoaded, torpedoLoading, _, _, _, _, _, _ := ship.snapshot()
	if torpedoLoading || torpedoLoaded != 6 {
		t.Fatalf("torpedo completion = loaded:%d loading:%v, want loaded 6 and loading false", torpedoLoaded, torpedoLoading)
	}

	if !controls.handleMouse(size, image.Point{X: alarmRect.Min.X + 1, Y: alarmRect.Min.Y}, true) {
		t.Fatal("clicking alarm trigger should be handled")
	}
	alarmArea := controls.alarmArea(size)
	alarmOption := image.Point{X: alarmArea.Min.X + 2, Y: alarmArea.Min.Y + 9}
	if !controls.handleMouse(size, alarmOption, true) {
		t.Fatal("clicking alarm option should be handled")
	}
	_, _, _, _, _, _, _, _, alarmThreshold, _, _ := ship.snapshot()
	if alarmThreshold != 550 {
		t.Fatalf("alarm threshold = %d, want 550", alarmThreshold)
	}

	if !controls.handleMouse(size, cloakArea.Min, true) {
		t.Fatal("clicking cloak checkbox should be handled")
	}
	_, _, _, _, _, cloakOnline, cloakEnabling, cloakDisabling, _, _, version := ship.snapshot()
	if cloakOnline || !cloakEnabling || cloakDisabling {
		t.Fatalf("cloak after enable click = online:%v enabling:%v disabling:%v", cloakOnline, cloakEnabling, cloakDisabling)
	}
	ship.finishCloakToggle(version, true)
	if !controls.handleMouse(size, cloakArea.Min, true) {
		t.Fatal("clicking cloak checkbox again should be handled")
	}
	_, _, _, _, _, cloakOnline, cloakEnabling, cloakDisabling, _, _, version = ship.snapshot()
	if !cloakOnline || cloakEnabling || !cloakDisabling {
		t.Fatalf("cloak after disable click = online:%v enabling:%v disabling:%v", cloakOnline, cloakEnabling, cloakDisabling)
	}
	ship.finishCloakToggle(version, false)

	ft = faketerm.MustNew(size)
	controls.draw(ft, size, true)
	buffer = ft.BackBuffer()
	if got := buffer[sliderArea.Min.X][sliderArea.Min.Y].Rune; got != '█' {
		t.Fatalf("slider first cell = %q, want %q", got, '█')
	}
	if got := buffer[sliderArea.Max.X-1][sliderArea.Min.Y].Rune; got != '●' {
		t.Fatalf("slider knob cell = %q, want %q", got, '●')
	}
	lines := strings.Split(ft.String(), "\n")
	if got := string([]rune(lines[warpArea.Min.Y])[warpArea.Min.X : warpArea.Min.X+12]); got != "◉ ON   ○ OFF" {
		t.Fatalf("warp overlay = %q, want %q", got, "◉ ON   ○ OFF")
	}
	if got := string([]rune(lines[torpedoRect.Min.Y])[torpedoRect.Min.X : torpedoRect.Min.X+6]); got != "[06 ▼]" {
		t.Fatalf("torpedo overlay = %q, want %q", got, "[06 ▼]")
	}
	if got := string([]rune(lines[alarmRect.Min.Y])[alarmRect.Min.X : alarmRect.Min.X+7]); got != "[550 ▼]" {
		t.Fatalf("alarm overlay = %q, want %q", got, "[550 ▼]")
	}
	if got := string([]rune(lines[cloakArea.Min.Y])[cloakArea.Min.X : cloakArea.Min.X+3]); got != "[ ]" {
		t.Fatalf("cloak overlay = %q, want %q", got, "[ ]")
	}
}

func TestScheduleFocusModeRestoreReEnablesFocusMode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	selectMode := true
	version := 1
	done := make(chan struct{}, 1)

	snapshot := func() (bool, int) {
		mu.Lock()
		defer mu.Unlock()
		return selectMode, version
	}
	setMode := func(next bool) {
		mu.Lock()
		selectMode = next
		version++
		mu.Unlock()
	}

	scheduleFocusModeRestore(
		ctx,
		time.Millisecond,
		snapshot,
		setMode,
		func() {},
		func(bool) { done <- struct{}{} },
	)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for focus mode restore")
	}

	selected, _ := snapshot()
	if selected {
		t.Fatal("mouse select mode still enabled after restore")
	}
}

func TestSparklineReadoutRequiresRenderedLine(t *testing.T) {
	size := image.Point{X: 100, Y: 40}
	chart := mustDemoLinechart(t)
	if err := chart.Series("subspace", []float64{140, 220, 360, 280, 420, 320}); err != nil {
		t.Fatalf("Series => unexpected error: %v", err)
	}
	hit, miss := demoLinechartPoints(t, chart, size)

	if got := sparklineReadout(chart, size, hit); !strings.Contains(got, "X:") || !strings.Contains(got, "Y:") {
		t.Fatalf("sparklineReadout(hit) = %q, want X/Y readout", got)
	}
	if got := sparklineReadout(chart, size, miss); got != "" {
		t.Fatalf("sparklineReadout(miss) = %q, want empty", got)
	}
}

func TestSparkProbeAddBatchCapsHistory(t *testing.T) {
	probe := &sparkProbe{}
	values := make([]int, 600)
	for i := range values {
		values[i] = i
	}
	probe.addBatch(values)

	probe.mu.Lock()
	defer probe.mu.Unlock()
	if got := len(probe.values); got != 512 {
		t.Fatalf("probe length = %d, want 512", got)
	}
	if got := probe.values[0]; got != 88 {
		t.Fatalf("first retained sample = %d, want 88", got)
	}
	if got := probe.values[len(probe.values)-1]; got != 599 {
		t.Fatalf("last retained sample = %d, want 599", got)
	}
}

func TestFabricatedPingSampleRangeAndMotion(t *testing.T) {
	for _, profile := range []telemetryDataProfile{
		dataProfileSine,
		dataProfileBurst,
		dataProfileSaw,
		dataProfilePulse,
		dataProfileStorm,
	} {
		minV := 1000
		maxV := -1
		prev := fabricatedPingSample(profile, 0)
		changed := false
		for i := 1; i < 240; i++ {
			v := fabricatedPingSample(profile, float64(i)*0.14)
			if v != prev {
				changed = true
			}
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
			prev = v
		}
		if minV < 120 || maxV > 700 {
			t.Fatalf("profile %v range = [%d,%d], want [120,700]", profile, minV, maxV)
		}
		if !changed {
			t.Fatalf("profile %v did not change across frames", profile)
		}
		if maxV-minV < 120 {
			t.Fatalf("profile %v dynamic range = %d, want >= 120", profile, maxV-minV)
		}
	}
}

func TestApplySparklineBatchUpdatesAlertState(t *testing.T) {
	s, err := linechart.New(linechart.BrailleOnly(), linechart.DownsampleLTTB(), linechart.XAxisUnscaled(), linechart.YAxisCustomScale(120, 700))
	if err != nil {
		t.Fatalf("linechart.New => unexpected error: %v", err)
	}
	probe := &sparkProbe{}
	ship := newShipState()
	ship.setAlarmThreshold(500)
	style := sparkStyleForProfile(styleProfileBraille)

	applySparklineBatch(s, probe, ship, []int{340, 520}, style)
	_, _, _, _, _, _, _, _, _, red, _ := ship.snapshot()
	if !red {
		t.Fatal("red alert after over-threshold batch = false, want true")
	}

	applySparklineBatch(s, probe, ship, []int{280, 300}, style)
	_, _, _, _, _, _, _, _, _, red, _ = ship.snapshot()
	if red {
		t.Fatal("red alert after under-threshold batch = true, want false")
	}
}

func TestTelemetryProfilesCycle(t *testing.T) {
	profiles := newTelemetryProfiles()
	data, style, lastKey := profiles.snapshot()
	if data != dataProfileSine || style != styleProfileBraille {
		t.Fatalf("initial profiles = (%v,%v), want (%v,%v)", data, style, dataProfileSine, styleProfileBraille)
	}
	if lastKey != "" {
		t.Fatalf("initial last key = %q, want empty", lastKey)
	}

	now := time.Now()
	for i := 0; i < 5; i++ {
		profiles.cycleData('d', now.Add(time.Duration(i)*200*time.Millisecond))
	}
	data, _, lastKey = profiles.snapshot()
	if data != dataProfileSine {
		t.Fatalf("data profile after wrap = %v, want %v", data, dataProfileSine)
	}
	if lastKey != "d" {
		t.Fatalf("last key after data cycle = %q, want %q", lastKey, "d")
	}

	for i := 0; i < 5; i++ {
		profiles.cycleStyle('s', now.Add(time.Duration(i)*200*time.Millisecond))
	}
	_, style, lastKey = profiles.snapshot()
	if style != styleProfileBraille {
		t.Fatalf("style profile after wrap = %v, want %v", style, styleProfileBraille)
	}
	if lastKey != "s" {
		t.Fatalf("last key after style cycle = %q, want %q", lastKey, "s")
	}
}

func TestProfileKeyMatchers(t *testing.T) {
	for _, key := range []keyboard.Key{'1', '3', '5', '7', '9', keyboard.KeyEnd, keyboard.KeyPgDn, keyboard.KeyHome, keyboard.KeyPgUp} {
		if !isDataProfileKey(key) {
			t.Fatalf("isDataProfileKey(%v) = false, want true", key)
		}
	}
	for _, key := range []keyboard.Key{'2', '4', '6', '8', '0', keyboard.KeyArrowDown, keyboard.KeyArrowLeft, keyboard.KeyArrowRight, keyboard.KeyArrowUp, keyboard.KeyInsert} {
		if !isStyleProfileKey(key) {
			t.Fatalf("isStyleProfileKey(%v) = false, want true", key)
		}
	}
	if isDataProfileKey('2') || isStyleProfileKey('1') {
		t.Fatal("profile key matchers overlap unexpectedly")
	}
}

func TestHandleGraphProfileKey(t *testing.T) {
	profiles := newTelemetryProfiles()
	if !handleGraphProfileKey('1', profiles) {
		t.Fatal("handleGraphProfileKey(1) = false, want true")
	}
	data, style, lastKey := profiles.snapshot()
	if data != dataProfileBurst || style != styleProfileBraille || lastKey != "1" {
		t.Fatalf("after 1 = (%v,%v,%q), want (%v,%v,%q)", data, style, lastKey, dataProfileBurst, styleProfileBraille, "1")
	}
	if !handleGraphProfileKey('2', profiles) {
		t.Fatal("handleGraphProfileKey(2) = false, want true")
	}
	data, style, lastKey = profiles.snapshot()
	if data != dataProfileBurst || style != styleProfileNeedle || lastKey != "2" {
		t.Fatalf("after 2 = (%v,%v,%q), want (%v,%v,%q)", data, style, lastKey, dataProfileBurst, styleProfileNeedle, "2")
	}

	profiles = newTelemetryProfiles()
	if !handleGraphProfileKey('d', profiles) {
		t.Fatal("handleGraphProfileKey(d) = false, want true")
	}
	data, style, lastKey = profiles.snapshot()
	if data != dataProfileBurst || style != styleProfileBraille || lastKey != "d" {
		t.Fatalf("after d = (%v,%v,%q), want (%v,%v,%q)", data, style, lastKey, dataProfileBurst, styleProfileBraille, "d")
	}
	if !handleGraphProfileKey('s', profiles) {
		t.Fatal("handleGraphProfileKey(s) = false, want true")
	}
	data, style, lastKey = profiles.snapshot()
	if data != dataProfileBurst || style != styleProfileNeedle || lastKey != "s" {
		t.Fatalf("after s = (%v,%v,%q), want (%v,%v,%q)", data, style, lastKey, dataProfileBurst, styleProfileNeedle, "s")
	}
	if handleGraphProfileKey('q', profiles) {
		t.Fatal("handleGraphProfileKey(q) = true, want false")
	}
}

func TestTooltipOrigin(t *testing.T) {
	size := image.Point{X: 20, Y: 10}

	got, ok := tooltipOrigin(size, image.Point{X: 10, Y: 3}, " X: 064 \n Y: 042 ")
	if !ok {
		t.Fatal("tooltipOrigin returned !ok for centered tooltip")
	}
	if got.Y != 4 {
		t.Fatalf("tooltip Y = %d, want 4", got.Y)
	}

	got, ok = tooltipOrigin(size, image.Point{X: 1, Y: 9}, " X: 064 \n Y: 042 ")
	if !ok {
		t.Fatal("tooltipOrigin returned !ok for bottom edge tooltip")
	}
	if got.Y != 7 {
		t.Fatalf("bottom edge tooltip Y = %d, want 7", got.Y)
	}
	if got.X != 0 {
		t.Fatalf("left-clamped tooltip X = %d, want 0", got.X)
	}
}

func TestRefreshHoverTooltipUpdatesReadoutAtCurrentAnchor(t *testing.T) {
	size := image.Point{X: 100, Y: 40}
	chart := mustDemoLinechart(t)
	tooltip := &hoverTooltip{}
	if err := chart.Series("subspace", []float64{140, 220, 360, 280, 420, 320}); err != nil {
		t.Fatalf("Series => unexpected error: %v", err)
	}
	anchor, _ := demoLinechartPoints(t, chart, size)

	tooltip.update(anchor, sparklineReadout(chart, size, anchor))
	if got := tooltip.text; !strings.Contains(got, "X:") || !strings.Contains(got, "Y:") {
		t.Fatalf("initial tooltip text = %q, want X/Y readout", got)
	}

	refreshHoverTooltip(tooltip, chart, size)

	if got := tooltip.text; !strings.Contains(got, "X:") || !strings.Contains(got, "Y:") {
		t.Fatalf("refreshed tooltip text = %q, want X/Y readout", got)
	}
}

func mustDemoLinechart(t *testing.T) *linechart.LineChart {
	t.Helper()

	chart, err := linechart.New(
		linechart.BrailleOnly(),
		linechart.DownsampleLTTB(),
		linechart.ThresholdLine(500, cell.FgColor(cell.ColorRed)),
		linechart.XAxisUnscaled(),
		linechart.YAxisCustomScale(120, 700),
		linechart.YAxisAdaptive(),
	)
	if err != nil {
		t.Fatalf("linechart.New => unexpected error: %v", err)
	}
	return chart
}

func demoLinechartPoints(t *testing.T, chart *linechart.LineChart, size image.Point) (image.Point, image.Point) {
	t.Helper()

	graph := sparklineGraphRect(size)
	var (
		hit    image.Point
		miss   image.Point
		hitOK  bool
		missOK bool
	)
	for y := graph.Min.Y; y < graph.Max.Y; y++ {
		for x := graph.Min.X; x < graph.Max.X; x++ {
			pos := image.Point{X: x, Y: y}
			_, ok := chart.ValueAt(
				image.Point{X: graph.Dx(), Y: graph.Dy()},
				image.Point{X: x - graph.Min.X, Y: y - graph.Min.Y},
			)
			if ok && !hitOK {
				hit = pos
				hitOK = true
			}
			if !ok && !missOK {
				miss = pos
				missOK = true
			}
			if hitOK && missOK {
				return hit, miss
			}
		}
	}
	t.Fatal("demoLinechartPoints => missing rendered or empty graph cell")
	return image.Point{}, image.Point{}
}

func mustDemoTestContainer(t *testing.T, ft *faketerm.Terminal) *container.Container {
	t.Helper()

	newText := func() *text.Text {
		w, err := text.New()
		if err != nil {
			t.Fatalf("text.New => unexpected error: %v", err)
		}
		return w
	}

	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(55,
			grid.ColWidthPerc(55,
				grid.Widget(newText(),
					container.Border(linestyle.Round),
					container.Focused(),
					container.ID(idSensors),
				),
			),
			grid.ColWidthPerc(45,
				grid.Widget(newText(),
					container.Border(linestyle.Round),
					container.ID(idLCARS),
				),
			),
		),
		grid.RowHeightPerc(30,
			grid.ColWidthPerc(55,
				grid.Widget(newText(),
					container.Border(linestyle.Round),
					container.ID(idWarp),
				),
			),
			grid.ColWidthPerc(45,
				grid.Widget(newText(),
					container.Border(linestyle.Round),
					container.ID(idComms),
				),
			),
		),
		grid.RowHeightPerc(15,
			grid.Widget(newText(),
				container.Border(linestyle.Round),
				container.ID(idHelp),
			),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		t.Fatalf("grid.Build => unexpected error: %v", err)
	}
	c, err := container.New(ft, gridOpts...)
	if err != nil {
		t.Fatalf("container.New => unexpected error: %v", err)
	}
	return c
}
