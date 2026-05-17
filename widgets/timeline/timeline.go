// Package timeline provides a scrollable event-log widget for termdash.
//
// A Timeline displays a list of timestamped Event entries and supports keyboard
// and mouse navigation.  It implements widgetapi.Widget and can be placed
// directly into any termdash container — no separate manager or event-handler
// types are needed.
//
// # Basic usage
//
//	tl, _ := timeline.New()
//	tl.SetEvents([]timeline.Event{...})
//	container.PlaceWidget(tl)
//
// # With options
//
//	tl, _ := timeline.New(
//	    timeline.FollowTail(),      // pin view to newest event
//	    timeline.MaxEvents(500),    // cap ring buffer to prevent unbounded growth
//	)
package timeline

import (
	"fmt"
	"image"
	"strings"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Severity indicates the importance level of an Event.
type Severity int

const (
	// SeverityDebug is the lowest tier — routine diagnostic noise.
	SeverityDebug Severity = iota
	// SeverityInfo marks normal operational events.
	SeverityInfo
	// SeverityWarn flags conditions that may require attention.
	SeverityWarn
	// SeverityError indicates a failure that needs investigation.
	SeverityError
	// SeverityCritical is the highest tier — rings the terminal bell on arrival.
	SeverityCritical
)

// SeverityGlyph returns the single-rune prefix that identifies the severity tier.
func SeverityGlyph(s Severity) string {
	switch s {
	case SeverityDebug:
		return " "
	case SeverityInfo:
		return "●"
	case SeverityWarn:
		return "▲"
	case SeverityError:
		return "✖"
	case SeverityCritical:
		return "⚡"
	default:
		return " "
	}
}

// SeverityColor returns the foreground cell color for each severity tier.
func SeverityColor(s Severity) cell.Color {
	switch s {
	case SeverityDebug:
		return cell.ColorGray
	case SeverityInfo:
		return cell.ColorCyan
	case SeverityWarn:
		return cell.ColorYellow
	case SeverityError:
		return cell.ColorRed
	case SeverityCritical:
		return cell.ColorMagenta
	default:
		return cell.ColorWhite
	}
}

// SeverityName returns a fixed-width (5-char) display name for the severity
// tier, suitable for use in aligned columns (e.g. "DEBUG", "INFO ", "CRIT ").
func SeverityName(s Severity) string {
	switch s {
	case SeverityDebug:
		return "DEBUG"
	case SeverityInfo:
		return "INFO "
	case SeverityWarn:
		return "WARN "
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRIT "
	default:
		return "?????"
	}
}

// FormatMiniBar returns a short ASCII bar of the given width whose filled
// portion is proportional to count/total.  Filled cells use "▪", empty cells
// use "·".  Returns all "·" when total or count is zero.
func FormatMiniBar(count, total, width int) string {
	if total == 0 || count == 0 {
		return strings.Repeat("·", width)
	}
	filled := count * width / total
	if filled < 1 {
		filled = 1
	}
	if filled > width {
		filled = width
	}
	return strings.Repeat("▪", filled) + strings.Repeat("·", width-filled)
}

// Event is a single entry in the timeline.
type Event struct {
	Time        string    // display string shown in the log row
	Title       string
	Description string
	// Severity controls the row color and glyph prefix.  Defaults to SeverityDebug.
	Severity Severity
	// Timestamp is used for time-range filtering and by TimeRangePicker.
	// If zero, the event is never filtered out by SetTimeFilter.
	Timestamp time.Time
}

// ── Options ───────────────────────────────────────────────────────────────────

// Option configures a Timeline at construction time.
// Pass options to New():
//
//	tl, _ := timeline.New(timeline.FollowTail(), timeline.MaxEvents(200))
type Option interface {
	set(*timelineOpts)
}

// option adapts a plain function into an Option.
type option func(*timelineOpts)

func (o option) set(opts *timelineOpts) { o(opts) }

// timelineOpts holds the values set by Option instances.
type timelineOpts struct {
	followTail bool
	maxEvents  int
}

func newTimelineOpts(opts []Option) timelineOpts {
	o := timelineOpts{} // followTail=false, maxEvents=0 (unlimited) by default
	for _, opt := range opts {
		opt.set(&o)
	}
	return o
}

// FollowTail pins the view to the newest event so it behaves like a live log.
// The user can still scroll up; any new AddEvent call re-pins the view.
// Equivalent to calling SetFollowTail(true) after construction.
func FollowTail() Option {
	return option(func(o *timelineOpts) { o.followTail = true })
}

// MaxEvents caps the in-memory event ring buffer.  When the cap is reached the
// oldest events are discarded on each AddEvent call.  Values ≤ 0 mean unlimited
// (the default).  A reasonable production value is 500–2000.
func MaxEvents(n int) Option {
	return option(func(o *timelineOpts) {
		if n > 0 {
			o.maxEvents = n
		}
	})
}

// ── Widget struct ─────────────────────────────────────────────────────────────

// Timeline displays a scrollable list of timestamped events.
// It implements widgetapi.Widget and can be placed directly in any container.
type Timeline struct {
	mu            sync.Mutex
	events        []Event   // all events ever appended (capped to maxEvents when set)
	displayed     []Event   // events currently visible (filtered or same as events)
	scrollOffset  int       // index of the first visible row in displayed
	selectedIndex int       // -1 means no selection; index into displayed
	canvasHeight  int       // updated each Draw call; used by Keyboard/Mouse for bounds
	canvasWidth   int       // updated each Draw call; used for bounds checking
	followTail    bool      // when true, Draw always pins the view to the last event
	maxEvents     int       // 0 = unlimited; >0 caps the events ring buffer
	filterActive  bool
	filterStart   time.Time
	filterEnd     time.Time
	// drag state for click-and-drag scrolling
	dragActive   bool
	dragStartY   int // widget-relative Y where the drag started
	dragStartOff int // scrollOffset at drag start
	// animation
	drawFrame int // incremented every Draw call; drives the scrollbar sweep
}

// New creates a new Timeline widget.
// Pass functional options to configure initial behaviour:
//
//	tl, err := timeline.New(timeline.FollowTail(), timeline.MaxEvents(500))
func New(opts ...Option) (*Timeline, error) {
	o := newTimelineOpts(opts)
	return &Timeline{
		selectedIndex: -1,
		followTail:    o.followTail,
		maxEvents:     o.maxEvents,
	}, nil
}

// activeSlice returns the slice that Draw/keyboard/mouse should operate on.
// When a filter is active it returns the pre-built displayed subset;
// otherwise it returns the full events slice directly.
// Must be called with t.mu held.
func (t *Timeline) activeSlice() []Event {
	if t.filterActive {
		return t.displayed
	}
	return t.events
}

// SetFollowTail enables or disables tail-following mode.  When enabled the
// view is automatically pinned to the last event after every AddEvent call,
// giving the widget a live-log / ops-dashboard feel.  Manual keyboard/mouse
// scroll still works — call SetFollowTail(false) to stop following.
func (t *Timeline) SetFollowTail(follow bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.followTail = follow
}

// SetEvents replaces all events, resetting scroll position and selection.
func (t *Timeline) SetEvents(events []Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = make([]Event, len(events))
	copy(t.events, events)
	t.scrollOffset = 0
	t.selectedIndex = -1
	t.rebuildDisplayed()
}

// SetTimeFilter restricts the visible rows to events whose Timestamp falls
// within [start, end].  Events with a zero Timestamp are always hidden when a
// filter is active.  Call ClearTimeFilter to show all events again.
func (t *Timeline) SetTimeFilter(start, end time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.filterActive = true
	t.filterStart = start
	t.filterEnd = end
	t.scrollOffset = 0
	t.selectedIndex = -1
	t.rebuildDisplayed()
}

// ClearTimeFilter removes any active time filter and shows all events.
func (t *Timeline) ClearTimeFilter() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.filterActive = false
	t.scrollOffset = 0
	t.selectedIndex = -1
	t.rebuildDisplayed()
}

// rebuildDisplayed rebuilds t.displayed from t.events applying the current
// filter.  When no filter is active t.displayed is left nil (activeSlice
// returns t.events directly in that case).  Must be called with t.mu held.
func (t *Timeline) rebuildDisplayed() {
	if !t.filterActive {
		t.displayed = nil
		return
	}
	// Always allocate fresh to avoid aliasing t.events's backing array.
	filtered := make([]Event, 0, len(t.events))
	for _, e := range t.events {
		if e.Timestamp.IsZero() {
			continue
		}
		if !e.Timestamp.Before(t.filterStart) && !e.Timestamp.After(t.filterEnd) {
			filtered = append(filtered, e)
		}
	}
	t.displayed = filtered
}

// AddEvent appends a single event.
//
// If the event has SeverityCritical the terminal bell (\a) is sounded.
// If FollowTail or SetFollowTail(true) is active the scroll offset is advanced
// so the new event will be visible on the next Draw (exact clamping happens in
// Draw because only Draw knows the current canvas height).
// If MaxEvents was set and the buffer is full, the oldest event is discarded.
func (t *Timeline) AddEvent(e Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, e)

	// Enforce the ring-buffer cap: discard the oldest event if over limit.
	if t.maxEvents > 0 && len(t.events) > t.maxEvents {
		t.events = t.events[len(t.events)-t.maxEvents:]
		// Rebuild the filter view since indices shifted.
		t.rebuildDisplayed()
	} else if t.filterActive && !e.Timestamp.IsZero() &&
		!e.Timestamp.Before(t.filterStart) && !e.Timestamp.After(t.filterEnd) {
		// Fast-path: no cap triggered; just append to the displayed slice.
		t.displayed = append(t.displayed, e)
	}

	if e.Severity == SeverityCritical {
		fmt.Print("\a") // ring the terminal bell
	}
	if t.followTail {
		// Use a value larger than any valid maxOffset; Draw clamps it down.
		t.scrollOffset = len(t.activeSlice())
	}
}

// SelectedEvent returns a copy of the currently selected event, or nil.
func (t *Timeline) SelectedEvent() *Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	active := t.activeSlice()
	if t.selectedIndex < 0 || t.selectedIndex >= len(active) {
		return nil
	}
	e := active[t.selectedIndex]
	return &e
}

// EventCount returns the number of currently displayed events (filtered or all).
func (t *Timeline) EventCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.activeSlice())
}

// SeverityCounts returns per-severity counts for the currently displayed
// events (which may be filtered).  The returned array is indexed by Severity:
// index 0 = SeverityDebug … index 4 = SeverityCritical.
func (t *Timeline) SeverityCounts() [5]int {
	t.mu.Lock()
	defer t.mu.Unlock()
	var counts [5]int
	for _, e := range t.activeSlice() {
		idx := int(e.Severity)
		if idx >= 0 && idx < 5 {
			counts[idx]++
		}
	}
	return counts
}

// Draw renders the timeline onto cvs, painting one event per row.
//
// When content overflows the canvas height a right-edge scrollbar is drawn in
// the last column: ░ for the track, █ for the thumb (cyan), with ▲/▼ arrows
// at the top and bottom.  Event text is truncated one column earlier so it
// never overlaps the scrollbar.
//
// Each row is prefixed with a severity glyph and coloured according to its
// Severity tier.  The selected row is additionally bolded and its leading
// column is replaced with ">".
func (t *Timeline) Draw(cvs *canvas.Canvas, _ *widgetapi.Meta) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	height := cvs.Area().Dy()
	width := cvs.Area().Dx()
	t.canvasHeight = height
	t.canvasWidth = width

	active := t.activeSlice()

	// When following the tail always snap to the last event.
	if t.followTail && len(active) > 0 {
		maxOffset := len(active) - height
		if maxOffset < 0 {
			maxOffset = 0
		}
		t.scrollOffset = maxOffset
	}

	// Clamp scroll offset to the valid range.
	maxOffset := len(active) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if t.scrollOffset > maxOffset {
		t.scrollOffset = maxOffset
	}

	// Decide whether a scrollbar is needed and reserve the rightmost column.
	needsScrollbar := len(active) > height
	contentWidth := width
	if needsScrollbar {
		contentWidth = width - 1
	}

	// Render each visible event.
	for row := 0; row < height; row++ {
		idx := t.scrollOffset + row
		if idx >= len(active) {
			break
		}
		e := active[idx]

		glyph := SeverityGlyph(e.Severity)
		color := SeverityColor(e.Severity)

		var leader string
		if idx == t.selectedIndex {
			leader = "▶"
		} else {
			leader = " "
		}
		line := fmt.Sprintf("%s%s [%s] %s: %s", leader, glyph, e.Time, e.Title, e.Description)

		var opts []cell.Option
		if idx == t.selectedIndex {
			// Bright white bold text on a dark navy background — unmistakable.
			opts = append(opts,
				cell.FgColor(cell.ColorNumber(255)),
				cell.Bold(),
				cell.BgColor(cell.ColorNumber(17)),
			)
		} else {
			opts = append(opts, cell.FgColor(color))
		}
		t.writeLine(cvs, row, contentWidth, line, opts...)
	}

	// ── Right-edge scrollbar (animated) ──────────────────────────────────────
	t.drawFrame++
	if needsScrollbar && width >= 2 {
		sbCol := width - 1
		trackColor := cell.FgColor(cell.ColorNumber(234)) // near-black track
		arrowColor := cell.FgColor(cell.ColorNumber(240)) // mid-dark gray arrows

		// Arrow caps.
		cvs.SetCell(image.Point{X: sbCol, Y: 0}, '▲', arrowColor)
		cvs.SetCell(image.Point{X: sbCol, Y: height - 1}, '▼', arrowColor)

		// Thumb: proportional, clamped to at least 1 row.
		trackRows := height - 2
		if trackRows < 1 {
			trackRows = 1
		}
		thumbSize := trackRows * height / len(active)
		if thumbSize < 1 {
			thumbSize = 1
		}
		if thumbSize > trackRows {
			thumbSize = trackRows
		}
		thumbTop := 1
		if maxOffset > 0 {
			thumbTop += t.scrollOffset * (trackRows - thumbSize) / maxOffset
		}

		// Sweep shimmer: a bright highlight travels from the top to the bottom
		// of the thumb over sweepPeriod frames, then restarts.
		const sweepPeriod = 56
		sweepRow := (t.drawFrame % sweepPeriod) * thumbSize / sweepPeriod

		for row := 1; row < height-1; row++ {
			if row >= thumbTop && row < thumbTop+thumbSize {
				thumbIdx := row - thumbTop

				// Distance from sweep peak → determines brightness.
				d := thumbIdx - sweepRow
				if d < 0 {
					d = -d
				}
				var c cell.Color
				switch {
				case d == 0:
					c = cell.ColorNumber(159) // near white-cyan peak
				case d == 1:
					c = cell.ColorNumber(123) // light cyan
				case d <= 3:
					c = cell.ColorNumber(87) // pale cyan
				default:
					c = cell.ColorNumber(51) // standard cyan body
				}

				// Rounded cap runes give a "pill" shape to the thumb.
				var r rune
				switch {
				case thumbSize == 1:
					r = '▪'
				case thumbIdx == 0:
					r = '▀' // top cap
				case thumbIdx == thumbSize-1:
					r = '▄' // bottom cap
				default:
					r = '█'
				}
				cvs.SetCell(image.Point{X: sbCol, Y: row}, r, cell.FgColor(c))
			} else {
				cvs.SetCell(image.Point{X: sbCol, Y: row}, '░', trackColor)
			}
		}
	}

	return nil
}

// writeLine writes text to a single row of the canvas, truncating at width.
func (t *Timeline) writeLine(cvs *canvas.Canvas, row, width int, text string, opts ...cell.Option) {
	col := 0
	for _, r := range text {
		if col >= width {
			break
		}
		w, err := cvs.SetCell(image.Point{X: col, Y: row}, r, opts...)
		if err != nil {
			return
		}
		col += w
	}
}

// Keyboard handles arrow-key navigation.
//
//   - ArrowDown  — move selection one row down; scroll if needed.
//   - ArrowUp    — move selection one row up; scroll if needed.
//   - Enter / ' ' — confirm selection (initialises to the top visible row if
//     nothing is selected yet).
func (t *Timeline) Keyboard(k *terminalapi.Keyboard, _ *widgetapi.EventMeta) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	active := t.activeSlice()
	switch k.Key {
	case keyboard.KeyArrowDown:
		if t.selectedIndex == -1 {
			// Nothing selected: initialise to the topmost visible event.
			t.selectedIndex = t.scrollOffset
		} else if t.selectedIndex < len(active)-1 {
			t.selectedIndex++
		}
		// Scroll down if the selection moved below the visible window.
		if t.canvasHeight > 0 && t.selectedIndex >= t.scrollOffset+t.canvasHeight {
			t.scrollOffset = t.selectedIndex - t.canvasHeight + 1
		}

	case keyboard.KeyArrowUp:
		if t.selectedIndex == -1 {
			t.selectedIndex = t.scrollOffset
		} else if t.selectedIndex > 0 {
			t.selectedIndex--
		}
		// Scroll up if the selection moved above the visible window.
		if t.selectedIndex < t.scrollOffset {
			t.scrollOffset = t.selectedIndex
		}

	case keyboard.KeyEnter, ' ':
		// Initialise selection if nothing is selected yet.
		if t.selectedIndex == -1 && len(active) > 0 {
			t.selectedIndex = t.scrollOffset
		}
	}

	return nil
}

// Mouse handles scroll-wheel, click, and drag events.
//
// The widget uses MouseScopeGlobal so it performs its own bounds check —
// events outside the canvas area are silently ignored (mirroring the
// linechart pattern).
//
//   - WheelUp/WheelDown  — scroll the visible window (3 lines per tick).
//   - ButtonLeft press   — start a drag-to-scroll gesture; also remembers the
//     press position to distinguish a click from a drag.
//   - ButtonLeft hold+drag (ButtonLeft events with movement) — scroll by
//     dragging: moving the mouse up scrolls down, moving down scrolls up,
//     exactly like grabbing content and pulling it.
//   - ButtonLeft release (ButtonRelease with no drag movement) — select the
//     event at the clicked row (single-click behaviour preserved).
func (t *Timeline) Mouse(m *terminalapi.Mouse, _ *widgetapi.EventMeta) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Termdash delivers mouse positions in widget-relative coordinates (0,0 =
	// top-left of the widget's canvas area).  MouseScopeGlobal means we may
	// receive events while the cursor is outside our area, so we guard every
	// branch with an explicit bounds check — same pattern as linechart.
	inBounds := t.canvasHeight > 0 &&
		m.Position.X >= 0 && m.Position.X < t.canvasWidth &&
		m.Position.Y >= 0 && m.Position.Y < t.canvasHeight

	active := t.activeSlice()
	maxOffset := func() int {
		o := len(active) - t.canvasHeight
		if o < 0 {
			return 0
		}
		return o
	}

	switch m.Button {
	case mouse.ButtonWheelUp:
		if !inBounds {
			break
		}
		t.followTail = false
		if t.scrollOffset > 3 {
			t.scrollOffset -= 3
		} else {
			t.scrollOffset = 0
		}

	case mouse.ButtonWheelDown:
		if !inBounds {
			break
		}
		max := maxOffset()
		if t.scrollOffset+3 < max {
			t.scrollOffset += 3
		} else {
			t.scrollOffset = max
		}

	case mouse.ButtonLeft:
		if !inBounds {
			break
		}
		// Record drag anchor — subsequent ButtonLeft events with different Y
		// are drag moves (termdash reports held-and-moved as repeated ButtonLeft).
		if !t.dragActive {
			t.dragActive = true
			t.dragStartY = m.Position.Y
			t.dragStartOff = t.scrollOffset
		} else {
			// Drag in progress: delta in rows.
			delta := t.dragStartY - m.Position.Y // positive → scrolled down
			newOff := t.dragStartOff + delta
			if newOff < 0 {
				newOff = 0
			}
			if mo := maxOffset(); newOff > mo {
				newOff = mo
			}
			t.scrollOffset = newOff
		}

	case mouse.ButtonRelease:
		if t.dragActive {
			// If mouse barely moved, treat as a click → select the row.
			if abs(m.Position.Y-t.dragStartY) <= 1 && inBounds {
				idx := t.scrollOffset + m.Position.Y
				if idx >= 0 && idx < len(active) {
					t.selectedIndex = idx
				}
			}
			t.dragActive = false
		}
	}

	return nil
}

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Options returns the widget's termdash options.
func (t *Timeline) Options() widgetapi.Options {
	return widgetapi.Options{
		WantKeyboard: widgetapi.KeyScopeGlobal,
		// MouseScopeGlobal mirrors the linechart pattern: receive all mouse
		// events and apply a bounds check inside Mouse() so the widget handles
		// scroll-wheel and drag correctly regardless of focus state.
		WantMouse: widgetapi.MouseScopeGlobal,
	}
}

// Ensure Timeline implements widgetapi.Widget at compile time.
var _ widgetapi.Widget = (*Timeline)(nil)
