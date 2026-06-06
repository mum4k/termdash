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

package modal

import (
	"context"
	"errors"
	"image"
	"strings"
	"testing"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/testcanvas"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// TestNewOptionsIgnoresLegacyLogging verifies the old logging options remain harmless.
func TestNewOptionsIgnoresLegacyLogging(t *testing.T) {
	opts := NewOptions(EnableLogging(true), LogPrefix("[legacy]"))
	if !opts.Border {
		t.Fatal("default Border = false, want true")
	}
	if got, want := opts.MinimumSize, (image.Point{X: 10, Y: 5}); got != want {
		t.Fatalf("MinimumSize = %v, want %v", got, want)
	}
}

// TestModalDrawUsesTopmostZIndex verifies higher z-index widgets render on top.
func TestModalDrawUsesTopmostZIndex(t *testing.T) {
	opts := NewOptions(Border(false), MinimumSize(image.Point{X: 1, Y: 1}))
	back := NewDraggableWidget("back", &fillWidget{r: 'a'}, 1, 1, 5, 3, opts)
	front := NewDraggableWidget("front", &fillWidget{r: 'b'}, 3, 2, 4, 2, opts)
	front.ZIndex = 2

	modal := NewModal("modal", []*DraggableWidget{back, front}, opts)
	ft := drawModal(t, modal, image.Point{X: 12, Y: 8})
	lines := strings.Split(ft.String(), "\n")
	if got := rune(lines[3][3]); got != 'b' {
		t.Fatalf("overlap cell = %q, want %q", got, 'b')
	}
}

// TestModalDrawTitleBar verifies draggable windows render a visible title bar.
func TestModalDrawTitleBar(t *testing.T) {
	opts := NewOptions(Border(true), MinimumSize(image.Point{X: 1, Y: 1}))
	item := NewDraggableWidget("signal-panel", &fillWidget{r: 'x'}, 1, 1, 12, 6, opts)
	modal := NewModal("modal", []*DraggableWidget{item}, opts)

	ft := drawModal(t, modal, image.Point{X: 20, Y: 10})
	if got := strings.Split(ft.String(), "\n")[1]; !strings.Contains(got, "Signal Pa") {
		t.Fatalf("title bar row = %q, want trimmed title text", got)
	}
}

// TestModalDrawResizeNeeded verifies undersized canvases show the standard marker.
func TestModalDrawResizeNeeded(t *testing.T) {
	modal := NewModal("modal", nil, NewOptions(MinimumSize(image.Point{X: 8, Y: 4})))
	ft := drawModal(t, modal, image.Point{X: 3, Y: 2})
	if got := string([]rune(strings.Split(ft.String(), "\n")[0])[:1]); got != "⇄" {
		t.Fatalf("resize marker = %q, want %q", got, "⇄")
	}
}

// TestModalDrawSkipsUndersizedChild verifies undersized child widgets draw the
// standard resize marker instead of being asked to draw on an invalid canvas.
func TestModalDrawSkipsUndersizedChild(t *testing.T) {
	opts := NewOptions(Border(true), MinimumSize(image.Point{X: 1, Y: 1}))
	item := NewDraggableWidget("tiny", &minSizeWidget{}, 0, 0, 6, 4, opts)
	modal := NewModal("modal", []*DraggableWidget{item}, opts)

	ft := drawModal(t, modal, image.Point{X: 12, Y: 8})
	if !strings.Contains(ft.String(), "⇄") {
		t.Fatalf("modal render = %q, want resize marker", ft.String())
	}
}

// TestModalDrawIgnoresClippedBorderSlivers verifies partially visible bordered
// windows do not error when resize clipping reduces the visible area below the
// minimum supported border size.
func TestModalDrawIgnoresClippedBorderSlivers(t *testing.T) {
	opts := NewOptions(Border(true), MinimumSize(image.Point{X: 1, Y: 1}))
	item := NewDraggableWidget("clipped", &fillWidget{r: 'x'}, 4, 1, 8, 6, opts)
	modal := NewModal("modal", []*DraggableWidget{item}, opts)

	ft := drawModal(t, modal, image.Point{X: 5, Y: 4})
	if ft == nil {
		t.Fatal("drawModal returned nil terminal")
	}
}

// TestDraggableWidgetHandleMouseMovesWithinBounds verifies title-bar dragging is clamped to the canvas.
func TestDraggableWidgetHandleMouseMovesWithinBounds(t *testing.T) {
	opts := NewOptions(Border(false), MinimumSize(image.Point{X: 1, Y: 1}))
	item := NewDraggableWidget("drag", &fillWidget{r: 'x'}, 0, 0, 4, 2, opts)
	modal := NewModal("modal", []*DraggableWidget{item}, opts)

	_ = drawModal(t, modal, image.Point{X: 10, Y: 6})

	modal.HandleMouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 1, Y: 0}})
	modal.HandleMouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 20, Y: 20}})
	modal.HandleMouse(&terminalapi.Mouse{Button: mouse.ButtonRelease, Position: image.Point{X: 20, Y: 20}})

	if got, want := item.X, 6; got != want {
		t.Fatalf("item.X = %d, want %d", got, want)
	}
	if got, want := item.Y, 4; got != want {
		t.Fatalf("item.Y = %d, want %d", got, want)
	}
}

// TestDraggableWidgetContentClickDoesNotDrag verifies dragging only starts from the title bar.
func TestDraggableWidgetContentClickDoesNotDrag(t *testing.T) {
	opts := NewOptions(Border(true), MinimumSize(image.Point{X: 1, Y: 1}))
	item := NewDraggableWidget("drag", &fillWidget{r: 'x'}, 1, 1, 8, 6, opts)
	modal := NewModal("modal", []*DraggableWidget{item}, opts)

	_ = drawModal(t, modal, image.Point{X: 20, Y: 12})
	modal.HandleMouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 3, Y: 3}})
	modal.HandleMouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 10, Y: 8}})
	modal.HandleMouse(&terminalapi.Mouse{Button: mouse.ButtonRelease, Position: image.Point{X: 10, Y: 8}})

	if got, want := item.X, 1; got != want {
		t.Fatalf("item.X = %d, want %d", got, want)
	}
	if got, want := item.Y, 1; got != want {
		t.Fatalf("item.Y = %d, want %d", got, want)
	}
}

func TestDraggableWidgetForwardsMouseToContent(t *testing.T) {
	opts := NewOptions(Border(true), MinimumSize(image.Point{X: 1, Y: 1}))
	child := &mouseWidget{}
	item := NewDraggableWidget("clickable", child, 1, 1, 12, 7, opts)
	modal := NewModal("modal", []*DraggableWidget{item}, opts)

	_ = drawModal(t, modal, image.Point{X: 24, Y: 12})
	if err := modal.Mouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 3, Y: 4}}, &widgetapi.EventMeta{}); err != nil {
		t.Fatalf("Mouse => unexpected error: %v", err)
	}

	if got, want := child.clicks, 1; got != want {
		t.Fatalf("child clicks = %d, want %d", got, want)
	}
	if got, want := child.last, (image.Point{X: 1, Y: 1}); got != want {
		t.Fatalf("child last point = %v, want %v", got, want)
	}
}

// TestMinimizeAndRestoreDock verifies windows minimize into the bottom dock and restore on click.
func TestMinimizeAndRestoreDock(t *testing.T) {
	opts := NewOptions(Border(true), MinimumSize(image.Point{X: 1, Y: 1}), DockGap(2))
	item := NewDraggableWidget("dock-me", &fillWidget{r: 'x'}, 2, 2, 10, 6, opts)
	modal := NewModal("modal", []*DraggableWidget{item}, opts)

	_ = drawModal(t, modal, image.Point{X: 30, Y: 12})
	minimize := image.Point{X: item.X + item.Width - 2, Y: item.Y}
	modal.HandleMouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: minimize})

	if !item.minimized {
		t.Fatal("item.minimized = false, want true")
	}
	if got, want := item.Y, 10; got != want {
		t.Fatalf("docked Y = %d, want %d", got, want)
	}

	modal.HandleMouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: item.X + 1, Y: item.Y}})
	if item.minimized {
		t.Fatal("item.minimized = true, want false")
	}
	if got, want := item.X, 2; got != want {
		t.Fatalf("restored X = %d, want %d", got, want)
	}
	if got, want := item.Y, 2; got != want {
		t.Fatalf("restored Y = %d, want %d", got, want)
	}
}

// TestMinimizedDockShowsCompactTitle verifies docked windows keep a readable
// title label and make truncation explicit.
func TestMinimizedDockShowsCompactTitle(t *testing.T) {
	opts := NewOptions(Border(true), MinimumSize(image.Point{X: 1, Y: 1}), DockGap(1))
	item := NewDraggableWidget("observability-dashboard-cluster-window", &fillWidget{r: 'x'}, 1, 1, 18, 6, opts)
	item.Title = "Observability Dashboard Cluster Window"
	modal := NewModal("modal", []*DraggableWidget{item}, opts)

	_ = drawModal(t, modal, image.Point{X: 36, Y: 10})
	minimize := image.Point{X: item.X + item.Width - 2, Y: item.Y}
	modal.HandleMouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: minimize})
	ft := drawModal(t, modal, image.Point{X: 36, Y: 10})

	if got, want := item.Width, maxDockedWidth; got != want {
		t.Fatalf("docked width = %d, want compact width %d", got, want)
	}
	dockLine := strings.Split(ft.String(), "\n")[item.Y]
	if !strings.Contains(dockLine, "Observability") || !strings.Contains(dockLine, "...") {
		t.Fatalf("docked title line = %q, want compact title with ellipsis", dockLine)
	}
}

// TestManagerShowHide verifies the manager can show and clear a modal.
func TestManagerShowHide(t *testing.T) {
	root := newModalTestContainer(t)
	manager := &Manager{}
	modal := NewModal("modal", nil, NewOptions())

	if err := manager.ShowModal(modal, root); err != nil {
		t.Fatalf("ShowModal => unexpected error: %v", err)
	}
	if !manager.HasActiveModal() {
		t.Fatal("HasActiveModal = false, want true")
	}
	if err := manager.HideModal(root); err != nil {
		t.Fatalf("HideModal => unexpected error: %v", err)
	}
	if manager.HasActiveModal() {
		t.Fatal("HasActiveModal = true, want false")
	}
}

// TestEventHandlerHandlesEscAndQuit verifies escape hides the modal and q cancels the demo.
func TestEventHandlerHandlesEscAndQuit(t *testing.T) {
	root := newModalTestContainer(t)
	manager := &Manager{}
	modal := NewModal("modal", nil, NewOptions())
	if err := manager.ShowModal(modal, root); err != nil {
		t.Fatalf("ShowModal => unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	handler := NewEventHandler(ctx, cancel, root, manager)
	handler.HandleKeyboard(&terminalapi.Keyboard{Key: keyboard.KeyEsc})
	if manager.HasActiveModal() {
		t.Fatal("modal still active after Esc")
	}

	handler.HandleKeyboard(&terminalapi.Keyboard{Key: 'q'})
	select {
	case <-ctx.Done():
	default:
		t.Fatal("context not canceled after q")
	}
}

// TestEventHandlerHandleMouseDoesNotMutateModal verifies demo-global mouse
// routing does not interfere with widget-local modal dragging.
func TestEventHandlerHandleMouseDoesNotMutateModal(t *testing.T) {
	root := newModalTestContainer(t)
	manager := &Manager{}
	opts := NewOptions(Border(true), MinimumSize(image.Point{X: 1, Y: 1}))
	item := NewDraggableWidget("drag", &fillWidget{r: 'x'}, 2, 2, 8, 6, opts)
	modal := NewModal("modal", []*DraggableWidget{item}, opts)
	if err := manager.ShowModal(modal, root); err != nil {
		t.Fatalf("ShowModal => unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	handler := NewEventHandler(ctx, cancel, root, manager)
	handler.HandleMouse(&terminalapi.Mouse{Button: mouse.ButtonLeft, Position: image.Point{X: 30, Y: 12}})

	if got, want := item.X, 2; got != want {
		t.Fatalf("item.X = %d, want %d", got, want)
	}
	if got, want := item.Y, 2; got != want {
		t.Fatalf("item.Y = %d, want %d", got, want)
	}
}

// drawModal renders the modal into a fake terminal for assertions.
func drawModal(t *testing.T, modal *Modal, size image.Point) *faketerm.Terminal {
	t.Helper()

	ft := faketerm.MustNew(size)
	cvs := testcanvas.MustNew(ft.Area())
	if err := modal.Draw(cvs, &widgetapi.Meta{Focused: true}); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	testcanvas.MustApply(cvs, ft)
	return ft
}

// newModalTestContainer creates a root container with a host slot for modal tests.
func newModalTestContainer(t *testing.T) *container.Container {
	t.Helper()

	term := faketerm.MustNew(image.Point{X: 40, Y: 20})
	root, err := container.New(
		term,
		container.ID("root"),
		container.SplitHorizontal(
			container.Top(),
			container.Bottom(container.ID("modal")),
			container.SplitPercent(10),
		),
	)
	if err != nil {
		t.Fatalf("container.New() => unexpected error: %v", err)
	}
	return root
}

// fillWidget draws a single rune across its full canvas.
type fillWidget struct {
	r rune
}

// Draw fills the widget canvas with the configured rune.
func (fw *fillWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	return cvs.SetAreaCells(cvs.Area(), fw.r)
}

// Keyboard ignores keyboard input for test purposes.
func (fw *fillWidget) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return nil
}

// Mouse ignores mouse input for test purposes.
func (fw *fillWidget) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	return nil
}

// Options reports the test widget's interaction and size requirements.
func (fw *fillWidget) Options() widgetapi.Options {
	return widgetapi.Options{}
}

type mouseWidget struct {
	clicks int
	last   image.Point
}

// Draw fills the widget canvas for test purposes.
func (mw *mouseWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	return cvs.SetAreaCells(cvs.Area(), 'm')
}

// Keyboard ignores keyboard input for test purposes.
func (mw *mouseWidget) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return nil
}

// Mouse records forwarded modal mouse events.
func (mw *mouseWidget) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	mw.clicks++
	mw.last = m.Position
	return nil
}

// Options reports that the test widget wants mouse events.
func (mw *mouseWidget) Options() widgetapi.Options {
	return widgetapi.Options{WantMouse: widgetapi.MouseScopeWidget}
}

// minSizeWidget refuses to draw unless the modal honors its minimum size.
type minSizeWidget struct{}

// Draw fails if called, because undersized widgets should be skipped.
func (mw *minSizeWidget) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	return errors.New("minSizeWidget.Draw should not be called")
}

// Keyboard ignores keyboard input for test purposes.
func (mw *minSizeWidget) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return nil
}

// Mouse ignores mouse input for test purposes.
func (mw *minSizeWidget) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	return nil
}

// Options reports a minimum size larger than the modal's child content area.
func (mw *minSizeWidget) Options() widgetapi.Options {
	return widgetapi.Options{MinimumSize: image.Point{X: 8, Y: 4}}
}
