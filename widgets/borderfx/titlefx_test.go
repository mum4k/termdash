package borderfx

import (
	"context"
	"testing"
	"time"

	spin "github.com/mum4k/termdash/widgets/spinner"
)

func TestTitleSpecHelpers(t *testing.T) {
	spec := TitleSpec{
		Base:      " Warp Core Flux ",
		Charset:   "01$#*+-",
		LeftSpin:  spin.Must("star"),
		RightSpin: spin.Must("star"),
	}

	if got := spec.Plain(); got != " Warp Core Flux " {
		t.Fatalf("Plain() = %q", got)
	}
	if got := spec.Decorated(0); got != "✶ Warp Core Flux ✶" {
		t.Fatalf("Decorated(0) = %q", got)
	}
	if got := spec.Decorated(3); got != "✺ Warp Core Flux ✺" {
		t.Fatalf("Decorated(3) = %q", got)
	}
	if got := spec.Scrambled(0); got == spec.Base {
		t.Fatalf("Scrambled(0) = %q, want scrambled title", got)
	}
	if got := spec.Scrambled(100); got != spec.Base {
		t.Fatalf("Scrambled(100) = %q, want base title", got)
	}
}

func TestDecryptingTitleHelper(t *testing.T) {
	spec := DecryptingTitle(" LCARS Telemetry ", DecryptCharsets.LCARSTelemetry).
		WithRightSpinner(spin.Must("dots_10"))

	if got := spec.Base; got != " LCARS Telemetry " {
		t.Fatalf("Base = %q, want %q", got, " LCARS Telemetry ")
	}
	if got := spec.Charset; got != DecryptCharsets.LCARSTelemetry {
		t.Fatalf("Charset = %q, want %q", got, DecryptCharsets.LCARSTelemetry)
	}
	if got := spec.Decorated(0); got == spec.Base {
		t.Fatalf("Decorated(0) = %q, want right spinner decoration", got)
	}
	if got := spec.Scrambled(0); got == spec.Base {
		t.Fatalf("Scrambled(0) = %q, want decrypting title behavior", got)
	}
}

func TestDecryptCharsetGroups(t *testing.T) {
	if got := DecryptCharsets.WarpCoreFlux; got != "01$#*+-" {
		t.Fatalf("DecryptCharsets.WarpCoreFlux = %q", got)
	}
	if got := DecryptCharsets.LCARSTelemetry; got != "><=|/-" {
		t.Fatalf("DecryptCharsets.LCARSTelemetry = %q", got)
	}
	if got := DecryptCharsets.Comms; got != "[]{}:;?" {
		t.Fatalf("DecryptCharsets.Comms = %q", got)
	}
	if got := DecryptCharsets.Controls; got != ".,:*~^" {
		t.Fatalf("DecryptCharsets.Controls = %q", got)
	}
	if got := DecryptCharsets.Default; got != "⠿⣾⣶⣤⣀⠒" {
		t.Fatalf("DecryptCharsets.Default = %q", got)
	}
	if got := DecryptCharsets.Sensors; got != DecryptCharsets.Default {
		t.Fatalf("DecryptCharsets.Sensors = %q, want %q", got, DecryptCharsets.Default)
	}
	if TitleCharsets != DecryptCharsets {
		t.Fatal("TitleCharsets compatibility alias drifted from DecryptCharsets")
	}
}

func TestTitleSpecLeftSpinnerAndEmptyCharset(t *testing.T) {
	spec := TitleSpec{
		Base:             " LCARS ",
		Spinner:          spin.Must("pulse"),
		SpinnerPlacement: TitleSpinnerLeft,
	}

	if got := spec.Decorated(3); got != "⎽ LCARS " {
		t.Fatalf("Decorated(3) = %q", got)
	}
	if got := spec.Scrambled(0); got != spec.Base {
		t.Fatalf("Scrambled(0) = %q, want base title when charset is empty", got)
	}
}

func TestTitleSpecLegacySpinnerPlacementStillWorks(t *testing.T) {
	spec := TitleSpec{
		Base:             " Comms ",
		Spinner:          spin.Must("spin_2"),
		SpinnerPlacement: TitleSpinnerRight,
	}

	if got := spec.Decorated(2); got != " Comms ◑" {
		t.Fatalf("Decorated(2) = %q", got)
	}
}

func TestTitleControllerRunHandlesNilContainer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tc := NewTitleController(nil, nil, nil, nil)
	tc.Run(ctx, nil)
}

func TestNewTitleControllerDefaults(t *testing.T) {
	tc := NewTitleController(nil, nil, nil, nil)
	if got := tc.pollInterval; got != 35*time.Millisecond {
		t.Fatalf("pollInterval = %v, want 35ms", got)
	}
	if got := tc.revealDelay; got != 38*time.Millisecond {
		t.Fatalf("revealDelay = %v, want 38ms", got)
	}
}
