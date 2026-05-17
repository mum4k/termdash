package borderfx

import (
	"image"
	"strings"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// LoadingBackgroundOption configures a loading background renderer.
type LoadingBackgroundOption interface {
	// set applies the option to the loading background.
	set(*LoadingBackground)
}

// LoadingBackground draws a striped loading surface for widgets that are still
// booting or hydrating content.
type LoadingBackground struct {
	primary   cell.Color
	secondary cell.Color
	text      cell.Color
}

// loadingBackgroundOption adapts a function into a LoadingBackgroundOption.
type loadingBackgroundOption func(*LoadingBackground)

// set implements LoadingBackgroundOption.set.
func (o loadingBackgroundOption) set(lb *LoadingBackground) {
	o(lb)
}

// NewLoadingBackground returns a loading background with the supplied options.
func NewLoadingBackground(opts ...LoadingBackgroundOption) LoadingBackground {
	lb := LoadingBackground{
		primary:   cell.ColorNumber(236),
		secondary: cell.ColorNumber(239),
		text:      cell.ColorNumber(117),
	}
	for _, opt := range opts {
		opt.set(&lb)
	}
	return lb
}

// InterlacedLoadingBackground alternates two background colors row-by-row.
func InterlacedLoadingBackground(primary, secondary cell.Color) LoadingBackgroundOption {
	return loadingBackgroundOption(func(lb *LoadingBackground) {
		lb.primary = primary
		lb.secondary = secondary
	})
}

// LoadingTextColor sets the foreground color used for loading copy rendered on
// top of the background stripes.
func LoadingTextColor(color cell.Color) LoadingBackgroundOption {
	return loadingBackgroundOption(func(lb *LoadingBackground) {
		lb.text = color
	})
}

// RowCellOpts returns the cell options for the supplied row of the loading
// background.
func (lb LoadingBackground) RowCellOpts(row int) []cell.Option {
	return []cell.Option{cell.BgColor(lb.rowColor(row))}
}

// Draw renders the striped loading background into rect and then draws the
// provided loading frame on top of it.
func (lb LoadingBackground) Draw(t terminalapi.Terminal, rect image.Rectangle, frame string) error {
	if t == nil || rect.Empty() {
		return nil
	}

	for row := 0; row < rect.Dy(); row++ {
		opts := lb.RowCellOpts(row)
		y := rect.Min.Y + row
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if err := t.SetCell(image.Point{X: x, Y: y}, ' ', opts...); err != nil {
				return err
			}
		}
	}

	lines := strings.Split(frame, "\n")
	for row, line := range lines {
		if row >= rect.Dy() {
			break
		}
		x := rect.Min.X
		opts := append([]cell.Option{cell.FgColor(lb.text)}, lb.RowCellOpts(row)...)
		for _, r := range line {
			width := runewidth.RuneWidth(r)
			if width <= 0 {
				continue
			}
			if x+width > rect.Max.X {
				break
			}
			if err := t.SetCell(image.Point{X: x, Y: rect.Min.Y + row}, r, opts...); err != nil {
				return err
			}
			x += width
		}
	}
	return nil
}

// rowColor returns the background color for the supplied row.
func (lb LoadingBackground) rowColor(row int) cell.Color {
	if row%2 == 0 {
		return lb.primary
	}
	return lb.secondary
}

// ── LoadingOverlay ────────────────────────────────────────────────────────────

// PanelRectFunc maps a terminal size and panel ID to the drawable content
// rectangle for that panel.  For a panel with a 1-cell border, shrink the
// outer border rectangle by 1 on all sides using rect.Inset(1).
type PanelRectFunc func(termSize image.Point, id string) image.Rectangle

// LoadingOverlay wraps a terminal and paints an interlaced loading background
// over one or more named panels during the boot or hydration phase of your app.
//
// It implements terminalapi.Terminal — pass it wherever a terminal is expected
// (container.New, termdash.Run, etc.).  All method calls are forwarded to the
// wrapped terminal; only Flush() is intercepted to paint the overlay.
//
// Quickstart — three steps:
//
//  1. Wrap your terminal:
//
//	lo := borderfx.WrapWithLoading(t, func(size image.Point, id string) image.Rectangle {
//	    outer := myLayout.BorderRect(size, id) // the panel's outer border rect
//	    return outer.Inset(1)                  // strip the 1-cell border
//	})
//
//  2. Set loading content and wire the animator (one call does both):
//
//	lo.SetContent("sensors", borderfx.LoadingText.BootSequence)
//	fx.ApplyInterlacedLoadingContent(lo, "sensors", "comms")
//
//  3. Hide when boot is done:
//
//	lo.Hide()
type LoadingOverlay struct {
	terminalapi.Terminal        // all non-overridden calls delegate here
	mu      sync.RWMutex
	bg      LoadingBackground
	rectFn  PanelRectFunc
	frames  map[string]string  // panel ID → loading text currently displayed
	visible bool
}

// WrapWithLoading wraps t and returns a LoadingOverlay ready to display an
// interlaced boot screen.
//
// rectFn maps (terminal size, panel ID) to the drawable inner rectangle for
// that panel.  It is called on every Flush so the overlay stays correct after
// terminal resize.  For a bordered panel the inner rect is the outer border
// rect inset by 1:
//
//	func(size image.Point, id string) image.Rectangle {
//	    return outerBorderRects(size)[id].Inset(1)
//	}
func WrapWithLoading(t terminalapi.Terminal, rectFn PanelRectFunc) *LoadingOverlay {
	return &LoadingOverlay{
		Terminal: t,
		bg:       NewLoadingBackground(),
		rectFn:   rectFn,
		frames:   make(map[string]string),
	}
}

// SetContent sets the loading text displayed inside a named panel.
// Call before Show() to pre-populate panels, or at any time during the loading
// phase to update the copy.  Use "\n" to separate lines.
//
// The LoadingText variable provides ready-made templates:
//
//	lo.SetContent("sensors", borderfx.LoadingText.BootSequence)
func (lo *LoadingOverlay) SetContent(id, text string) {
	lo.mu.Lock()
	lo.frames[id] = text
	lo.mu.Unlock()
}

// Show makes the loading overlay visible.  It is called automatically by
// Animator.ApplyInterlacedLoadingContent — you only need to call it directly
// if you are managing the overlay lifecycle manually.
func (lo *LoadingOverlay) Show() {
	lo.mu.Lock()
	lo.visible = true
	lo.mu.Unlock()
}

// Hide removes the loading overlay, revealing the live widget content beneath.
// Call this when your application has finished booting or loading.
func (lo *LoadingOverlay) Hide() {
	lo.mu.Lock()
	lo.visible = false
	lo.mu.Unlock()
}

// Flush flushes the wrapped terminal's back buffer, then — when the overlay is
// visible — paints the interlaced loading background into each registered panel.
// This method satisfies terminalapi.Terminal and overrides the embedded one.
func (lo *LoadingOverlay) Flush() error {
	if err := lo.Terminal.Flush(); err != nil {
		return err
	}

	lo.mu.RLock()
	visible := lo.visible
	bg := lo.bg
	frames := make(map[string]string, len(lo.frames))
	for id, f := range lo.frames {
		frames[id] = f
	}
	size := lo.Terminal.Size()
	rectFn := lo.rectFn
	lo.mu.RUnlock()

	if !visible || rectFn == nil {
		return nil
	}
	for id, text := range frames {
		rect := rectFn(size, id)
		if !rect.Empty() {
			_ = bg.Draw(lo.Terminal, rect, text)
		}
	}
	return nil
}

// ── LoadingText ───────────────────────────────────────────────────────────────

// LoadingText provides ready-made boot-screen text templates for use with
// LoadingOverlay.SetContent.  Each template is multi-line copy that fits
// comfortably inside a typical bordered panel.
//
// Example:
//
//	lo.SetContent("sensors", borderfx.LoadingText.BootSequence)
//	lo.SetContent("db",      borderfx.LoadingText.DataSync)
var LoadingText = struct {
	// BootSequence mimics an RF carrier-lock + telemetry initialisation.
	// Good for sensor, signal, or graph panels.
	BootSequence string

	// DataSync shows a database / schema hydration sequence.
	// Good for panels that load data from a backend.
	DataSync string

	// Initializing shows a generic component-wiring startup sequence.
	// Good for catch-all or secondary panels.
	Initializing string

	// Standby shows a minimal "waiting for signal" state.
	// Good for panels that depend on an upstream source not yet ready.
	Standby string

	// NetworkBoot shows a network stack initialization sequence.
	// Good for panels that require connectivity.
	NetworkBoot string
}{
	BootSequence: "" +
		"  :: carrier lock ::\n\n" +
		"  preparing window .............\n\n" +
		"  signal lattice ...............\n\n" +
		"  telemetry uplink .............\n",

	DataSync: "" +
		"  :: data sync ::\n\n" +
		"  connecting .....................\n\n" +
		"  fetching schema ................\n\n" +
		"  hydrating cache ................\n",

	Initializing: "" +
		"  :: initializing ::\n\n" +
		"  loading components .............\n\n" +
		"  wiring dependencies ............\n\n" +
		"  ready check ....................\n",

	Standby: "" +
		"  :: standby ::\n\n" +
		"  waiting for signal .............\n\n" +
		"  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  \n\n" +
		"  ·  ·  ·  ·  ·  ·  ·  ·  ·  ·  \n",

	NetworkBoot: "" +
		"  :: network boot ::\n\n" +
		"  resolving address ..............\n\n" +
		"  establishing link ..............\n\n" +
		"  syncing clock ..................\n",
}
