package borderfx

import (
	"context"
	"image"
	"testing"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/private/runewidth"
)

func TestScannerBeamMoves(t *testing.T) {
	e := Scanner(cell.ColorNumber(51), cell.ColorNumber(236))
	bc := container.BorderCell{
		Point:  image.Point{0, 0},
		Border: image.Rect(0, 0, 10, 5),
		Rune:   '─',
		Index:  0,
		Length: 26,
	}

	first := e.nextStyler()(bc)
	second := e.nextStyler()(bc)
	third := e.nextStyler()(bc)
	fourth := e.nextStyler()(bc)

	if first.Rune != '█' {
		t.Fatalf("first frame rune = %q, want beam peak", first.Rune)
	}
	if second.Rune != '█' {
		t.Fatalf("second frame rune = %q, want beam peak", second.Rune)
	}
	if third.Rune != '▓' {
		t.Fatalf("third frame rune = %q, want trailing beam", third.Rune)
	}
	if fourth.Rune != '▒' {
		t.Fatalf("fourth frame rune = %q, want trailing beam", fourth.Rune)
	}
}

func TestEffectsNotEmpty(t *testing.T) {
	for name, e := range map[string]*Effect{
		"fire":  Fire(),
		"ice":   Ice(),
		"warp":  Warp(),
		"pulse": Pulse(cell.ColorBlue, cell.ColorCyan),
		"glow":  Glow(cell.ColorGreen),
		"neon":  Neon(cell.ColorNumber(51)),
		"cycle": Cycle([]cell.Color{cell.ColorRed, cell.ColorBlue}),
		"empty": Cycle(nil),
	} {
		if name == "empty" {
			if got := e.Next(); got != cell.ColorDefault {
				t.Fatalf("%s produced %v, want ColorDefault", name, got)
			}
			continue
		}
		if c := e.Next(); c == cell.ColorDefault {
			t.Fatalf("%s produced default color", name)
		}
	}
}

func TestReset(t *testing.T) {
	e := Cycle([]cell.Color{cell.ColorRed, cell.ColorBlue})
	first := e.Next()
	e.Next()
	e.Reset()
	if after := e.Next(); after != first {
		t.Fatalf("color after reset = %v, want %v", after, first)
	}
}

func TestPresetsAndLegacyEffectsRenderSingleCellBorders(t *testing.T) {
	p := Colors(cell.ColorNumber(51), cell.ColorNumber(24), cell.ColorNumber(236))
	for name, effect := range map[string]*Effect{
		"scanner":        Scanner(p.Bright, p.Dim),
		"dual":           DualScanner(p.Bright, p.Mid, p.Dim),
		"interlace":      InterlacedScanner(p.Bright, p.Mid, p.Dim),
		"braided":        BraidedScanner(p.Bright, p.Mid, p.Dim),
		"shard":          ShardScanner(p.Bright, p.Mid, p.Dim),
		"orbit":          OrbitScanner(p.Bright, p.Mid, p.Dim),
		"focus_rail":     FocusRail(p.Bright, p.Mid, p.Dim),
		"focus_brace":    FocusBrace(p.Bright, p.Mid, p.Dim),
		"focus_matrix":   FocusMatrix(p.Bright, p.Mid, p.Dim),
		"focus_ribbon":   FocusRibbon(p.Bright, p.Mid, p.Dim),
		"focus":          FocusPins(p.Bright, p.Mid, p.Dim),
		"rail":           FocusPinsRail(p.Bright, p.Mid, p.Dim),
		"decode":         FocusPinsMatrix(p.Bright, p.Mid, p.Dim),
		"power":          FocusPinsPower(p.Bright, p.Mid, p.Dim),
		"rain":           RainScanner(p.Bright, p.Mid, p.Dim),
		"braille":        BrailleDrift(p.Bright, p.Mid, p.Dim),
		"bracket":        BracketScan(p.Bright, p.Mid, p.Dim),
		"ticks":          DataTicks(p.Bright, p.Mid, p.Dim),
		"noise":          StaticNoise(p.Bright, p.Mid, p.Dim),
		"spin_pulse":     SpinnerPulse(p.Bright, p.Mid, p.Dim),
		"dots6":          Dots6Spinner(p.Bright, p.Mid, p.Dim),
		"dots10":         Dots10Spinner(p.Bright, p.Mid, p.Dim),
		"ribbon":         FocusPinsRibbon(p.Bright, p.Mid, p.Dim),
		"brace":          FocusPinsBrace(p.Bright, p.Mid, p.Dim),
		"emoji":          FocusPinsEmoji(p.Bright, p.Mid, p.Dim),
		"focus_shard":    FocusPinsShard(p.Bright, p.Mid, p.Dim),
		"pulse":          FocusPinsPulse(p.Bright, p.Mid, p.Dim),
		"palette_fire":   Fire(),
		"palette_ice":    Ice(),
		"palette_rain":   Rainbow(),
		"palette_neon":   Neon(p.Bright),
		"palette_glow":   Glow(p.Bright),
		"palette_matrix": Matrix(),
		"palette_synth":  Synthwave(),
		"palette_cycle":  Cycle([]cell.Color{cell.ColorRed, cell.ColorBlue}),
		"palette_warp":   Warp(),
		"macro_focus":    Presets.Focus.With(p),
		"macro_power":    Presets.Power.With(p),
		"macro_rain":     Presets.Rain.With(p),
		"macro_braille":  Presets.Braille.With(p),
		"macro_bracket":  Presets.Bracket.With(p),
		"macro_ticks":    Presets.Ticks.With(p),
		"macro_noise":    Presets.Noise.With(p),
		"macro_pulse":    Presets.SpinPulse.With(p),
		"macro_dots6":    Presets.Dots6.With(p),
		"macro_dots10":   Presets.Dots10.With(p),
		"macro_emoji":    Presets.Emoji.With(p),
		"macro_scan":     Presets.Scanner.With(p),
		"macro_fire":     Presets.Fire.With(p),
	} {
		if got := effect.Next(); got == cell.ColorDefault && name != "palette_cycle" {
			t.Fatalf("%s Next() returned default color", name)
		}
		for frame := 0; frame < 96; frame++ {
			styler := effect.nextStyler()
			for _, bc := range sampleBorderCells(image.Rect(0, 0, 16, 7)) {
				style := styler(bc)
				if style.Rune != 0 && runewidth.RuneWidth(style.Rune) != 1 {
					t.Fatalf("%s frame %d produced rune %q with width %d", name, frame, style.Rune, runewidth.RuneWidth(style.Rune))
				}
			}
		}
	}
}

func TestFocusPinsEffectsLoop(t *testing.T) {
	e := FocusPinsPower(cell.ColorNumber(51), cell.ColorNumber(24), cell.ColorNumber(236))
	bc := container.BorderCell{
		Point:  image.Point{0, 0},
		Border: image.Rect(0, 0, 20, 6),
		Rune:   '─',
		Index:  0,
		Length: 48,
	}

	first := e.nextStyler()(bc)
	for i := 0; i < 107; i++ {
		_ = e.nextStyler()(bc)
	}
	looped := e.nextStyler()(bc)

	if first.Rune != looped.Rune {
		t.Fatalf("looped rune = %q, want %q after one full cycle", looped.Rune, first.Rune)
	}
}

func TestFocusPinsRailLeavesTitleUntouched(t *testing.T) {
	bright := cell.ColorNumber(51)
	mid := cell.ColorNumber(24)
	dim := cell.ColorNumber(236)
	e := FocusPinsRail(bright, mid, dim)
	topMid := borderMidpoints(image.Rect(0, 0, 11, 5))[0]
	titleCell := container.BorderCell{
		Point:  image.Point{X: 5, Y: 0},
		Border: image.Rect(0, 0, 11, 5),
		Rune:   'T',
		Index:  topMid,
		Length: 28,
		Title:  true,
	}

	style := e.nextStyler()(titleCell)
	if style.Rune != 0 || style.CellOpts != nil {
		t.Fatalf("title orbit style = %+v, want no override", style)
	}
}

func TestFocusPinsRailKeepsCornerPegs(t *testing.T) {
	bright := cell.ColorNumber(51)
	mid := cell.ColorNumber(24)
	dim := cell.ColorNumber(236)
	e := FocusPinsRail(bright, mid, dim)
	corner := container.BorderCell{
		Point:  image.Point{X: 0, Y: 0},
		Border: image.Rect(0, 0, 11, 5),
		Rune:   '╭',
		Index:  0,
		Length: 28,
	}

	style := e.nextStyler()(corner)
	if got := style.Rune; got != '○' {
		t.Fatalf("corner peg rune = %q, want softened peg corner", got)
	}
	if got := cell.NewOptions(style.CellOpts...).FgColor; got == dim {
		t.Fatalf("corner peg color = %v, want brighter than dim", got)
	}
}

func TestInterlacedScannerDimsTitleDuringStripePass(t *testing.T) {
	bright := cell.ColorNumber(51)
	mid := cell.ColorNumber(24)
	dim := cell.ColorNumber(236)
	e := InterlacedScanner(bright, mid, dim)
	titleCell := container.BorderCell{
		Point:  image.Point{X: 1, Y: 0},
		Border: image.Rect(0, 0, 11, 5),
		Rune:   'T',
		Index:  0,
		Length: 28,
		Title:  true,
	}

	style := e.nextStyler()(titleCell)
	if style.Rune != 0 || style.CellOpts != nil {
		t.Fatalf("interlace title style = %+v, want no override", style)
	}
}

func TestRainScannerLeavesTitleUntouched(t *testing.T) {
	bright := cell.ColorNumber(51)
	mid := cell.ColorNumber(24)
	dim := cell.ColorNumber(236)
	e := RainScanner(bright, mid, dim)
	titleCell := container.BorderCell{
		Point:  image.Point{X: 5, Y: 0},
		Border: image.Rect(0, 0, 11, 5),
		Rune:   'T',
		Index:  5,
		Length: 28,
		Title:  true,
	}

	style := e.nextStyler()(titleCell)
	if style.Rune != 0 || style.CellOpts != nil {
		t.Fatalf("rain title style = %+v, want no override", style)
	}
}

func TestBrailleDriftUsesMovingSpinnerBlocks(t *testing.T) {
	e := BrailleDrift(cell.ColorNumber(117), cell.ColorNumber(251), cell.ColorNumber(233))
	bc := container.BorderCell{
		Point:  image.Point{X: 4, Y: 0},
		Border: image.Rect(0, 0, 16, 6),
		Rune:   '─',
		Index:  4,
		Length: 40,
	}

	seen := map[rune]bool{}
	for i := 0; i < 24; i++ {
		style := e.nextStyler()(bc)
		seen[style.Rune] = true
	}
	for _, want := range []rune{'░', '▒', '▓'} {
		if !seen[want] {
			t.Fatalf("BrailleDrift never produced spinner rune %q, saw %v", want, seen)
		}
	}
}

func TestSpinnerFrameEffectsUseSpinner2Frames(t *testing.T) {
	for name, tc := range map[string]struct {
		e    *Effect
		want []rune
	}{
		"pulse": {
			e:    SpinnerPulse(cell.ColorNumber(117), cell.ColorNumber(251), cell.ColorNumber(233)),
			want: []rune{'⎺', '⎻', '⎼', '⎽', '⎼', '⎻'},
		},
		"dots6": {
			e:    Dots6Spinner(cell.ColorNumber(117), cell.ColorNumber(251), cell.ColorNumber(233)),
			want: []rune{'⠁', '⠉', '⠙', '⠚', '⠒', '⠂', '⠂', '⠒', '⠲', '⠴', '⠤', '⠄', '⠄', '⠤', '⠴', '⠲', '⠒', '⠂', '⠂', '⠒', '⠚', '⠙', '⠉', '⠁'},
		},
		"dots10": {
			e:    Dots10Spinner(cell.ColorNumber(117), cell.ColorNumber(251), cell.ColorNumber(233)),
			want: []rune{'⢄', '⢂', '⢁', '⡁', '⡈', '⡐', '⡠'},
		},
	} {
		t.Run(name, func(t *testing.T) {
			bc := container.BorderCell{
				Point:  image.Point{X: 3, Y: 0},
				Border: image.Rect(0, 0, 16, 6),
				Rune:   '─',
				Index:  0,
				Length: 40,
			}

			for _, want := range tc.want {
				style := tc.e.nextStyler()(bc)
				if style.Rune != want {
					t.Fatalf("frame rune = %q, want %q", style.Rune, want)
				}
			}
		})
	}
}

func TestFocusPinsTitleOnlyDimsUnderHorizontalSweep(t *testing.T) {
	bright := cell.ColorNumber(51)
	mid := cell.ColorNumber(24)
	dim := cell.ColorNumber(236)
	variant := pinVariant{
		horizontalFrames: 11,
		verticalFrames:   4,
		holdFrames:       3,
		lag:              1,
		topHead:          '▣',
		bottomHead:       '▣',
		sideHead:         '│',
		cornerRune:       '◉',
		cornerTrail:      '●',
	}
	titleCell := container.BorderCell{
		Point:  image.Point{X: 5, Y: 0},
		Border: image.Rect(0, 0, 11, 4),
		Rune:   'T',
		Index:  5,
		Length: 26,
		Title:  true,
	}

	resting := parkedFocusStyle(0, titleCell, bright, mid, dim, focusProfile{
		parked:   []rune{'◉'},
		parkMode: parkCorners,
	})
	if resting.Rune != 0 || resting.CellOpts != nil {
		t.Fatalf("resting title style = %+v, want no override", resting)
	}

	away := horizontalSweepPins(0, titleCell, bright, mid, dim, variant)
	if away.Rune != 0 || away.CellOpts != nil {
		t.Fatalf("away title sweep style = %+v, want no override", away)
	}

	underSweep := horizontalSweepPins(5, titleCell, bright, mid, dim, variant)
	if underSweep.Rune != 0 {
		t.Fatalf("under-sweep title rune = %q, want unchanged title rune", underSweep.Rune)
	}
	if got := cell.NewOptions(underSweep.CellOpts...).FgColor; got != dim {
		t.Fatalf("under-sweep title color = %v, want %v", got, dim)
	}
}

func TestRoundedCornersStayRoundedDuringPins(t *testing.T) {
	bright := cell.ColorNumber(51)
	mid := cell.ColorNumber(24)
	dim := cell.ColorNumber(236)
	bc := container.BorderCell{
		Point:  image.Point{X: 0, Y: 0},
		Border: image.Rect(0, 0, 11, 4),
		Rune:   '╭',
		Index:  0,
		Length: 26,
	}

	parked := parkedFocusStyle(0, bc, bright, mid, dim, focusProfile{
		parked:   []rune{'◉'},
		parkMode: parkCorners,
	})
	if got := parked.Rune; got != '○' {
		t.Fatalf("parked corner rune = %q, want softened rounded corner", got)
	}
	if got := cell.NewOptions(parked.CellOpts...).FgColor; got != lerp(dim, bright, 0.58) {
		t.Fatalf("parked corner color = %v, want softened rounded glow", got)
	}

	variant := pinVariant{
		horizontalFrames: 11,
		verticalFrames:   4,
		holdFrames:       3,
		lag:              1,
		topHead:          '▣',
		bottomHead:       '▣',
		sideHead:         '│',
		cornerRune:       '◉',
		cornerTrail:      '●',
	}
	sweeping := horizontalSweepPins(0, bc, bright, mid, dim, variant)
	if got := sweeping.Rune; got != '○' {
		t.Fatalf("horizontal sweep corner rune = %q, want softened rounded corner", got)
	}
	if got := cell.NewOptions(sweeping.CellOpts...).FgColor; got != lerp(dim, bright, 0.62) {
		t.Fatalf("horizontal sweep corner color = %v, want softened rounded glow", got)
	}

	vertical := verticalSweepPins(0, bc, bright, mid, dim, variant)
	if got := vertical.Rune; got != '○' {
		t.Fatalf("vertical sweep corner rune = %q, want softened rounded corner", got)
	}
	if got := cell.NewOptions(vertical.CellOpts...).FgColor; got != lerp(dim, bright, 0.62) {
		t.Fatalf("vertical sweep corner color = %v, want softened rounded glow", got)
	}
}

func TestPaletteHelpersAndMacros(t *testing.T) {
	duo := Duo(cell.ColorNumber(51), cell.ColorNumber(236))
	if duo.Mid == cell.ColorDefault || duo.Mid == duo.Bright || duo.Mid == duo.Dim {
		t.Fatalf("Duo() built unexpected palette: %+v", duo)
	}

	if got := Presets.Power.Name(); got != "power" {
		t.Fatalf("Presets.Power.Name() = %q, want power", got)
	}
	if Presets.Power.With(duo) == nil {
		t.Fatal("Presets.Power.With returned nil")
	}
	if Colors(cell.ColorRed, cell.ColorBlue, cell.ColorGreen).Apply(Presets.Focus) == nil {
		t.Fatal("Palette.Apply returned nil")
	}

	var zero Macro
	if zero.With(duo) != nil {
		t.Fatal("zero macro returned non-nil effect")
	}
	zero.Register(nil, "noop", duo)
}

func TestNamedPalettes(t *testing.T) {
	for name, p := range map[string]Palette{
		"cyan":      Palettes.Cyan,
		"amber":     Palettes.Amber,
		"matrix":    Palettes.Matrix,
		"synthwave": Palettes.Synthwave,
		"ice":       Palettes.Ice,
		"silver":    Palettes.Silver,
	} {
		if p.Bright == cell.ColorDefault || p.Dim == cell.ColorDefault {
			t.Fatalf("%s palette is incomplete: %+v", name, p)
		}
	}
}

func TestMustHalfWidthRunesPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("mustHalfWidthRunes did not panic on wide rune")
		}
	}()
	mustHalfWidthRunes('⚡')
}

func TestAnimatorHelpers(t *testing.T) {
	ft := faketerm.MustNew(image.Point{X: 30, Y: 10})
	root, err := container.New(ft,
		container.SplitVertical(
			container.Left(container.ID("left"), container.Focused()),
			container.Right(container.ID("right")),
		),
	)
	if err != nil {
		t.Fatalf("container.New() error = %v", err)
	}

	a := NewAnimator(root)
	a.SetTickRate(time.Millisecond)
	a.SetInactiveStyle(func(id string, bc container.BorderCell) container.BorderCellStyle {
		return container.BorderCellStyle{Rune: bc.Rune}
	})
	a.RegisterMacro("left", Presets.Power, Palettes.Cyan)
	Presets.Emoji.Register(a, "right", Palettes.Ice)
	a.tick()
	a.Unregister("right")
	a.tick()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := a.Run(ctx); err == nil {
		t.Fatal("Run on canceled context returned nil")
	}

	NewAnimator(nil).tick()

	a2 := NewAnimator(root)
	a2.Register("left", Presets.Focus.With(Palettes.Cyan))
	a2.Register("right", Presets.Emoji.With(Palettes.Ice))
	a2.tick()

	ctx2, cancel2 := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- a2.Run(ctx2)
	}()
	time.Sleep(3 * time.Millisecond)
	cancel2()
	if err := <-done; err == nil {
		t.Fatal("Run after tick returned nil")
	}
}

func TestHelpers(t *testing.T) {
	if got := circularDistance(1, 9, 10); got != 2 {
		t.Fatalf("circularDistance = %d, want 2", got)
	}
	if got := circularDistance(1, 3, 10); got != 2 {
		t.Fatalf("circularDistance direct = %d, want 2", got)
	}
	if got := lerp(cell.ColorNumber(20), cell.ColorNumber(40), -1); got != cell.ColorNumber(20) {
		t.Fatalf("lerp below bounds = %v", got)
	}
	if got := lerp(cell.ColorNumber(20), cell.ColorNumber(40), 2); got != cell.ColorNumber(40) {
		t.Fatalf("lerp above bounds = %v", got)
	}
	if got := lerp(cell.Color(-20), cell.ColorNumber(20), 0.25); got != cell.ColorNumber(0) {
		t.Fatalf("lerp low clamp = %v", got)
	}
	if got := lerp(cell.ColorNumber(250), cell.Color(400), 0.9); got != cell.ColorNumber(255) {
		t.Fatalf("lerp high clamp = %v", got)
	}
	if got := hue256(420); got == cell.ColorDefault {
		t.Fatal("hue256 returned default color")
	}
	for _, deg := range []float64{0, 60, 120, 180, 240, 300} {
		if got := hue256(deg); got == cell.ColorDefault {
			t.Fatalf("hue256(%v) returned default color", deg)
		}
	}
	if got := abs(-5); got != 5 {
		t.Fatalf("abs(-5) = %d", got)
	}
	if got := max(3, 9); got != 9 {
		t.Fatalf("max(3, 9) = %d", got)
	}
	if got := max(9, 3); got != 9 {
		t.Fatalf("max(9, 3) = %d", got)
	}
	if got := pulseColor(3, cell.ColorNumber(51), cell.ColorNumber(236), 12); got == cell.ColorDefault {
		t.Fatal("pulseColor returned default color")
	}
}

func TestParkedIndices(t *testing.T) {
	if got := parkedIndices(0, parkTopBrackets); got != nil {
		t.Fatalf("parkedIndices(0) = %v, want nil", got)
	}
	if got := len(parkedIndices(32, parkSideBraces)); got != 2 {
		t.Fatalf("side braces markers = %d, want 2", got)
	}
	if got := len(parkedIndices(32, parkTopCenter)); got != 4 {
		t.Fatalf("top center markers = %d, want 4", got)
	}
	if got := len(parkedIndices(32, parkBottomCenter)); got != 3 {
		t.Fatalf("bottom center markers = %d, want 3", got)
	}
	if got := parkedIndices(32, parkCorners); got != nil {
		t.Fatalf("corner markers = %v, want nil", got)
	}
}

func sampleBorderCells(border image.Rectangle) []container.BorderCell {
	width := border.Dx()
	height := border.Dy()
	if width < 2 || height < 2 {
		return nil
	}

	length := 2*width + 2*height - 4
	var cells []container.BorderCell
	index := 0

	appendCell := func(x, y int, r rune) {
		cells = append(cells, container.BorderCell{
			Point:  image.Point{X: x, Y: y},
			Border: border,
			Rune:   r,
			Index:  index,
			Length: length,
		})
		index++
	}

	appendCell(border.Min.X, border.Min.Y, '╭')
	for x := border.Min.X + 1; x < border.Max.X-1; x++ {
		appendCell(x, border.Min.Y, '─')
	}
	appendCell(border.Max.X-1, border.Min.Y, '╮')
	for y := border.Min.Y + 1; y < border.Max.Y-1; y++ {
		appendCell(border.Max.X-1, y, '│')
	}
	appendCell(border.Max.X-1, border.Max.Y-1, '╯')
	for x := border.Max.X - 2; x > border.Min.X; x-- {
		appendCell(x, border.Max.Y-1, '─')
	}
	appendCell(border.Min.X, border.Max.Y-1, '╰')
	for y := border.Max.Y - 2; y > border.Min.Y; y-- {
		appendCell(border.Min.X, y, '│')
	}

	return cells
}

func TestZeroEffectHelpers(t *testing.T) {
	e := newEffect(nil)
	if got := e.nextStyler()(container.BorderCell{}); got.Rune != 0 || got.CellOpts != nil {
		t.Fatalf("zero styler = %+v, want empty style", got)
	}
	if got := e.Next(); got != cell.ColorDefault {
		t.Fatalf("zero effect color = %v, want default", got)
	}
}

func TestLoadingBackgroundDrawInterlacesRows(t *testing.T) {
	ft := faketerm.MustNew(image.Point{X: 10, Y: 4})
	background := NewLoadingBackground()

	if err := background.Draw(ft, image.Rect(0, 0, 10, 4), "BOOT\nSYNC"); err != nil {
		t.Fatalf("LoadingBackground.Draw => unexpected error: %v", err)
	}

	buffer := ft.BackBuffer()
	if got, want := buffer[0][0].Opts.BgColor, cell.ColorNumber(236); got != want {
		t.Fatalf("row 0 background = %v, want %v", got, want)
	}
	if got, want := buffer[0][1].Opts.BgColor, cell.ColorNumber(239); got != want {
		t.Fatalf("row 1 background = %v, want %v", got, want)
	}
	if got, want := buffer[0][0].Opts.FgColor, cell.ColorNumber(117); got != want {
		t.Fatalf("row 0 text color = %v, want %v", got, want)
	}
	if got := buffer[0][0].Rune; got != 'B' {
		t.Fatalf("row 0 first rune = %q, want %q", got, 'B')
	}
	if got := buffer[0][1].Rune; got != 'S' {
		t.Fatalf("row 1 first rune = %q, want %q", got, 'S')
	}
}
