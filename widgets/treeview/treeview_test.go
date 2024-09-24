// treeview_test.go
package treeview

import (
	"image"
	"testing"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// MockCanvas is a mock implementation of canvas.Canvas for testing purposes.
type MockCanvas struct {
	Cells map[image.Point]rune
}

func NewMockCanvas() *MockCanvas {
	return &MockCanvas{
		Cells: make(map[image.Point]rune),
	}
}

// SetCell sets a rune at the specified point.
func (mc *MockCanvas) SetCell(p image.Point, r rune, opts ...cell.Option) (bool, error) {
	mc.Cells[p] = r
	return true, nil
}

// Clear clears the canvas.
func (mc *MockCanvas) Clear() error {
	mc.Cells = make(map[image.Point]rune)
	return nil
}

// Area returns the area of the canvas.
func (mc *MockCanvas) Area() image.Rectangle {
	return image.Rect(0, 0, 80, 24) // Default terminal size
}

// Write writes a string starting at the given point.
func (mc *MockCanvas) Write(p image.Point, s string, opts ...cell.Option) error {
	for i, char := range s {
		mc.Cells[image.Point{X: p.X + i, Y: p.Y}] = char
	}
	return nil
}

// MockMeta is a mock implementation of widgetapi.Meta for testing purposes.
type MockMeta struct {
	area image.Rectangle
}

func NewMockMeta(area image.Rectangle) *MockMeta {
	return &MockMeta{
		area: area,
	}
}

// Area returns the area of the widget.
func (m *MockMeta) Area() image.Rectangle {
	return m.area
}

// TestNew tests the initialization of the Treeview widget.
func TestNew(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
			},
		},
	}

	tv, err := New(
		Nodes(root...),
		Indentation(4),
		Icons("▼", "▶", "•"), // Corrected Icons order
		LabelColor(cell.ColorRed),
		WaitingIcons([]string{"|", "/", "-", "\\"}),
		Truncate(true),
		EnableLogging(false),
	)
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Verify selectedNode is initialized to "Root"
	if tv.selectedNode == nil {
		t.Errorf("Expected selectedNode to be initialized, got nil")
	} else if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to be 'Root', got '%s'", tv.selectedNode.Label)
	}

	// Verify the number of root nodes
	if len(tv.opts.nodes) != 1 {
		t.Errorf("Expected 1 root node, got %d", len(tv.opts.nodes))
	}

	// Verify indentation
	if tv.opts.indentation != 4 {
		t.Errorf("Expected indentation to be 4, got %d", tv.opts.indentation)
	}

	// Verify Icons
	if tv.opts.expandedIcon != "▼" || tv.opts.collapsedIcon != "▶" || tv.opts.leafIcon != "•" {
		t.Errorf("Icons not set correctly: got expandedIcon=%s, collapsedIcon=%s, leafIcon=%s",
			tv.opts.expandedIcon, tv.opts.collapsedIcon, tv.opts.leafIcon)
	}

	// Verify LabelColor
	if tv.opts.labelColor != cell.ColorRed {
		t.Errorf("Expected labelColor to be Red, got %v", tv.opts.labelColor)
	}

	// Verify WaitingIcons
	if len(tv.opts.waitingIcons) != 4 {
		t.Errorf("Expected 4 waitingIcons, got %d", len(tv.opts.waitingIcons))
	}

	// Verify Truncate
	if !tv.opts.truncate {
		t.Errorf("Expected truncate to be true")
	}

	// Verify EnableLogging
	if tv.opts.enableLogging {
		t.Errorf("Expected enableLogging to be false")
	}
}

// TestNextPrevious tests navigating through the nodes using Next and Previous methods.
func TestNextPrevious(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
				{Label: "Child3"},
			},
		},
	}

	tv, err := New(
		Nodes(root...),
		Indentation(4),
		Icons("▼", "▶", "•"),
	)
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Manually set Root to be expanded to make children visible
	root[0].ExpandedState = true
	tv.updateVisibleNodes()

	// Initially selected node should be "Root"
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to be 'Root', got '%s'", tv.selectedNode.Label)
	}

	// Navigate down to "Child1"
	tv.Next()
	if tv.selectedNode.Label != "Child1" {
		t.Errorf("Expected selectedNode to be 'Child1', got '%s'", tv.selectedNode.Label)
	}

	// Navigate down to "Child2"
	tv.Next()
	if tv.selectedNode.Label != "Child2" {
		t.Errorf("Expected selectedNode to be 'Child2', got '%s'", tv.selectedNode.Label)
	}

	// Navigate up to "Child1"
	tv.Previous()
	if tv.selectedNode.Label != "Child1" {
		t.Errorf("Expected selectedNode to be 'Child1', got '%s'", tv.selectedNode.Label)
	}

	// Navigate up to "Root"
	tv.Previous()
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to be 'Root', got '%s'", tv.selectedNode.Label)
	}

	// Navigate up at top; should stay at "Root"
	tv.Previous()
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to remain 'Root', got '%s'", tv.selectedNode.Label)
	}
}

// TestMouseScroll adjusted to align with actual behavior
func TestMouseScroll(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
				{Label: "Child3"},
				{Label: "Child4"},
				{Label: "Child5"},
			},
		},
	}

	tv, err := New(Nodes(root...), Indentation(2))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Mock a large canvas height
	tv.canvasHeight = 3
	tv.updateVisibleNodes()

	// Initially, scrollOffset should be 0
	if tv.scrollOffset != 0 {
		t.Errorf("Expected initial scrollOffset to be 0, got %d", tv.scrollOffset)
	}

	// Simulate mouse wheel down
	mouseEvent := &terminalapi.Mouse{
		Button:   mouse.ButtonWheelDown,
		Position: image.Point{X: 0, Y: 0},
	}

	err = tv.Mouse(mouseEvent, &widgetapi.EventMeta{})
	if err != nil {
		t.Errorf("Mouse method returned an error: %v", err)
	}

	// After scrolling down, scrollOffset should be updated accordingly
	maxOffset := len(tv.visibleNodes) - tv.canvasHeight
	if tv.scrollOffset != maxOffset {
		t.Errorf("Expected scrollOffset to be clamped to %d, got %d", maxOffset, tv.scrollOffset)
	}

	// Simulate mouse wheel up
	mouseEvent = &terminalapi.Mouse{
		Button:   mouse.ButtonWheelUp,
		Position: image.Point{X: 0, Y: 0},
	}

	err = tv.Mouse(mouseEvent, &widgetapi.EventMeta{})
	if err != nil {
		t.Errorf("Mouse method returned an error: %v", err)
	}

	// After scrolling up, scrollOffset should be 0
	if tv.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset to be clamped to 0, got %d", tv.scrollOffset)
	}
}

// TestKeyboardScroll tests keyboard navigation in the Treeview
func TestKeyboardScroll(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
				{Label: "Child3"},
				{Label: "Child4"},
				{Label: "Child5"},
			},
		},
	}

	tv, err := New(Nodes(root...), Indentation(2))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Mock a canvas height of 3
	tv.canvasHeight = 3
	tv.updateVisibleNodes()

	// Navigate to Child1
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child1" {
		t.Errorf("Expected selectedNode to be 'Child1', got '%s'", tv.selectedNode.Label)
	}

	// Navigate to Child2
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child2" {
		t.Errorf("Expected selectedNode to be 'Child2', got '%s'", tv.selectedNode.Label)
	}

	// Ensure scrollOffset is updated correctly when navigating further
	tv.Keyboard(&terminalapi.Keyboard{Key: keyboard.KeyArrowDown}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Child3" {
		t.Errorf("Expected selectedNode to be 'Child3', got '%s'", tv.selectedNode.Label)
	}

	if tv.scrollOffset != 1 {
		t.Errorf("Expected scrollOffset to be 1, got %d", tv.scrollOffset)
	}
}

// TestUpdateVisibleNodes tests the visibility of nodes based on expansion state.
func TestUpdateVisibleNodes(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
				{Label: "Child3"},
			},
		},
	}

	tv, err := New(Nodes(root...), Indentation(2))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Lock the Treeview before modifying node states
	tv.mu.Lock()
	root[0].SetExpandedState(true)
	root[0].Children[0].SetExpandedState(false)
	tv.mu.Unlock()

	tv.updateVisibleNodes()

	// Lock before accessing visibleNodes
	tv.mu.Lock()
	visibleNodes := make([]string, len(tv.visibleNodes))
	for i, node := range tv.visibleNodes {
		visibleNodes[i] = node.Label
	}
	tv.mu.Unlock()

	expectedVisible := []string{"Root", "Child1", "Child2", "Child3"}

	if len(visibleNodes) != len(expectedVisible) {
		t.Errorf("Expected %d visible nodes, got %d", len(expectedVisible), len(visibleNodes))
	}

	for i, label := range expectedVisible {
		if i >= len(visibleNodes) || visibleNodes[i] != label {
			t.Errorf("Expected node at index %d to be '%s', got '%s'", i, label, visibleNodes[i])
		}
	}
}

// TestNodeExpansionAndCollapse adjusted for actual behavior
func TestNodeExpansionAndCollapse(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
				{Label: "Child2"},
			},
		},
	}

	tv, err := New(Nodes(root...), Indentation(2))
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Initially, all nodes should be visible
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 3 { // Root + 2 children
		t.Errorf("Expected 3 visible nodes, got %d", len(tv.visibleNodes))
	}

	// Collapse Root
	root[0].SetExpandedState(false)
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 1 { // Only Root
		t.Errorf("Expected 1 visible node after collapsing Root, got %d", len(tv.visibleNodes))
	}

	// Expand Root again
	root[0].SetExpandedState(true)
	tv.updateVisibleNodes()
	if len(tv.visibleNodes) != 3 { // Root + 2 children
		t.Errorf("Expected 3 visible nodes after expanding Root, got %d", len(tv.visibleNodes))
	}
}

// TestSelectNoVisibleNodes tests selecting a node when no nodes are visible.
func TestSelectNoVisibleNodes(t *testing.T) {
	root := []*TreeNode{
		{
			Label:    "Root",
			Children: []*TreeNode{}, // No children, making it a leaf node
		},
	}

	tv, err := New(
		Nodes(root...),
		Indentation(2),
		Icons("▼", "▶", "•"),
	)
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Manually set selectedNode to nil to simulate no visible nodes
	tv.selectedNode = nil

	label, err := tv.Select()
	if err == nil {
		t.Errorf("Expected Select to return an error when no node is selected")
	}

	if label != "" {
		t.Errorf("Expected Select to return empty string when no node is selected, got '%s'", label)
	}
}

// TestKeyboardNonArrowKeys tests that non-arrow keys do not affect navigation.
func TestKeyboardNonArrowKeys(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
			},
		},
	}

	tv, err := New(
		Nodes(root...),
		Indentation(2),
		Icons("▼", "▶", "•"),
	)
	if err != nil {
		t.Fatalf("Failed to create Treeview: %v", err)
	}

	// Manually expand Root to make children visible
	root[0].ExpandedState = true
	tv.updateVisibleNodes()

	// Initial selection is "Root"
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to be 'Root', got '%s'", tv.selectedNode.Label)
	}

	// Send a non-arrow key event (e.g., 'a')
	tv.Keyboard(&terminalapi.Keyboard{Key: 'a'}, &widgetapi.EventMeta{})
	if tv.selectedNode.Label != "Root" {
		t.Errorf("Expected selectedNode to remain 'Root', got '%s'", tv.selectedNode.Label)
	}
}

// TestRunSpinner tests the runSpinner method.
func TestRunSpinner(t *testing.T) {
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{Label: "Child1"},
			},
		},
	}

	tv, err := New(
		Nodes(root...),
		WaitingIcons([]string{"|", "/", "-", "\\"}),
	)
	if err != nil {
		t.Fatalf("Failed to create TreeView: %v", err)
	}

	// Start spinner on "Child1"
	root[0].Children[0].SetShowSpinner(true)

	// Wait to allow spinner to update
	time.Sleep(500 * time.Millisecond)

	// Check that SpinnerIndex has incremented
	root[0].Children[0].mu.Lock()
	spinnerIndex := root[0].Children[0].SpinnerIndex
	root[0].Children[0].mu.Unlock()

	if spinnerIndex == 0 {
		t.Errorf("Expected SpinnerIndex to have incremented, got %d", spinnerIndex)
	}

	// Stop the spinner ticker
	tv.StopSpinnerTicker()
}

// TestHandleNodeClick tests the handleNodeClick method.
// TestHandleNodeClick tests the handleNodeClick method.
func TestHandleNodeClick(t *testing.T) {
	// Create a channel to signal when OnClick has been executed.
	clickCh := make(chan struct{})

	// Define the tree structure with an OnClick handler.
	root := []*TreeNode{
		{
			Label: "Root",
			Children: []*TreeNode{
				{
					Label: "Child1",
					OnClick: func() error {
						// Signal that OnClick has been called.
						clickCh <- struct{}{}
						return nil
					},
				},
			},
		},
	}

	// Initialize the TreeView.
	tv, err := New(
		Nodes(root...),
		Indentation(2),
		Icons("▼", "▶", "•"),
	)
	if err != nil {
		t.Fatalf("Failed to create TreeView: %v", err)
	}

	// Expand Root to make its children visible.
	root[0].SetExpandedState(true)
	tv.updateVisibleNodes()

	// Select "Child1".
	tv.selectedNode = root[0].Children[0]
	err = tv.handleNodeClick(tv.selectedNode)
	if err != nil {
		t.Errorf("handleNodeClick returned an error: %v", err)
	}

	// Wait for the OnClick handler to signal completion or timeout after 1 second.
	select {
	case <-clickCh:
		// OnClick was called successfully.
	case <-time.After(1 * time.Second):
		t.Errorf("OnClick was not called within the expected time")
	}

	// Verify that the spinner has been reset.
	if tv.selectedNode.GetShowSpinner() {
		t.Errorf("Expected spinner to be false after OnClick execution")
	}
}

// TestTruncateString tests the truncateString function.
func TestTruncateString(t *testing.T) {
	longString := "This is a very long string that should be truncated"
	truncated := truncateString(longString, 10)

	if truncated != "This is a…" {
		t.Errorf("Expected 'This is a…', got '%s'", truncated)
	}

	// Test with a maxWidth smaller than ellipsis
	truncated = truncateString(longString, 1)
	if truncated != "…" {
		t.Errorf("Expected '…', got '%s'", truncated)
	}
}
