package timeline

import (
	"image"
	"strings"
	"testing"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// sampleEvents returns a small set of events for use in tests.
func sampleEvents() []Event {
	return []Event{
		{Time: "09:00", Title: "Event 1", Description: "First"},
		{Time: "09:05", Title: "Event 2", Description: "Second"},
		{Time: "09:10", Title: "Event 3", Description: "Third"},
		{Time: "09:15", Title: "Event 4", Description: "Fourth"},
		{Time: "09:20", Title: "Event 5", Description: "Fifth"},
	}
}

func newTestTimeline(t *testing.T) *Timeline {
	t.Helper()
	tl, err := New()
	if err != nil {
		t.Fatalf("New() => unexpected error: %v", err)
	}
	tl.SetEvents(sampleEvents())
	return tl
}

// drawToString renders the widget onto a canvas of the given size and returns
// the text of every non-blank cell as a concatenated string (row-major order).
func drawToString(t *testing.T, tl *Timeline, w, h int) string {
	t.Helper()
	cvs, err := canvas.New(image.Rect(0, 0, w, h))
	if err != nil {
		t.Fatalf("canvas.New() => unexpected error: %v", err)
	}
	if err := tl.Draw(cvs, &widgetapi.Meta{}); err != nil {
		t.Fatalf("Draw() => unexpected error: %v", err)
	}
	var sb strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c, err := cvs.Cell(image.Point{X: x, Y: y})
			if err != nil {
				t.Fatalf("Cell(%d,%d) => unexpected error: %v", x, y, err)
			}
			if c.Rune != 0 {
				sb.WriteRune(c.Rune)
			}
		}
	}
	return sb.String()
}

// TestNew verifies that New returns a non-nil widget.
func TestNew(t *testing.T) {
	tl, err := New()
	if err != nil {
		t.Fatalf("New() => unexpected error: %v", err)
	}
	if tl == nil {
		t.Fatal("New() returned nil widget")
	}
}

// TestSetEventsAndAddEvent verifies event mutation helpers.
func TestSetEventsAndAddEvent(t *testing.T) {
	tl := newTestTimeline(t)

	tl.AddEvent(Event{Time: "09:25", Title: "Event 6", Description: "Sixth"})

	tl.mu.Lock()
	got := len(tl.events)
	tl.mu.Unlock()

	if got != 6 {
		t.Fatalf("len(events) = %d, want 6", got)
	}
}

// TestDrawRendersEvents verifies that Draw writes event text to the canvas.
func TestDrawRendersEvents(t *testing.T) {
	tl := newTestTimeline(t)
	content := drawToString(t, tl, 60, 10)

	if !strings.Contains(content, "Event 1") {
		t.Errorf("rendered output does not contain 'Event 1'; got: %q", content)
	}
	if !strings.Contains(content, "09:05") {
		t.Errorf("rendered output does not contain '09:05'; got: %q", content)
	}
}

// TestDrawSelectedHighlight verifies that the selected event uses the current
// selection marker rune.
func TestDrawSelectedHighlight(t *testing.T) {
	tl := newTestTimeline(t)

	tl.mu.Lock()
	tl.selectedIndex = 1
	tl.mu.Unlock()

	content := drawToString(t, tl, 60, 10)
	if !strings.Contains(content, "▶") {
		t.Errorf("selected row should start with '▶'; got: %q", content)
	}
}

// TestKeyboardArrowDownMovesSelection checks that ArrowDown initialises the
// selection on the first press and moves it downward on subsequent presses.
func TestKeyboardArrowDownMovesSelection(t *testing.T) {
	tl := newTestTimeline(t)

	// First press: initialise to scrollOffset (0) rather than move.
	if err := tl.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, nil); err != nil {
		t.Fatalf("Keyboard(ArrowDown) => unexpected error: %v", err)
	}
	if got := tl.selectedIndex; got != 0 {
		t.Fatalf("selectedIndex after first ArrowDown = %d, want 0", got)
	}

	// Second press: move down to index 1.
	if err := tl.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, nil); err != nil {
		t.Fatalf("Keyboard(ArrowDown) => unexpected error: %v", err)
	}
	if got := tl.selectedIndex; got != 1 {
		t.Fatalf("selectedIndex after second ArrowDown = %d, want 1", got)
	}
}

// TestKeyboardArrowUpMovesSelection checks ArrowUp navigation.
func TestKeyboardArrowUpMovesSelection(t *testing.T) {
	tl := newTestTimeline(t)
	tl.mu.Lock()
	tl.selectedIndex = 2
	tl.mu.Unlock()

	if err := tl.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowUp}, nil); err != nil {
		t.Fatalf("Keyboard(ArrowUp) => unexpected error: %v", err)
	}
	if got := tl.selectedIndex; got != 1 {
		t.Fatalf("selectedIndex = %d, want 1", got)
	}
}

// TestKeyboardArrowUpAtTopDoesNotUnderflow ensures ArrowUp at index 0 stays at 0.
func TestKeyboardArrowUpAtTopDoesNotUnderflow(t *testing.T) {
	tl := newTestTimeline(t)
	tl.mu.Lock()
	tl.selectedIndex = 0
	tl.mu.Unlock()

	if err := tl.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowUp}, nil); err != nil {
		t.Fatalf("Keyboard(ArrowUp) => unexpected error: %v", err)
	}
	if got := tl.selectedIndex; got != 0 {
		t.Fatalf("selectedIndex = %d, want 0 (no underflow)", got)
	}
}

// TestKeyboardArrowDownAtBottomDoesNotOverflow ensures ArrowDown at the last
// event stays at the last index.
func TestKeyboardArrowDownAtBottomDoesNotOverflow(t *testing.T) {
	tl := newTestTimeline(t)
	tl.mu.Lock()
	tl.selectedIndex = len(tl.events) - 1
	tl.mu.Unlock()

	if err := tl.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, nil); err != nil {
		t.Fatalf("Keyboard(ArrowDown) => unexpected error: %v", err)
	}
	tl.mu.Lock()
	want := len(tl.events) - 1
	tl.mu.Unlock()
	if got := tl.selectedIndex; got != want {
		t.Fatalf("selectedIndex = %d, want %d (no overflow)", got, want)
	}
}

// TestMouseWheelScrolls verifies wheel-up/down scrolling.
func TestMouseWheelScrolls(t *testing.T) {
	tl, err := New()
	if err != nil {
		t.Fatalf("New() => unexpected error: %v", err)
	}
	// Use 20 events so scrolling is actually possible.
	events := make([]Event, 20)
	for i := range events {
		events[i] = Event{Title: "E", Time: "00:00", Description: "x"}
	}
	tl.SetEvents(events)
	tl.mu.Lock()
	tl.canvasHeight = 5
	tl.canvasWidth = 40
	tl.mu.Unlock()

	// Scroll down. The widget scrolls in 3-row steps.
	if err := tl.Mouse(&terminalapi.Mouse{Button: mouse.ButtonWheelDown}, nil); err != nil {
		t.Fatalf("Mouse(WheelDown) => unexpected error: %v", err)
	}
	if got := tl.scrollOffset; got != 3 {
		t.Fatalf("scrollOffset after WheelDown = %d, want 3", got)
	}

	// Scroll back up to the top.
	if err := tl.Mouse(&terminalapi.Mouse{Button: mouse.ButtonWheelUp}, nil); err != nil {
		t.Fatalf("Mouse(WheelUp) => unexpected error: %v", err)
	}
	if got := tl.scrollOffset; got != 0 {
		t.Fatalf("scrollOffset after WheelUp = %d, want 0", got)
	}
}

// TestMouseClickSelectsEvent verifies that a press/release click selects the
// event at the clicked row.
func TestMouseClickSelectsEvent(t *testing.T) {
	tl := newTestTimeline(t)
	tl.mu.Lock()
	tl.canvasHeight = 10
	tl.canvasWidth = 40
	tl.mu.Unlock()

	press := &terminalapi.Mouse{
		Button:   mouse.ButtonLeft,
		Position: image.Point{X: 0, Y: 2}, // row 2 → event index 2
	}
	if err := tl.Mouse(press, nil); err != nil {
		t.Fatalf("Mouse(ButtonLeft) => unexpected error: %v", err)
	}
	release := &terminalapi.Mouse{
		Button:   mouse.ButtonRelease,
		Position: image.Point{X: 0, Y: 2},
	}
	if err := tl.Mouse(release, nil); err != nil {
		t.Fatalf("Mouse(ButtonRelease) => unexpected error: %v", err)
	}
	if got := tl.selectedIndex; got != 2 {
		t.Fatalf("selectedIndex = %d, want 2", got)
	}
}

// TestSelectedEventReturnsCorrectEvent checks SelectedEvent after navigation.
func TestSelectedEventReturnsCorrectEvent(t *testing.T) {
	tl := newTestTimeline(t)
	tl.mu.Lock()
	tl.selectedIndex = 1
	tl.mu.Unlock()

	e := tl.SelectedEvent()
	if e == nil {
		t.Fatal("SelectedEvent() = nil, want Event 2")
	}
	if e.Title != "Event 2" {
		t.Fatalf("SelectedEvent().Title = %q, want %q", e.Title, "Event 2")
	}
}

// TestSelectedEventNilWhenUnset verifies no selection returns nil.
func TestSelectedEventNilWhenUnset(t *testing.T) {
	tl := newTestTimeline(t)
	if e := tl.SelectedEvent(); e != nil {
		t.Fatalf("SelectedEvent() = %+v, want nil", e)
	}
}

// TestOptionsWantsKeyboardAndMouse verifies the Options return value.
func TestOptionsWantsKeyboardAndMouse(t *testing.T) {
	tl := newTestTimeline(t)
	opts := tl.Options()
	if opts.WantKeyboard == widgetapi.KeyScopeNone {
		t.Errorf("Options().WantKeyboard = %v, want non-zero", opts.WantKeyboard)
	}
	if opts.WantMouse == widgetapi.MouseScopeNone {
		t.Errorf("Options().WantMouse = %v, want non-zero", opts.WantMouse)
	}
}
