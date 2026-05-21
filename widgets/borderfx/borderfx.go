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

// Package borderfx provides animated border effects for Termdash containers.
//
// # Tier 1 — Profiles (simplest)
//
// Profiles are fully pre-configured effects: colors, tick rate, and inactive
// style all baked in.  One call wires everything:
//
//	fx := borderfx.NewAnimator(cont)
//	fx.ApplyProfile(borderfx.Profiles.GradientArc, "root", "sidebar", "chart")
//	go fx.Run(ctx)
//
// Built-in profiles:
//
//	GradientArc        – purple→lavender→blue arc sweeping clockwise
//	LoadingSweepWhite  – white highlight scanning across the border title
//	FuturisticSweep    – braille interlaced scanner in bright cyan
//	NeonPulse          – magenta neon-sign flicker
//	AmberTelemetry     – amber tick telemetry (professional ops look)
//
// # Tier 1 — Loading overlay (simplest)
//
// ApplyInterlacedLoadingContent shows the interlaced boot-screen from the
// borderfx demo with one call, and hides it when your app is ready:
//
//	lo := borderfx.WrapWithLoading(t, func(size image.Point, id string) image.Rectangle {
//	    return panelRects(size)[id].Inset(1) // inner content area
//	})
//	lo.SetContent("sensors", borderfx.LoadingText.BootSequence)
//	lo.SetContent("data",    borderfx.LoadingText.DataSync)
//
//	fx := borderfx.NewAnimator(cont)
//	fx.ApplyInterlacedLoadingContent(lo, "sensors", "data")
//
//	cont, _ = container.New(lo, ...)  // pass lo in place of the raw terminal
//	go fx.Run(ctx)
//	// ... when boot completes:
//	lo.Hide()
//
// # Tier 2 — Macro + Palette (intermediate)
//
// For fine-grained color control, combine a Macro with a Palette:
//
//	fx.RegisterMacro("sensors", borderfx.Presets.Orbit, borderfx.Palettes.Cyan)
//
// # Tier 3 — Raw Effect (advanced)
//
// For maximum flexibility, build and register an Effect directly:
//
//	fx.Register("sensors", borderfx.GradientArcN(stops, dim, 0.45))
package borderfx

import (
	"context"
	"image"
	"math"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/private/runewidth"
	spinners "github.com/mum4k/termdash/widgets/spinner"
)

// Palette groups the colors used by the higher-level borderfx presets.
type Palette struct {
	Bright cell.Color
	Mid    cell.Color
	Dim    cell.Color
}

// Colors builds a palette from explicit bright, mid, and dim colors.
func Colors(bright, mid, dim cell.Color) Palette {
	return Palette{
		Bright: bright,
		Mid:    mid,
		Dim:    dim,
	}
}

// Duo builds a palette from a bright accent and dim resting color.
// The mid tone is derived automatically.
func Duo(bright, dim cell.Color) Palette {
	return Palette{
		Bright: bright,
		Mid:    lerpRGB(dim, bright, 0.58),
		Dim:    dim,
	}
}

// Apply renders the supplied macro using this palette.
func (p Palette) Apply(m Macro) *Effect {
	return m.With(p)
}

// Macro is a reusable, high-level border animation preset.
type Macro struct {
	name  string
	build func(Palette) *Effect
}

func newMacro(name string, build func(Palette) *Effect) Macro {
	return Macro{name: name, build: build}
}

// Name returns the stable preset name.
func (m Macro) Name() string {
	return m.name
}

// With renders the macro with the supplied palette.
func (m Macro) With(p Palette) *Effect {
	if m.build == nil {
		return nil
	}
	return m.build(p)
}

// Register applies the macro to an animator in one call.
func (m Macro) Register(a *Animator, id string, p Palette) {
	if a == nil {
		return
	}
	a.Register(id, m.With(p))
}

// Presets exposes the package's high-level reusable animation presets.
var Presets = struct {
	Scanner     Macro
	Dual        Macro
	Interlace   Macro
	Braided     Macro
	Shard       Macro
	Orbit       Macro
	Focus       Macro
	Rail        Macro
	Decode      Macro
	Power       Macro
	Rain        Macro
	Braille     Macro
	Bracket     Macro
	Ticks       Macro
	Noise       Macro
	SpinPulse   Macro
	Dots6       Macro
	Dots10      Macro
	Ribbon      Macro
	Brace       Macro
	Emoji       Macro
	Pulse       Macro
	Fire        Macro
	Ice         Macro
	Rainbow     Macro
	Neon        Macro
	Matrix      Macro
	Glow        Macro
	Warp        Macro
	TextSweep   Macro
	GradArc     Macro // single gradient arc sweeping clockwise; uses Palette.Bright→Mid, dim background
	DualGradArc Macro // two opposing gradient arcs; uses Palette.Bright vs Palette.Mid
}{
	Scanner: newMacro("scanner", func(p Palette) *Effect {
		return Scanner(p.Bright, p.Dim)
	}),
	Dual: newMacro("dual", func(p Palette) *Effect {
		return DualScanner(p.Bright, p.Mid, p.Dim)
	}),
	Interlace: newMacro("interlace", func(p Palette) *Effect {
		return InterlacedScanner(p.Bright, p.Mid, p.Dim)
	}),
	Braided: newMacro("braided", func(p Palette) *Effect {
		return BraidedScanner(p.Bright, p.Mid, p.Dim)
	}),
	Shard: newMacro("shard", func(p Palette) *Effect {
		return ShardScanner(p.Bright, p.Mid, p.Dim)
	}),
	Orbit: newMacro("orbit", func(p Palette) *Effect {
		return OrbitScanner(p.Bright, p.Mid, p.Dim)
	}),
	Focus: newMacro("focus", func(p Palette) *Effect {
		return FocusPins(p.Bright, p.Mid, p.Dim)
	}),
	Rail: newMacro("rail", func(p Palette) *Effect {
		return FocusPinsRail(p.Bright, p.Mid, p.Dim)
	}),
	Decode: newMacro("decode", func(p Palette) *Effect {
		return FocusPinsMatrix(p.Bright, p.Mid, p.Dim)
	}),
	Power: newMacro("power", func(p Palette) *Effect {
		return FocusPinsPower(p.Bright, p.Mid, p.Dim)
	}),
	Rain: newMacro("rain", func(p Palette) *Effect {
		return RainScanner(p.Bright, p.Mid, p.Dim)
	}),
	Braille: newMacro("braille", func(p Palette) *Effect {
		return BrailleDrift(p.Bright, p.Mid, p.Dim)
	}),
	Bracket: newMacro("bracket", func(p Palette) *Effect {
		return BracketScan(p.Bright, p.Mid, p.Dim)
	}),
	Ticks: newMacro("ticks", func(p Palette) *Effect {
		return DataTicks(p.Bright, p.Mid, p.Dim)
	}),
	Noise: newMacro("noise", func(p Palette) *Effect {
		return StaticNoise(p.Bright, p.Mid, p.Dim)
	}),
	SpinPulse: newMacro("spin_pulse", func(p Palette) *Effect {
		return SpinnerPulse(p.Bright, p.Mid, p.Dim)
	}),
	Dots6: newMacro("dots6", func(p Palette) *Effect {
		return Dots6Spinner(p.Bright, p.Mid, p.Dim)
	}),
	Dots10: newMacro("dots10", func(p Palette) *Effect {
		return Dots10Spinner(p.Bright, p.Mid, p.Dim)
	}),
	Ribbon: newMacro("ribbon", func(p Palette) *Effect {
		return FocusPinsRibbon(p.Bright, p.Mid, p.Dim)
	}),
	Brace: newMacro("brace", func(p Palette) *Effect {
		return FocusPinsBrace(p.Bright, p.Mid, p.Dim)
	}),
	Emoji: newMacro("emoji", func(p Palette) *Effect {
		return FocusPinsEmoji(p.Bright, p.Mid, p.Dim)
	}),
	Pulse: newMacro("pulse", func(p Palette) *Effect {
		return FocusPinsPulse(p.Bright, p.Mid, p.Dim)
	}),
	Fire: newMacro("fire", func(Palette) *Effect {
		return Fire()
	}),
	Ice: newMacro("ice", func(Palette) *Effect {
		return Ice()
	}),
	Rainbow: newMacro("rainbow", func(Palette) *Effect {
		return Rainbow()
	}),
	Neon: newMacro("neon", func(p Palette) *Effect {
		return Neon(p.Bright)
	}),
	Matrix: newMacro("matrix", func(Palette) *Effect {
		return Matrix()
	}),
	Glow: newMacro("glow", func(p Palette) *Effect {
		return Glow(p.Bright)
	}),
	Warp: newMacro("warp", func(Palette) *Effect {
		return Warp()
	}),
	TextSweep: newMacro("text_sweep", func(p Palette) *Effect {
		return TextSweep(cell.ColorWhite, cell.ColorNumber(245))
	}),
	GradArc: newMacro("grad_arc", func(p Palette) *Effect {
		return GradientArc(p.Bright, p.Mid, p.Dim, 0.35)
	}),
	DualGradArc: newMacro("dual_grad_arc", func(p Palette) *Effect {
		return DualGradientArc(p.Bright, p.Mid, p.Dim, 0.35)
	}),
}

// Palettes exposes a few ready-made color palettes for common borderfx looks.
var Palettes = struct {
	Cyan      Palette
	Amber     Palette
	Matrix    Palette
	Synthwave Palette
	Ice       Palette
	Silver    Palette
}{
	Cyan:      Colors(cell.ColorNumber(51), cell.ColorNumber(24), cell.ColorNumber(236)),
	Amber:     Colors(cell.ColorNumber(214), cell.ColorNumber(136), cell.ColorNumber(236)),
	Matrix:    Colors(cell.ColorNumber(123), cell.ColorNumber(39), cell.ColorNumber(236)),
	Synthwave: Colors(cell.ColorNumber(201), cell.ColorNumber(90), cell.ColorNumber(236)),
	Ice:       Colors(cell.ColorNumber(220), cell.ColorNumber(178), cell.ColorNumber(236)),
	Silver:    Duo(cell.ColorNumber(250), cell.ColorNumber(239)),
}

// ── Profiles ──────────────────────────────────────────────────────────────────

// Profile is a fully pre-configured border effect.  Colors, animation speed,
// and the recommended inactive-panel style are all baked in — no palette
// selection required.
//
// Quickstart:
//
//	fx := borderfx.NewAnimator(cont)
//	fx.ApplyProfile(borderfx.Profiles.GradientArc, "root", "panel1", "panel2")
//	go fx.Run(ctx)
type Profile struct {
	name          string
	description   string
	tickRate      time.Duration
	effectFn      func() *Effect
	inactiveColor cell.Color
}

// Name returns the profile's stable identifier.
func (p Profile) Name() string { return p.name }

// Description returns a human-readable summary of the visual effect.
func (p Profile) Description() string { return p.description }

// TickRate returns the recommended animation tick interval.
func (p Profile) TickRate() time.Duration { return p.tickRate }

// InactiveColor returns the color used to style unfocused panels.
func (p Profile) InactiveColor() cell.Color { return p.inactiveColor }

// New returns a fresh *Effect instance for this profile.
// Each panel must get its own instance so animation phases are independent.
func (p Profile) New() *Effect { return p.effectFn() }

// Profiles contains named, fully pre-configured border effect profiles.
// Each profile bundles an effect, color scheme, tick rate, and inactive style
// into a single value you can apply in one call.
//
// All five built-in profiles are described below.  For the full catalogue of
// lower-level effects and palette primitives see Presets and Palettes.
//
//	Profile             Visual character
//	──────────────────  ─────────────────────────────────────────────────────
//	GradientArc         Purple→lavender→blue→purple arc sweeping clockwise
//	LoadingSweepWhite   White highlight scanning across the border title text
//	FuturisticSweep     Braille-interlaced multi-lane scanner in bright cyan
//	NeonPulse           Magenta neon-sign flicker with occasional dark frames
//	AmberTelemetry      Amber measured-tick telemetry sweep (ops-dashboard look)
var Profiles = struct {
	// GradientArc sweeps a purple→lavender→blue→purple gradient arc clockwise.
	// The focused panel animates; inactive panels show a warm-gold static border.
	//
	// Origin: extracted from the ops-dashboard timeline demo (timelinedemo.go).
	//
	// Recommended for: general-purpose dashboards, ops panels, any widget where
	// you want a premium animated border with a clear focus indicator.
	GradientArc Profile

	// LoadingSweepWhite sweeps a bright-white highlight across the border title
	// text, leaving non-title cells in near-black.  Creates a clean "loading" or
	// "scanning" feel.
	//
	// Origin: extracted from the ThreeD tab of the tabdemo (tabdemo.go).
	//
	// Recommended for: panels that boot/load data, 3-D / render stages, any
	// widget whose title text deserves a spotlight moment.
	LoadingSweepWhite Profile

	// FuturisticSweep runs a multi-lane braille interlaced scanner in bright cyan.
	// Dense and high-energy — the border looks like live signal telemetry.
	//
	// Origin: extracted from the Spectrum Analyzer pane of the tabdemo.
	//
	// Recommended for: signal visualization, sensor feeds, spectrum analyzers,
	// any high-frequency data panel.
	FuturisticSweep Profile

	// NeonPulse flickers a magenta neon-sign effect around the full border.
	// Occasional dark frames mimic the gas-discharge instability of a real tube.
	//
	// Recommended for: alert panels, critical-status widgets, notification areas,
	// anything that should draw the eye immediately.
	NeonPulse Profile

	// AmberTelemetry sweeps measured block ticks in amber/gold across the border.
	// Clean and readable — a professional ops look inspired by hardware consoles.
	//
	// Recommended for: production dashboards, metrics panels, timeline widgets,
	// anywhere you want animation that doesn't distract from the data.
	AmberTelemetry Profile
}{
	GradientArc: Profile{
		name:        "gradient_arc",
		description: "Purple→lavender→blue→purple gradient arc sweeping clockwise. Inactive: warm gold.",
		tickRate:    32 * time.Millisecond,
		effectFn: func() *Effect {
			// xterm-256 cube path: 129(3,0,5) vivid purple → 147(3,3,5) lavender →
			// 93(2,0,5) blue-violet → 57(1,0,5) deep blue → 129 closes the cycle.
			// dim=54(1,0,2) dark indigo keeps the trail visually grounded.
			return GradientArcN([]cell.Color{
				cell.ColorNumber(129), // (3,0,5) RGB(175,  0,255) vivid purple   HEAD
				cell.ColorNumber(147), // (3,3,5) RGB(175,175,255) lavender
				cell.ColorNumber(93),  // (2,0,5) RGB(135,  0,255) blue-violet
				cell.ColorNumber(57),  // (1,0,5) RGB( 95,  0,255) deep blue
				cell.ColorNumber(129), // vivid purple again                      TAIL
			}, cell.ColorNumber(54), 0.50)
		},
		inactiveColor: cell.ColorNumber(178), // (4,3,0) RGB(215,175,0) warm gold
	},

	LoadingSweepWhite: Profile{
		name:        "loading_sweep_white",
		description: "White highlight sweeps across the border title; dim grey elsewhere. Clean loading feel.",
		tickRate:    64 * time.Millisecond,
		effectFn: func() *Effect {
			return TextSweep(cell.ColorWhite, cell.ColorNumber(236))
		},
		inactiveColor: cell.ColorNumber(240), // mid grey
	},

	FuturisticSweep: Profile{
		name:        "futuristic_sweep",
		description: "Multi-lane braille interlaced scanner in bright cyan. Dense sci-fi aesthetic.",
		tickRate:    64 * time.Millisecond,
		effectFn: func() *Effect {
			return InterlacedScanner(
				cell.ColorNumber(75),  // bright cyan-blue
				cell.ColorNumber(243), // mid grey
				cell.ColorNumber(236), // near-black dim
			)
		},
		inactiveColor: cell.ColorNumber(240), // mid grey
	},

	NeonPulse: Profile{
		name:        "neon_pulse",
		description: "Magenta neon-sign flicker. Occasional dark frames mimic a real neon tube.",
		tickRate:    45 * time.Millisecond,
		effectFn: func() *Effect {
			return Neon(cell.ColorNumber(201)) // hot magenta #d700ff
		},
		inactiveColor: cell.ColorNumber(54), // dark indigo — keeps the moody purple theme at rest
	},

	AmberTelemetry: Profile{
		name:        "amber_telemetry",
		description: "Amber measured-tick telemetry sweep. Clean and professional for ops dashboards.",
		tickRate:    64 * time.Millisecond,
		effectFn: func() *Effect {
			return DataTicks(
				cell.ColorNumber(214), // amber bright
				cell.ColorNumber(136), // amber mid
				cell.ColorNumber(236), // near-black dim
			)
		},
		inactiveColor: cell.ColorNumber(94), // dark amber/brown
	},
}

// Animator drives border animations on containers by ID.
type Animator struct {
	mu            sync.RWMutex
	root          *container.Container
	effects       map[string]*Effect
	tickRate      time.Duration
	alwaysActive  bool // when true all registered panels animate regardless of focus
	inactiveStyle func(id string, bc container.BorderCell) container.BorderCellStyle
}

// NewAnimator creates an animator bound to the root container.
func NewAnimator(root *container.Container) *Animator {
	return &Animator{
		root:     root,
		effects:  make(map[string]*Effect),
		tickRate: 45 * time.Millisecond,
	}
}

// SetTickRate changes animation speed.
func (a *Animator) SetTickRate(d time.Duration) {
	a.mu.Lock()
	a.tickRate = d
	a.mu.Unlock()
}

// Register assigns an effect to a container ID. Replaces any existing effect.
func (a *Animator) Register(id string, e *Effect) {
	a.mu.Lock()
	a.effects[id] = e
	a.mu.Unlock()
}

// Unregister removes animation from a container ID.
func (a *Animator) Unregister(id string) {
	a.mu.Lock()
	delete(a.effects, id)
	a.mu.Unlock()
}

// RegisterMacro applies a high-level macro preset to a container ID.
func (a *Animator) RegisterMacro(id string, m Macro, p Palette) {
	a.Register(id, m.With(p))
}

// ApplyProfile configures the animator from a pre-baked Profile in one call.
// It sets the tick rate to the profile's recommended value, registers a fresh
// Effect instance for each supplied container ID, and wires up the inactive
// style so unfocused panels show the profile's recommended resting color.
//
// Each ID receives its own independent Effect so animation phases never lock
// together — a panel's arc position is not shared with its neighbours.
//
// Example:
//
//	fx := borderfx.NewAnimator(cont)
//	fx.ApplyProfile(borderfx.Profiles.GradientArc, "root", "sidebar", "chart")
//	go fx.Run(ctx)
func (a *Animator) ApplyProfile(p Profile, ids ...string) {
	a.SetTickRate(p.TickRate())
	for _, id := range ids {
		a.Register(id, p.New())
	}
	inactive := p.InactiveColor()
	a.SetInactiveStyle(func(_ string, _ container.BorderCell) container.BorderCellStyle {
		return container.BorderCellStyle{
			CellOpts: []cell.Option{cell.FgColor(inactive)},
		}
	})
}

// ApplyInterlacedLoadingContent wires the FuturisticSweep border profile to
// each of the given container IDs and activates the interlaced loading overlay.
//
// This is the single call that makes a panel look like the borderfxdemo boot
// screen — animated braille-interlaced cyan borders with a striped background
// and loading copy text inside the panel content area.
//
// Typical usage:
//
//	t, _ := tcell.New()
//	lo := borderfx.WrapWithLoading(t, func(size image.Point, id string) image.Rectangle {
//	    return panelBorderRects(size)[id].Inset(1)
//	})
//	lo.SetContent("sensors", borderfx.LoadingText.BootSequence)
//	lo.SetContent("data",    borderfx.LoadingText.DataSync)
//
//	fx := borderfx.NewAnimator(cont)
//	fx.ApplyInterlacedLoadingContent(lo, "sensors", "data", "sidebar")
//
//	cont, _ = container.New(lo, ...)  // pass lo as the terminal
//	go fx.Run(ctx)
//	go termdash.Run(ctx, lo, cont, ...)
//
//	// When your app finishes loading:
//	lo.Hide()
func (a *Animator) ApplyInterlacedLoadingContent(lo *LoadingOverlay, ids ...string) {
	a.ApplyProfile(Profiles.FuturisticSweep, ids...)
	lo.Show()
}

// SetInactiveStyle sets a styler applied to unfocused registered windows.
func (a *Animator) SetInactiveStyle(styler func(id string, bc container.BorderCell) container.BorderCellStyle) {
	a.mu.Lock()
	a.inactiveStyle = styler
	a.mu.Unlock()
}

// SetAlwaysActive makes every registered effect animate at full brightness on
// every tick, regardless of which container is focused.  When false (default),
// only the focused container's effect plays; others receive the inactive style
// or are cleared.
func (a *Animator) SetAlwaysActive(v bool) {
	a.mu.Lock()
	a.alwaysActive = v
	a.mu.Unlock()
}

// Run starts the animation loop. Blocks until ctx is done.
func (a *Animator) Run(ctx context.Context) error {
	for {
		a.mu.RLock()
		rate := a.tickRate
		a.mu.RUnlock()
		ticker := time.NewTicker(rate)

		select {
		case <-ctx.Done():
			ticker.Stop()
			return ctx.Err()
		case <-ticker.C:
			ticker.Stop()
			a.tick()
		}
	}
}

func (a *Animator) tick() {
	a.mu.RLock()
	snap := make(map[string]*Effect, len(a.effects))
	for k, v := range a.effects {
		snap[k] = v
	}
	root := a.root
	inactiveStyle := a.inactiveStyle
	alwaysActive := a.alwaysActive
	a.mu.RUnlock()
	if root == nil {
		return
	}
	activeID := root.ActiveID()
	for id, e := range snap {
		if alwaysActive || id == activeID {
			_ = root.Update(id, container.BorderCellStyleFunc(e.nextStyler()))
			continue
		}
		if inactiveStyle == nil {
			_ = root.Update(id, container.BorderCellStyleFunc(nil))
			continue
		}
		_ = root.Update(id, container.BorderCellStyleFunc(func(bc container.BorderCell) container.BorderCellStyle {
			return inactiveStyle(id, bc)
		}))
	}
}

// Effect holds one animated border treatment.
type Effect struct {
	mu      sync.Mutex
	frame   int
	styler  func(frame int, bc container.BorderCell) container.BorderCellStyle
	colors  []cell.Color
	colorAt func(frame int) cell.Color
}

func newEffect(styler func(frame int, bc container.BorderCell) container.BorderCellStyle) *Effect {
	return &Effect{styler: styler}
}

func (e *Effect) nextStyler() container.BorderCellStyler {
	e.mu.Lock()
	frame := e.frame
	e.frame++
	styler := e.styler
	e.mu.Unlock()

	return func(bc container.BorderCell) container.BorderCellStyle {
		if styler == nil {
			return container.BorderCellStyle{}
		}
		return styler(frame, bc)
	}
}

// Next advances one frame and returns the effect's primary color.
func (e *Effect) Next() cell.Color {
	e.mu.Lock()
	defer e.mu.Unlock()
	frame := e.frame
	e.frame++
	if e.colorAt != nil {
		return e.colorAt(frame)
	}
	if len(e.colors) == 0 {
		return cell.ColorDefault
	}
	return e.colors[frame%len(e.colors)]
}

// Reset restarts from frame 0.
func (e *Effect) Reset() {
	e.mu.Lock()
	e.frame = 0
	e.mu.Unlock()
}

// Scanner simulates a scanning beam traveling clockwise around the full border.
// Top and bottom edges therefore move in opposite visual directions.
func Scanner(bright, dim cell.Color) *Effect {
	const speed = 1
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		return beamStyle(bc, (frame*speed)%max(1, bc.Length), bright, dim)
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 42)
	}
	return e
}

// DualScanner runs two beams in opposite directions for a denser sci-fi trace.
func DualScanner(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		length := max(1, bc.Length)
		a := beamStyle(bc, (frame*2)%length, bright, dim)
		b := beamStyle(bc, (length-1-(frame*3+length/3)%length+length)%length, mid, dim)
		if beamRank(a.Rune) >= beamRank(b.Rune) {
			return a
		}
		return b
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 64)
	}
	return e
}

// InterlacedScanner runs a denser multi-lane tracer with a braille shimmer tail.
func InterlacedScanner(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		return scrollingInterlaceStyle(frame, bc, bright, mid, dim)
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 56)
	}
	return e
}

// BraidedScanner runs two luminous tracers with a softer, more segmented tail.
func BraidedScanner(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		length := max(1, bc.Length)
		headA := (frame*2 + length/9) % length
		headB := (length - 1 - (frame*2+length/3)%length + length) % length

		a := braidedStyle(bc, headA, bright, mid, dim, frame, 0)
		b := braidedStyle(bc, headB, mid, bright, dim, frame, 1)

		if interlaceRank(b) > interlaceRank(a) {
			return b
		}
		return a
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 48)
	}
	return e
}

// ShardScanner runs a sharper segmented scanner with longer travel streaks.
func ShardScanner(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		length := max(1, bc.Length)
		headA := (frame * 3) % length
		headB := (frame*2 + length/2) % length

		a := shardStyle(bc, headA, bright, mid, dim, frame, 0)
		b := shardStyle(bc, headB, mid, bright, dim, frame, 1)

		if interlaceRank(b) > interlaceRank(a) {
			return b
		}
		return a
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 52)
	}
	return e
}

// OrbitScanner runs a denser, tri-lane scanner for high-energy panels.
func OrbitScanner(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		length := max(1, bc.Length)
		headA := (frame*2 + length/7) % length
		headB := (frame*4 + length/2) % length
		headC := (length - 1 - (frame*3+length/4)%length + length) % length

		a := orbitStyle(bc, headA, bright, mid, dim, frame, 0)
		b := orbitStyle(bc, headB, lerpRGB(mid, bright, 0.75), mid, dim, frame, 1)
		c := orbitStyle(bc, headC, mid, bright, dim, frame, 2)

		best := a
		if interlaceRank(b) > interlaceRank(best) {
			best = b
		}
		if interlaceRank(c) > interlaceRank(best) {
			best = c
		}
		return best
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 44)
	}
	return e
}

// FocusRail performs a brief dual-rail activation sweep, then parks as top-edge markers.
func FocusRail(bright, mid, dim cell.Color) *Effect {
	return stagedFocusEffect(bright, mid, dim, focusProfile{
		moveA:     2,
		moveB:     3,
		offsetB:   9,
		trail:     4,
		heads:     []rune{'▰', '▰', '▣', '▣'},
		tail:      []rune{'▪', '▫', '▪', '▫'},
		parked:    []rune{'▰', '▰', '▰'},
		parkMode:  parkTopBrackets,
		moveFrame: 24,
	})
}

// FocusBrace performs a vertical brace sweep, then parks on the side rails.
func FocusBrace(bright, mid, dim cell.Color) *Effect {
	return stagedFocusEffect(bright, mid, dim, focusProfile{
		moveA:     1,
		moveB:     2,
		offsetB:   17,
		trail:     5,
		heads:     []rune{'▐', '▌', '▐', '▌'},
		tail:      []rune{'•', '·', '•', '·'},
		parked:    []rune{'▐', '▌', '▐', '▌'},
		parkMode:  parkSideBraces,
		moveFrame: 26,
	})
}

// FocusMatrix performs a decode-style sweep, then parks as matrix glyph locks.
func FocusMatrix(bright, mid, dim cell.Color) *Effect {
	return stagedFocusEffect(bright, mid, dim, focusProfile{
		moveA:     2,
		moveB:     4,
		offsetB:   13,
		trail:     5,
		heads:     []rune{'0', '1', '▓', '▒'},
		tail:      []rune{'1', '0', '·', ':'},
		parked:    []rune{'[', '0', '1', ']'},
		parkMode:  parkTopCenter,
		moveFrame: 22,
	})
}

// FocusRibbon performs a braided sweep, then parks as a centered ribbon.
func FocusRibbon(bright, mid, dim cell.Color) *Effect {
	return stagedFocusEffect(bright, mid, dim, focusProfile{
		moveA:     3,
		moveB:     2,
		offsetB:   11,
		trail:     4,
		heads:     []rune{'◆', '◈', '◆', '◈'},
		tail:      []rune{'•', '∙', '•', '∙'},
		parked:    []rune{'◈', '◆', '◈'},
		parkMode:  parkBottomCenter,
		moveFrame: 24,
	})
}

// FocusPins performs a crisp activation pass, then parks as corner pins.
func FocusPins(bright, mid, dim cell.Color) *Effect {
	return focusPinsVariant(bright, mid, dim, pinVariant{
		horizontalFrames: 34,
		verticalFrames:   18,
		holdFrames:       22,
		lag:              4,
		topHead:          '▣',
		bottomHead:       '▣',
		sideHead:         '│',
		cornerRune:       '◉',
		cornerTrail:      '●',
	})
}

// FocusPinsRail uses heavier rail blocks during activation before pinning corners.
func FocusPinsRail(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		return splitOrbitRailStyle(frame, bc, bright, mid, dim)
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 64)
	}
	return e
}

// FocusPinsMatrix uses denser decode-like blocks during activation.
func FocusPinsMatrix(bright, mid, dim cell.Color) *Effect {
	return focusPinsVariant(bright, mid, dim, pinVariant{
		horizontalFrames: 46,
		verticalFrames:   24,
		holdFrames:       18,
		lag:              2,
		topHead:          '▓',
		bottomHead:       '░',
		sideHead:         '║',
		cornerRune:       '◉',
		cornerTrail:      '●',
	})
}

// FocusPinsPower uses electrical glyphs and a sine-wave sweep before pinning corners.
func FocusPinsPower(bright, mid, dim cell.Color) *Effect {
	return focusPinsVariant(bright, mid, dim, pinVariant{
		horizontalFrames: 56,
		verticalFrames:   24,
		holdFrames:       28,
		lag:              5,
		topHead:          '⌁',
		bottomHead:       '≋',
		sideHead:         'Ϟ',
		cornerRune:       'Ϟ',
		cornerTrail:      '•',
	})
}

// RainScanner runs a restrained telemetry shimmer with sparse side markers and a
// faint bottom accent.
func RainScanner(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		return rainBorderStyle(frame, bc, bright, mid, dim)
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 72)
	}
	return e
}

// BrailleDrift runs a continuous shaded-block spinner around the border.
func BrailleDrift(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		return brailleDriftStyle(frame, bc, bright, mid, dim)
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 76)
	}
	return e
}

// BracketScan runs restrained bracket-like ticks that sweep along the rails.
func BracketScan(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		return bracketScanStyle(frame, bc, bright, mid, dim)
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 68)
	}
	return e
}

// DataTicks runs measured block ticks that move like professional telemetry.
func DataTicks(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		return dataTicksStyle(frame, bc, bright, mid, dim)
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 64)
	}
	return e
}

// StaticNoise flickers clean border glyphs with a cool noisy signal field,
// without any moving artifact or traveling tracer.
func StaticNoise(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		return staticNoiseStyle(frame, bc, bright, mid, dim)
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 84)
	}
	return e
}

// SpinnerPulse runs the spinner2 pulse frames across the border.
func SpinnerPulse(bright, mid, dim cell.Color) *Effect {
	return spinnerFrameEffect(
		mustSpinnerRunes("pulse"),
		bright,
		mid,
		dim,
		72,
	)
}

// Dots6Spinner runs the spinner2 dots_6 frames across the border.
func Dots6Spinner(bright, mid, dim cell.Color) *Effect {
	return spinnerFrameEffect(
		mustSpinnerRunes("dots_6"),
		bright,
		mid,
		dim,
		80,
	)
}

// Dots10Spinner runs the spinner2 dots_10 frames across the border.
func Dots10Spinner(bright, mid, dim cell.Color) *Effect {
	return spinnerFrameEffect(
		mustSpinnerRunes("dots_10"),
		bright,
		mid,
		dim,
		70,
	)
}

// spinnerFrameEffect builds a border effect from a spinner frame set.
func spinnerFrameEffect(frames []rune, bright, mid, dim cell.Color, period int) *Effect {
	mustHalfWidthRunes(frames...)
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		return spinnerFrameStyle(frame, bc, frames, bright, mid, dim)
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, period)
	}
	return e
}

func mustSpinnerRunes(name string) []rune {
	frames, ok := spinners.Must(name).RuneFrames()
	if !ok {
		panic("borderfx requires single-cell spinner frames: " + name)
	}
	return frames
}

// FocusPinsRibbon uses diamond ribbons during activation.
func FocusPinsRibbon(bright, mid, dim cell.Color) *Effect {
	return focusPinsVariant(bright, mid, dim, pinVariant{
		horizontalFrames: 36,
		verticalFrames:   20,
		holdFrames:       22,
		lag:              4,
		topHead:          '✦',
		bottomHead:       '✧',
		sideHead:         '│',
		cornerRune:       '◉',
		cornerTrail:      '●',
	})
}

// FocusPinsBrace uses brace-like side rails during activation.
func FocusPinsBrace(bright, mid, dim cell.Color) *Effect {
	return focusPinsVariant(bright, mid, dim, pinVariant{
		horizontalFrames: 28,
		verticalFrames:   28,
		holdFrames:       18,
		lag:              6,
		topHead:          '▌',
		bottomHead:       '▐',
		sideHead:         '▐',
		cornerRune:       '◉',
		cornerTrail:      '●',
	})
}

// FocusPinsEmoji uses symbol-heavy sweeps and energized corner pins.
func FocusPinsEmoji(bright, mid, dim cell.Color) *Effect {
	return focusPinsVariant(bright, mid, dim, pinVariant{
		horizontalFrames: 42,
		verticalFrames:   22,
		holdFrames:       24,
		lag:              4,
		topHead:          '✦',
		bottomHead:       '✺',
		sideHead:         '✸',
		cornerRune:       '✹',
		cornerTrail:      '✧',
	})
}

// FocusPinsShard uses compact shard markers during activation.
func FocusPinsShard(bright, mid, dim cell.Color) *Effect {
	return focusPinsVariant(bright, mid, dim, pinVariant{
		horizontalFrames: 24,
		verticalFrames:   14,
		holdFrames:       18,
		lag:              2,
		topHead:          '■',
		bottomHead:       '▪',
		sideHead:         '▎',
		cornerRune:       '◉',
		cornerTrail:      '●',
	})
}

// FocusPinsPulse uses softer round markers during activation.
func FocusPinsPulse(bright, mid, dim cell.Color) *Effect {
	return focusPinsVariant(bright, mid, dim, pinVariant{
		horizontalFrames: 40,
		verticalFrames:   18,
		holdFrames:       26,
		lag:              4,
		topHead:          '●',
		bottomHead:       '•',
		sideHead:         '╎',
		cornerRune:       '◉',
		cornerTrail:      '●',
	})
}

type pinVariant struct {
	horizontalFrames int
	verticalFrames   int
	holdFrames       int
	lag              int
	topHead          rune
	bottomHead       rune
	sideHead         rune
	cornerRune       rune
	cornerTrail      rune
}

func focusPinsVariant(bright, mid, dim cell.Color, variant pinVariant) *Effect {
	mustHalfWidthRunes(
		variant.topHead,
		variant.bottomHead,
		variant.sideHead,
		variant.cornerRune,
		variant.cornerTrail,
	)

	moveFrames := variant.horizontalFrames + variant.verticalFrames
	holdFrames := max(1, variant.holdFrames)
	cycleFrames := moveFrames + holdFrames

	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		phase := frame % cycleFrames
		if phase < variant.horizontalFrames {
			return horizontalSweepPins(phase, bc, bright, mid, dim, variant)
		}
		if phase < moveFrames {
			return verticalSweepPins(phase-variant.horizontalFrames, bc, bright, mid, dim, variant)
		}
		return parkedFocusStyle(phase-moveFrames, bc, bright, mid, dim, focusProfile{
			parked:   []rune{variant.cornerRune},
			parkMode: parkCorners,
		})
	})
	e.colorAt = func(frame int) cell.Color {
		phase := frame % cycleFrames
		if phase < moveFrames {
			return pulseColor(phase, bright, dim, moveFrames)
		}
		return lerpRGB(dim, bright, 0.45)
	}
	return e
}

func mustHalfWidthRunes(runes ...rune) {
	for _, r := range runes {
		if runewidth.RuneWidth(r) != 1 {
			panic("borderfx requires single-cell border runes")
		}
	}
}

// Rainbow cycles through the full hue spectrum.
func Rainbow() *Effect {
	return paletteEffect(72, func(i int) cell.Color {
		return hue256(float64(i) * 5)
	})
}

// Pulse smoothly fades between two colors and back.
func Pulse(a, b cell.Color) *Effect {
	return paletteEffect(32, func(i int) cell.Color {
		t := (math.Sin(float64(i)*2*math.Pi/32) + 1) / 2
		return lerpRGB(a, b, t)
	})
}

// Glow pulses a single color between bright and nearly-dark.
func Glow(c cell.Color) *Effect {
	return Pulse(cell.ColorNumber(235), c)
}

// Fire flickers reds, oranges, yellows.
func Fire() *Effect {
	return Cycle([]cell.Color{
		cell.ColorNumber(196), cell.ColorNumber(202),
		cell.ColorNumber(208), cell.ColorNumber(214),
		cell.ColorNumber(220), cell.ColorNumber(214),
		cell.ColorNumber(208), cell.ColorNumber(202),
		cell.ColorNumber(196), cell.ColorNumber(160),
		cell.ColorNumber(196), cell.ColorNumber(202),
	})
}

// Ice shimmers blues and cyans.
func Ice() *Effect {
	return Cycle([]cell.Color{
		cell.ColorNumber(27), cell.ColorNumber(33),
		cell.ColorNumber(39), cell.ColorNumber(45),
		cell.ColorNumber(51), cell.ColorNumber(123),
		cell.ColorNumber(159), cell.ColorNumber(123),
		cell.ColorNumber(51), cell.ColorNumber(45),
		cell.ColorNumber(39), cell.ColorNumber(33),
	})
}

// Neon flickers like a neon sign with occasional dark frames.
func Neon(bright cell.Color) *Effect {
	d := cell.ColorNumber(233)
	return Cycle([]cell.Color{
		bright, bright, bright, bright,
		bright, bright, d, bright,
		bright, d, bright, bright,
		bright, bright, bright, bright,
	})
}

// Matrix pulses dark-to-bright green.
func Matrix() *Effect {
	return Cycle([]cell.Color{
		cell.ColorNumber(22), cell.ColorNumber(28),
		cell.ColorNumber(34), cell.ColorNumber(40),
		cell.ColorNumber(46), cell.ColorNumber(82),
		cell.ColorNumber(46), cell.ColorNumber(40),
		cell.ColorNumber(34), cell.ColorNumber(28),
	})
}

// Synthwave cycles purple, pink, cyan.
func Synthwave() *Effect {
	return Cycle([]cell.Color{
		cell.ColorNumber(129), cell.ColorNumber(135),
		cell.ColorNumber(141), cell.ColorNumber(177),
		cell.ColorNumber(213), cell.ColorNumber(207),
		cell.ColorNumber(201), cell.ColorNumber(165),
		cell.ColorNumber(129), cell.ColorNumber(93),
		cell.ColorNumber(57), cell.ColorNumber(51),
		cell.ColorNumber(45), cell.ColorNumber(51),
		cell.ColorNumber(93), cell.ColorNumber(129),
	})
}

// Cycle steps through any custom color list you provide.
func Cycle(colors []cell.Color) *Effect {
	cp := make([]cell.Color, len(colors))
	copy(cp, colors)
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		if len(cp) == 0 {
			return container.BorderCellStyle{}
		}
		return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(cp[frame%len(cp)])}}
	})
	e.colors = cp
	return e
}

// Warp simulates warp core energy: blue-white pulse with afterglow.
func Warp() *Effect {
	return Cycle([]cell.Color{
		cell.ColorNumber(17), cell.ColorNumber(18),
		cell.ColorNumber(19), cell.ColorNumber(20),
		cell.ColorNumber(21), cell.ColorNumber(27),
		cell.ColorNumber(33), cell.ColorNumber(39),
		cell.ColorNumber(45), cell.ColorNumber(51),
		cell.ColorNumber(87), cell.ColorNumber(123),
		cell.ColorNumber(159), cell.ColorNumber(195),
		cell.ColorNumber(159), cell.ColorNumber(123),
		cell.ColorNumber(87), cell.ColorNumber(51),
		cell.ColorNumber(45), cell.ColorNumber(39),
		cell.ColorNumber(33), cell.ColorNumber(27),
		cell.ColorNumber(21), cell.ColorNumber(17),
	})
}

// ColorScanner simulates a scanning beam travelling clockwise around the
// border using only color changes — the original box-drawing rune characters
// are never replaced.  This matches a CSS conic-gradient rotation effect.
func ColorScanner(bright, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		length := max(1, bc.Length)
		head := (frame * 2) % length
		dist := circularDistance(bc.Index, head, length)
		var c cell.Color
		switch dist {
		case 0:
			c = bright
		case 1:
			c = lerpRGB(dim, bright, 0.72)
		case 2:
			c = lerpRGB(dim, bright, 0.46)
		case 3:
			c = lerpRGB(dim, bright, 0.24)
		default:
			c = dim
		}
		return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(c)}}
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 42)
	}
	return e
}

// DualColorScanner runs two color beams in opposite directions around the
// border, preserving all original rune characters.  Gives the rotating
// conic-gradient illusion of the HTML demo preview.
func DualColorScanner(bright, mid, dim cell.Color) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		length := max(1, bc.Length)
		headA := (frame * 2) % length
		headB := positiveMod(-(frame*3 + length/3), length)

		score := func(dist int, accent cell.Color) (cell.Color, int) {
			switch dist {
			case 0:
				return accent, 4
			case 1:
				return lerpRGB(dim, accent, 0.72), 3
			case 2:
				return lerpRGB(dim, accent, 0.44), 2
			case 3:
				return lerpRGB(dim, accent, 0.22), 1
			default:
				return dim, 0
			}
		}
		ca, sa := score(circularDistance(bc.Index, headA, length), bright)
		cb, sb := score(circularDistance(bc.Index, headB, length), mid)
		c := ca
		if sb > sa {
			c = cb
		}
		return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(c)}}
	})
	e.colorAt = func(frame int) cell.Color {
		return pulseColor(frame, bright, dim, 64)
	}
	return e
}

// GradientArc renders a single arc of arcFraction of the total border length
// that sweeps clockwise.  The leading edge of the arc glows with colorA, the
// trailing edge fades to colorB, and cells outside the arc receive dim.
//
// For visually smooth gradients choose colorA and colorB that are adjacent in
// the xterm-256 cube (e.g. the pure-hue column: 21, 57, 93, 129, 165, 201 —
// each 36 apart).  arcFraction is clamped to [0.05, 0.95].
func GradientArc(colorA, colorB, dim cell.Color, arcFraction float64) *Effect {
	if arcFraction < 0.05 {
		arcFraction = 0.05
	}
	if arcFraction > 0.95 {
		arcFraction = 0.95
	}
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		length := max(1, bc.Length)
		arcLen := int(float64(length)*arcFraction + 0.5)
		if arcLen < 2 {
			arcLen = 2
		}
		head := (frame * 2) % length
		// behind: how many steps behind the head this cell sits (clockwise from head)
		behind := positiveMod(head-bc.Index, length)
		if behind >= arcLen {
			return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(dim)}}
		}
		t := float64(behind) / float64(arcLen-1)
		return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(lerpRGB(colorA, colorB, t))}}
	})
	e.colorAt = func(frame int) cell.Color { return colorA }
	return e
}

// GradientArcN is like GradientArc but accepts an ordered slice of color stops.
// stops[0] is the leading-edge color; stops[len-1] is the trailing-edge color.
// Cells outside the arc receive dim.  arcFraction is clamped to [0.05, 0.95].
//
// Example — purple→indigo→blue with stops from the xterm-256 pure-hue column:
//
//	GradientArcN([]cell.Color{
//	    cell.ColorNumber(201), // magenta
//	    cell.ColorNumber(129), // violet
//	    cell.ColorNumber(57),  // indigo
//	    cell.ColorNumber(21),  // blue
//	}, cell.ColorNumber(17), 0.4)
func GradientArcN(stops []cell.Color, dim cell.Color, arcFraction float64) *Effect {
	if len(stops) == 0 {
		stops = []cell.Color{dim}
	}
	cp := make([]cell.Color, len(stops))
	copy(cp, stops)
	if arcFraction < 0.05 {
		arcFraction = 0.05
	}
	if arcFraction > 0.95 {
		arcFraction = 0.95
	}
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		length := max(1, bc.Length)
		arcLen := int(float64(length)*arcFraction + 0.5)
		if arcLen < 2 {
			arcLen = 2
		}
		head := (frame * 2) % length
		behind := positiveMod(head-bc.Index, length)
		if behind >= arcLen {
			return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(dim)}}
		}
		// Map position-within-arc to stops.
		t := float64(behind) / float64(arcLen-1)
		scaled := t * float64(len(cp)-1)
		lo := int(scaled)
		if lo >= len(cp)-1 {
			return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(cp[len(cp)-1])}}
		}
		frac := scaled - float64(lo)
		return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(lerpRGB(cp[lo], cp[lo+1], frac))}}
	})
	e.colorAt = func(frame int) cell.Color { return cp[0] }
	return e
}

// DualGradientArc runs two gradient arcs simultaneously: one sweeping clockwise
// lit with colorA, one counter-clockwise lit with colorB.  Where arcs overlap
// the brighter cell (closest to its head) wins.  Cells outside both arcs receive dim.
// arcFraction is clamped to [0.05, 0.95].
func DualGradientArc(colorA, colorB, dim cell.Color, arcFraction float64) *Effect {
	if arcFraction < 0.05 {
		arcFraction = 0.05
	}
	if arcFraction > 0.95 {
		arcFraction = 0.95
	}
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		length := max(1, bc.Length)
		arcLen := int(float64(length)*arcFraction + 0.5)
		if arcLen < 2 {
			arcLen = 2
		}

		// CW arc: head advances forward.
		headCW := (frame * 2) % length
		behindCW := positiveMod(headCW-bc.Index, length)

		// CCW arc: head retreats (counter-clockwise → index increases for "behind").
		headCCW := positiveMod(-(frame * 2), length)
		behindCCW := positiveMod(bc.Index-headCCW, length)

		colorFor := func(behind int, accent cell.Color) (cell.Color, bool) {
			if behind >= arcLen {
				return dim, false
			}
			t := float64(behind) / float64(arcLen-1)
			return lerpRGB(accent, dim, t), true
		}

		cA, inA := colorFor(behindCW, colorA)
		cB, inB := colorFor(behindCCW, colorB)

		switch {
		case !inA && !inB:
			return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(dim)}}
		case inA && !inB:
			return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(cA)}}
		case !inA && inB:
			return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(cB)}}
		default:
			// Both in range: pick whichever is closer to its head (lower "behind" index = brighter).
			if behindCW <= behindCCW {
				return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(cA)}}
			}
			return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(cB)}}
		}
	})
	e.colorAt = func(frame int) cell.Color { return colorA }
	return e
}

func paletteEffect(n int, color func(int) cell.Color) *Effect {
	colors := make([]cell.Color, n)
	for i := range colors {
		colors[i] = color(i)
	}
	return Cycle(colors)
}

func beamStyle(bc container.BorderCell, head int, bright, dim cell.Color) container.BorderCellStyle {
	dist := circularDistance(bc.Index, head, max(1, bc.Length))
	switch dist {
	case 0, 1:
		return container.BorderCellStyle{Rune: '█', CellOpts: []cell.Option{cell.FgColor(bright)}}
	case 2:
		return container.BorderCellStyle{Rune: '▓', CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.78))}}
	case 3:
		return container.BorderCellStyle{Rune: '▒', CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.45))}}
	case 4:
		return container.BorderCellStyle{Rune: '░', CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.22))}}
	default:
		return container.BorderCellStyle{Rune: softenedBorderRune(bc), CellOpts: []cell.Option{cell.FgColor(dim)}}
	}
}

func interlaceStyle(bc container.BorderCell, head int, bright, mid, dim cell.Color, frame, lane int) container.BorderCellStyle {
	dist := circularDistance(bc.Index, head, max(1, bc.Length))
	phase := (bc.Index + frame + lane) % 4
	switch dist {
	case 0:
		return container.BorderCellStyle{Rune: []rune{'⣾', '⣷', '⣯', '⣟'}[phase], CellOpts: []cell.Option{cell.FgColor(bright)}}
	case 1:
		return container.BorderCellStyle{Rune: []rune{'⣶', '⣝', '⣻', '⢿'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.82))}}
	case 2:
		return container.BorderCellStyle{Rune: []rune{'⣴', '⣤', '⣦', '⣆'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.58))}}
	case 3:
		return container.BorderCellStyle{Rune: []rune{'⣀', '⣄', '⡄', '⠤'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.34))}}
	case 4:
		return container.BorderCellStyle{Rune: []rune{'⠒', '⠤', '⠢', '⠔'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, 0.4))}}
	default:
		return container.BorderCellStyle{Rune: bc.Rune, CellOpts: []cell.Option{cell.FgColor(dim)}}
	}
}

func braidedStyle(bc container.BorderCell, head int, bright, mid, dim cell.Color, frame, lane int) container.BorderCellStyle {
	dist := circularDistance(bc.Index, head, max(1, bc.Length))
	phase := (bc.Index + frame + lane) % 4
	switch dist {
	case 0:
		return container.BorderCellStyle{Rune: []rune{'⢿', '⣻', '⣽', '⣾'}[phase], CellOpts: []cell.Option{cell.FgColor(bright)}}
	case 1:
		return container.BorderCellStyle{Rune: []rune{'⢻', '⢽', '⣹', '⣝'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.8))}}
	case 2:
		return container.BorderCellStyle{Rune: []rune{'⠿', '⣟', '⡿', '⣯'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.6))}}
	case 3:
		return container.BorderCellStyle{Rune: []rune{'⠶', '⠷', '⠧', '⠳'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, 0.42))}}
	default:
		return container.BorderCellStyle{Rune: bc.Rune, CellOpts: []cell.Option{cell.FgColor(dim)}}
	}
}

func shardStyle(bc container.BorderCell, head int, bright, mid, dim cell.Color, frame, lane int) container.BorderCellStyle {
	dist := circularDistance(bc.Index, head, max(1, bc.Length))
	phase := (bc.Index + frame + lane) % 4
	switch dist {
	case 0:
		return container.BorderCellStyle{Rune: []rune{'⣾', '⣷', '⡿', '⢿'}[phase], CellOpts: []cell.Option{cell.FgColor(bright)}}
	case 1:
		return container.BorderCellStyle{Rune: []rune{'⣶', '⣦', '⢾', '⢶'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.78))}}
	case 2:
		return container.BorderCellStyle{Rune: []rune{'⣤', '⣄', '⠶', '⠤'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.55))}}
	case 3:
		return container.BorderCellStyle{Rune: []rune{'⠒', '⠢', '⠔', '⠂'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, 0.36))}}
	default:
		return container.BorderCellStyle{Rune: bc.Rune, CellOpts: []cell.Option{cell.FgColor(dim)}}
	}
}

func orbitStyle(bc container.BorderCell, head int, bright, mid, dim cell.Color, frame, lane int) container.BorderCellStyle {
	dist := circularDistance(bc.Index, head, max(1, bc.Length))
	phase := (bc.Index + frame + lane) % 4
	switch dist {
	case 0:
		return container.BorderCellStyle{Rune: []rune{'⣾', '⣽', '⣻', '⢿'}[phase], CellOpts: []cell.Option{cell.FgColor(bright)}}
	case 1:
		return container.BorderCellStyle{Rune: []rune{'⣷', '⣯', '⣟', '⡿'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.82))}}
	case 2:
		return container.BorderCellStyle{Rune: []rune{'⣶', '⣦', '⣆', '⡆'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.62))}}
	case 3:
		return container.BorderCellStyle{Rune: []rune{'⠿', '⠷', '⠯', '⠟'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, 0.46))}}
	case 4:
		return container.BorderCellStyle{Rune: []rune{'⠒', '⠤', '⠢', '⠔'}[phase], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, 0.3))}}
	default:
		return container.BorderCellStyle{Rune: bc.Rune, CellOpts: []cell.Option{cell.FgColor(dim)}}
	}
}

type parkMode int

const (
	parkTopBrackets parkMode = iota + 1
	parkSideBraces
	parkTopCenter
	parkBottomCenter
	parkCorners
)

type focusProfile struct {
	moveA     int
	moveB     int
	offsetB   int
	trail     int
	heads     []rune
	tail      []rune
	parked    []rune
	parkMode  parkMode
	moveFrame int
}

func stagedFocusEffect(bright, mid, dim cell.Color, profile focusProfile) *Effect {
	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		if frame < profile.moveFrame {
			return movingFocusStyle(frame, bc, bright, mid, dim, profile)
		}
		return parkedFocusStyle(frame-profile.moveFrame, bc, bright, mid, dim, profile)
	})
	e.colorAt = func(frame int) cell.Color {
		if frame < profile.moveFrame {
			return pulseColor(frame, bright, dim, max(8, profile.moveFrame))
		}
		return lerpRGB(dim, bright, 0.45)
	}
	return e
}

func movingFocusStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color, profile focusProfile) container.BorderCellStyle {
	length := max(1, bc.Length)
	headA := (frame * max(1, profile.moveA)) % length
	headB := (length - 1 - (frame*max(1, profile.moveB)+profile.offsetB)%length + length) % length

	a := focusedTracerStyle(bc, headA, bright, mid, dim, frame, profile.heads, profile.tail, profile.trail)
	b := focusedTracerStyle(bc, headB, lerpRGB(mid, bright, 0.8), mid, dim, frame+1, profile.heads, profile.tail, profile.trail)
	if tracerRank(b.Rune) > tracerRank(a.Rune) {
		return b
	}
	return a
}

func parkedFocusStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color, profile focusProfile) container.BorderCellStyle {
	if bc.Title {
		return container.BorderCellStyle{}
	}

	if profile.parkMode == parkCorners {
		if isCornerCell(bc) {
			glow := cornerGlowStrength(bc, 0.58, 0.42)
			if frame%18 < 4 {
				glow = cornerGlowStrength(bc, 0.82, 0.58)
			}
			return container.BorderCellStyle{
				Rune:     preserveCornerRune(bc, profile.parked[0]),
				CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, glow))},
			}
		}
		return container.BorderCellStyle{Rune: bc.Rune, CellOpts: []cell.Option{cell.FgColor(dim)}}
	}

	markers := parkedIndices(bc.Length, profile.parkMode)
	for i, idx := range markers {
		if bc.Index != idx {
			continue
		}
		r := profile.parked[i%len(profile.parked)]
		glow := 0.58
		if frame%18 < 4 {
			glow = 0.82
		}
		return container.BorderCellStyle{Rune: r, CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, glow))}}
	}

	for _, idx := range markers {
		if circularDistance(bc.Index, idx, max(1, bc.Length)) == 1 {
			return container.BorderCellStyle{Rune: softenedBorderRune(bc), CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, 0.34))}}
		}
	}
	return container.BorderCellStyle{Rune: softenedBorderRune(bc), CellOpts: []cell.Option{cell.FgColor(dim)}}
}

func focusedTracerStyle(bc container.BorderCell, head int, bright, mid, dim cell.Color, frame int, heads, tail []rune, trail int) container.BorderCellStyle {
	dist := circularDistance(bc.Index, head, max(1, bc.Length))
	if bc.Title {
		if dist <= max(1, trail/2) {
			return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(dim)}}
		}
		return container.BorderCellStyle{}
	}

	switch {
	case dist == 0:
		return container.BorderCellStyle{Rune: heads[frame%len(heads)], CellOpts: []cell.Option{cell.FgColor(bright)}}
	case dist == 1:
		return container.BorderCellStyle{Rune: heads[(frame+1)%len(heads)], CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.82))}}
	case dist <= trail:
		return container.BorderCellStyle{Rune: tail[(frame+dist)%len(tail)], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.18+0.12*float64(trail-dist+1)))}}
	default:
		return container.BorderCellStyle{Rune: softenedBorderRune(bc), CellOpts: []cell.Option{cell.FgColor(dim)}}
	}
}

// splitOrbitRailStyle renders a continuous dual-color orbit that originates from
// each side midpoint and circles the full border in opposite directions.
func splitOrbitRailStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color) container.BorderCellStyle {
	length := max(1, bc.Length)
	midpoints := borderMidpoints(bc.Border)
	trail := 4
	best := container.BorderCellStyle{
		Rune:     softenedBorderRune(bc),
		CellOpts: []cell.Option{cell.FgColor(dim)},
	}
	bestRank := -1

	for lane, seed := range midpoints {
		cwHead := positiveMod(seed+frame, length)
		ccwHead := positiveMod(seed-frame, length)

		cw := splitOrbitTracerStyle(bc, cwHead, bright, mid, dim, frame+lane, trail, false)
		if rank := orbitTracerRank(cw); rank > bestRank {
			best = cw
			bestRank = rank
		}

		ccw := splitOrbitTracerStyle(bc, ccwHead, lerpRGB(mid, bright, 0.84), mid, dim, frame+lane+1, trail, true)
		if rank := orbitTracerRank(ccw); rank > bestRank {
			best = ccw
			bestRank = rank
		}
	}

	if bc.Title {
		return container.BorderCellStyle{}
	}

	if isCornerCell(bc) {
		return orbitCornerStyle(bc, best, bright, dim)
	}
	return best
}

// scrollingInterlaceStyle renders an interlaced two-color border that appears to
// scroll continuously while keeping the corner pegs anchored.
func scrollingInterlaceStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color) container.BorderCellStyle {
	shadow := grayscaleBridge(dim, mid, 0.45)
	progress, segmentLength, segmentOK := interlaceSegmentProgress(bc)
	glow := dim
	stripeA := positiveMod(bc.Index+frame, 2) == 0
	hot := false
	if segmentOK {
		glow, hot = interlaceSegmentColor(frame, progress, segmentLength, dim, shadow, mid, bright)
	}

	if bc.Title {
		return container.BorderCellStyle{}
	}

	if isCornerCell(bc) {
		glow := 0.42
		if hot {
			glow = 0.68
		}
		return container.BorderCellStyle{
			Rune:     preserveCornerRune(bc, '◉'),
			CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, glow))},
		}
	}

	r := interlaceBorderRune(bc, frame, stripeA)
	return container.BorderCellStyle{
		Rune:     r,
		CellOpts: []cell.Option{cell.FgColor(glow)},
	}
}

// rainBorderStyle paints a cool rain-on-glass border treatment for the LCARS pane.
func rainBorderStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color) container.BorderCellStyle {
	if bc.Title {
		return container.BorderCellStyle{}
	}
	if isCornerCell(bc) {
		return container.BorderCellStyle{
			Rune:     preserveCornerRune(bc, '◉'),
			CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.54))},
		}
	}

	minX, maxX := bc.Border.Min.X, bc.Border.Max.X-1
	minY, maxY := bc.Border.Min.Y, bc.Border.Max.Y-1
	if bc.Point.Y == minY {
		return rainTopStyle(frame, bc, bright, mid, dim)
	}
	if bc.Point.X == minX || bc.Point.X == maxX {
		return rainSideStyle(frame, bc, bright, mid, dim)
	}
	if bc.Point.Y == maxY {
		return rainBottomStyle(frame, bc, bright, mid, dim)
	}
	return container.BorderCellStyle{Rune: softenedBorderRune(bc), CellOpts: []cell.Option{cell.FgColor(dim)}}
}

// rainTopStyle keeps the title rail calm so the side rain reads clearly.
func rainTopStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color) container.BorderCellStyle {
	_ = frame
	_ = bright
	_ = mid
	return container.BorderCellStyle{Rune: softenedBorderRune(bc), CellOpts: []cell.Option{cell.FgColor(dim)}}
}

// rainSideStyle drives sparse telemetry markers down the side rails.
func rainSideStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color) container.BorderCellStyle {
	height := max(1, bc.Border.Dy()-1)
	dropletA := positiveMod(frame+2, height)
	dropletB := positiveMod(frame*2+height/3, height)
	localY := bc.Point.Y - bc.Border.Min.Y

	best := container.BorderCellStyle{Rune: softenedBorderRune(bc), CellOpts: []cell.Option{cell.FgColor(dim)}}
	for lane, head := range []int{dropletA, dropletB} {
		dist := localY - head
		if dist < 0 {
			dist = -dist
		}
		switch dist {
		case 0:
			return container.BorderCellStyle{Rune: []rune{'✦', '✧'}[lane%2], CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.82))}}
		case 1:
			best = container.BorderCellStyle{Rune: []rune{'•', '◦'}[(frame+lane)%2], CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, 0.58))}}
		case 2:
			if cell.NewOptions(best.CellOpts...).FgColor == dim {
				best = container.BorderCellStyle{Rune: '·', CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, 0.4))}}
			}
		}
	}
	return best
}

// rainBottomStyle gives the border a faint telemetry accent along the bottom edge.
func rainBottomStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color) container.BorderCellStyle {
	phase := positiveMod(frame+bc.Index, 28)
	switch {
	case phase == 0:
		return container.BorderCellStyle{Rune: '✦', CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.58))}}
	case phase == 1:
		return container.BorderCellStyle{Rune: '•', CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, 0.42))}}
	default:
		return container.BorderCellStyle{Rune: softenedBorderRune(bc), CellOpts: []cell.Option{cell.FgColor(dim)}}
	}
}

// brailleDriftStyle paints the warp-style spinner field with shaded block cells.
func brailleDriftStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color) container.BorderCellStyle {
	return fluxSpinnerStyle(frame, bc, bright, mid, dim)
}

// bracketScanStyle paints measured bracket and rail segments for a utility look.
func bracketScanStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color) container.BorderCellStyle {
	return instrumentTrackStyle(frame, bc, bright, mid, dim, instrumentGlyphs{
		horizontal: []rune{'╍', '┄'},
		vertical:   []rune{'╎', '┆'},
		active:     []rune{'▏', '▎', '▍', '▎'},
	})
}

// dataTicksStyle paints short measured block ticks with a subtle trailing echo.
func dataTicksStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color) container.BorderCellStyle {
	return instrumentTrackStyle(frame, bc, bright, mid, dim, instrumentGlyphs{
		horizontal: []rune{'╾', '╼'},
		vertical:   []rune{'┊', '┆'},
		active:     []rune{'▁', '▂', '▃', '▂'},
	})
}

// staticNoiseStyle keeps the border geometry intact and varies only the color
// intensity per cell, producing a crisp “alive” edge with no sweep artifact.
func staticNoiseStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color) container.BorderCellStyle {
	if bc.Title {
		return container.BorderCellStyle{}
	}
	if isCornerCell(bc) {
		return container.BorderCellStyle{
			Rune:     preserveCornerRune(bc, '◉'),
			CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.72))},
		}
	}

	seed := positiveMod(bc.Index*13+bc.Point.X*5+bc.Point.Y*7+(frame/2)*9, 29)
	var c cell.Color
	switch {
	case seed <= 2:
		c = bright
	case seed <= 8:
		c = lerpRGB(mid, bright, 0.68)
	case seed <= 17:
		c = mid
	case seed <= 23:
		c = lerpRGB(dim, mid, 0.58)
	default:
		c = dim
	}
	return container.BorderCellStyle{
		Rune:     softenedBorderRune(bc),
		CellOpts: []cell.Option{cell.FgColor(c)},
	}
}

// spinnerFrameStyle renders one frame-shifted spinner cell on the border.
func spinnerFrameStyle(frame int, bc container.BorderCell, frames []rune, bright, mid, dim cell.Color) container.BorderCellStyle {
	if bc.Title {
		return container.BorderCellStyle{}
	}
	if isCornerCell(bc) {
		return container.BorderCellStyle{
			Rune:     preserveCornerRune(bc, '◉'),
			CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.44))},
		}
	}
	if len(frames) == 0 {
		return container.BorderCellStyle{Rune: softenedBorderRune(bc), CellOpts: []cell.Option{cell.FgColor(dim)}}
	}

	phase := positiveMod(frame+bc.Index, len(frames))
	return container.BorderCellStyle{
		Rune:     frames[phase],
		CellOpts: []cell.Option{cell.FgColor(spinnerFrameColor(phase, len(frames), bright, mid, dim))},
	}
}

// spinnerFrameColor keeps spinner-derived borders inside the cool instrument
// palette while still giving the active frame a visible crest.
func spinnerFrameColor(phase, total int, bright, mid, dim cell.Color) cell.Color {
	if total <= 1 {
		return mid
	}
	center := total / 2
	dist := abs(phase - center)
	if dist > total-dist {
		dist = total - dist
	}
	switch dist {
	case 0:
		return bright
	case 1:
		return mid
	case 2:
		return grayscaleBridge(dim, mid, 0.45)
	default:
		return dim
	}
}

// fluxSpinnerStyle makes every border cell participate in a moving shaded-block
// spinner wave.
func fluxSpinnerStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color) container.BorderCellStyle {
	if bc.Title {
		return container.BorderCellStyle{}
	}
	if isCornerCell(bc) {
		return container.BorderCellStyle{
			Rune:     preserveCornerRune(bc, '◉'),
			CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.42))},
		}
	}

	phase := positiveMod(frame+bc.Index, 18)
	r, color := fluxSpinnerCell(phase, bright, mid, dim)
	return container.BorderCellStyle{
		Rune:     r,
		CellOpts: []cell.Option{cell.FgColor(color)},
	}
}

// fluxSpinnerCell returns the shaded block and stepped color for one spinner
// phase, keeping the motion inside the graphite/light-blue palette.
func fluxSpinnerCell(phase int, bright, mid, dim cell.Color) (rune, cell.Color) {
	gray := grayscaleBridge(dim, mid, 0.45)
	switch phase {
	case 0, 1, 2, 3, 4, 5, 13, 14, 15, 16, 17:
		return '▒', dim
	case 6, 12:
		return '░', gray
	case 7, 11:
		return '▒', mid
	case 8, 10:
		return '▒', bright
	case 9:
		return '▓', bright
	default:
		return '▒', dim
	}
}

type instrumentGlyphs struct {
	horizontal []rune
	vertical   []rune
	active     []rune
}

// instrumentTrackStyle is the shared professional border treatment used by the
// quieter demo variants.
func instrumentTrackStyle(frame int, bc container.BorderCell, bright, mid, dim cell.Color, glyphs instrumentGlyphs) container.BorderCellStyle {
	if bc.Title {
		return container.BorderCellStyle{}
	}
	if isCornerCell(bc) {
		return container.BorderCellStyle{
			Rune:     preserveCornerRune(bc, '◉'),
			CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.46))},
		}
	}

	r := professionalBaseRune(frame, bc, glyphs)
	shadow := grayscaleBridge(dim, mid, 0.45)
	progress, segmentLength, ok := interlaceSegmentProgress(bc)
	if !ok {
		return container.BorderCellStyle{Rune: r, CellOpts: []cell.Option{cell.FgColor(dim)}}
	}

	glow, hot := interlaceSegmentColor(frame, progress, segmentLength, dim, shadow, mid, bright)
	if hot {
		r = glyphs.active[positiveMod(frame+progress, len(glyphs.active))]
	}
	return container.BorderCellStyle{Rune: r, CellOpts: []cell.Option{cell.FgColor(glow)}}
}

// professionalBaseRune chooses the restrained resting line glyph for a border cell.
func professionalBaseRune(frame int, bc container.BorderCell, glyphs instrumentGlyphs) rune {
	if bc.Point.Y == bc.Border.Min.Y || bc.Point.Y == bc.Border.Max.Y-1 {
		return glyphs.horizontal[positiveMod(frame+bc.Index, len(glyphs.horizontal))]
	}
	return glyphs.vertical[positiveMod(frame+bc.Index, len(glyphs.vertical))]
}

// interlaceSegmentProgress maps a border cell onto one of the four quarter-tracks
// that run from the top and bottom title joins outward toward the side midpoints.
func interlaceSegmentProgress(bc container.BorderCell) (progress, segmentLength int, ok bool) {
	minX, maxX := bc.Border.Min.X, bc.Border.Max.X-1
	minY, maxY := bc.Border.Min.Y, bc.Border.Max.Y-1
	topMidX := minX + (bc.Border.Dx()-1)/2
	bottomMidX := topMidX
	leftMidY := minY + (bc.Border.Dy()-1)/2
	rightMidY := leftMidY

	switch {
	case bc.Point.X == minX && bc.Point.Y <= leftMidY:
		segmentLength = (leftMidY - minY) + (topMidX - minX)
		return segmentLength - (leftMidY - bc.Point.Y), segmentLength, true
	case bc.Point.Y == minY && bc.Point.X <= topMidX:
		segmentLength = (leftMidY - minY) + (topMidX - minX)
		return segmentLength - ((leftMidY - minY) + (bc.Point.X - minX)), segmentLength, true
	case bc.Point.X == maxX && bc.Point.Y <= rightMidY:
		segmentLength = (rightMidY - minY) + (maxX - topMidX)
		return segmentLength - (rightMidY - bc.Point.Y), segmentLength, true
	case bc.Point.Y == minY && bc.Point.X >= topMidX:
		segmentLength = (rightMidY - minY) + (maxX - topMidX)
		return segmentLength - ((rightMidY - minY) + (maxX - bc.Point.X)), segmentLength, true
	case bc.Point.X == minX && bc.Point.Y >= leftMidY:
		segmentLength = (maxY - leftMidY) + (bottomMidX - minX)
		return segmentLength - (bc.Point.Y - leftMidY), segmentLength, true
	case bc.Point.Y == maxY && bc.Point.X <= bottomMidX:
		segmentLength = (maxY - leftMidY) + (bottomMidX - minX)
		return segmentLength - ((maxY - leftMidY) + (bc.Point.X - minX)), segmentLength, true
	case bc.Point.X == maxX && bc.Point.Y >= rightMidY:
		segmentLength = (maxY - rightMidY) + (maxX - bottomMidX)
		return segmentLength - (bc.Point.Y - rightMidY), segmentLength, true
	case bc.Point.Y == maxY && bc.Point.X >= bottomMidX:
		segmentLength = (maxY - rightMidY) + (maxX - bottomMidX)
		return segmentLength - ((maxY - rightMidY) + (maxX - bc.Point.X)), segmentLength, true
	default:
		return 0, 0, false
	}
}

// interlaceSegmentColor returns the stepped color profile for one moving artifact
// traveling along a quarter-track and whether the artifact is actively over the cell.
func interlaceSegmentColor(frame, progress, segmentLength int, dark, gray, light, accent cell.Color) (cell.Color, bool) {
	profileCenter := 18
	head := positiveMod(frame, max(1, segmentLength+1))
	index := profileCenter + progress - head
	color, hot := interlaceArtifactColor(index, dark, gray, light, accent)
	return color, hot
}

// splitOrbitTracerStyle renders one of the moving orbit tracers for a single cell.
func splitOrbitTracerStyle(bc container.BorderCell, head int, bright, mid, dim cell.Color, frame, trail int, reverse bool) container.BorderCellStyle {
	dist := circularDistance(bc.Index, head, max(1, bc.Length))
	if dist > trail {
		return container.BorderCellStyle{
			Rune:     softenedBorderRune(bc),
			CellOpts: []cell.Option{cell.FgColor(dim)},
		}
	}

	if dist == 0 {
		return container.BorderCellStyle{
			Rune:     orbitHeadRune(bc, frame, reverse),
			CellOpts: []cell.Option{cell.FgColor(bright)},
		}
	}

	if dist == 1 {
		return container.BorderCellStyle{
			Rune:     orbitTraceRune(bc, frame, reverse, 1),
			CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.76))},
		}
	}

	intensity := 0.2 + 0.12*float64(trail-dist+1)
	return container.BorderCellStyle{
		Rune:     orbitTraceRune(bc, frame, reverse, dist),
		CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, intensity))},
	}
}

// orbitCornerStyle keeps corner pegs visible while letting passing tracers boost
// their glow.
func orbitCornerStyle(bc container.BorderCell, base container.BorderCellStyle, bright, dim cell.Color) container.BorderCellStyle {
	glow := 0.5
	if orbitTracerRank(base) >= 5 {
		glow = 0.84
	} else if orbitTracerRank(base) >= 3 {
		glow = 0.68
	}
	return container.BorderCellStyle{
		Rune:     preserveCornerRune(bc, '◉'),
		CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, cornerGlowStrength(bc, glow, glow*0.78)))},
	}
}

// borderMidpoints returns the perimeter indices for the midpoint of each side in
// top, right, bottom, left order.
func borderMidpoints(border image.Rectangle) [4]int {
	width := max(2, border.Dx())
	height := max(2, border.Dy())
	top := width / 2
	right := (width - 1) + (height-1)/2
	bottom := (width - 1) + (height - 1) + (width-1)/2
	left := (width - 1) + (height - 1) + (width - 1) + (height-2)/2
	return [4]int{top, right, bottom, left}
}

// orbitHeadRune returns the brightest glyph for a moving orbit head.
func orbitHeadRune(bc container.BorderCell, frame int, reverse bool) rune {
	if bc.Point.Y == bc.Border.Min.Y || bc.Point.Y == bc.Border.Max.Y-1 {
		if reverse {
			return []rune{'⣯', '⣟', '⣷', '⣾'}[positiveMod(frame, 4)]
		}
		return []rune{'⣾', '⣷', '⣯', '⣟'}[positiveMod(frame, 4)]
	}
	if reverse {
		return []rune{'⣻', '⢿', '⣽', '⣾'}[positiveMod(frame, 4)]
	}
	return []rune{'⣾', '⣽', '⣻', '⢿'}[positiveMod(frame, 4)]
}

// orbitTraceRune returns a subtler glyph for the tracer tail.
func orbitTraceRune(bc container.BorderCell, frame int, reverse bool, dist int) rune {
	if dist == 1 {
		if bc.Point.Y == bc.Border.Min.Y || bc.Point.Y == bc.Border.Max.Y-1 {
			if reverse {
				return []rune{'⣶', '⣦', '⣤', '⣄'}[positiveMod(frame, 4)]
			}
			return []rune{'⣄', '⣤', '⣦', '⣶'}[positiveMod(frame, 4)]
		}
		if reverse {
			return []rune{'⣧', '⣇', '⡇', '⢸'}[positiveMod(frame, 4)]
		}
		return []rune{'⢸', '⡇', '⣇', '⣧'}[positiveMod(frame, 4)]
	}
	if bc.Point.Y == bc.Border.Min.Y || bc.Point.Y == bc.Border.Max.Y-1 {
		return []rune{'⠿', '⠷', '⠯', '⠟'}[positiveMod(frame+dist, 4)]
	}
	return []rune{'⡇', '⢸', '⡇', '⢸'}[positiveMod(frame+dist, 4)]
}

// interlaceBorderRune picks the alternating barber-pole glyph for the current cell.
func interlaceBorderRune(bc container.BorderCell, frame int, stripeA bool) rune {
	phase := positiveMod(frame+bc.Index, 4)
	if bc.Point.Y == bc.Border.Min.Y || bc.Point.Y == bc.Border.Max.Y-1 {
		if stripeA {
			return []rune{'╍', '┅', '╍', '┅'}[phase]
		}
		return []rune{'┅', '╍', '┅', '╍'}[phase]
	}
	if stripeA {
		return []rune{'╏', '┆', '╏', '┆'}[phase]
	}
	return []rune{'┆', '╏', '┆', '╏'}[phase]
}

// orbitTracerRank reports how visually dominant a style is when multiple tracers
// overlap the same cell.
func orbitTracerRank(style container.BorderCellStyle) int {
	switch style.Rune {
	case '⣾', '⣷', '⣯', '⣟', '⣽', '⣻', '⢿':
		return 6
	case '⣶', '⣦', '⣤', '⣄', '⣧', '⣇', '⡇', '⢸':
		return 5
	case '⠿', '⠷', '⠯', '⠟':
		return 3
	default:
		return interlaceRank(style)
	}
}

// positiveMod returns a positive modulo result for signed animation offsets.
func positiveMod(v, mod int) int {
	if mod <= 0 {
		return 0
	}
	v %= mod
	if v < 0 {
		v += mod
	}
	return v
}

// interlaceArtifactColor keeps the interlaced scanner on a deliberate stepped
// loop rather than interpolating through unrelated xterm palette hues.
func interlaceArtifactColor(index int, dark, gray, light, accent cell.Color) (cell.Color, bool) {
	switch index {
	case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10:
		return dark, false
	case 11, 12, 13, 14:
		return gray, true
	case 15, 16, 17:
		return light, true
	case 18, 19, 20, 21, 22, 23, 24, 25, 26:
		return accent, true
	case 27, 28, 29:
		return light, true
	case 30, 31, 32:
		return gray, true
	default:
		return dark, false
	}
}

// grayscaleBridge derives an in-between neutral tone for grayscale-heavy palettes.
// When the palette is not on the grayscale ramp, this still biases toward the dim
// side and avoids blending directly into accent colors.
func grayscaleBridge(dim, light cell.Color, t float64) cell.Color {
	if t <= 0 {
		return dim
	}
	if t >= 1 {
		return light
	}
	return lerpRGB(dim, light, t)
}

func horizontalSweepPins(frame int, bc container.BorderCell, bright, mid, dim cell.Color, variant pinVariant) container.BorderCellStyle {
	width := max(1, bc.Border.Dx()-1)
	topProgress := float64(frame) / float64(max(1, variant.horizontalFrames-1))
	bottomFrame := frame - variant.lag
	bottomProgress := 0.0
	if bottomFrame > 0 {
		bottomProgress = float64(bottomFrame) / float64(max(1, variant.horizontalFrames-1))
	}
	topX := bc.Border.Min.X + int(math.Round(topProgress*float64(width)))
	bottomX := bc.Border.Min.X + int(math.Round(bottomProgress*float64(width)))

	if bc.Point.Y == bc.Border.Min.Y {
		dx := abs(bc.Point.X - topX)
		if bc.Title {
			if dx <= 2 {
				return container.BorderCellStyle{CellOpts: []cell.Option{cell.FgColor(dim)}}
			}
			return container.BorderCellStyle{}
		}
		if isCornerCell(bc) && dx <= 1 {
			r := preserveCornerRune(bc, '●')
			if dx == 0 {
				r = preserveCornerRune(bc, '◉')
			}
			glow := cornerGlowStrength(bc, 1.0, 0.62)
			if dx == 1 {
				glow = cornerGlowStrength(bc, 0.84, 0.48)
			}
			return container.BorderCellStyle{
				Rune:     r,
				CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, glow))},
			}
		}
		if dx == 0 {
			return container.BorderCellStyle{
				Rune:     variant.topHead,
				CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.9))},
			}
		}
		if dx == 1 {
			return container.BorderCellStyle{
				Rune:     softenedBorderRune(bc),
				CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.56))},
			}
		}
		if dx == 2 {
			return container.BorderCellStyle{
				Rune:     softenedBorderRune(bc),
				CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, 0.36))},
			}
		}
	}

	if bc.Point.Y == bc.Border.Max.Y-1 {
		dx := abs(bc.Point.X - bottomX)
		if isCornerCell(bc) && dx <= 1 {
			r := preserveCornerRune(bc, '●')
			if dx == 0 {
				r = preserveCornerRune(bc, '◉')
			}
			glow := cornerGlowStrength(bc, 0.84, 0.56)
			if dx == 1 {
				glow = cornerGlowStrength(bc, 0.72, 0.42)
			}
			return container.BorderCellStyle{
				Rune:     r,
				CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, glow))},
			}
		}
		if dx == 0 {
			return container.BorderCellStyle{
				Rune:     variant.bottomHead,
				CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.8))},
			}
		}
		if dx == 1 {
			return container.BorderCellStyle{
				Rune:     softenedBorderRune(bc),
				CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.46))},
			}
		}
	}

	return container.BorderCellStyle{
		Rune:     softenedBorderRune(bc),
		CellOpts: []cell.Option{cell.FgColor(dim)},
	}
}

func verticalSweepPins(frame int, bc container.BorderCell, bright, mid, dim cell.Color, variant pinVariant) container.BorderCellStyle {
	if bc.Title {
		return container.BorderCellStyle{}
	}

	height := max(1, bc.Border.Dy()-1)
	progress := float64(frame) / float64(max(1, variant.verticalFrames-1))
	headY := bc.Border.Min.Y + int(math.Round(progress*float64(height)))
	dy := abs(bc.Point.Y - headY)

	if isCornerCell(bc) && dy <= 1 {
		r := preserveCornerRune(bc, variant.cornerTrail)
		if dy == 0 {
			r = preserveCornerRune(bc, variant.cornerRune)
		}
		glow := cornerGlowStrength(bc, 1.0, 0.62)
		if dy == 1 {
			glow = cornerGlowStrength(bc, 0.84, 0.48)
		}
		return container.BorderCellStyle{
			Rune:     r,
			CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, glow))},
		}
	}

	if (bc.Point.X == bc.Border.Min.X || bc.Point.X == bc.Border.Max.X-1) && dy == 0 {
		return container.BorderCellStyle{
			Rune:     variant.sideHead,
			CellOpts: []cell.Option{cell.FgColor(lerpRGB(mid, bright, 0.86))},
		}
	}

	if (bc.Point.X == bc.Border.Min.X || bc.Point.X == bc.Border.Max.X-1) && dy == 1 {
		return container.BorderCellStyle{
			Rune:     softenedBorderRune(bc),
			CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, bright, 0.52))},
		}
	}

	if (bc.Point.X == bc.Border.Min.X || bc.Point.X == bc.Border.Max.X-1) && dy == 2 {
		return container.BorderCellStyle{
			Rune:     softenedBorderRune(bc),
			CellOpts: []cell.Option{cell.FgColor(lerpRGB(dim, mid, 0.34))},
		}
	}

	return container.BorderCellStyle{
		Rune:     softenedBorderRune(bc),
		CellOpts: []cell.Option{cell.FgColor(dim)},
	}
}

// TextSweep creates a left-to-right sweep effect on the border title text.
// The title text is held at grey (restColor) and a white (sweepColor) window
// quickly sweeps across it left to right. Text returns to grey once the
// window passes. The sweep repeats continuously to signal active focus or
// loading. The sweep head advances 3 characters per frame for a fast,
// snappy feel.
func TextSweep(sweepColor, restColor cell.Color) *Effect {
	return TextSweepWidth(sweepColor, restColor, 2, 1)
}

// TextSweepWidth is like TextSweep but lets the caller choose the sweep
// window width and speed. Speed is how many characters the sweep head
// advances per frame (higher = faster).
func TextSweepWidth(sweepColor, restColor cell.Color, windowWidth, speed int) *Effect {
	if windowWidth < 1 {
		windowWidth = 1
	}
	if speed < 1 {
		speed = 1
	}

	e := newEffect(func(frame int, bc container.BorderCell) container.BorderCellStyle {
		if bc.Title {
			charPos := bc.Point.X - bc.Border.Min.X
			titleWidth := max(1, bc.Border.Dx()-1)

			// Sweep head moves left to right. Multiply by 115 and
			// divide by 100 to get ~15% faster than 1 char/frame.
			cycle := titleWidth + windowWidth/2
			head := (frame * speed * 115 / 100) % cycle

			// Characters inside the sweep window turn white.
			if charPos <= head && charPos > head-windowWidth {
				return container.BorderCellStyle{
					CellOpts: []cell.Option{cell.FgColor(sweepColor)},
				}
			}
			// Everything else is grey.
			return container.BorderCellStyle{
				CellOpts: []cell.Option{cell.FgColor(restColor)},
			}
		}

		// Non-title border cells: dim.
		return container.BorderCellStyle{
			Rune:     softenedBorderRune(bc),
			CellOpts: []cell.Option{cell.FgColor(restColor)},
		}
	})
	e.colorAt = func(frame int) cell.Color {
		return sweepColor
	}
	return e
}

func parkedIndices(length int, mode parkMode) []int {
	if length <= 0 {
		return nil
	}
	switch mode {
	case parkSideBraces:
		return []int{length / 4, (3 * length) / 4}
	case parkTopCenter:
		center := length / 8
		return []int{center, center + 1, center + 2, center + 3}
	case parkBottomCenter:
		center := (length / 2) + (length / 8)
		return []int{center - 1, center, center + 1}
	case parkCorners:
		return nil
	default:
		return []int{2, 3, max(4, length-4), max(5, length-3)}
	}
}

func isCornerCell(bc container.BorderCell) bool {
	return (bc.Point.X == bc.Border.Min.X || bc.Point.X == bc.Border.Max.X-1) &&
		(bc.Point.Y == bc.Border.Min.Y || bc.Point.Y == bc.Border.Max.Y-1)
}

func preserveCornerRune(bc container.BorderCell, fallback rune) rune {
	if isRoundedCornerRune(bc.Rune) {
		return softenedBorderRune(bc)
	}
	return fallback
}

func softenedBorderRune(bc container.BorderCell) rune {
	switch bc.Rune {
	case '╭':
		return '○'
	case '╮':
		return '○'
	case '╰':
		return '○'
	case '╯':
		return '○'
	default:
		return bc.Rune
	}
}

func cornerGlowStrength(bc container.BorderCell, normal, rounded float64) float64 {
	if isRoundedCornerRune(bc.Rune) {
		return rounded
	}
	return normal
}

func isRoundedCornerRune(r rune) bool {
	switch r {
	case '╭', '╮', '╰', '╯':
		return true
	default:
		return false
	}
}

func tracerRank(r rune) int {
	switch r {
	case '◉', '●', '◆', '◈', '▣', '▰':
		return 6
	case '▐', '▌', '▓', '▒':
		return 5
	case '▪', '▫', '•', '∙', '·', ':':
		return 4
	default:
		return interlaceRank(container.BorderCellStyle{Rune: r})
	}
}

func beamRank(r rune) int {
	switch r {
	case '█':
		return 4
	case '▓':
		return 3
	case '▒':
		return 2
	case '░':
		return 1
	default:
		return 0
	}
}

func interlaceRank(style container.BorderCellStyle) int {
	switch style.Rune {
	case '⣾', '⣷', '⣯', '⣟':
		return 6
	case '⣶', '⣝', '⣻', '⢿':
		return 5
	case '⣴', '⣤', '⣦', '⣆':
		return 4
	case '⣀', '⣄', '⡄':
		return 3
	case '⠒', '⠤', '⠢', '⠔':
		return 2
	default:
		return beamRank(style.Rune)
	}
}

func circularDistance(a, b, length int) int {
	d := abs(a - b)
	if d > length-d {
		return length - d
	}
	return d
}

func pulseColor(frame int, bright, dim cell.Color, period int) cell.Color {
	t := (math.Sin(float64(frame)*2*math.Pi/float64(period)) + 1) / 2
	return lerpRGB(dim, bright, math.Pow(t, 3))
}

// lerpIdx interpolates linearly between two color indices.  It is the raw
// index-based fallback used internally by lerpRGB.  Prefer lerpRGB in all
// public-facing effect code so gradients stay on-hue in the xterm-256 cube.
func lerpIdx(a, b cell.Color, t float64) cell.Color {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	ai, bi := int(a), int(b)
	r := ai + int(math.Round(float64(bi-ai)*t))
	if r < 0 {
		r = 0
	}
	if r > 255 {
		r = 255
	}
	return cell.ColorNumber(r)
}

// lerp is preserved as a small compatibility alias for older internal tests
// and helpers that still refer to the original interpolation name.
func lerp(a, b cell.Color, t float64) cell.Color {
	return lerpIdx(a, b, t)
}

// xterm6ToRGB decodes the (r, g, b) cube components of an xterm-256 cube
// color.  idx must be in [16, 231]; components are in [0, 5].
func xterm6ToRGB(idx int) (r, g, b int) {
	idx -= 16
	b = idx % 6
	idx /= 6
	g = idx % 6
	r = idx / 6
	return
}

// lerpRGB interpolates between two xterm-256 cube colors (indices 16–231) in
// RGB space, producing visually correct gradients regardless of how far apart
// the two indices are.
//
// Why not lerpIdx? lerpIdx(51=cyan, 17=dim-blue) passes through index 34 =
// (0,3,0) = dark green — completely wrong hue.  lerpRGB(51, 16=black) instead
// travels through RGB space: (0,5,5)→(0,4,4)→(0,3,3)→(0,2,2)→(0,0,0)
// = cyan → teal → dark-teal → near-black.  Always correct.
//
// Falls back to lerpIdx for non-cube colors (index < 16 or > 231).
func lerpRGB(a, b cell.Color, t float64) cell.Color {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	ai, bi := int(a), int(b)
	if ai >= 16 && ai <= 231 && bi >= 16 && bi <= 231 {
		ar, ag, ab := xterm6ToRGB(ai)
		br, bg, bb := xterm6ToRGB(bi)
		ri := int(math.Round(float64(ar) + float64(br-ar)*t))
		gi := int(math.Round(float64(ag) + float64(bg-ag)*t))
		bli := int(math.Round(float64(ab) + float64(bb-ab)*t))
		if ri < 0 {
			ri = 0
		}
		if ri > 5 {
			ri = 5
		}
		if gi < 0 {
			gi = 0
		}
		if gi > 5 {
			gi = 5
		}
		if bli < 0 {
			bli = 0
		}
		if bli > 5 {
			bli = 5
		}
		return cell.ColorNumber(16 + 36*ri + 6*gi + bli)
	}
	return lerpIdx(a, b, t)
}

func hue256(deg float64) cell.Color {
	for deg >= 360 {
		deg -= 360
	}
	h := deg / 60
	s := int(h) % 6
	f := h - float64(int(h))
	var r, g, b float64
	switch s {
	case 0:
		r, g, b = 1, f, 0
	case 1:
		r, g, b = 1-f, 1, 0
	case 2:
		r, g, b = 0, 1, f
	case 3:
		r, g, b = 0, 1-f, 1
	case 4:
		r, g, b = f, 0, 1
	case 5:
		r, g, b = 1, 0, 1-f
	}
	return cell.ColorNumber(16 + 36*int(r*5) + 6*int(g*5) + int(b*5))
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
