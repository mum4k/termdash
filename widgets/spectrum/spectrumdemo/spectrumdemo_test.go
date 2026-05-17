package main

import (
	"context"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/spectrum"
)

// TestSpectrumDemoLayoutBuilds ensures the demo grid composes without errors.
func TestSpectrumDemoLayoutBuilds(t *testing.T) {
	_ = newTestDemoContainer(t)
}

// newTestDemoContainer builds the same container tree used by the demo tests.
func newTestDemoContainer(t *testing.T) *container.Container {
	t.Helper()

	stereo, err := spectrum.New()
	if err != nil {
		t.Fatalf("spectrum.New(stereo) => unexpected error: %v", err)
	}
	network, err := spectrum.New(spectrum.HalfDuplex())
	if err != nil {
		t.Fatalf("spectrum.New(network) => unexpected error: %v", err)
	}
	horizontal, err := spectrum.New(spectrum.Horizontal())
	if err != nil {
		t.Fatalf("spectrum.New(horizontal) => unexpected error: %v", err)
	}

	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(50,
			grid.ColWidthPerc(50,
				grid.Widget(stereo, paneOptions(idHarmonic, "Stereo", false)...),
			),
			grid.ColWidthPerc(50,
				grid.Widget(horizontal, paneOptions(idVectors, "Horizontal", false)...),
			),
		),
		grid.RowHeightPerc(50,
			grid.Widget(network, paneOptions(idSubspace, "Half Duplex", true)...),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		t.Fatalf("grid.Build => unexpected error: %v", err)
	}
	ft := faketerm.MustNew(image.Point{X: 80, Y: 20})
	c, err := container.New(ft, append(gridOpts,
		container.KeyFocusNext(keyboard.KeyTab),
		container.KeyFocusPrevious(keyboard.KeyBacktab),
	)...)
	if err != nil {
		t.Fatalf("container.New => unexpected error: %v", err)
	}
	return c
}

// TestSpectrumDemoInstructionTitles verifies titles expose the key shortcuts.
func TestSpectrumDemoInstructionTitles(t *testing.T) {
	profiles := newDemoProfiles()
	if got, want := dataInstructionTitle(profiles), "DATA SHAPE | Press 1, 3, 5, 7, 9 to change data shape"; got != want {
		t.Fatalf("dataInstructionTitle = %q, want %q", got, want)
	}
	if got, want := styleInstructionTitle(profiles), "STYLE PROFILE | Press 2, 4, 6, 8, 0 to change style"; got != want {
		t.Fatalf("styleInstructionTitle = %q, want %q", got, want)
	}
	if got := focusInstructionTitle(profiles); !strings.Contains(got, "ACTIVE SIGNAL") || !strings.Contains(got, "LATTICE / ANALYZER") {
		t.Fatalf("focusInstructionTitle = %q, want focus and active profiles", got)
	}

	profiles.setData(dataProfileStorm)
	profiles.setStyle(styleProfileMatrix)
	if got, want := dataInstructionTitle(profiles), "DATA SHAPE | Press 1, 3, 5, 7, 9 to change data shape"; got != want {
		t.Fatalf("dataInstructionTitle after update = %q, want %q", got, want)
	}
	if got, want := styleInstructionTitle(profiles), "STYLE PROFILE | Press 2, 4, 6, 8, 0 to change style"; got != want {
		t.Fatalf("styleInstructionTitle after update = %q, want %q", got, want)
	}
}

// TestSpectrumDemoTitlesApply verifies live title updates target every pane.
func TestSpectrumDemoTitlesApply(t *testing.T) {
	c := newTestDemoContainer(t)
	profiles := newDemoProfiles()
	profiles.setData(dataProfileCarrier)
	profiles.setStyle(styleProfileWire)

	if err := applySpectrumTitles(c, profiles); err != nil {
		t.Fatalf("applySpectrumTitles => unexpected error: %v", err)
	}

	ft := faketerm.MustNew(image.Point{X: 20, Y: 10})
	missing, err := container.New(ft)
	if err != nil {
		t.Fatalf("container.New(missing IDs) => unexpected error: %v", err)
	}
	if err := applySpectrumTitles(missing, profiles); err == nil {
		t.Fatal("applySpectrumTitles without pane IDs => nil error, want error")
	}
	if err := applySpectrumTitles(newPartialDemoContainer(t, idHarmonic), profiles); err == nil {
		t.Fatal("applySpectrumTitles without vector pane => nil error, want error")
	}
	if err := applySpectrumTitles(newPartialDemoContainer(t, idHarmonic, idVectors), profiles); err == nil {
		t.Fatal("applySpectrumTitles without subspace pane => nil error, want error")
	}
}

// TestConfigureFocusSweep verifies the focus sweep animator can be constructed.
func TestConfigureFocusSweep(t *testing.T) {
	c := newTestDemoContainer(t)
	fx := configureFocusSweep(c)
	if fx == nil {
		t.Fatal("configureFocusSweep returned nil")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := fx.Run(ctx); err == nil {
		t.Fatal("Run on canceled focus sweep returned nil")
	}
}

// TestInactiveBorderStyle verifies unfocused windows are subdued but readable.
func TestInactiveBorderStyle(t *testing.T) {
	borderCell := container.BorderCell{Rune: '─'}
	style := inactiveBorderStyle(idHarmonic, borderCell)
	if got := cell.NewOptions(style.CellOpts...).FgColor; got != cell.ColorNumber(238) {
		t.Fatalf("inactive border color = %v, want %v", got, cell.ColorNumber(238))
	}
	if style.Rune != borderCell.Rune {
		t.Fatalf("inactive border rune = %q, want %q", style.Rune, borderCell.Rune)
	}

	roundedCorner := container.BorderCell{Rune: '╭'}
	style = inactiveBorderStyle(idHarmonic, roundedCorner)
	if got, want := style.Rune, '○'; got != want {
		t.Fatalf("inactive rounded corner rune = %q, want %q", got, want)
	}

	titleCell := container.BorderCell{Rune: 'T', Title: true}
	style = inactiveBorderStyle(idVectors, titleCell)
	if got := cell.NewOptions(style.CellOpts...).FgColor; got != cell.ColorNumber(244) {
		t.Fatalf("inactive title color = %v, want %v", got, cell.ColorNumber(244))
	}
}

// newPartialDemoContainer builds deliberately incomplete pane trees for errors.
func newPartialDemoContainer(t *testing.T, ids ...string) *container.Container {
	t.Helper()

	switch len(ids) {
	case 1:
		c, err := container.New(
			faketerm.MustNew(image.Point{X: 40, Y: 10}),
			container.ID(ids[0]),
			container.PlaceWidget(newTestSpectrum(t)),
		)
		if err != nil {
			t.Fatalf("container.New(partial one) => unexpected error: %v", err)
		}
		return c
	case 2:
		c, err := container.New(
			faketerm.MustNew(image.Point{X: 40, Y: 10}),
			container.SplitVertical(
				container.Left(
					container.ID(ids[0]),
					container.PlaceWidget(newTestSpectrum(t)),
				),
				container.Right(
					container.ID(ids[1]),
					container.PlaceWidget(newTestSpectrum(t)),
				),
			),
		)
		if err != nil {
			t.Fatalf("container.New(partial two) => unexpected error: %v", err)
		}
		return c
	default:
		t.Fatalf("newPartialDemoContainer received unsupported ID count %d", len(ids))
	}
	return nil
}

// newTestSpectrum returns a Spectrum widget for container wiring tests.
func newTestSpectrum(t *testing.T) *spectrum.Spectrum {
	t.Helper()

	s, err := spectrum.New()
	if err != nil {
		t.Fatalf("spectrum.New => unexpected error: %v", err)
	}
	return s
}

// TestSpectrumDemoNumberKeysSelectProfiles covers laptop number row shortcuts.
func TestSpectrumDemoNumberKeysSelectProfiles(t *testing.T) {
	profiles := newDemoProfiles()

	for _, tc := range []struct {
		key  keyboard.Key
		want dataProfile
	}{
		{key: '1', want: dataProfileLattice},
		{key: '3', want: dataProfileComb},
		{key: '5', want: dataProfilePulse},
		{key: '7', want: dataProfileCarrier},
		{key: '9', want: dataProfileStorm},
		{key: keyboard.KeyEnd, want: dataProfileLattice},
		{key: keyboard.KeyPgDn, want: dataProfileComb},
		{key: keyboard.KeyHome, want: dataProfileCarrier},
		{key: keyboard.KeyPgUp, want: dataProfileStorm},
	} {
		if !handleSpectrumProfileKey(tc.key, profiles) {
			t.Fatalf("handleSpectrumProfileKey(%v) = false, want true", tc.key)
		}
		got, _ := profiles.snapshot()
		if got != tc.want {
			t.Fatalf("data profile after %v = %v, want %v", tc.key, got, tc.want)
		}
	}

	for _, tc := range []struct {
		key  keyboard.Key
		want styleProfile
	}{
		{key: '2', want: styleProfileNeedle},
		{key: '4', want: styleProfileLCARS},
		{key: '6', want: styleProfileMatrix},
		{key: '8', want: styleProfileWire},
		{key: '0', want: styleProfileAnalyzer},
		{key: keyboard.KeyArrowDown, want: styleProfileNeedle},
		{key: keyboard.KeyArrowLeft, want: styleProfileLCARS},
		{key: keyboard.KeyArrowRight, want: styleProfileMatrix},
		{key: keyboard.KeyArrowUp, want: styleProfileWire},
		{key: keyboard.KeyInsert, want: styleProfileAnalyzer},
	} {
		if !handleSpectrumProfileKey(tc.key, profiles) {
			t.Fatalf("handleSpectrumProfileKey(%v) = false, want true", tc.key)
		}
		_, got := profiles.snapshot()
		if got != tc.want {
			t.Fatalf("style profile after %v = %v, want %v", tc.key, got, tc.want)
		}
	}

	if handleSpectrumProfileKey('x', profiles) {
		t.Fatal("handleSpectrumProfileKey(x) = true, want false")
	}
}

// TestSpectrumDemoProfileApplication verifies live style updates are valid.
func TestSpectrumDemoProfileApplication(t *testing.T) {
	stereo, err := spectrum.New()
	if err != nil {
		t.Fatalf("spectrum.New(stereo) => unexpected error: %v", err)
	}
	network, err := spectrum.New(spectrum.HalfDuplex())
	if err != nil {
		t.Fatalf("spectrum.New(network) => unexpected error: %v", err)
	}
	horizontal, err := spectrum.New(spectrum.Horizontal())
	if err != nil {
		t.Fatalf("spectrum.New(horizontal) => unexpected error: %v", err)
	}
	profiles := newDemoProfiles()
	profiles.setData(dataProfileStorm)
	profiles.setStyle(styleProfileMatrix)

	if err := applySpectrumProfiles(stereo, network, horizontal, profiles); err != nil {
		t.Fatalf("applySpectrumProfiles => unexpected error: %v", err)
	}

	for _, style := range []styleProfile{
		styleProfileAnalyzer,
		styleProfileNeedle,
		styleProfileLCARS,
		styleProfileMatrix,
		styleProfileWire,
		styleProfile(99),
	} {
		profiles.setStyle(style)
		if err := applySpectrumProfiles(stereo, network, horizontal, profiles); err != nil {
			t.Fatalf("applySpectrumProfiles(style %v) => unexpected error: %v", style, err)
		}
	}
}

// TestAlertControlDrivesNetworkGraph verifies dropdown choices update state.
func TestAlertControlDrivesNetworkGraph(t *testing.T) {
	network, err := spectrum.New(spectrum.HalfDuplex(), spectrum.MaxValue(spectrumMax), spectrum.ChannelLabels("", ""))
	if err != nil {
		t.Fatalf("spectrum.New(network) => unexpected error: %v", err)
	}
	control, err := spectrum.NewAlertControl(thresholdMin, spectrumMax, thresholdStep, defaultThreshold, func(value int) error {
		return applyNetworkThreshold(network, value)
	})
	if err != nil {
		t.Fatalf("NewAlertControl => unexpected error: %v", err)
	}
	profiles := newDemoProfiles()
	graph := networkWidgetArea(image.Point{X: 100, Y: 32})
	primaryLabel := networkPrimaryLabel(profiles)
	menu := alertControlMenuRect(graph, primaryLabel)

	if !control.HandleMouse(menu.Min.Add(image.Point{X: 1, Y: 0}), graph, primaryLabel) {
		t.Fatal("clicking threshold trigger was not handled")
	}
	optionPos := menu.Min.Add(image.Point{X: 2, Y: 5})
	if !control.HandleMouse(optionPos, graph, primaryLabel) {
		t.Fatal("clicking threshold option was not handled")
	}
	if got, want := control.Threshold(), 350; got != want {
		t.Fatalf("control threshold = %d, want %d", got, want)
	}

	if err := network.Update([]int{spectrumMax, spectrumMax, spectrumMax}, nil); err != nil {
		t.Fatalf("network.Update => unexpected error: %v", err)
	}
	ft := faketerm.MustNew(image.Point{X: 8, Y: 8})
	c, err := container.New(ft, container.PlaceWidget(network))
	if err != nil {
		t.Fatalf("container.New => unexpected error: %v", err)
	}
	if err := c.Draw(); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	if !strings.Contains(ft.String(), "─") {
		t.Fatalf("rendered network missing threshold line: %q", ft.String())
	}
}

// TestAlertOverlayDraws verifies the overlay writes useful labels and tooltips.
func TestAlertOverlayDraws(t *testing.T) {
	control, err := spectrum.NewAlertControl(thresholdMin, spectrumMax, thresholdStep, defaultThreshold, nil)
	if err != nil {
		t.Fatalf("NewAlertControl => unexpected error: %v", err)
	}
	ft := faketerm.MustNew(image.Point{X: 100, Y: 32})
	if err := control.Draw(ft, networkWidgetArea(ft.Size()), networkPrimaryLabel(newDemoProfiles())); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	rendered := ft.String()
	if !strings.Contains(rendered, "ALARM") || !strings.Contains(rendered, "[ ]") || !strings.Contains(rendered, "500") {
		t.Fatalf("threshold overlay = %q, want label, checkbox, and default value", rendered)
	}

	term := newSpectrumTerminal(ft, control, newDemoProfiles())
	term.setActiveIDProvider(func() string { return idSubspace })
	term.tooltip.update(image.Point{X: 5, Y: 5}, " X: 001 \n Y: 500 ")
	if err := term.Flush(); err != nil {
		t.Fatalf("spectrumTerminal.Flush => unexpected error: %v", err)
	}
	if rendered := ft.String(); !strings.Contains(rendered, "X: 001") {
		t.Fatalf("terminal overlay missing tooltip text: %q", rendered)
	}
}

// TestAlertOverlayCompactLayout covers compact layouts and out-of-bounds clicks.
func TestAlertOverlayCompactLayout(t *testing.T) {
	control, err := spectrum.NewAlertControl(thresholdMin, spectrumMax, thresholdStep, defaultThreshold, nil)
	if err != nil {
		t.Fatalf("NewAlertControl => unexpected error: %v", err)
	}
	if err := control.Draw(faketerm.MustNew(image.Point{X: 3, Y: 3}), image.Rectangle{}, "PING"); err != nil {
		t.Fatalf("Draw on tiny area => unexpected error: %v", err)
	}
	if control.HandleMouse(image.Point{}, networkWidgetArea(image.Point{X: 100, Y: 32}), networkPrimaryLabel(newDemoProfiles())) {
		t.Fatal("HandleMouse outside dropdown returned true, want false")
	}
}

// TestAlertControlAlarmToggleAndBanner verifies the checkbox and warning banner.
func TestAlertControlAlarmToggleAndBanner(t *testing.T) {
	control, err := spectrum.NewAlertControl(thresholdMin, spectrumMax, thresholdStep, defaultThreshold, nil)
	if err != nil {
		t.Fatalf("NewAlertControl => unexpected error: %v", err)
	}
	profiles := newDemoProfiles()
	graph := networkWidgetArea(image.Point{X: 100, Y: 32})
	primaryLabel := networkPrimaryLabel(profiles)
	if !control.HandleMouse(alertControlCheckboxPos(graph, primaryLabel), graph, primaryLabel) {
		t.Fatal("clicking alarm checkbox was not handled")
	}
	if !control.Enabled() {
		t.Fatal("alarm checkbox = false, want true")
	}

	control.UpdateSamples([]int{100, 200, 550})
	message := control.AlertMessage()
	if !strings.Contains(message, "Warning: data exceeds 500 threshold") {
		t.Fatalf("alertMessage = %q, want threshold warning", message)
	}

	ft := faketerm.MustNew(image.Point{X: 100, Y: 32})
	control.DrawAlert(ft, spectrumPaneRects(ft.Size())[idSubspace], true)
	if rendered := ft.String(); !strings.Contains(rendered, "Warning: data exceeds 500 threshold") {
		t.Fatalf("drawAlert output = %q, want warning banner", rendered)
	}
	ft = faketerm.MustNew(image.Point{X: 100, Y: 32})
	control.DrawAlert(ft, spectrumPaneRects(ft.Size())[idSubspace], false)
	if rendered := ft.String(); strings.Contains(rendered, "Warning: data exceeds 500 threshold") {
		t.Fatalf("drawAlert for unfocused pane = %q, want no warning banner", rendered)
	}

	control.UpdateSamples([]int{100, 200, 300})
	if got := control.AlertMessage(); got != "" {
		t.Fatalf("alertMessage below threshold = %q, want empty", got)
	}
}

// TestNetworkHoverReadout verifies bottom graph samples are formatted for clicks.
func TestNetworkHoverReadout(t *testing.T) {
	network, err := spectrum.New(spectrum.HalfDuplex(), spectrum.ChannelLabels("PING", "LATENCY"))
	if err != nil {
		t.Fatalf("spectrum.New(network) => unexpected error: %v", err)
	}
	values := make([]int, networkBins)
	for i := range values {
		values[i] = i * 10
	}
	if err := network.Update(values, nil); err != nil {
		t.Fatalf("network.Update => unexpected error: %v", err)
	}
	size := image.Point{X: 100, Y: 32}
	hit, miss := firstNetworkHoverPoints(t, network, size)
	text := networkHoverReadout(network, size, hit)
	if !strings.Contains(text, "X:") || !strings.Contains(text, "Y:") {
		t.Fatalf("networkHoverReadout = %q, want X and Y lines", text)
	}
	if got := networkHoverReadout(network, size, miss); got != "" {
		t.Fatalf("networkHoverReadout miss = %q, want empty", got)
	}
	if got := networkHoverReadout(network, size, image.Point{}); got != "" {
		t.Fatalf("networkHoverReadout outside = %q, want empty", got)
	}
}

// TestSpectrumMouseFocusAndTooltip verifies click focus and click-to-reveal plumbing.
func TestSpectrumMouseFocusAndTooltip(t *testing.T) {
	c := newTestDemoContainer(t)
	size := image.Point{X: 80, Y: 24}
	if !focusSpectrumPaneAt(c, size, spectrumPaneRects(size)[idVectors].Min.Add(image.Point{X: 1, Y: 1})) {
		t.Fatal("focusSpectrumPaneAt returned false for vectors pane")
	}
	if got, want := c.ActiveID(), idVectors; got != want {
		t.Fatalf("active pane = %q, want %q", got, want)
	}

	network, err := spectrum.New(spectrum.HalfDuplex())
	if err != nil {
		t.Fatalf("spectrum.New(network) => unexpected error: %v", err)
	}
	if err := network.Update([]int{100, 200, 300, 400, 500, 600}, nil); err != nil {
		t.Fatalf("network.Update => unexpected error: %v", err)
	}
	tooltip := &spectrumTooltip{}
	hit, miss := firstNetworkHoverPoints(t, network, size)
	showNetworkTooltip(tooltip, network, size, hit)
	if _, visible := tooltip.snapshot(); !visible {
		t.Fatal("tooltip not visible after clicking graph")
	}
	showNetworkTooltip(tooltip, network, size, miss)
	if _, visible := tooltip.snapshot(); visible {
		t.Fatal("tooltip visible after clicking empty graph space")
	}
	showNetworkTooltip(tooltip, network, size, image.Point{})
	if _, visible := tooltip.snapshot(); visible {
		t.Fatal("tooltip visible after leaving graph")
	}
}

// TestSpectrumSingleClickFocusThenReveal verifies the graph only reveals values
// after the subspace pane is already focused.
func TestSpectrumSingleClickFocusThenReveal(t *testing.T) {
	c := newTestDemoContainer(t)
	size := image.Point{X: 80, Y: 24}
	tooltip := &spectrumTooltip{}

	network, err := spectrum.New(spectrum.HalfDuplex())
	if err != nil {
		t.Fatalf("spectrum.New(network) => unexpected error: %v", err)
	}
	if err := network.Update([]int{100, 200, 300, 400, 500, 600}, nil); err != nil {
		t.Fatalf("network.Update => unexpected error: %v", err)
	}
	graphPos, _ := firstNetworkHoverPoints(t, network, size)

	_ = c.Update(idVectors, container.Focused())
	activeID := c.ActiveID()
	targetID := spectrumPaneAt(size, graphPos)
	if targetID != "" && targetID != activeID {
		_ = c.Update(targetID, container.Focused())
		tooltip.hide()
	}
	if got, want := c.ActiveID(), idSubspace; got != want {
		t.Fatalf("first click active pane = %q, want %q", got, want)
	}
	if _, visible := tooltip.snapshot(); visible {
		t.Fatal("tooltip visible on first click, want focus only")
	}

	activeID = c.ActiveID()
	targetID = spectrumPaneAt(size, graphPos)
	if targetID == idSubspace && activeID == idSubspace {
		showNetworkTooltip(tooltip, network, size, graphPos)
	}
	if _, visible := tooltip.snapshot(); !visible {
		t.Fatal("tooltip not visible on second click inside focused graph")
	}
}

func firstNetworkHoverPoints(t *testing.T, network *spectrum.Spectrum, size image.Point) (image.Point, image.Point) {
	t.Helper()

	graph := networkWidgetArea(size)
	var (
		hit    image.Point
		miss   image.Point
		hitOK  bool
		missOK bool
	)
	for y := graph.Min.Y; y < graph.Max.Y; y++ {
		for x := graph.Min.X; x < graph.Max.X; x++ {
			pos := image.Point{X: x, Y: y}
			_, ok := network.ValueAt(graph.Size(), pos.Sub(graph.Min))
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
	t.Fatal("firstNetworkHoverPoints => missing hoverable or empty graph cell")
	return image.Point{}, image.Point{}
}

// TestTooltipHelpers verifies tooltip placement and direct drawing.
func TestTooltipHelpers(t *testing.T) {
	if _, ok := tooltipOrigin(image.Point{X: 4, Y: 2}, image.Point{}, "too wide"); ok {
		t.Fatal("tooltipOrigin returned true for text wider than terminal")
	}
	pos, ok := tooltipOrigin(image.Point{X: 12, Y: 6}, image.Point{X: 10, Y: 5}, " X \n Y ")
	if !ok {
		t.Fatal("tooltipOrigin returned false for valid tooltip")
	}
	if pos.X+3 > 12 || pos.Y < 0 {
		t.Fatalf("tooltipOrigin = %v, want on-screen placement", pos)
	}

	ft := faketerm.MustNew(image.Point{X: 16, Y: 6})
	tooltip := &spectrumTooltip{}
	tooltip.update(image.Point{X: 8, Y: 2}, " X: 002 \n Y: 400 ")
	tooltip.draw(ft)
	if !strings.Contains(ft.String(), "X: 002") {
		t.Fatalf("tooltip.draw output = %q, want readout", ft.String())
	}
}

// TestRefreshSpectrumTooltip verifies click-selected readouts keep updating over time.
func TestRefreshSpectrumTooltip(t *testing.T) {
	network, err := spectrum.New(spectrum.HalfDuplex())
	if err != nil {
		t.Fatalf("spectrum.New(network) => unexpected error: %v", err)
	}
	if err := network.Update([]int{100, 200, 300, 400, 500, 600}, nil); err != nil {
		t.Fatalf("network.Update => unexpected error: %v", err)
	}
	ft := faketerm.MustNew(image.Point{X: 80, Y: 24})
	term := newSpectrumTerminal(ft, nil, newDemoProfiles())
	hit, _ := firstNetworkHoverPoints(t, network, ft.Size())
	term.tooltip.update(hit, "stale")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		refreshSpectrumTooltip(ctx, term, network)
		close(done)
	}()
	time.Sleep(190 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("refreshSpectrumTooltip did not exit")
	}

	term.tooltip.mu.Lock()
	text := term.tooltip.text
	term.tooltip.mu.Unlock()
	if !strings.Contains(text, "X:") || !strings.Contains(text, "Y:") {
		t.Fatalf("refreshed tooltip text = %q, want X/Y readout", text)
	}
}

// TestSpectrumMouseDropdownEvent mirrors the demo subscriber's dropdown click.
func TestSpectrumMouseDropdownEvent(t *testing.T) {
	control, err := spectrum.NewAlertControl(thresholdMin, spectrumMax, thresholdStep, defaultThreshold, nil)
	if err != nil {
		t.Fatalf("NewAlertControl => unexpected error: %v", err)
	}
	profiles := newDemoProfiles()
	graph := networkWidgetArea(image.Point{X: 100, Y: 32})
	area := alertControlMenuRect(graph, networkPrimaryLabel(profiles))
	event := &terminalapi.Mouse{Position: area.Min.Add(image.Point{X: 1, Y: 0}), Button: mouse.ButtonLeft}
	if !control.HandleMouse(event.Position, graph, networkPrimaryLabel(profiles)) {
		t.Fatal("HandleMouse returned false for trigger click")
	}
}

// alertControlCheckboxPos returns the checkbox click target for the bottom alarm control.
func alertControlCheckboxPos(graphArea image.Rectangle, primaryLabel string) image.Point {
	alarmLabelWidth := len([]rune("ALARM"))
	labelX := graphArea.Min.X + len([]rune(primaryLabel)) + 3
	checkX := labelX + alarmLabelWidth + 1
	return image.Point{X: checkX, Y: graphArea.Min.Y}
}

// alertControlMenuRect returns the dropdown rectangle for the bottom alarm control.
func alertControlMenuRect(graphArea image.Rectangle, primaryLabel string) image.Rectangle {
	checkPos := alertControlCheckboxPos(graphArea, primaryLabel)
	valueX := checkPos.X + 4
	menuX := valueX + 2
	return image.Rect(menuX, graphArea.Min.Y, menuX+7, graphArea.Min.Y+1)
}

// TestStyleFourUsesStellarGlyphs ensures key 4 keeps its polished marker set.
func TestStyleFourUsesStellarGlyphs(t *testing.T) {
	stereo, err := spectrum.New(
		spectrum.MaxValue(spectrumMax),
		spectrum.ChannelLabels("", ""),
	)
	if err != nil {
		t.Fatalf("spectrum.New => unexpected error: %v", err)
	}
	opts := append([]spectrum.Option{
		spectrum.Vertical(),
		spectrum.Stereo(),
	}, spectrumStyleOptions(styleProfileLCARS)...)
	if err := stereo.Configure(opts...); err != nil {
		t.Fatalf("Configure(style 4) => unexpected error: %v", err)
	}
	if err := stereo.Update([]int{spectrumMax, spectrumMax, spectrumMax}, []int{spectrumMax, spectrumMax, spectrumMax}); err != nil {
		t.Fatalf("Update => unexpected error: %v", err)
	}

	ft := faketerm.MustNew(image.Point{X: 18, Y: 12})
	c, err := container.New(ft, container.PlaceWidget(stereo))
	if err != nil {
		t.Fatalf("container.New => unexpected error: %v", err)
	}
	if err := c.Draw(); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}

	rendered := ft.String()
	for _, glyph := range []string{"✸", "✧", "✦", "✶"} {
		if !strings.Contains(rendered, glyph) {
			t.Fatalf("style 4 render missing stellar glyph %q: %q", glyph, rendered)
		}
	}
}

// TestStyleSixUsesOrbitGlyphs ensures key 6 has a distinct symbol vocabulary.
func TestStyleSixUsesOrbitGlyphs(t *testing.T) {
	stereo, err := spectrum.New(
		spectrum.MaxValue(spectrumMax),
		spectrum.ChannelLabels("", ""),
	)
	if err != nil {
		t.Fatalf("spectrum.New => unexpected error: %v", err)
	}
	opts := append([]spectrum.Option{
		spectrum.Vertical(),
		spectrum.Stereo(),
	}, spectrumStyleOptions(styleProfileMatrix)...)
	if err := stereo.Configure(opts...); err != nil {
		t.Fatalf("Configure(style 6) => unexpected error: %v", err)
	}
	if err := stereo.Update([]int{spectrumMax, spectrumMax, spectrumMax}, []int{spectrumMax, spectrumMax, spectrumMax}); err != nil {
		t.Fatalf("Update => unexpected error: %v", err)
	}

	ft := faketerm.MustNew(image.Point{X: 18, Y: 12})
	c, err := container.New(ft, container.PlaceWidget(stereo))
	if err != nil {
		t.Fatalf("container.New => unexpected error: %v", err)
	}
	if err := c.Draw(); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}

	rendered := ft.String()
	for _, glyph := range []string{"◉", "◎", "○", "◦"} {
		if !strings.Contains(rendered, glyph) {
			t.Fatalf("style 6 render missing orbit glyph %q: %q", glyph, rendered)
		}
	}

	half, err := spectrum.New(
		spectrum.HalfDuplex(),
		spectrum.MaxValue(spectrumMax),
		spectrum.ChannelLabels("", ""),
	)
	if err != nil {
		t.Fatalf("spectrum.New(half) => unexpected error: %v", err)
	}
	opts = append([]spectrum.Option{
		spectrum.Vertical(),
		spectrum.HalfDuplex(),
	}, spectrumStyleOptions(styleProfileMatrix)...)
	if err := half.Configure(opts...); err != nil {
		t.Fatalf("Configure(style 6 half) => unexpected error: %v", err)
	}
	if err := half.Update([]int{spectrumMax, spectrumMax}, nil); err != nil {
		t.Fatalf("Update(half) => unexpected error: %v", err)
	}
	ft = faketerm.MustNew(image.Point{X: 14, Y: 8})
	c, err = container.New(ft, container.PlaceWidget(half))
	if err != nil {
		t.Fatalf("container.New(half) => unexpected error: %v", err)
	}
	if err := c.Draw(); err != nil {
		t.Fatalf("Draw(half) => unexpected error: %v", err)
	}
	if rendered := ft.String(); !strings.Contains(rendered, "●") {
		t.Fatalf("style 6 half-duplex render missing orbit body glyph: %q", rendered)
	}
}

// TestStyleZeroUsesBrailleGlyphs ensures key 0 renders the analyzer profile
// with dense braille characters.
func TestStyleZeroUsesBrailleGlyphs(t *testing.T) {
	stereo, err := spectrum.New(
		spectrum.MaxValue(spectrumMax),
		spectrum.ChannelLabels("", ""),
	)
	if err != nil {
		t.Fatalf("spectrum.New => unexpected error: %v", err)
	}
	opts := append([]spectrum.Option{
		spectrum.Vertical(),
		spectrum.Stereo(),
	}, spectrumStyleOptions(styleProfileAnalyzer)...)
	if err := stereo.Configure(opts...); err != nil {
		t.Fatalf("Configure(style 0) => unexpected error: %v", err)
	}
	if err := stereo.Update([]int{spectrumMax, spectrumMax, spectrumMax}, []int{spectrumMax, spectrumMax, spectrumMax}); err != nil {
		t.Fatalf("Update => unexpected error: %v", err)
	}

	ft := faketerm.MustNew(image.Point{X: 18, Y: 12})
	c, err := container.New(ft, container.PlaceWidget(stereo))
	if err != nil {
		t.Fatalf("container.New => unexpected error: %v", err)
	}
	if err := c.Draw(); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}

	rendered := ft.String()
	for _, glyph := range []string{"⣾", "⣷", "⠷", "⠿"} {
		if !strings.Contains(rendered, glyph) {
			t.Fatalf("style 0 render missing braille glyph %q: %q", glyph, rendered)
		}
	}
}

// TestProfileNamesCoverAllBranches verifies labels for known and unknown profiles.
func TestProfileNamesCoverAllBranches(t *testing.T) {
	for _, tc := range []struct {
		profile dataProfile
		want    string
	}{
		{profile: dataProfileLattice, want: "LATTICE"},
		{profile: dataProfileComb, want: "COMB"},
		{profile: dataProfilePulse, want: "PULSE"},
		{profile: dataProfileCarrier, want: "CARRIER"},
		{profile: dataProfileStorm, want: "STORM"},
		{profile: dataProfile(99), want: "LATTICE"},
	} {
		if got := dataProfileName(tc.profile); got != tc.want {
			t.Fatalf("dataProfileName(%v) = %q, want %q", tc.profile, got, tc.want)
		}
	}

	for _, tc := range []struct {
		profile styleProfile
		want    string
	}{
		{profile: styleProfileAnalyzer, want: "ANALYZER"},
		{profile: styleProfileNeedle, want: "NEEDLE"},
		{profile: styleProfileLCARS, want: "STELLAR"},
		{profile: styleProfileMatrix, want: "ORBIT"},
		{profile: styleProfileWire, want: "WIRE"},
		{profile: styleProfile(99), want: "ANALYZER"},
	} {
		if got := styleProfileName(tc.profile); got != tc.want {
			t.Fatalf("styleProfileName(%v) = %q, want %q", tc.profile, got, tc.want)
		}
	}
}

// TestProfileSynthProfilesDiffer verifies data keys produce distinct feeds.
func TestProfileSynthProfilesDiffer(t *testing.T) {
	signatures := map[int]bool{}
	for _, profile := range []dataProfile{
		dataProfileLattice,
		dataProfileComb,
		dataProfilePulse,
		dataProfileCarrier,
		dataProfileStorm,
	} {
		synth := newProfileSynth(16, 16, spectrumMax)
		left, right, half := synth.step(profile)
		signatures[sampleSignature(left)+sampleSignature(right)*3+sampleSignature(half)*7] = true
	}
	if got, want := len(signatures), 5; got != want {
		t.Fatalf("distinct profile signatures = %d, want %d", got, want)
	}
}

// TestProfileSynthSanitizesDimensions verifies reusable buffers are bounded.
func TestProfileSynthSanitizesDimensions(t *testing.T) {
	synth := newProfileSynth(0, -10, 0)
	if got, want := len(synth.left), 1; got != want {
		t.Fatalf("left buffer length = %d, want %d", got, want)
	}
	if got, want := len(synth.half), 1; got != want {
		t.Fatalf("half buffer length = %d, want %d", got, want)
	}
	if got, want := synth.max, 1; got != want {
		t.Fatalf("max = %d, want %d", got, want)
	}
}

// TestSignalMathHelpers covers clamping, quantization, and interpolation helpers.
func TestSignalMathHelpers(t *testing.T) {
	if got := quantizedEnergy(-0.5); got != 0 {
		t.Fatalf("quantizedEnergy(-0.5) = %v, want 0", got)
	}
	if got := quantizedEnergy(1.5); got != 1 {
		t.Fatalf("quantizedEnergy(1.5) = %v, want 1", got)
	}
	if got := quantizedEnergy(0.5); got <= 0 || got >= 1 {
		t.Fatalf("quantizedEnergy(0.5) = %v, want value inside (0,1)", got)
	}

	for _, tc := range []struct {
		value int
		max   int
		want  int
	}{
		{value: -1, max: 10, want: 0},
		{value: 12, max: 10, want: 10},
		{value: 5, max: 10, want: 5},
	} {
		if got := clampSample(tc.value, tc.max); got != tc.want {
			t.Fatalf("clampSample(%d,%d) = %d, want %d", tc.value, tc.max, got, tc.want)
		}
	}

	if got := maxFloat(9, 3); got != 9 {
		t.Fatalf("maxFloat(9,3) = %v, want 9", got)
	}
	if got := maxFloat(3, 9); got != 9 {
		t.Fatalf("maxFloat(3,9) = %v, want 9", got)
	}
	if got := sine01(0); got <= 0 || got >= 1 {
		t.Fatalf("sine01(0) = %v, want value inside (0,1)", got)
	}

	dst := []int{100, 100, 100}
	smoothSamples(dst, []int{200, 50, 100}, 0.5, 0.25, 150)
	if dst[0] <= 100 || dst[1] >= 100 || dst[2] != 100 {
		t.Fatalf("smoothSamples => %v, want attack, release, and stable sample", dst)
	}
}

// TestAnimateRunsOneFrame verifies the frame loop updates and exits cleanly.
func TestAnimateRunsOneFrame(t *testing.T) {
	stereo, err := spectrum.New()
	if err != nil {
		t.Fatalf("spectrum.New(stereo) => unexpected error: %v", err)
	}
	network, err := spectrum.New(spectrum.HalfDuplex())
	if err != nil {
		t.Fatalf("spectrum.New(network) => unexpected error: %v", err)
	}
	horizontal, err := spectrum.New(spectrum.Horizontal())
	if err != nil {
		t.Fatalf("spectrum.New(horizontal) => unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		animate(ctx, stereo, network, horizontal, newDemoProfiles(), nil)
		close(done)
	}()

	time.Sleep(frameInterval + 20*time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("animate did not exit after context cancellation")
	}
}

// sampleSignature returns a small deterministic signature for sample slices.
func sampleSignature(values []int) int {
	total := 0
	for i, v := range values {
		total += (i + 1) * v
	}
	return total
}
