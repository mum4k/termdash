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

// Package tab contains tests for the tabbed interface helpers.
package tab

import (
	"context"
	"image"
	"sync"
	"testing"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

// mockTerminal satisfies terminalapi.Terminal for focused tab tests.
type mockTerminal struct{}

// Size returns a stable fake terminal size.
func (m *mockTerminal) Size() image.Point { return image.Point{X: 80, Y: 24} }

// Clear is a no-op.
func (m *mockTerminal) Clear(opts ...cell.Option) error { return nil }

// Flush is a no-op.
func (m *mockTerminal) Flush() error { return nil }

// SetCursor is a no-op.
func (m *mockTerminal) SetCursor(p image.Point) {}

// HideCursor is a no-op.
func (m *mockTerminal) HideCursor() {}

// SetCell is a no-op.
func (m *mockTerminal) SetCell(p image.Point, r rune, opts ...cell.Option) error { return nil }

// Event is unused in these tests.
func (m *mockTerminal) Event(ctx context.Context) terminalapi.Event { return nil }

// Close is a no-op.
func (m *mockTerminal) Close() {}

// TestManagerNavigation verifies the tab manager wraps cleanly in both directions.
func TestManagerNavigation(t *testing.T) {
	tm := NewManager(newTestTab(t, "Overview"), newTestTab(t, "Controls"))

	if got := tm.GetTabNum(); got != 2 {
		t.Fatalf("GetTabNum() = %d, want 2", got)
	}
	if got := tm.GetActiveTab().Name; got != "Overview" {
		t.Fatalf("initial active tab = %q, want Overview", got)
	}

	tm.NextTab()
	if got := tm.GetActiveTab().Name; got != "Controls" {
		t.Fatalf("active tab after NextTab() = %q, want Controls", got)
	}

	tm.PreviousTab()
	if got := tm.GetActiveTab().Name; got != "Overview" {
		t.Fatalf("active tab after PreviousTab() = %q, want Overview", got)
	}

	tm.PreviousTab()
	if got := tm.GetActiveTab().Name; got != "Controls" {
		t.Fatalf("wrapped active tab after PreviousTab() = %q, want Controls", got)
	}
}

// TestManagerSnapshotReflectsNotification verifies snapshots expose active notifications.
func TestManagerSnapshotReflectsNotification(t *testing.T) {
	tm := NewManager(newTestTab(t, "Overview"), newTestTab(t, "Signals"))

	if !tm.SetNotification(1, true, 20*time.Millisecond) {
		t.Fatal("SetNotification() = false, want true")
	}

	snapshot, active := tm.Snapshot()
	if active != 0 {
		t.Fatalf("active index = %d, want 0", active)
	}
	if len(snapshot) != 2 || !snapshot[1].Notification {
		t.Fatalf("snapshot = %#v, want notification on second tab", snapshot)
	}

	time.Sleep(30 * time.Millisecond)
	snapshot, _ = tm.Snapshot()
	if snapshot[1].Notification {
		t.Fatalf("expired notification still visible in snapshot: %#v", snapshot[1])
	}
}

// TestHeaderClickTargets verifies the tab strip computes stable mouse targets.
func TestHeaderClickTargets(t *testing.T) {
	tm := NewManager(newTestTab(t, "Overview"), newTestTab(t, "Controls"), newTestTab(t, "Signals"))
	header, err := NewHeader(tm, NewOptions())
	if err != nil {
		t.Fatalf("NewHeader() => unexpected error: %v", err)
	}

	if got := len(header.tabRectangles); got != 3 {
		t.Fatalf("len(tabRectangles) = %d, want 3", got)
	}
	for i, rect := range header.tabRectangles {
		if rect.Dx() <= 0 {
			t.Fatalf("tabRectangles[%d] = %v, want positive width", i, rect)
		}
		center := image.Point{X: rect.Min.X + rect.Dx()/2, Y: 0}
		if got := header.GetClickedTab(center); got != i {
			t.Fatalf("GetClickedTab(%v) = %d, want %d", center, got, i)
		}
	}

	if got := header.GetClickedTab(image.Point{X: 999, Y: 0}); got != -1 {
		t.Fatalf("GetClickedTab(outside) = %d, want -1", got)
	}
}

// TestHeaderAdvanceKeepsClickTargetsStable verifies animation doesn't disturb hit testing.
func TestHeaderAdvanceKeepsClickTargetsStable(t *testing.T) {
	tm := NewManager(newTestTab(t, "Overview"), newTestTab(t, "Controls"))
	header, err := NewHeader(tm, NewOptions(AnimatedActiveTab(true)))
	if err != nil {
		t.Fatalf("NewHeader() => unexpected error: %v", err)
	}

	target := header.tabRectangles[0]
	if err := header.Advance(); err != nil {
		t.Fatalf("Advance() => unexpected error: %v", err)
	}
	if got := header.GetClickedTab(image.Point{X: target.Min.X + 1, Y: 0}); got != 0 {
		t.Fatalf("GetClickedTab() after Advance = %d, want 0", got)
	}
}

// TestContentUpdateHandlesEmptyManager verifies empty tab sets are harmless.
func TestContentUpdateHandlesEmptyManager(t *testing.T) {
	term := &mockTerminal{}
	cont, err := container.New(term, container.ID("tabContent"))
	if err != nil {
		t.Fatalf("container.New() => unexpected error: %v", err)
	}
	if err := NewContent(NewManager()).Update(cont); err != nil {
		t.Fatalf("Content.Update() => unexpected error: %v", err)
	}
}

// TestEventHandlerKeyboardSwitchesTabs verifies keyboard navigation updates manager state.
func TestEventHandlerKeyboardSwitchesTabs(t *testing.T) {
	eh, tm, _ := newTestEventHandler(t)

	eh.HandleKeyboard(&terminalapi.Keyboard{Key: keyboard.KeyTab})
	if got := tm.GetActiveTab().Name; got != "Controls" {
		t.Fatalf("active tab after Tab = %q, want Controls", got)
	}

	eh.HandleKeyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowLeft})
	if got := tm.GetActiveTab().Name; got != "Overview" {
		t.Fatalf("active tab after Left = %q, want Overview", got)
	}
}

// TestEventHandlerMouseSwitchesTabs verifies header clicks activate the target tab only.
func TestEventHandlerMouseSwitchesTabs(t *testing.T) {
	eh, tm, header := newTestEventHandler(t)

	target := header.tabRectangles[1]
	eh.HandleMouse(&terminalapi.Mouse{
		Button:   mouse.ButtonLeft,
		Position: image.Point{X: target.Min.X + 1, Y: 0},
	})
	if got := tm.GetActiveTab().Name; got != "Controls" {
		t.Fatalf("active tab after click = %q, want Controls", got)
	}

	eh.HandleMouse(&terminalapi.Mouse{
		Button:   mouse.ButtonLeft,
		Position: image.Point{X: 1, Y: 5},
	})
	if got := tm.GetActiveTab().Name; got != "Controls" {
		t.Fatalf("active tab after click outside header = %q, want Controls", got)
	}
}

// TestEventHandlerRefreshConcurrent verifies external refreshes serialize cleanly.
func TestEventHandlerRefreshConcurrent(t *testing.T) {
	eh, _, _ := newTestEventHandler(t)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eh.Refresh()
		}()
	}
	wg.Wait()
}

// newTestEventHandler wires a minimal tab tree for keyboard and mouse tests.
func newTestEventHandler(t *testing.T) (*EventHandler, *Manager, *Header) {
	t.Helper()

	tm := NewManager(newTestTab(t, "Overview"), newTestTab(t, "Controls"))
	header, err := NewHeader(tm, NewOptions())
	if err != nil {
		t.Fatalf("NewHeader() => unexpected error: %v", err)
	}
	content := NewContent(tm)
	term := &mockTerminal{}
	cont, err := container.New(term, container.ID("tabContent"))
	if err != nil {
		t.Fatalf("container.New() => unexpected error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return NewEventHandler(ctx, term, tm, header, content, cont, cancel, NewOptions()), tm, header
}

// newTestTab returns a small tab with valid widget content.
func newTestTab(t *testing.T, name string) *Tab {
	t.Helper()

	w, err := text.New(text.WrapAtWords())
	if err != nil {
		t.Fatalf("text.New() => unexpected error: %v", err)
	}
	if err := w.Write(name); err != nil {
		t.Fatalf("Write() => unexpected error: %v", err)
	}
	return &Tab{Name: name, Content: container.PlaceWidget(w)}
}
