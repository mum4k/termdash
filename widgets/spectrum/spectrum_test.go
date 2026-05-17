package spectrum

import (
	"image"
	"reflect"
	"strings"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// TestNew validates option parsing and constructor invariants.
func TestNew(t *testing.T) {
	tests := []struct {
		desc    string
		opts    []Option
		wantErr bool
	}{
		{desc: "accepts defaults"},
		{desc: "rejects negative height", opts: []Option{Height(-1)}, wantErr: true},
		{desc: "rejects negative max", opts: []Option{MaxValue(-1)}, wantErr: true},
		{desc: "rejects empty gradient", opts: []Option{Gradient()}, wantErr: true},
		{desc: "rejects negative threshold", opts: []Option{Threshold(-1)}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			s, err := New(tc.opts...)
			if tc.wantErr {
				if err == nil {
					t.Fatal("New => nil error, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}
			if s == nil {
				t.Fatal("New => nil spectrum")
			}
		})
	}
}

// TestConfigure verifies live option changes preserve data and validate input.
func TestConfigure(t *testing.T) {
	s, err := New(MaxValue(10))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := s.SetStereo([]int{1, 2, 3}, []int{3, 2, 1}); err != nil {
		t.Fatalf("SetStereo => unexpected error: %v", err)
	}

	if err := s.Configure(
		Horizontal(),
		HalfDuplex(),
		ChannelLabels("PING", "JITTER"),
		Gradient(cell.ColorRed, cell.ColorYellow),
		Threshold(5),
		ThresholdLineColor(cell.ColorCyan),
		AlertColor(cell.ColorMagenta),
		PeakRunes('x', 'y'),
		PrimaryRunes('a', 'b'),
		SecondaryRunes('c', 'd'),
		HorizontalRunes('e', 'f'),
		HalfDuplexRune('z'),
	); err != nil {
		t.Fatalf("Configure => unexpected error: %v", err)
	}
	if got, want := s.opts.orientation, OrientationHorizontal; got != want {
		t.Fatalf("orientation = %v, want %v", got, want)
	}
	if got, want := s.mode, ModeHalfDuplex; got != want {
		t.Fatalf("mode = %v, want %v", got, want)
	}
	if got, want := s.opts.primaryLabel, "PING"; got != want {
		t.Fatalf("primary label = %q, want %q", got, want)
	}
	if got, want := s.opts.threshold, 5; got != want {
		t.Fatalf("threshold = %d, want %d", got, want)
	}
	if got, want := s.opts.thresholdLine, cell.ColorCyan; got != want {
		t.Fatalf("threshold line color = %v, want %v", got, want)
	}
	if got, want := s.opts.alertColor, cell.ColorMagenta; got != want {
		t.Fatalf("alert color = %v, want %v", got, want)
	}
	if got, want := s.opts.secondLabel, "JITTER"; got != want {
		t.Fatalf("secondary label = %q, want %q", got, want)
	}
	if got, want := s.opts.primaryPeak, 'x'; got != want {
		t.Fatalf("primary peak = %q, want %q", got, want)
	}
	if got, want := s.opts.secondPeak, 'y'; got != want {
		t.Fatalf("secondary peak = %q, want %q", got, want)
	}
	if got, want := s.opts.halfRune, 'z'; got != want {
		t.Fatalf("half rune = %q, want %q", got, want)
	}
	if got, want := string(s.opts.primaryRunes), "ab"; got != want {
		t.Fatalf("primary runes = %q, want %q", got, want)
	}
	if got, want := string(s.opts.secondRunes), "cd"; got != want {
		t.Fatalf("secondary runes = %q, want %q", got, want)
	}
	if got, want := string(s.opts.horizRunes), "ef"; got != want {
		t.Fatalf("horizontal runes = %q, want %q", got, want)
	}
	if got, want := s.primary, []int{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("primary samples = %v, want %v", got, want)
	}

	before := s.opts.primaryLabel
	if err := s.Configure(Gradient()); err == nil {
		t.Fatal("Configure(empty Gradient) => nil error, want error")
	}
	if got := s.opts.primaryLabel; got != before {
		t.Fatalf("primary label after failed configure = %q, want %q", got, before)
	}
}

// TestSettersAndCapacity verifies channel setters and visible span reporting.
func TestSettersAndCapacity(t *testing.T) {
	s, err := New(MaxValue(10))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := s.SetStereo([]int{1, 2, 3}, []int{4, 5, 6}); err != nil {
		t.Fatalf("SetStereo => unexpected error: %v", err)
	}
	if err := s.SetHalfDuplex([]int{2, 4, 6}); err != nil {
		t.Fatalf("SetHalfDuplex => unexpected error: %v", err)
	}
	if err := s.SetHalfDuplex([]int{-1}); err == nil {
		t.Fatal("SetHalfDuplex => nil error for negative sample")
	}
	if err := s.SetStereo([]int{1}, []int{-1}); err == nil {
		t.Fatal("SetStereo => nil error for negative secondary sample")
	}
	if err := s.SetStereo([]int{-1}, []int{1}); err == nil {
		t.Fatal("SetStereo => nil error for negative primary sample")
	}
	if err := s.Update([]int{1, 2, 3}, []int{4, 5, 6}); err != nil {
		t.Fatalf("Update(stereo) => unexpected error: %v", err)
	}
	if err := s.Update([]int{2, 4, 6}, nil); err != nil {
		t.Fatalf("Update(half) => unexpected error: %v", err)
	}

	drawSpectrum(t, s, image.Point{X: 10, Y: 6})
	if got := s.ValueCapacity(); got != 10 {
		t.Fatalf("ValueCapacity = %d, want 10", got)
	}
}

// TestOptionsAndSizing ensures option knobs are reflected in internal state.
func TestOptionsAndSizing(t *testing.T) {
	s, err := New(
		Vertical(),
		Stereo(),
		Height(7),
		MaxValue(64),
		ChannelLabels("PING", "LAT"),
		LabelCellOpts(cell.FgColor(cell.ColorCyan)),
		AxisCellOpts(cell.FgColor(cell.ColorRed)),
		Gradient(cell.ColorGreen, cell.ColorYellow),
		PeakRunes('A', 'V'),
		HalfDuplexRune('#'),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if got := s.Options().MinimumSize.Y; got != 7 {
		t.Fatalf("Options.MinimumSize.Y = %d, want 7", got)
	}
	if got := s.Options().MaximumSize.Y; got != 7 {
		t.Fatalf("Options.MaximumSize.Y = %d, want 7", got)
	}
	if got := s.opts.primaryLabel; got != "PING" {
		t.Fatalf("primary label = %q, want %q", got, "PING")
	}
	if got := s.opts.secondLabel; got != "LAT" {
		t.Fatalf("secondary label = %q, want %q", got, "LAT")
	}
	if got := s.opts.primaryPeak; got != 'A' {
		t.Fatalf("primary peak = %q, want %q", got, 'A')
	}
	if got := s.opts.secondPeak; got != 'V' {
		t.Fatalf("secondary peak = %q, want %q", got, 'V')
	}
	if got := s.opts.halfRune; got != '#' {
		t.Fatalf("half-duplex rune = %q, want %q", got, '#')
	}
	if len(s.opts.labelCellOpts) != 1 || len(s.opts.axisCellOpts) != 1 {
		t.Fatal("expected label and axis cell options to be set")
	}
	ft := faketerm.MustNew(image.Point{X: 20, Y: 10})
	if got := s.area(testcanvas.MustNew(ft.Area())); got.Min.Y != 3 {
		t.Fatalf("area.Min.Y = %d, want 3", got.Min.Y)
	}
}

// TestDrawVerticalStereo checks mirrored vertical stereo rendering basics.
func TestDrawVerticalStereo(t *testing.T) {
	s, err := New(ChannelLabels("LEFT", "RIGHT"), MaxValue(8))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := s.SetStereo([]int{2, 4, 8}, []int{1, 3, 6}); err != nil {
		t.Fatalf("SetStereo => unexpected error: %v", err)
	}

	ft := drawSpectrum(t, s, image.Point{X: 8, Y: 8})
	lines := strings.Split(ft.String(), "\n")
	if got := string([]rune(lines[0])[:4]); got != "LEFT" {
		t.Fatalf("top label = %q, want %q", got, "LEFT")
	}
	if got := string([]rune(lines[7])[:5]); got != "RIGHT" {
		t.Fatalf("bottom label = %q, want %q", got, "RIGHT")
	}
	if got := string([]rune(lines[3])); !strings.Contains(got, "──") {
		t.Fatalf("axis row = %q, want horizontal axis", got)
	}
	if got := string([]rune(lines[1])); !strings.Contains(got, "^") {
		t.Fatalf("upper bars = %q, want peak marker", got)
	}
	if got := string([]rune(lines[6])); !strings.Contains(got, "v") {
		t.Fatalf("lower bars = %q, want lower peak marker", got)
	}
	aboveAxis := []rune(lines[2])
	belowAxis := []rune(lines[4])
	if aboveAxis[0] == ' ' && belowAxis[0] == ' ' {
		t.Fatalf("left edge cells = %q/%q, want stereo data spanning from left edge", aboveAxis[0], belowAxis[0])
	}
}

// TestDrawVerticalHalfDuplex checks vertical half-duplex layout and axis.
func TestDrawVerticalHalfDuplex(t *testing.T) {
	s, err := New(HalfDuplex(), ChannelLabels("PING", "LATENCY"), MaxValue(8))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := s.SetHalfDuplex([]int{2, 4, 8}); err != nil {
		t.Fatalf("SetHalfDuplex => unexpected error: %v", err)
	}

	ft := drawSpectrum(t, s, image.Point{X: 8, Y: 7})
	lines := strings.Split(ft.String(), "\n")
	if got := string([]rune(lines[0])[:4]); got != "PING" {
		t.Fatalf("top label = %q, want %q", got, "PING")
	}
	if got := string([]rune(lines[6])[:7]); got != "LATENCY"[:7] {
		t.Fatalf("bottom label = %q, want prefix of LATENCY", got)
	}
	if got := string([]rune(lines[5])); !strings.Contains(got, "──") {
		t.Fatalf("half-duplex axis row = %q, want horizontal axis", got)
	}
	if got := []rune(lines[4])[0]; got == ' ' {
		t.Fatalf("left edge cell above axis = %q, want half-duplex data to span from left edge", got)
	}
}

// TestDrawThreshold verifies threshold line and alert coloring.
func TestDrawThreshold(t *testing.T) {
	s, err := New(
		HalfDuplex(),
		ChannelLabels("", ""),
		MaxValue(100),
		Gradient(cell.ColorGreen),
		Threshold(50),
		ThresholdLineColor(cell.ColorRed),
		AlertColor(cell.ColorRed),
		HalfDuplexRune('|'),
	)
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := s.SetHalfDuplex([]int{100, 100, 100, 100}); err != nil {
		t.Fatalf("SetHalfDuplex => unexpected error: %v", err)
	}

	ft := drawSpectrum(t, s, image.Point{X: 4, Y: 6})
	foundLine := false
	foundAlert := false
	for x := 0; x < 4; x++ {
		for y := 0; y < 6; y++ {
			cellAt := ft.BackBuffer()[x][y]
			if cellAt.Opts == nil || cellAt.Opts.FgColor != cell.ColorRed {
				continue
			}
			switch cellAt.Rune {
			case '─':
				foundLine = true
			case '^', '|':
				foundAlert = true
			}
		}
	}
	if !foundLine {
		t.Fatal("threshold line with alert color not found")
	}
	if !foundAlert {
		t.Fatal("bar cell above threshold with alert color not found")
	}

	stereo, err := New(
		ChannelLabels("", ""),
		MaxValue(100),
		Threshold(50),
		ThresholdLineColor(cell.ColorRed),
		AlertColor(cell.ColorRed),
	)
	if err != nil {
		t.Fatalf("New(stereo) => unexpected error: %v", err)
	}
	if err := stereo.SetStereo([]int{100, 100, 100, 100}, []int{100, 100, 100, 100}); err != nil {
		t.Fatalf("SetStereo => unexpected error: %v", err)
	}
	ft = drawSpectrum(t, stereo, image.Point{X: 4, Y: 8})
	if !hasColoredRune(ft, '─', cell.ColorRed) {
		t.Fatal("vertical stereo threshold line with alert color not found")
	}
}

// TestDrawHorizontalThreshold verifies threshold markers for horizontal modes.
func TestDrawHorizontalThreshold(t *testing.T) {
	half, err := New(
		Horizontal(),
		HalfDuplex(),
		ChannelLabels("", ""),
		MaxValue(100),
		Threshold(50),
		ThresholdLineColor(cell.ColorRed),
		AlertColor(cell.ColorRed),
	)
	if err != nil {
		t.Fatalf("New(half) => unexpected error: %v", err)
	}
	if err := half.SetHalfDuplex([]int{100, 100, 100}); err != nil {
		t.Fatalf("SetHalfDuplex => unexpected error: %v", err)
	}
	ft := drawSpectrum(t, half, image.Point{X: 8, Y: 5})
	if !hasColoredRune(ft, '│', cell.ColorRed) {
		t.Fatal("horizontal half-duplex threshold column with alert color not found")
	}

	stereo, err := New(
		Horizontal(),
		ChannelLabels("", ""),
		MaxValue(100),
		Threshold(50),
		ThresholdLineColor(cell.ColorRed),
		AlertColor(cell.ColorRed),
	)
	if err != nil {
		t.Fatalf("New(stereo) => unexpected error: %v", err)
	}
	if err := stereo.SetStereo([]int{100, 100, 100}, []int{100, 100, 100}); err != nil {
		t.Fatalf("SetStereo => unexpected error: %v", err)
	}
	ft = drawSpectrum(t, stereo, image.Point{X: 10, Y: 5})
	if !hasColoredRune(ft, '│', cell.ColorRed) {
		t.Fatal("horizontal stereo threshold column with alert color not found")
	}
}

// TestValueAt verifies mouse-position readouts can map to visible samples.
func TestValueAt(t *testing.T) {
	s, err := New(HalfDuplex(), ChannelLabels("PING", "LATENCY"), MaxValue(100))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := s.SetHalfDuplex([]int{10, 20, 30, 40, 50, 60}); err != nil {
		t.Fatalf("SetHalfDuplex => unexpected error: %v", err)
	}

	got, ok := s.ValueAt(image.Point{X: 6, Y: 7}, image.Point{X: 4, Y: 3})
	if !ok {
		t.Fatal("ValueAt returned false, want sample")
	}
	if want := (Sample{X: 5, Y: 50}); got != want {
		t.Fatalf("ValueAt = %+v, want %+v", got, want)
	}
	if _, ok := s.ValueAt(image.Point{X: 6, Y: 7}, image.Point{X: 4, Y: 0}); ok {
		t.Fatal("ValueAt above half-duplex bar returned true, want false")
	}
	if _, ok := s.ValueAt(image.Point{X: 6, Y: 7}, image.Point{X: 99, Y: 2}); ok {
		t.Fatal("ValueAt outside canvas returned true, want false")
	}
	if _, ok := s.ValueAt(image.Point{}, image.Point{}); ok {
		t.Fatal("ValueAt zero size returned true, want false")
	}

	horizontal, err := New(Horizontal(), ChannelLabels("LEFT", "RIGHT"), MaxValue(100))
	if err != nil {
		t.Fatalf("New(horizontal) => unexpected error: %v", err)
	}
	if err := horizontal.SetStereo([]int{10, 20, 30, 40}, []int{50, 60, 70, 80}); err != nil {
		t.Fatalf("SetStereo(horizontal) => unexpected error: %v", err)
	}
	got, ok = horizontal.ValueAt(image.Point{X: 10, Y: 5}, image.Point{X: 3, Y: 2})
	if !ok {
		t.Fatal("ValueAt horizontal returned false, want sample")
	}
	if want := (Sample{X: 2, Y: 20}); got != want {
		t.Fatalf("ValueAt horizontal = %+v, want %+v", got, want)
	}
	if _, ok := horizontal.ValueAt(image.Point{X: 10, Y: 5}, image.Point{X: 2, Y: 2}); ok {
		t.Fatal("ValueAt horizontal empty space returned true, want false")
	}
}

// TestDrawHorizontalStereo checks mirrored horizontal stereo rendering basics.
func TestDrawHorizontalStereo(t *testing.T) {
	s, err := New(Horizontal(), ChannelLabels("LEFT", "RIGHT"), MaxValue(8), PeakRunes('<', '>'))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := s.SetStereo([]int{2, 4, 8}, []int{1, 3, 6}); err != nil {
		t.Fatalf("SetStereo => unexpected error: %v", err)
	}

	ft := drawSpectrum(t, s, image.Point{X: 12, Y: 6})
	lines := strings.Split(ft.String(), "\n")
	if got := string([]rune(lines[0])[:4]); got != "LEFT" {
		t.Fatalf("top row left label = %q, want %q", got, "LEFT")
	}
	if got := string([]rune(lines[0])[7:12]); got != "RIGHT" {
		t.Fatalf("top row right label = %q, want %q", got, "RIGHT")
	}
	body := strings.Join(lines[1:6], "\n")
	if !strings.Contains(body, "│") {
		t.Fatalf("horizontal body = %q, want vertical axis", body)
	}
	if !strings.Contains(body, "<") || !strings.Contains(body, ">") {
		t.Fatalf("horizontal body = %q, want mirrored peak markers", body)
	}
}

// TestDrawHorizontalHalfDuplex checks horizontal half-duplex rendering basics.
func TestDrawHorizontalHalfDuplex(t *testing.T) {
	s, err := New(Horizontal(), HalfDuplex(), ChannelLabels("PING", "LATENCY"), MaxValue(8), PeakRunes('>', '<'))
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := s.SetHalfDuplex([]int{2, 4, 8}); err != nil {
		t.Fatalf("SetHalfDuplex => unexpected error: %v", err)
	}

	ft := drawSpectrum(t, s, image.Point{X: 12, Y: 5})
	body := strings.Join(strings.Split(ft.String(), "\n")[1:5], "\n")
	if !strings.Contains(body, "│") {
		t.Fatalf("horizontal half-duplex body = %q, want axis", body)
	}
	if !strings.Contains(body, ">") {
		t.Fatalf("horizontal half-duplex body = %q, want peak marker", body)
	}
}

// TestDrawResizeNeeded confirms undersized canvases show resize markers.
func TestDrawResizeNeeded(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	ft := faketerm.MustNew(image.Point{X: 2, Y: 2})
	cvs := testcanvas.MustNew(ft.Area())
	if err := s.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	if got := string([]rune(strings.Split(ft.String(), "\n")[0])[:1]); got != "⇄" {
		t.Fatalf("resize marker = %q, want %q", got, "⇄")
	}
}

// TestHelpersAndUnsupportedInput exercises helpers and unsupported event paths.
func TestHelpersAndUnsupportedInput(t *testing.T) {
	if got := tailInts([]int{1, 2, 3}, 2); len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Fatalf("tailInts = %v, want [2 3]", got)
	}
	dst := make([]int, 2, 4)
	copy(dst, []int{9, 9})
	if got := copyInts(dst, []int{1, 2, 3}); len(got) != 3 || got[2] != 3 {
		t.Fatalf("copyInts = %v, want [1 2 3]", got)
	}
	if got := tailInts([]int{1, 2}, 0); got != nil {
		t.Fatalf("tailInts zero max = %v, want nil", got)
	}
	if got := fitSamples([]int{2, 4, 8}, 0); got != nil {
		t.Fatalf("fitSamples zero span = %v, want nil", got)
	}
	if got := fitSamples(nil, 5); got != nil {
		t.Fatalf("fitSamples nil input = %v, want nil", got)
	}
	if got := fitSamples([]int{2, 4, 8}, 1); len(got) != 1 || got[0] != 8 {
		t.Fatalf("fitSamples span one = %v, want [8]", got)
	}
	if got := fitSamples([]int{2, 4, 8}, 6); len(got) != 6 || got[0] != 2 || got[len(got)-1] != 8 {
		t.Fatalf("fitSamples expanded = %v, want bounded interpolation", got)
	}
	if got := fitSamples([]int{1, 2, 3, 4}, 2); len(got) != 2 || got[0] != 3 || got[1] != 4 {
		t.Fatalf("fitSamples tail path = %v, want [3 4]", got)
	}
	if got, ok := sampleAtColumn([]int{10, 20, 30}, 2, 1); !ok || got.X != 3 || got.Y != 30 {
		t.Fatalf("sampleAtColumn tail = %+v/%v, want X=3 Y=30 true", got, ok)
	}
	if _, ok := sampleAtColumn([]int{10}, 2, 2); ok {
		t.Fatal("sampleAtColumn outside span = true, want false")
	}
	if got := maxSample(9, []int{1, 2}, []int{3, 4}); got != 9 {
		t.Fatalf("maxSample fixed = %d, want 9", got)
	}
	if got := maxSample(0, []int{1, 7}, []int{3, 4}); got != 7 {
		t.Fatalf("maxSample dynamic = %d, want 7", got)
	}
	if got := maxSample(0); got != 1 {
		t.Fatalf("maxSample empty = %d, want 1", got)
	}
	if got := upperRune(0, 3); got != 'i' {
		t.Fatalf("upperRune = %q, want %q", got, 'i')
	}
	if got := upperRune(1, 2); got != '!' {
		t.Fatalf("upperRune near peak = %q, want %q", got, '!')
	}
	if got := upperRune(1, 4); got != '|' {
		t.Fatalf("upperRune middle = %q, want %q", got, '|')
	}
	if got := lowerRune(1, 3); got != '¡' {
		t.Fatalf("lowerRune = %q, want %q", got, '¡')
	}
	if got := lowerRune(0, 1); got != 'i' {
		t.Fatalf("lowerRune single = %q, want %q", got, 'i')
	}
	if got := lowerRune(1, 4); got != '|' {
		t.Fatalf("lowerRune middle = %q, want %q", got, '|')
	}
	if got := horizontalBodyRune(1, 3); got != '=' {
		t.Fatalf("horizontalBodyRune = %q, want %q", got, '=')
	}
	if got := channelRune(0, 3, nil, upperRune); got != 'i' {
		t.Fatalf("channelRune fallback = %q, want %q", got, 'i')
	}
	if got := channelRune(0, 1, []rune{'x', 'y'}, upperRune); got != 'x' {
		t.Fatalf("channelRune single cell = %q, want %q", got, 'x')
	}
	if got := channelRune(2, 4, []rune{'a', 'b', 'c'}, upperRune); got != 'c' {
		t.Fatalf("channelRune scaled = %q, want %q", got, 'c')
	}
	if got := gradientColor(nil, 0.5); got != cell.ColorGreen {
		t.Fatalf("gradientColor nil = %v, want %v", got, cell.ColorGreen)
	}
	if got := gradientColor([]cell.Color{cell.ColorGreen}, 0.5); got != cell.ColorGreen {
		t.Fatalf("gradientColor single = %v, want %v", got, cell.ColorGreen)
	}
	if got := gradientColor([]cell.Color{cell.ColorGreen, cell.ColorYellow, cell.ColorRed}, 0); got != cell.ColorGreen {
		t.Fatalf("gradientColor low = %v, want %v", got, cell.ColorGreen)
	}
	if got := gradientColor([]cell.Color{cell.ColorGreen, cell.ColorYellow, cell.ColorRed}, 0.5); got != cell.ColorYellow {
		t.Fatalf("gradientColor mid = %v, want %v", got, cell.ColorYellow)
	}
	if got := gradientColor([]cell.Color{cell.ColorGreen, cell.ColorRed}, 1); got != cell.ColorRed {
		t.Fatalf("gradientColor = %v, want %v", got, cell.ColorRed)
	}
	if got := gradientColor([]cell.Color{cell.ColorGreen, cell.ColorRed}, 2); got != cell.ColorRed {
		t.Fatalf("gradientColor high = %v, want %v", got, cell.ColorRed)
	}
	if got := scaledCells(0, 8, 4); got != 0 {
		t.Fatalf("scaledCells zero = %d, want 0", got)
	}
	if got := scaledCells(8, 8, 4); got != 4 {
		t.Fatalf("scaledCells full = %d, want 4", got)
	}
	if got := scaledCells(4, 8, 4); got != 2 {
		t.Fatalf("scaledCells = %d, want 2", got)
	}
	if got := cellRatio(1, 1); got != 1 {
		t.Fatalf("cellRatio single = %v, want 1", got)
	}
	if got := cellRatio(2, 4); got != 1.0/3.0 {
		t.Fatalf("cellRatio span = %v, want %v", got, 1.0/3.0)
	}
	if got := maxInt(2, 4); got != 4 {
		t.Fatalf("maxInt = %d, want 4", got)
	}
	if got := maxInt(4, 2); got != 4 {
		t.Fatalf("maxInt reverse = %d, want 4", got)
	}

	s, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	if err := s.Keyboard(&terminalapi.Keyboard{}, &widgetapi.EventMeta{}); err == nil {
		t.Fatal("Keyboard => nil error, want unsupported error")
	}
	if err := s.Mouse(&terminalapi.Mouse{}, &widgetapi.EventMeta{}); err == nil {
		t.Fatal("Mouse => nil error, want unsupported error")
	}
}

// TestAxisHelpers validates direct axis drawing helper behavior.
func TestAxisHelpers(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("New => unexpected error: %v", err)
	}
	ft := faketerm.MustNew(image.Point{X: 6, Y: 4})
	cvs := testcanvas.MustNew(ft.Area())
	if err := s.drawHorizontalAxis(cvs, 1, 5, 2); err != nil {
		t.Fatalf("drawHorizontalAxis => unexpected error: %v", err)
	}
	if err := s.drawVerticalAxis(cvs, 3, 0, 4); err != nil {
		t.Fatalf("drawVerticalAxis => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	lines := strings.Split(ft.String(), "\n")
	if got := string([]rune(lines[2])[1:5]); !strings.Contains(got, "─") {
		t.Fatalf("horizontal axis = %q, want line", got)
	}
	if got := string([]rune(lines[0])[3:4]); got != "│" {
		t.Fatalf("vertical axis = %q, want %q", got, "│")
	}
}

// drawSpectrum renders a widget to a fake terminal for assertions.
func drawSpectrum(t *testing.T, s *Spectrum, size image.Point) *faketerm.Terminal {
	t.Helper()

	ft := faketerm.MustNew(size)
	cvs := testcanvas.MustNew(ft.Area())
	if err := s.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	return ft
}

// hasColoredRune reports whether ft contains r with the requested foreground.
func hasColoredRune(ft *faketerm.Terminal, r rune, color cell.Color) bool {
	size := ft.Size()
	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			cellAt := ft.BackBuffer()[x][y]
			if cellAt.Rune == r && cellAt.Opts != nil && cellAt.Opts.FgColor == color {
				return true
			}
		}
	}
	return false
}
