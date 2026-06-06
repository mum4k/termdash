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

package timeline

// TimeRangePicker is a horizontal timeline scrubber widget.
//
// Visual layout (7 rows):
//
//	Row 0   15:00:00      15:02:00      15:04:00      15:06:00
//	Row 1  ───────────────┬─────────────┬─────────────┬──────────
//	Row 2                 │      ▼      │      ▼      │
//	Row 3                 │      │      │      │      │
//	Row 4   ●  ▲  ⚡  ●   │  ▲  ●  ⚡   │  ▲  ●  ✖   │
//	Row 5   ░░░░░░░░░░░░░░│█████████████│░░░░░░░░░░░░░░░░░░░░░░
//	Row 6   ▼ 15:02:00 ──── ▼ 15:04:00  ·  34 events  ·  [r]
//
// Row 0 – centered timestamps, no decoration.
// Row 1 – box-drawing ruler: ─ across the full width, ┬ at each tick column.
//         Selection-boundary columns show cyan ┬ so they anchor into the axis.
// Row 2 – notch A: │ at tick columns, ▼ at selection boundaries (cyan).
// Row 3 – notch B: │ at tick columns and selection boundaries.
// Row 4 – event glyph strip: highest-severity glyph per column; │ at ticks/selection.
// Row 5 – fill bar: ░ unselected, █ selected, │ at boundaries.
// Row 6 – status / instructions.
//
// Interaction:
//
//	First left-click  – set range start (▼ appears)
//	Second left-click – set range end   (range highlighted, OnChange fires)
//	Third left-click  – reset selection
//	r / R / Escape    – reset selection
//
// OnChange is called on every state transition and can be wired directly to
// Timeline.SetTimeFilter / ClearTimeFilter.

import (
	"fmt"
	"image"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// TimeRangePicker is a horizontal timeline scrubber widget.
type TimeRangePicker struct {
	mu         sync.Mutex
	events     []Event   // snapshot updated by SetPickerEvents / AddPickerEvent
	timeMin    time.Time // earliest timestamp across all events
	timeMax    time.Time // latest  timestamp across all events
	selStart   time.Time
	selEnd     time.Time
	clickCount int // 0=none 1=start set 2=range set
	width      int // canvas width captured during last Draw

	// drag state
	dragActive   bool // true while ButtonLeft is held
	dragStartX   int  // x-column where the drag began
	dragCurrentX int  // x-column of the latest ButtonLeft move (for ghost pin)

	// onChange is called (in its own goroutine) whenever the selection changes.
	onChange func(start, end time.Time, hasRange bool)
}

// NewTimeRangePicker creates an empty TimeRangePicker.
//
// onChange is called whenever the selection changes: first click sets the start
// pin, drag or second click sets the end pin, r/R/Esc resets.  Pass nil if you
// prefer to poll Selection() instead.
//
//	picker, err := timeline.NewTimeRangePicker(func(start, end time.Time, hasRange bool) {
//	    if hasRange {
//	        tl.SetTimeFilter(start, end)
//	    } else {
//	        tl.ClearTimeFilter()
//	    }
//	})
func NewTimeRangePicker(onChange func(start, end time.Time, hasRange bool)) (*TimeRangePicker, error) {
	return &TimeRangePicker{onChange: onChange}, nil
}

// SetPickerEvents replaces the event snapshot and recalculates the time span.
func (p *TimeRangePicker) SetPickerEvents(events []Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = make([]Event, len(events))
	copy(p.events, events)
	p.recalcSpan()
}

// AddPickerEvent appends one event and expands the time span if needed.
func (p *TimeRangePicker) AddPickerEvent(e Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, e)
	if e.Timestamp.IsZero() {
		return
	}
	if p.timeMin.IsZero() || e.Timestamp.Before(p.timeMin) {
		p.timeMin = e.Timestamp
	}
	if e.Timestamp.After(p.timeMax) {
		p.timeMax = e.Timestamp
	}
}

// HasRange reports whether a complete (start+end) range is selected.
func (p *TimeRangePicker) HasRange() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.clickCount == 2
}

// Selection returns the current selection endpoints.
// hasRange is false if no complete range has been set.
func (p *TimeRangePicker) Selection() (start, end time.Time, hasRange bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.selStart, p.selEnd, p.clickCount == 2
}

// ── internal helpers ─────────────────────────────────────────────────────────

func (p *TimeRangePicker) recalcSpan() {
	p.timeMin = time.Time{}
	p.timeMax = time.Time{}
	for _, e := range p.events {
		if e.Timestamp.IsZero() {
			continue
		}
		if p.timeMin.IsZero() || e.Timestamp.Before(p.timeMin) {
			p.timeMin = e.Timestamp
		}
		if e.Timestamp.After(p.timeMax) {
			p.timeMax = e.Timestamp
		}
	}
}

// xToTime maps column x (0…width-1) → time in [timeMin, timeMax].
func (p *TimeRangePicker) xToTime(x, w int) time.Time {
	if w <= 1 || p.timeMin.IsZero() || p.timeMax.Equal(p.timeMin) {
		return p.timeMin
	}
	span := p.timeMax.Sub(p.timeMin)
	offset := time.Duration(float64(span) * float64(x) / float64(w-1))
	return p.timeMin.Add(offset)
}

// timeToX maps t → column x (0…width-1).
func (p *TimeRangePicker) timeToX(t time.Time, w int) int {
	if w <= 1 || p.timeMin.IsZero() || p.timeMax.Equal(p.timeMin) {
		return 0
	}
	span := p.timeMax.Sub(p.timeMin)
	ratio := float64(t.Sub(p.timeMin)) / float64(span)
	x := int(ratio * float64(w-1))
	if x < 0 {
		return 0
	}
	if x >= w {
		return w - 1
	}
	return x
}

func (p *TimeRangePicker) reset() {
	p.clickCount = 0
	p.selStart = time.Time{}
	p.selEnd = time.Time{}
	p.dragActive = false
	p.dragCurrentX = -1
}

func (p *TimeRangePicker) fireOnChange() {
	if p.onChange == nil {
		return
	}
	start, end, has := p.selStart, p.selEnd, p.clickCount == 2
	go p.onChange(start, end, has)
}

// ── widgetapi.Widget ─────────────────────────────────────────────────────────

// Draw renders the seven-row scrubber.
func (p *TimeRangePicker) Draw(cvs *canvas.Canvas, _ *widgetapi.Meta) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	w := cvs.Area().Dx()
	h := cvs.Area().Dy()
	p.width = w

	if w < 8 || h < 1 {
		return nil
	}

	noSpan := p.timeMin.IsZero() || p.timeMax.Equal(p.timeMin)

	// Pre-compute selection boundary columns (–1 = not set).
	sx, ex := -1, -1
	if p.clickCount >= 1 {
		sx = p.timeToX(p.selStart, w)
	}
	if p.clickCount >= 2 {
		ex = p.timeToX(p.selEnd, w)
	}
	isBoundary := func(x int) bool { return x >= 0 && (x == sx || x == ex) }

	// Ghost pin: shown whenever a drag is active and the cursor has moved right
	// of the start.  clickCount stays 1 during the drag; it only flips to 2 on
	// ButtonRelease, at which point dragActive becomes false and gx drops to -1.
	gx := -1
	if p.dragActive && p.dragCurrentX > p.dragStartX {
		gx = p.dragCurrentX
		if gx >= w {
			gx = w - 1
		}
	}
	ghost := []cell.Option{cell.FgColor(cell.ColorNumber(240))} // dim grey

	// ── Build ruler ticks ─────────────────────────────────────────────────────
	type tick struct {
		x     int
		label string
	}
	var ticks []tick
	tickAt := make(map[int]bool)
	if !noSpan {
		const labelWidth = 8
		const minGap = labelWidth + 4
		numTicks := w / minGap
		if numTicks < 2 {
			numTicks = 2
		}
		if numTicks > 10 {
			numTicks = 10
		}
		span := p.timeMax.Sub(p.timeMin)
		for i := 0; i < numTicks; i++ {
			ratio := float64(i) / float64(numTicks-1)
			col := int(ratio * float64(w-1))
			t := p.timeMin.Add(time.Duration(float64(span) * ratio))
			ticks = append(ticks, tick{x: col, label: t.Format("15:04:05")})
			tickAt[col] = true
		}
	}

	// Convenience wrappers.
	gray := cell.FgColor(cell.ColorGray)
	cyan := []cell.Option{cell.FgColor(cell.ColorCyan), cell.Bold()}

	put := func(x, y int, r rune, opts ...cell.Option) {
		if x >= 0 && x < w && y >= 0 && y < h {
			cvs.SetCell(image.Point{X: x, Y: y}, r, opts...)
		}
	}
	// putV draws a vertical glyph: cyan if x is a selection boundary, gray otherwise.
	putV := func(x, y int, normalR, selR rune) {
		if isBoundary(x) {
			put(x, y, selR, cyan...)
		} else {
			put(x, y, normalR, gray)
		}
	}

	// ── Row 0: labels, centered above each tick ───────────────────────────────
	if h >= 1 {
		if noSpan {
			p.writeAt(cvs, 0, 0, "  waiting for events…", gray)
		} else {
			for _, tk := range ticks {
				start := tk.x - len(tk.label)/2
				if start < 0 {
					start = 0
				}
				if start+len(tk.label) > w {
					start = w - len(tk.label)
				}
				p.writeAt(cvs, start, 0, tk.label, gray)
			}
		}
	}

	if noSpan {
		if h >= 7 {
			p.writeAt(cvs, 0, 6, "  waiting for event data…", gray)
		}
		return nil
	}

	// ── Row 1: ─── axis with ┬ at tick columns ───────────────────────────────
	// Selection boundaries get a cyan ┬; ghost column gets a dim-grey ┬.
	if h >= 2 {
		for x := 0; x < w; x++ {
			if isBoundary(x) {
				put(x, 1, '┬', cyan...)
			} else if x == gx {
				put(x, 1, '┬', ghost...)
			} else if tickAt[x] {
				put(x, 1, '┬', gray)
			} else {
				put(x, 1, '─', gray)
			}
		}
	}

	// ── Row 2: notch A — │ at ticks, ▼ at selection boundaries ──────────────
	if h >= 3 {
		for _, tk := range ticks {
			putV(tk.x, 2, '│', '▼')
		}
		if sx >= 0 {
			put(sx, 2, '▼', cyan...)
		}
		if ex >= 0 {
			put(ex, 2, '▼', cyan...)
		}
		// Ghost ▼ at drag cursor.
		if gx >= 0 {
			put(gx, 2, '▼', ghost...)
		}
	}

	// ── Row 3: notch B — │ at ticks and selection boundaries ─────────────────
	if h >= 4 {
		for _, tk := range ticks {
			putV(tk.x, 3, '│', '│')
		}
		if sx >= 0 {
			put(sx, 3, '│', cyan...)
		}
		if ex >= 0 {
			put(ex, 3, '│', cyan...)
		}
		if gx >= 0 {
			put(gx, 3, '│', ghost...)
		}
	}

	// ── Row 4: event glyph strip ──────────────────────────────────────────────
	if h >= 5 {
		type col struct {
			sev   Severity
			count int
		}
		cols := make([]col, w)
		for _, e := range p.events {
			if e.Timestamp.IsZero() {
				continue
			}
			x := p.timeToX(e.Timestamp, w)
			cols[x].count++
			if e.Severity > cols[x].sev {
				cols[x].sev = e.Severity
			}
		}
		for x := 0; x < w; x++ {
			if cols[x].count > 0 {
				r := []rune(SeverityGlyph(cols[x].sev))[0]
				put(x, 4, r, cell.FgColor(SeverityColor(cols[x].sev)))
			}
		}
		// Tick and boundary verticals overlay glyphs.
		for _, tk := range ticks {
			putV(tk.x, 4, '│', '│')
		}
		if sx >= 0 {
			put(sx, 4, '│', cyan...)
		}
		if ex >= 0 {
			put(ex, 4, '│', cyan...)
		}
		if gx >= 0 {
			put(gx, 4, '│', ghost...)
		}
	}

	// ── Row 5: fill bar ───────────────────────────────────────────────────────
	if h >= 6 {
		hasRange := p.clickCount == 2 && sx >= 0 && ex >= 0
		for x := 0; x < w; x++ {
			switch {
			case isBoundary(x):
				put(x, 5, '│', cyan...)
			case hasRange && x > sx && x < ex:
				put(x, 5, '█', cell.FgColor(cell.ColorCyan))
			// Ghost fill: dim ▒ between start pin and ghost cursor while dragging.
			case gx >= 0 && sx >= 0 && x > sx && x < gx:
				put(x, 5, '▒', ghost...)
			case gx >= 0 && x == gx:
				put(x, 5, '│', ghost...)
			default:
				put(x, 5, '░', gray)
			}
		}
	}

	// ── Row 6: status line ────────────────────────────────────────────────────
	if h >= 7 {
		var status string
		switch p.clickCount {
		case 0:
			status = "  click to set range start  ·  [r] reset"
		case 1:
			if gx >= 0 {
				// Actively dragging: show live ghost timestamp.
				ghostTime := p.xToTime(gx, w)
				status = fmt.Sprintf("  ▼ %s  ──  dragging ──  %s  ·  [r] reset",
					p.selStart.Format("15:04:05"), ghostTime.Format("15:04:05"))
			} else {
				status = fmt.Sprintf("  ▼ %s  ·  drag to set range end  ·  [r] reset",
					p.selStart.Format("15:04:05"))
			}
		case 2:
			n := 0
			for _, e := range p.events {
				if !e.Timestamp.IsZero() &&
					!e.Timestamp.Before(p.selStart) &&
					!e.Timestamp.After(p.selEnd) {
					n++
				}
			}
			status = fmt.Sprintf("  ▼ %s  ────  ▼ %s   %d events   [r] reset",
				p.selStart.Format("15:04:05"), p.selEnd.Format("15:04:05"), n)
		}
		p.writeAt(cvs, 0, 6, status, cell.FgColor(cell.ColorYellow))
	}

	return nil
}

// writeAt writes text at column x, row y, truncating at the canvas edge.
func (p *TimeRangePicker) writeAt(cvs *canvas.Canvas, x, y int, text string, opts ...cell.Option) {
	w := cvs.Area().Dx()
	for _, r := range text {
		if x >= w {
			break
		}
		adv, err := cvs.SetCell(image.Point{X: x, Y: y}, r, opts...)
		if err != nil {
			return
		}
		x += adv
	}
}

// Mouse handles click-and-drag to set a time range.
//
// Interaction model:
//   - ButtonLeft press  : starts a new drag session; pins the start at the
//     click column (any prior selection is cleared).
//   - ButtonLeft held + move right : live-updates the end pin so the selection
//     grows in real time as the user drags.
//   - ButtonRelease     : finalises the selection.
//     If the release column is more than 1 cell to the right of the press
//     column the range is committed (start+end both set).
//     If the mouse barely moved it is treated as a plain click that sets only
//     the start pin, leaving the user free to drag next time.
//
// WantMouse is MouseScopeGlobal so ButtonRelease is always delivered even when
// the pointer drifts outside the widget during the drag.
func (p *TimeRangePicker) Mouse(m *terminalapi.Mouse, _ *widgetapi.EventMeta) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.timeMin.IsZero() || p.width <= 0 {
		return nil
	}

	// clamp x to valid column range
	clamp := func(x int) int {
		if x < 0 {
			return 0
		}
		if x >= p.width {
			return p.width - 1
		}
		return x
	}

	switch m.Button {
	case mouse.ButtonLeft:
		x := clamp(m.Position.X)
		if !p.dragActive {
			// Press: start a fresh drag session regardless of prior state.
			p.dragActive = true
			p.dragStartX = x
			p.dragCurrentX = x
			p.selStart = p.xToTime(x, p.width)
			p.selEnd = time.Time{}
			p.clickCount = 1
			p.fireOnChange()
		} else {
			// Held + moved: only track cursor for the ghost pin.
			// Do NOT commit clickCount=2 here — that happens on ButtonRelease so
			// the ghost stays grey until the user lets go.
			p.dragCurrentX = x
		}

	case mouse.ButtonRelease:
		if !p.dragActive {
			return nil
		}
		x := clamp(m.Position.X)
		if x > p.dragStartX+1 {
			// Meaningful drag → commit the range as real cyan pins.
			p.selEnd = p.xToTime(x, p.width)
			p.clickCount = 2
		}
		// else: tiny movement → leave as start-only (clickCount stays 1).
		p.dragActive = false
		p.dragCurrentX = -1
		p.fireOnChange()

	case mouse.ButtonMiddle, mouse.ButtonRight:
		// r/R/Escape still handles reset via Keyboard; ignore other buttons.
	}
	return nil
}

// Keyboard handles r/R/Escape to reset the selection.
func (p *TimeRangePicker) Keyboard(k *terminalapi.Keyboard, _ *widgetapi.EventMeta) error {
	if k.Key != 'r' && k.Key != 'R' && k.Key != keyboard.KeyEsc {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.clickCount == 0 {
		return nil
	}
	p.reset()
	p.fireOnChange()
	return nil
}

// Options declares the widget's needs.
func (p *TimeRangePicker) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize: image.Point{X: 20, Y: 7},
		// MouseScopeGlobal ensures ButtonRelease is delivered even when the
		// pointer leaves the widget mid-drag.
		WantMouse:    widgetapi.MouseScopeGlobal,
		WantKeyboard: widgetapi.KeyScopeGlobal,
	}
}

// Ensure TimeRangePicker satisfies widgetapi.Widget.
var _ widgetapi.Widget = (*TimeRangePicker)(nil)
