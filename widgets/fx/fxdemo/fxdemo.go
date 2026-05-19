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

// Binary fxdemo showcases every built-in effect in the widgets/fx package.
//
// The demo starts LOCKED — all content is scrambled and a centered password
// modal appears.  Type "termdash" and press Enter to unlock.  The modal
// disappears and the full showcase is revealed with a Glitch + FadeIn
// entrance animation.
//
// Once unlocked, each content pane is wrapped in a FocusEffectWidget: click
// any pane with the mouse to trigger a fade-in / fade-out on focus change,
// with the pane's border color animating in sync.  Effects auto-advance every
// few seconds; manual navigation is also available.
//
// Controls (active after unlock)
//
//	1–13       Jump to effect by number (two-digit entry supported)
//	Tab / →    Next effect
//	← / h      Previous effect
//	Space      Replay current effect from the start
//	q / Q / Esc / Ctrl+C  Quit
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
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/fx"
	"github.com/mum4k/termdash/widgets/modal"
	"github.com/mum4k/termdash/widgets/text"
)

// ── timing ────────────────────────────────────────────────────────────────────

const (
	redrawTick  = 40 * time.Millisecond
	autoAdvance = 4 * time.Second

	// Container IDs used for hot-swapping widgets.
	// idModal hosts the draggable modal window set for the entire stage.
	idModal   = "fx-modal"
	idCatalog = "fx-catalog"
	idInfo    = "fx-info"
)

// ── effect catalogue ──────────────────────────────────────────────────────────

// effectSpec describes one demo entry.
type effectSpec struct {
	Name        string
	Tag         string // short classifier shown in catalog
	Description string
	Duration    time.Duration // total play time (sum of all steps)
	Build       func() []fx.Effect
}

// allEffects is the ordered list shown in the catalog.
var allEffects = []effectSpec{
	{
		Name:        "FadeIn",
		Tag:         "reveal",
		Description: "Transitions content from invisible through dim to full\nbrightness over the specified duration.\n\nWorks on all terminal color modes (16-color, 256,\ntrue-color).  The three-step dim ramp (blank → dim →\nbright) is imperceptible at normal frame rates.",
		Duration:    700 * time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{fx.FadeIn(700 * time.Millisecond)}
		},
	},
	{
		Name:        "FadeOut",
		Tag:         "hide",
		Description: "Dims and hides content to blank — the time-reverse\nof FadeIn.\n\nPaired here with a leading FadeIn so the cycle is\nvisible: content fades in, pauses, then fades out.",
		Duration:    400*time.Millisecond + 200*time.Millisecond + 700*time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{
				fx.FadeIn(400 * time.Millisecond),
				fx.FadeIn(200 * time.Millisecond),
				fx.FadeOut(700 * time.Millisecond),
			}
		},
	},
	{
		Name:        "ColorFadeIn",
		Tag:         "reveal",
		Description: "Interpolates foreground and background colors from\nblack to their target values.\n\nProduces the smoothest gradient when widgets use\ncell.ColorRGB6 colors (ColorMode256).  Falls back to\nthe standard three-step dim ramp for named colors.",
		Duration:    800 * time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{fx.ColorFadeIn(800 * time.Millisecond)}
		},
	},
	{
		Name:        "SweepLeft",
		Tag:         "reveal",
		Description: "Reveals the canvas column by column, left to right.",
		Duration:    600 * time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{fx.SweepLeft(600 * time.Millisecond)}
		},
	},
	{
		Name:        "SweepDown",
		Tag:         "reveal",
		Description: "Reveals the canvas row by row, top to bottom.\n\nCombining SweepDown with FadeIn via Parallel() gives\na very polished window-open transition.",
		Duration:    600 * time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{fx.SweepDown(600 * time.Millisecond)}
		},
	},
	{
		Name:        "SweepRight",
		Tag:         "reveal",
		Description: "Reveals the canvas right to left — useful for\nclosing or dismissal transitions.",
		Duration:    600 * time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{fx.SweepRight(600 * time.Millisecond)}
		},
	},
	{
		Name:        "SweepUp",
		Tag:         "reveal",
		Description: "Reveals the canvas bottom to top.\n\nPair with SweepDown in sequence for a venetian-blind\nopen/close effect.",
		Duration:    600 * time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{fx.SweepUp(600 * time.Millisecond)}
		},
	},
	{
		Name:        "WipeDiagonal",
		Tag:         "reveal",
		Description: "Reveals the canvas along a diagonal frontier sweeping\nfrom the top-left corner to the bottom-right.\n\nThe frontier is a line perpendicular to the main\ndiagonal.  Works at any aspect ratio.",
		Duration:    700 * time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{fx.WipeDiagonal(700 * time.Millisecond)}
		},
	},
	{
		Name:        "ScanLine",
		Tag:         "reveal",
		Description: "A bold horizontal scan line sweeps from top to\nbottom, illuminating each row as it passes and\nleaving it fully revealed behind.\n\nInspired by classic CRT phosphor scan effects.",
		Duration:    700 * time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{fx.ScanLine(700 * time.Millisecond)}
		},
	},
	{
		Name:        "Dissolve",
		Tag:         "reveal",
		Description: "Reveals cells in a deterministic pseudo-random order.\nThe same seed always produces the same pattern,\nmaking the effect reproducible across runs.",
		Duration:    900 * time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{fx.Dissolve(900*time.Millisecond, 0xdeadbeef)}
		},
	},
	{
		Name:        "Glitch",
		Tag:         "noise",
		Description: "Floods the canvas with box-drawing noise that peaks\nat the midpoint then decays as real content settles.\n\nNoise is derived from a fast integer hash — no\nallocations, fully goroutine-safe.",
		Duration:    900 * time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{fx.Glitch(900*time.Millisecond, 0xcafebabe)}
		},
	},
	{
		Name:        "Parallel: Fade + Sweep",
		Tag:         "combo",
		Description: "FadeIn and SweepDown applied simultaneously using\nParallel().  Both sub-effects share the same\ntimeline; the result is composited per-cell.\n\nUse Parallel() to combine any number of effects.",
		Duration:    750 * time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{
				fx.Parallel(
					fx.FadeIn(750*time.Millisecond),
					fx.SweepDown(750*time.Millisecond),
				),
			}
		},
	},
	{
		Name:        "Sequence: Glitch → FadeIn",
		Tag:         "combo",
		Description: "Glitch and FadeIn play back-to-back as a sequence.\n\nPass multiple effects to fx.New() or fx.NewLooping()\nto chain them; each runs for its own Duration before\nthe next begins.",
		Duration:    600*time.Millisecond + 500*time.Millisecond,
		Build: func() []fx.Effect {
			return []fx.Effect{
				fx.Glitch(600*time.Millisecond, 0x1337),
				fx.FadeIn(500 * time.Millisecond),
			}
		},
	},
}

// numEffects is the total number of entries in allEffects.
const numEffects = 13 // keep in sync with allEffects length

// defaultEffectIdx is the 0-based index of the effect shown after unlock.
// "Sequence: Glitch → FadeIn" gives a dramatic entrance animation.
const defaultEffectIdx = 12

// ── Lock state ────────────────────────────────────────────────────────────────

// lockState manages the password gate and post-unlock coordination.
type lockState struct {
	mu           sync.Mutex
	locked       bool
	input        string
	errMsg       string
	promptW      *text.Text // placed in idModal while locked
	justUnlocked bool       // prevents animateDemo double-switching on first tick
}

func newLockState() *lockState {
	promptW, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("lock screen widget: %v", err)
	}
	ls := &lockState{locked: true, promptW: promptW}
	_ = fillLockScreen(ls.promptW, "", "")
	return ls
}

func (ls *lockState) isLocked() bool {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.locked
}

// push appends a printable rune to the password buffer.
func (ls *lockState) push(r rune) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.input += string(r)
	ls.errMsg = ""
	_ = fillLockScreen(ls.promptW, ls.input, "")
}

// backspace removes the last rune from the password buffer.
func (ls *lockState) backspace() {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	runes := []rune(ls.input)
	if len(runes) > 0 {
		ls.input = string(runes[:len(runes)-1])
	}
	ls.errMsg = ""
	_ = fillLockScreen(ls.promptW, ls.input, "")
}

// submit checks the password.  Returns true (and clears locked) on match;
// returns false and shows an error on mismatch.
func (ls *lockState) submit() bool {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	if ls.input == "termdash" {
		ls.locked = false
		return true
	}
	ls.errMsg = "✗  incorrect password — try again"
	ls.input = ""
	_ = fillLockScreen(ls.promptW, "", ls.errMsg)
	return false
}

// setJustUnlocked marks that the unlock switch has already been performed.
func (ls *lockState) setJustUnlocked() {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.justUnlocked = true
}

// consumeJustUnlocked reads and clears the justUnlocked flag in one step.
func (ls *lockState) consumeJustUnlocked() bool {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	if ls.justUnlocked {
		ls.justUnlocked = false
		return true
	}
	return false
}

// fillLockScreen renders the password prompt into t.
// Kept compact so it fits inside the draggable modal window.
func fillLockScreen(t *text.Text, input, errMsg string) error {
	t.Reset()
	masked := strings.Repeat("●", len(input))

	_ = t.Write("\n  ✦  SECURE TERMINAL\n\n", wo(clrAccent), text.WriteCellOpts(cell.Bold()))
	_ = t.Write("  Content is encrypted.\n\n", wo(clrDim))
	_ = t.Write("  Password  ", wo(clrMid))
	_ = t.Write("│ "+masked+"▌\n\n", wo(clrAccent), text.WriteCellOpts(cell.Bold()))

	if errMsg != "" {
		_ = t.Write("  "+errMsg+"\n\n", wo(clrErr))
	} else {
		_ = t.Write("\n", wo(clrDim))
	}

	_ = t.Write("  ──────────────────────────────\n", wo(clrDim))
	_ = t.Write("  hint    │ ", wo(clrDim))
	_ = t.Write("termdash\n\n", wo(clrDim))
	_ = t.Write("  Enter · Backspace\n", wo(clrDim))
	return nil
}

// renderLockedCatalog fills the catalog sidebar with a locked-state message.
func renderLockedCatalog(t *text.Text) error {
	t.Reset()
	_ = t.Write("\n", wo(clrDim))
	_ = t.Write(" 🔒  DEMO LOCKED\n\n", wo(clrErr), text.WriteCellOpts(cell.Bold()))
	_ = t.Write(" 13 effects waiting.\n\n", wo(clrMid))
	_ = t.Write(" Type in the top\n", wo(clrDim))
	_ = t.Write(" pane, then Enter.\n", wo(clrDim))
	return nil
}

// renderLockedInfo fills the info sidebar with a locked-state message.
func renderLockedInfo(t *text.Text) error {
	t.Reset()
	_ = t.Write("\n Status: ", wo(clrDim))
	_ = t.Write("LOCKED\n\n", wo(clrErr), text.WriteCellOpts(cell.Bold()))
	_ = t.Write(" Unlock to browse\n", wo(clrDim))
	_ = t.Write(" 13 fx effects.\n", wo(clrDim))
	return nil
}

// ── numSelector ───────────────────────────────────────────────────────────────

// numSelector buffers digit keypresses and commits after a short delay,
// enabling two-digit effect selection (e.g. '1' then '3' → effect 13).
type numSelector struct {
	mu     sync.Mutex
	buf    string
	timer  *time.Timer
	action func(n int) // called with a 1-based effect number
}

func newNumSelector(action func(n int)) *numSelector {
	return &numSelector{action: action}
}

func (ns *numSelector) push(r rune) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.buf += string(r)

	if ns.timer != nil {
		ns.timer.Stop()
		ns.timer = nil
	}

	n := 0
	for _, ch := range ns.buf {
		n = n*10 + int(ch-'0')
	}

	if len(ns.buf) >= 2 || n*10 > numEffects {
		ns.doCommit(n)
		return
	}

	buf := ns.buf
	ns.timer = time.AfterFunc(350*time.Millisecond, func() {
		ns.mu.Lock()
		defer ns.mu.Unlock()
		if ns.buf == buf {
			n := 0
			for _, ch := range ns.buf {
				n = n*10 + int(ch-'0')
			}
			ns.doCommit(n)
		}
	})
}

func (ns *numSelector) doCommit(n int) {
	if ns.timer != nil {
		ns.timer.Stop()
		ns.timer = nil
	}
	ns.buf = ""
	if n < 1 || n > numEffects {
		return
	}
	action := ns.action
	go action(n)
}

// ── demo state ────────────────────────────────────────────────────────────────

type demoState struct {
	mu         sync.Mutex
	active     int
	lastSwitch time.Time
}

func (s *demoState) getActive() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

func (s *demoState) setActive(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = (idx%numEffects + numEffects) % numEffects
	s.lastSwitch = time.Now()
}

func (s *demoState) next() { s.setActive(s.getActive() + 1) }
func (s *demoState) prev() { s.setActive(s.getActive() - 1) }

func (s *demoState) sinceSwitch() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.lastSwitch)
}

// ── content widgets ───────────────────────────────────────────────────────────

type contentWidgets struct {
	log     *text.Text
	status  *text.Text
	console *text.Text
}

func newContentWidgets() (*contentWidgets, error) {
	logW, err := text.New()
	if err != nil {
		return nil, err
	}
	statusW, err := text.New()
	if err != nil {
		return nil, err
	}
	consoleW, err := text.New()
	if err != nil {
		return nil, err
	}
	if err := fillLog(logW); err != nil {
		return nil, err
	}
	if err := fillStatus(statusW); err != nil {
		return nil, err
	}
	if err := fillConsole(consoleW); err != nil {
		return nil, err
	}
	return &contentWidgets{log: logW, status: statusW, console: consoleW}, nil
}

// ── color palette ─────────────────────────────────────────────────────────────

var (
	clrOK     = cell.FgColor(cell.ColorNumber(46))
	clrWarn   = cell.FgColor(cell.ColorNumber(220))
	clrErr    = cell.FgColor(cell.ColorNumber(197))
	clrInfo   = cell.FgColor(cell.ColorNumber(39))
	clrDim    = cell.FgColor(cell.ColorNumber(240))
	clrMid    = cell.FgColor(cell.ColorNumber(250))
	clrBright = cell.FgColor(cell.ColorNumber(255))
	clrVal    = cell.FgColor(cell.ColorNumber(87))
	clrNum    = cell.FgColor(cell.ColorNumber(214))
	clrPurple = cell.FgColor(cell.ColorNumber(141))
	clrBlue   = cell.FgColor(cell.ColorNumber(69))
	clrAccent = cell.FgColor(cell.ColorNumber(159))
)

func wo(c cell.Option) text.WriteOption { return text.WriteCellOpts(c) }

func fillLog(t *text.Text) error {
	type entry struct {
		tag  string
		tagC cell.Option
		msg  string
	}
	entries := []entry{
		{"[BOOT] ", clrInfo, "kernel v4.2.1 loaded — 4 cores, 16 GiB"},
		{"[ OK ] ", clrOK, "network interfaces: eth0 eth1 lo"},
		{"[ OK ] ", clrOK, "storage: /dev/nvme0 512 GiB"},
		{"[ OK ] ", clrOK, "services started: 23 / 23"},
		{"[WARN] ", clrWarn, "memory pressure: 74 %"},
		{"[ OK ] ", clrOK, "watchdog armed — interval 5 s"},
		{"[ OK ] ", clrOK, "telemetry stream open on :9090"},
		{"[INFO] ", clrInfo, "mission elapsed: 02:14:07"},
		{"[ OK ] ", clrOK, "comms: signal –42 dBm, link stable"},
		{"[WARN] ", clrWarn, "nav drift: +0.003° — correcting"},
		{"[ OK ] ", clrOK, "propulsion: thrust nominal"},
		{"[INFO] ", clrInfo, "next waypoint: 847 km"},
	}
	if err := t.Write("  SYSTEM BOOT LOG\n", wo(clrBright), text.WriteCellOpts(cell.Bold())); err != nil {
		return err
	}
	if err := t.Write("  ─────────────────────────────────\n", wo(clrDim)); err != nil {
		return err
	}
	for _, e := range entries {
		if err := t.Write("  "+e.tag, wo(e.tagC)); err != nil {
			return err
		}
		if err := t.Write(e.msg+"\n", wo(clrMid)); err != nil {
			return err
		}
	}
	return nil
}

func fillStatus(t *text.Text) error {
	type bar struct {
		label string
		value int
		color cell.Option
	}
	bars := []bar{
		{"CPU   ", 78, clrOK},
		{"MEM   ", 91, clrWarn},
		{"NET ↑ ", 38, clrBlue},
		{"NET ↓ ", 64, clrBlue},
		{"DISK  ", 52, clrInfo},
		{"TEMP  ", 83, clrErr},
		{"GPU   ", 47, clrPurple},
		{"SWAP  ", 12, clrVal},
	}
	if err := t.Write("  SYSTEM METRICS\n", wo(clrBright), text.WriteCellOpts(cell.Bold())); err != nil {
		return err
	}
	if err := t.Write("  ─────────────────────────────────\n", wo(clrDim)); err != nil {
		return err
	}
	for _, b := range bars {
		const barWidth = 10
		filled := int(math.Round(float64(b.value) / 100.0 * barWidth))
		empty := barWidth - filled
		if err := t.Write("  "+b.label, wo(clrMid)); err != nil {
			return err
		}
		if err := t.Write("["+repeat('█', filled), wo(b.color)); err != nil {
			return err
		}
		if err := t.Write(repeat('░', empty)+"] ", wo(clrDim)); err != nil {
			return err
		}
		if err := t.Write(fmt.Sprintf("%3d%%\n", b.value), wo(clrNum)); err != nil {
			return err
		}
	}
	return nil
}

func fillConsole(t *text.Text) error {
	type subsystem struct {
		name  string
		state string
		stC   cell.Option
		info  string
	}
	subs := []subsystem{
		{"power_grid   ", "OK  ", clrOK, "14.2 V nominal"},
		{"comm_array   ", "OK  ", clrOK, "signal –42 dBm"},
		{"nav_system   ", "WARN", clrWarn, "drift +0.003°"},
		{"life_support ", "OK  ", clrOK, "O₂ 21.0 %"},
		{"propulsion   ", "OK  ", clrOK, "thrust nominal"},
		{"thermal_ctrl ", "WARN", clrWarn, "82 °C / 95 °C limit"},
		{"data_logger  ", "OK  ", clrOK, "4.2 GiB free"},
		{"gyroscope    ", "OK  ", clrOK, "0.001°/s drift"},
		{"solar_panels ", "OK  ", clrOK, "312 W input"},
		{"battery      ", "OK  ", clrOK, "87 % charge"},
	}
	if err := t.Write("  $ diagnostics --full --verbose\n", wo(clrAccent)); err != nil {
		return err
	}
	if err := t.Write("  > scanning subsystems…\n\n", wo(clrDim)); err != nil {
		return err
	}
	for i, s := range subs {
		prefix := "  ├─ "
		if i == len(subs)-1 {
			prefix = "  └─ "
		}
		if err := t.Write(prefix, wo(clrDim)); err != nil {
			return err
		}
		if err := t.Write(s.name, wo(clrMid)); err != nil {
			return err
		}
		if err := t.Write("["+s.state+"] ", wo(s.stC)); err != nil {
			return err
		}
		if err := t.Write(s.info+"\n", wo(clrVal)); err != nil {
			return err
		}
	}
	if err := t.Write("\n  $ _", wo(clrAccent)); err != nil {
		return err
	}
	return nil
}

func repeat(r rune, n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = r
	}
	return string(b)
}

// ── sidebar ───────────────────────────────────────────────────────────────────

func newSidebarWidgets() (*text.Text, *text.Text, *text.Text, error) {
	catalog, err := text.New()
	if err != nil {
		return nil, nil, nil, err
	}
	info, err := text.New(text.WrapAtWords())
	if err != nil {
		return nil, nil, nil, err
	}
	controls, err := text.New()
	if err != nil {
		return nil, nil, nil, err
	}
	if err := renderControls(controls); err != nil {
		return nil, nil, nil, err
	}
	return catalog, info, controls, nil
}

func renderCatalog(t *text.Text, active int) error {
	t.Reset()
	for i, e := range allEffects {
		var tag string
		var tagC cell.Option
		switch e.Tag {
		case "reveal":
			tag = "reveal"
			tagC = clrOK
		case "hide":
			tag = "hide  "
			tagC = clrWarn
		case "noise":
			tag = "noise "
			tagC = clrErr
		case "combo":
			tag = "combo "
			tagC = clrPurple
		default:
			tag = "      "
			tagC = clrDim
		}
		if i == active {
			if err := t.Write(fmt.Sprintf(" ► %2d. ", i+1), wo(clrAccent), text.WriteCellOpts(cell.Bold())); err != nil {
				return err
			}
			if err := t.Write(e.Name+"\n", wo(clrBright), text.WriteCellOpts(cell.Bold())); err != nil {
				return err
			}
		} else {
			if err := t.Write(fmt.Sprintf("   %2d. ", i+1), wo(clrDim)); err != nil {
				return err
			}
			if err := t.Write(e.Name, wo(clrMid)); err != nil {
				return err
			}
			if err := t.Write(fmt.Sprintf("  [%s]\n", tag), wo(tagC)); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderInfo(t *text.Text, active int) error {
	t.Reset()
	e := allEffects[active]
	if err := t.Write("Effect: ", wo(clrDim)); err != nil {
		return err
	}
	if err := t.Write(e.Name+"\n", wo(clrBright), text.WriteCellOpts(cell.Bold())); err != nil {
		return err
	}
	if err := t.Write(fmt.Sprintf("Duration: %v\n", e.Duration), wo(clrNum)); err != nil {
		return err
	}
	if err := t.Write(fmt.Sprintf("Tag: [%s]\n\n", e.Tag), wo(clrInfo)); err != nil {
		return err
	}
	if err := t.Write(e.Description+"\n", wo(clrMid)); err != nil {
		return err
	}
	return nil
}

func renderControls(t *text.Text) error {
	type row struct{ key, act string }
	rows := []row{
		{"1–13  ", "jump to effect"},
		{"Tab / →", "next effect"},
		{"← / h  ", "previous effect"},
		{"Space  ", "replay current"},
		{"q / Esc", "quit"},
	}
	for _, r := range rows {
		if err := t.Write(" "+r.key, wo(clrAccent)); err != nil {
			return err
		}
		if err := t.Write("  "+r.act+"\n", wo(clrMid)); err != nil {
			return err
		}
	}
	return nil
}

// ── EffectWidget + FocusEffectWidget factory ──────────────────────────────────

// paneWidgets bundles the layers of a content pane.
//
//   - focus  — outermost widget placed in the container
//   - framed — the FramedWidget that draws the self-owned border
//
// The stack is: text → FramedWidget → EffectWidget → FocusEffectWidget.
// Because FramedWidget paints its border onto the widget canvas, effects
// applied by EffectWidget and FocusEffectWidget cover the entire visual
// area — border characters and inner content together.
type paneWidgets struct {
	focus  *fx.FocusEffectWidget
	framed *fx.FramedWidget
}

// wrapContent builds a paneWidgets for each content widget using the effect
// at idx.  The containers for these panes must use linestyle.None so that
// FramedWidget's self-drawn border is the only border visible.
func wrapContent(idx int, cw *contentWidgets) (logP, statusP, consoleP paneWidgets, err error) {
	effects := allEffects[idx].Build()
	focusIn := []fx.Effect{fx.FadeIn(300 * time.Millisecond)}
	focusOut := []fx.Effect{fx.FadeOut(300 * time.Millisecond)}

	wrap := func(inner *text.Text, title string) (paneWidgets, error) {
		// FramedWidget: draws the border onto the widget canvas so the effect
		// animates border + content as a single unit.
		framed, err := fx.FramedNew(inner,
			fx.FramedTitle(title),
			fx.FramedBorderOpts(cell.FgColor(cell.ColorNumber(240))),
		)
		if err != nil {
			return paneWidgets{}, err
		}
		// EffectWidget: plays the selected looping animation on the whole area.
		ew, err := fx.NewLooping(framed, effects...)
		if err != nil {
			return paneWidgets{}, err
		}
		// FocusEffectWidget: overlays a fade on focus transitions.
		fw, err := fx.FocusNew(ew, focusIn, focusOut)
		if err != nil {
			return paneWidgets{}, err
		}
		return paneWidgets{focus: fw, framed: framed}, nil
	}

	if logP, err = wrap(cw.log, "Signal Log"); err != nil {
		return paneWidgets{}, paneWidgets{}, paneWidgets{}, fmt.Errorf("logW: %w", err)
	}
	if statusP, err = wrap(cw.status, "System Metrics"); err != nil {
		return paneWidgets{}, paneWidgets{}, paneWidgets{}, fmt.Errorf("statusW: %w", err)
	}
	if consoleP, err = wrap(cw.console, "Diagnostics Console"); err != nil {
		return paneWidgets{}, paneWidgets{}, paneWidgets{}, fmt.Errorf("consoleW: %w", err)
	}
	return
}

// setFocusCallbacks wires OnFocusChange so focus transitions animate the
// FramedWidget's border color (the border is now part of the widget canvas).
func setFocusCallbacks(logP, statusP, consoleP paneWidgets) {
	for _, p := range []paneWidgets{logP, statusP, consoleP} {
		p := p
		p.focus.OnFocusChange = func(gained bool) {
			fx.AnimateBorderFocus(gained, 300*time.Millisecond, 236, 255, 20, func(n int) {
				p.framed.SetBorderColor(cell.ColorNumber(n))
			})
		}
	}
}

// ── layout ────────────────────────────────────────────────────────────────────

func pane(title string, extra ...container.Option) []container.Option {
	base := []container.Option{
		container.Border(linestyle.Round),
		container.BorderTitle(" " + title + " "),
		container.BorderColor(cell.ColorNumber(240)),
	}
	return append(base, extra...)
}

// buildLayout constructs the container tree.
//
// Stage layout (left 68%):
//
//	idModal hosts draggable modal windows for password, log, status, and console.
//	The right sidebar remains fixed so effect selection and controls stay visible.
func buildLayout(
	term *tcell.Terminal,
	catalog, info, controls *text.Text,
) (*container.Container, error) {
	return container.New(
		term,
		container.ID("root"),
		container.Border(linestyle.Round),
		container.BorderTitle(" ✦ termdash / fx effects showcase ✦ "),
		container.BorderColor(cell.ColorNumber(69)),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					// ── Stage (left 68%) ──────────────────────────────────
					container.Left(
						container.Border(linestyle.None),
						container.ID(idModal),
						// modal.Manager.ShowModal places the draggable window set here.
					),
					// ── Sidebar (right 32%) ────────────────────────────────
					container.Right(
						container.SplitHorizontal(
							container.Top(
								pane("Effect Catalog",
									container.ID(idCatalog),
									container.PaddingLeft(1),
									container.PaddingTop(1),
									container.PlaceWidget(catalog),
								)...,
							),
							container.Bottom(
								container.SplitHorizontal(
									container.Top(
										pane("Effect Info",
											container.ID(idInfo),
											container.PaddingLeft(1),
											container.PaddingTop(1),
											container.PlaceWidget(info),
										)...,
									),
									container.Bottom(
										pane("Controls",
											container.PaddingLeft(1),
											container.PaddingTop(1),
											container.PlaceWidget(controls),
										)...,
									),
									container.SplitPercent(65),
								),
							),
							container.SplitPercent(42),
						),
					),
					container.SplitPercent(68),
				),
			),
			container.Bottom(
				container.PlaceWidget(mustFooter()),
			),
			container.SplitPercent(93),
		),
	)
}

func demoModalOptions() *modal.Options {
	return modal.NewOptions(
		modal.Border(true),
		modal.MinimumSize(image.Point{X: 36, Y: 12}),
		modal.TitleBarCellOpts(
			cell.BgColor(cell.ColorNumber(17)),
			cell.FgColor(cell.ColorNumber(214)),
		),
		modal.TitleCellOpts(
			cell.FgColor(cell.ColorNumber(214)),
			cell.Bold(),
		),
		modal.TitleControlCellOpts(
			cell.FgColor(cell.ColorNumber(159)),
			cell.Bold(),
		),
	)
}

func newStageWindow(id, title string, w widgetapi.Widget, x, y, width, height int, opts *modal.Options) *modal.DraggableWidget {
	item := modal.NewDraggableWidget(id, w, x, y, width, height, opts)
	item.Title = title
	return item
}

func mustFooter() *text.Text {
	t, err := text.New()
	if err != nil {
		log.Fatalf("footer: %v", err)
	}
	parts := []struct {
		s string
		c cell.Option
	}{
		{"  1–13", clrAccent},
		{" select   ", clrMid},
		{"Tab/→", clrAccent},
		{" next   ", clrMid},
		{"←/h", clrAccent},
		{" prev   ", clrMid},
		{"Space", clrAccent},
		{" replay   ", clrMid},
		{"click pane", clrAccent},
		{" focus fade   ", clrMid},
		{"q/Esc", clrAccent},
		{" quit", clrMid},
	}
	for _, p := range parts {
		if err := t.Write(p.s, wo(p.c)); err != nil {
			log.Fatalf("footer write: %v", err)
		}
	}
	return t
}

// ── switching ─────────────────────────────────────────────────────────────────

// switchEffect hot-swaps the three stage modal windows to use the effect at idx.
func switchEffect(
	idx int,
	c *container.Container,
	cw *contentWidgets,
	catalog, info *text.Text,
	manager *modal.Manager,
	modalOpts *modal.Options,
) error {
	logP, statusP, consoleP, err := wrapContent(idx, cw)
	if err != nil {
		return err
	}
	setFocusCallbacks(logP, statusP, consoleP)

	stage := modal.NewModal(idModal, []*modal.DraggableWidget{
		newStageWindow("signal-log", "Signal Log", logP.focus, 3, 1, 82, 21, modalOpts),
		newStageWindow("system-metrics", "System Metrics", statusP.focus, 5, 24, 44, 13, modalOpts),
		newStageWindow("diagnostics-console", "Diagnostics Console", consoleP.focus, 52, 24, 50, 13, modalOpts),
	}, modalOpts)
	if err := manager.ShowModal(stage, c); err != nil {
		return fmt.Errorf("show stage modal: %w", err)
	}
	if err := renderCatalog(catalog, idx); err != nil {
		return fmt.Errorf("catalog: %w", err)
	}
	if err := renderInfo(info, idx); err != nil {
		return fmt.Errorf("info: %w", err)
	}
	go pulseBorder(c, "root", 12, 22*time.Millisecond)
	return nil
}

// onUnlock swaps the locked window set for the animated fx window set.
func onUnlock(
	c *container.Container,
	cw *contentWidgets,
	ls *lockState,
	catalog, info *text.Text,
	state *demoState,
	manager *modal.Manager,
	modalOpts *modal.Options,
) {
	active := state.getActive()

	if err := switchEffect(active, c, cw, catalog, info, manager, modalOpts); err != nil {
		log.Printf("onUnlock switchEffect: %v", err)
		return
	}

	// Reset auto-advance timer for a fresh window after unlock.
	state.setActive(active)

	// Prevent animateDemo from re-switching on the very next tick.
	ls.setJustUnlocked()

	// Celebrate with a bright border pulse.
	go pulseBorder(c, "root", 18, 16*time.Millisecond)
}

// pulseBorder briefly animates a container's border color to signal a switch.
func pulseBorder(c *container.Container, id string, steps int, tick time.Duration) {
	for i := 0; i <= steps; i++ {
		t := math.Sin(float64(i) / float64(steps) * math.Pi)
		brightness := 232 + int(t*23)
		_ = c.Update(id, container.BorderColor(cell.ColorNumber(brightness)))
		time.Sleep(tick)
	}
	_ = c.Update(id, container.BorderColor(cell.ColorNumber(69)))
}

// ── animation loop ────────────────────────────────────────────────────────────

func animateDemo(
	ctx context.Context,
	c *container.Container,
	cw *contentWidgets,
	ls *lockState,
	catalog, info *text.Text,
	state *demoState,
	manager *modal.Manager,
	modalOpts *modal.Options,
) {
	ticker := time.NewTicker(redrawTick)
	defer ticker.Stop()

	lastActive := -1

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ls.isLocked() {
				lastActive = -1
				continue
			}

			// onUnlock already performed the initial switch; sync lastActive
			// and skip this tick to avoid a redundant double-switch.
			if ls.consumeJustUnlocked() {
				lastActive = state.getActive()
				continue
			}

			if state.sinceSwitch() >= autoAdvance {
				state.next()
			}

			active := state.getActive()
			if active != lastActive {
				if err := switchEffect(active, c, cw, catalog, info, manager, modalOpts); err != nil {
					log.Printf("switchEffect: %v", err)
				}
				lastActive = active
			}
		}
	}
}

// ── keyboard ──────────────────────────────────────────────────────────────────

func handleKeyboard(
	k *terminalapi.Keyboard,
	cancel context.CancelFunc,
	c *container.Container,
	cw *contentWidgets,
	ls *lockState,
	catalog, info *text.Text,
	state *demoState,
	sel *numSelector,
	manager *modal.Manager,
	modalOpts *modal.Options,
) {
	switch k.Key {
	case keyboard.KeyEsc, keyboard.KeyCtrlC:
		cancel()
		return
	}

	if ls.isLocked() {
		switch {
		case k.Key == keyboard.KeyEnter:
			if ls.submit() {
				onUnlock(c, cw, ls, catalog, info, state, manager, modalOpts)
			}
		case k.Key == keyboard.KeyBackspace || k.Key == keyboard.KeyBackspace2:
			ls.backspace()
		default:
			r := rune(k.Key)
			if r >= 32 && r <= 126 {
				ls.push(r)
			}
		}
		return
	}

	switch k.Key {
	case 'q', 'Q':
		cancel()
	case keyboard.KeyTab, keyboard.KeyArrowRight:
		state.next()
	case keyboard.KeyArrowLeft, 'h', 'H':
		state.prev()
	case ' ':
		active := state.getActive()
		if err := switchEffect(active, c, cw, catalog, info, manager, modalOpts); err != nil {
			log.Printf("replay: %v", err)
		}
	default:
		r := rune(k.Key)
		if r >= '0' && r <= '9' {
			sel.push(r)
		}
	}
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	term, err := tcell.New()
	if err != nil {
		log.Fatalf("terminal: %v", err)
	}
	defer term.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cw, err := newContentWidgets()
	if err != nil {
		log.Fatalf("content widgets: %v", err)
	}

	catalog, info, controls, err := newSidebarWidgets()
	if err != nil {
		log.Fatalf("sidebar widgets: %v", err)
	}

	// Initialise the lock gate and populate the locked sidebar.
	ls := newLockState()
	if err := renderLockedCatalog(catalog); err != nil {
		log.Fatalf("locked catalog: %v", err)
	}
	if err := renderLockedInfo(info); err != nil {
		log.Fatalf("locked info: %v", err)
	}
	modalOpts := demoModalOptions()

	// Default effect: "Sequence: Glitch → FadeIn" — a dramatic entrance.
	state := &demoState{}
	state.setActive(defaultEffectIdx)

	// Locked scramble widgets: content is scrambled but the FramedWidget draws
	// a clean border and title so pane labels remain legible while locked.
	//
	// Stack: text → EffectWidget(Scramble) → FramedWidget
	// Scramble runs inside Framed, so only the inner canvas is obfuscated.
	dimBorder := []fx.FramedOption{fx.FramedBorderOpts(cell.FgColor(cell.ColorNumber(238)))}

	scramStatEW, err := fx.NewLooping(cw.status, fx.Scramble(0xdeadbeef))
	if err != nil {
		log.Fatalf("scramble status: %v", err)
	}
	scramStat, err := fx.FramedNew(scramStatEW, append(dimBorder, fx.FramedTitle("System Metrics"))...)
	if err != nil {
		log.Fatalf("framed scramble status: %v", err)
	}

	scramConsEW, err := fx.NewLooping(cw.console, fx.Scramble(0xcafebabe))
	if err != nil {
		log.Fatalf("scramble console: %v", err)
	}
	scramCons, err := fx.FramedNew(scramConsEW, append(dimBorder, fx.FramedTitle("Diagnostics Console"))...)
	if err != nil {
		log.Fatalf("framed scramble console: %v", err)
	}

	// Build the fixed shell; idModal hosts the draggable stage windows.
	root, err := buildLayout(term, catalog, info, controls)
	if err != nil {
		log.Fatalf("layout: %v", err)
	}

	sel := newNumSelector(func(n int) {
		state.setActive(n - 1)
	})

	// Enable mouse motion so the modal window can be dragged.
	term.EnableMouseMotion()

	// Build the locked modal stage. Every stage pane is a draggable modal
	// window so minimize/restore and styling are consistent from the start.
	pwdItem := newStageWindow("password", "🔒  Access Restricted", ls.promptW, 4, 1, 42, 15, modalOpts)
	statusItem := newStageWindow("system-metrics", "System Metrics", scramStat, 4, 18, 46, 13, modalOpts)
	consoleItem := newStageWindow("diagnostics-console", "Diagnostics Console", scramCons, 53, 18, 52, 13, modalOpts)
	modalWidget := modal.NewModal(idModal, []*modal.DraggableWidget{pwdItem, statusItem, consoleItem}, modalOpts)

	manager := &modal.Manager{}
	if err := manager.ShowModal(modalWidget, root); err != nil {
		log.Fatalf("show modal: %v", err)
	}

	go animateDemo(ctx, root, cw, ls, catalog, info, state, manager, modalOpts)

	if err := termdash.Run(
		ctx,
		term,
		root,
		termdash.RedrawInterval(redrawTick),
		termdash.KeyboardSubscriber(func(k *terminalapi.Keyboard) {
			handleKeyboard(k, cancel, root, cw, ls, catalog, info, state, sel, manager, modalOpts)
		}),
	); err != nil {
		log.Fatalf("termdash: %v", err)
	}
}
