package treeview

import (
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Number of nodes to scroll per mouse wheel event.
const (
	ScrollStep = 5
)

// TreeNode represents a node in the treeview.
type TreeNode struct {
	// ID is the unique identifier for the node.
	ID string
	// Label is the display text of the node.
	Label string
	// Level is the depth level of the node in the tree.
	Level int
	// Parent is the parent node of this node.
	Parent *TreeNode
	// Children are the child nodes of this node.
	Children []*TreeNode
	// Value holds any data associated with the node.
	Value interface{}
	// ShowSpinner indicates whether to display a spinner for this node.
	ShowSpinner bool
	// OnClick is the function to execute when the node is clicked.
	OnClick func() error
	// ExpandedState indicates whether the node is expanded to show its children.
	ExpandedState bool
	// SpinnerIndex is the current index for the spinner icons.
	SpinnerIndex int
	// mu protects access to the node's fields.
	mu sync.Mutex
}

// SetShowSpinner safely sets the ShowSpinner flag.
func (tn *TreeNode) SetShowSpinner(value bool) {
	tn.mu.Lock()
	defer tn.mu.Unlock()
	tn.ShowSpinner = value
	if !value {
		tn.SpinnerIndex = 0 // Reset spinner index when spinner is turned off
	}
}

// GetShowSpinner safely retrieves the ShowSpinner flag.
func (tn *TreeNode) GetShowSpinner() bool {
	tn.mu.Lock()
	defer tn.mu.Unlock()
	return tn.ShowSpinner
}

// IncrementSpinner safely increments the SpinnerIndex.
func (tn *TreeNode) IncrementSpinner(totalIcons int) {
	tn.mu.Lock()
	defer tn.mu.Unlock()
	tn.SpinnerIndex = (tn.SpinnerIndex + 1) % totalIcons
}

// IsRoot checks if the node is a root node.
func (tn *TreeNode) IsRoot() bool {
	return tn.Parent == nil
}

// SetExpandedState safely sets the ExpandedState flag.
func (tn *TreeNode) SetExpandedState(value bool) {
	tn.mu.Lock()
	defer tn.mu.Unlock()
	tn.ExpandedState = value
}

// GetExpandedState safely retrieves the ExpandedState flag.
func (tn *TreeNode) GetExpandedState() bool {
	tn.mu.Lock()
	defer tn.mu.Unlock()
	return tn.ExpandedState
}

// TreeView represents the treeview widget.
type TreeView struct {
	// mu protects access to the TreeView's fields.
	mu sync.Mutex
	// position stores the widget's top-left position.
	position image.Point
	// opts holds the configuration options for the TreeView.
	opts *options
	// selectedNode is the currently selected node.
	selectedNode *TreeNode
	// visibleNodes is the list of currently visible nodes.
	visibleNodes []*TreeNode
	// logger logs debugging information.
	logger *log.Logger
	// spinnerTicker updates spinner indices periodically.
	spinnerTicker *time.Ticker
	// stopSpinner signals the spinner goroutine to stop.
	stopSpinner chan struct{}
	// expandedIcon is the icon used for expanded nodes.
	expandedIcon string
	// collapsedIcon is the icon used for collapsed nodes.
	collapsedIcon string
	// leafIcon is the icon used for leaf nodes.
	leafIcon string
	// scrollOffset is the current vertical scroll offset.
	scrollOffset int
	// indentationPerLevel is the number of spaces to indent per tree level.
	indentationPerLevel int
	// canvasWidth is the width of the canvas.
	canvasWidth int
	// canvasHeight is the height of the canvas.
	canvasHeight int
	// totalContentHeight is the total height of the content.
	totalContentHeight int
	// waitingIcons are the icons used for the spinner.
	waitingIcons []string
	// lastClickTime is the timestamp of the last handled click.
	lastClickTime time.Time
	// lastKeyTime is the timestamp for debouncing the enter key.
	lastKeyTime time.Time
}

// New creates a new TreeView instance.
func New(opts ...Option) (*TreeView, error) {
	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Set default leaf icon if not provided
	if options.leafIcon == "" {
		options.leafIcon = "→"
	}

	// Set default indentation if not provided
	if options.indentation == 0 {
		options.indentation = 2
	}

	for _, node := range options.nodes {
		setParentsAndAssignIDs(node, nil, 0, "")
	}

	// Create a logger to log debugging information to a file if logging is enabled
	var logger *log.Logger
	if options.enableLogging {
		file, err := os.OpenFile("treeview_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
		logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		// Create a dummy logger that discards all logs
		logger = log.New(io.Discard, "", 0)
	}

	tv := &TreeView{
		opts:                options,
		logger:              logger,
		stopSpinner:         make(chan struct{}),
		expandedIcon:        options.expandedIcon,
		collapsedIcon:       options.collapsedIcon,
		leafIcon:            options.leafIcon,
		scrollOffset:        0,
		indentationPerLevel: options.indentation,
		waitingIcons:        options.waitingIcons,
	}

	setInitialExpandedState(tv, true) // Expand root nodes by default

	if len(options.waitingIcons) > 0 {
		tv.spinnerTicker = time.NewTicker(200 * time.Millisecond)
		go tv.runSpinner()
	}
	tv.updateTotalHeight()

	// Set selectedNode to the first visible node
	visibleNodes := tv.getVisibleNodesList()
	if len(visibleNodes) > 0 {
		tv.selectedNode = visibleNodes[0]
	}
	return tv, nil
}

// generateNodeID creates a consistent node ID.
func generateNodeID(path string, label string) string {
	if path == "" {
		return label
	}
	return fmt.Sprintf("%s/%s", path, label)
}

// setParentsAndAssignIDs assigns parent references, levels, and IDs to nodes recursively.
func setParentsAndAssignIDs(tn *TreeNode, parent *TreeNode, level int, path string) {
	tn.Parent = parent
	tn.Level = level

	tn.ID = generateNodeID(path, tn.Label)

	for _, child := range tn.Children {
		setParentsAndAssignIDs(child, tn, level+1, tn.ID)
	}
}

// runSpinner updates spinner indices periodically.
func (tv *TreeView) runSpinner() {
	for {
		select {
		case <-tv.spinnerTicker.C:
			tv.mu.Lock()
			visibleNodes := tv.getVisibleNodesList()
			tv.mu.Unlock() // Release the TreeView lock before operating on individual nodes
			for _, tn := range visibleNodes {
				if tn.GetShowSpinner() && len(tv.waitingIcons) > 0 {
					tn.IncrementSpinner(len(tv.waitingIcons))
					tv.logger.Printf("Spinner updated for node: %s (SpinnerIndex: %d)", tn.Label, tn.SpinnerIndex)
				}
			}
		case <-tv.stopSpinner:
			return
		}
	}
}

// StopSpinnerTicker stops the spinner ticker.
func (tv *TreeView) StopSpinnerTicker() {
	if tv.spinnerTicker != nil {
		tv.spinnerTicker.Stop()
		close(tv.stopSpinner)
	}
}

// setInitialExpandedState sets the initial expanded state for root nodes.
func setInitialExpandedState(tv *TreeView, expandRoot bool) {
	for _, tn := range tv.opts.nodes {
		if tn.IsRoot() {
			tn.SetExpandedState(expandRoot)
		}
	}
	tv.updateTotalHeight()
}

// calculateHeight calculates the height of a node, including its children if expanded.
func (tv *TreeView) calculateHeight(tn *TreeNode) int {
	height := 1 // Start with the height of the current node
	if tn.ExpandedState {
		for _, child := range tn.Children {
			height += tv.calculateHeight(child)
		}
	}
	return height
}

// calculateTotalHeight calculates the total height of all visible nodes.
func (tv *TreeView) calculateTotalHeight() int {
	totalHeight := 0
	for _, rootNode := range tv.opts.nodes {
		totalHeight += tv.calculateHeight(rootNode)
	}
	return totalHeight
}

// updateTotalHeight updates the totalContentHeight based on visible nodes.
func (tv *TreeView) updateTotalHeight() {
	tv.totalContentHeight = tv.calculateTotalHeight()
}

// getVisibleNodesList retrieves a flat list of all currently visible nodes.
func (tv *TreeView) getVisibleNodesList() []*TreeNode {
	var list []*TreeNode
	var traverse func(tn *TreeNode)
	traverse = func(tn *TreeNode) {
		list = append(list, tn)
		tv.logger.Printf("Visible Node Added: '%s' at Level %d", tn.Label, tn.Level)
		if tn.GetExpandedState() { // Use getter with mutex
			for _, child := range tn.Children {
				traverse(child)
			}
		}
	}
	for _, root := range tv.opts.nodes {
		traverse(root)
	}
	return list
}

// getNodePrefix returns the appropriate prefix for a node based on its state.
func (tv *TreeView) getNodePrefix(tn *TreeNode) string {
	if tn.GetShowSpinner() && len(tv.waitingIcons) > 0 {
		return tv.waitingIcons[tn.SpinnerIndex]
	}

	if len(tn.Children) > 0 {
		if tn.ExpandedState {
			return tv.expandedIcon
		}
		return tv.collapsedIcon
	}

	return tv.leafIcon
}

// drawNode draws nodes based on the nodesToDraw slice.
func (tv *TreeView) drawNode(cvs *canvas.Canvas, nodesToDraw []*TreeNode) error {
	for y, tn := range nodesToDraw {
		// Determine if this node is selected
		isSelected := (tn.ID == tv.selectedNode.ID)

		// Get the prefix based on node state
		prefix := tv.getNodePrefix(tn)
		prefixWidth := runewidth.StringWidth(prefix)

		// Construct the label
		label := fmt.Sprintf("%s %s", prefix, tn.Label)
		labelWidth := runewidth.StringWidth(label)
		indentX := tn.Level * tv.indentationPerLevel
		availableWidth := tv.canvasWidth - indentX

		if tv.opts.truncate && labelWidth > availableWidth {
			// Truncate the label to fit within the available space
			truncatedLabel := truncateString(label, availableWidth)
			if truncatedLabel != label {
				label = truncatedLabel
			}
			labelWidth = runewidth.StringWidth(label)
		}

		// Log prefix width for debugging
		tv.logger.Printf("Drawing node '%s' with prefix width %d", tn.Label, prefixWidth)

		// Determine colors based on selection
		var fgColor cell.Color = tv.opts.labelColor
		var bgColor cell.Color = cell.ColorDefault
		if isSelected {
			fgColor = cell.ColorBlack
			bgColor = cell.ColorWhite
		}

		// Draw the label at the correct position
		if err := tv.drawLabel(cvs, label, indentX, y, fgColor, bgColor); err != nil {
			return err
		}
	}
	return nil
}

// findNodeByClick determines which node was clicked based on x and y coordinates.
func (tv *TreeView) findNodeByClick(x, y int, visibleNodes []*TreeNode) *TreeNode {
	clickedIndex := y + tv.scrollOffset // Adjust Y-coordinate based on scroll offset
	if clickedIndex < 0 || clickedIndex >= len(visibleNodes) {
		return nil
	}

	tn := visibleNodes[clickedIndex]

	label := fmt.Sprintf("%s %s", tv.getNodePrefix(tn), tn.Label)
	labelWidth := runewidth.StringWidth(label)
	indentX := tn.Level * tv.indentationPerLevel
	availableWidth := tv.canvasWidth - indentX

	if tv.opts.truncate && labelWidth > availableWidth {
		truncatedLabel := truncateString(label, availableWidth)
		labelWidth = runewidth.StringWidth(truncatedLabel)
		label = truncatedLabel
	}

	labelStartX := indentX
	labelEndX := labelStartX + labelWidth

	if x >= labelStartX && x < labelEndX {
		tv.logger.Printf("Node '%s' (ID: %s) clicked at [X:%d Y:%d]", tn.Label, tn.ID, x, y)
		return tn
	}

	return nil
}

// handleMouseClick processes mouse click at given x, y coordinates.
func (tv *TreeView) handleMouseClick(x, y int) error {
	tv.logger.Printf("Handling mouse click at (X:%d, Y:%d)", x, y)
	visibleNodes := tv.visibleNodes
	clickedNode := tv.findNodeByClick(x, y, visibleNodes)
	if clickedNode != nil {
		tv.logger.Printf("Node: %s (ID: %s) clicked, expanded: %v", clickedNode.Label, clickedNode.ID, clickedNode.ExpandedState)
		// Update selectedNode to the clicked node
		tv.selectedNode = clickedNode
		if err := tv.handleNodeClick(clickedNode); err != nil {
			tv.logger.Println("Error handling node click:", err)
		}
	} else {
		tv.logger.Printf("No node found at position: (X:%d, Y:%d)", x, y)
	}

	return nil
}

// handleNodeClick toggles the expansion state of a node and manages the spinner.
func (tv *TreeView) handleNodeClick(tn *TreeNode) error {
	tv.logger.Printf("Handling node click for: %s (ID: %s)", tn.Label, tn.ID)
	if len(tn.Children) > 0 {
		// Toggle expansion state
		tn.SetExpandedState(!tn.GetExpandedState())
		tv.updateTotalHeight()
		tv.logger.Printf("Toggled expansion for node: %s to %v", tn.Label, tn.ExpandedState)
		return nil
	}

	// Handle leaf node click
	if tn.OnClick != nil {
		tn.SetShowSpinner(true)
		tv.logger.Printf("Spinner started for node: %s", tn.Label)
		go func(n *TreeNode) {
			tv.logger.Printf("Executing OnClick for node: %s", n.Label)
			if err := n.OnClick(); err != nil {
				tv.logger.Printf("Error executing OnClick for node %s: %v", n.Label, err)
			}
			n.SetShowSpinner(false)
			tv.logger.Printf("Spinner stopped for node: %s", n.Label)
		}(tn)
	}

	return nil
}

// Mouse handles mouse events with debouncing for ButtonLeft clicks.
// It processes mouse press events and mouse wheel events.
func (tv *TreeView) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {

	// Ignore mouse release events to avoid handling multiple events per physical click
	if m.Button == mouse.ButtonRelease {
		return nil
	}

	// Adjust coordinates to be relative to the widget's position
	x := m.Position.X - tv.position.X
	y := m.Position.Y - tv.position.Y

	switch m.Button {
	case mouse.ButtonLeft:
		now := time.Now()
		if now.Sub(tv.lastClickTime) < 100*time.Millisecond {
			// Ignore duplicate click within 100ms
			tv.logger.Printf("Ignored duplicate ButtonLeft click at (X:%d, Y:%d)", x, y)
			return nil
		}
		tv.lastClickTime = now
		tv.logger.Printf("MouseDown event at position: (X:%d, Y:%d)", x, y)
		return tv.handleMouseClick(x, y)
	case mouse.ButtonWheelUp:
		tv.logger.Println("Mouse wheel up")
		if tv.scrollOffset >= ScrollStep {
			tv.scrollOffset -= ScrollStep
		} else {
			tv.scrollOffset = 0
		}
		tv.updateVisibleNodes()
		return nil
	case mouse.ButtonWheelDown:
		tv.logger.Println("Mouse wheel down")
		maxOffset := tv.totalContentHeight - tv.canvasHeight
		if maxOffset < 0 {
			maxOffset = 0
		}
		if tv.scrollOffset+ScrollStep <= maxOffset {
			tv.scrollOffset += ScrollStep
		} else {
			tv.scrollOffset = maxOffset
		}
		tv.updateVisibleNodes()
		return nil
	}
	return nil
}

// Keyboard handles keyboard events.
func (tv *TreeView) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	tv.mu.Lock()
	visibleNodes := tv.visibleNodes
	currentIndex := tv.getSelectedNodeIndex(visibleNodes)
	tv.mu.Unlock()
	if currentIndex == -1 {
		if len(visibleNodes) > 0 {
			tv.selectedNode = visibleNodes[0]
			currentIndex = 0
		} else {
			// No visible nodes to select
			return nil
		}
	}

	// Debounce Enter key to avoid rapid toggling
	now := time.Now()
	if k.Key == keyboard.KeyEnter || k.Key == ' ' {
		if now.Sub(tv.lastKeyTime) < 100*time.Millisecond {
			tv.logger.Printf("Ignored rapid Enter key press")
			return nil
		}
		tv.lastKeyTime = now
	}

	switch k.Key {
	case keyboard.KeyArrowDown:
		if currentIndex < len(visibleNodes)-1 {
			currentIndex++
			tv.selectedNode = visibleNodes[currentIndex]
			// Adjust scrollOffset to keep selectedNode in view
			if currentIndex >= tv.scrollOffset+tv.canvasHeight {
				tv.scrollOffset = currentIndex - tv.canvasHeight + 1
			}
		}
	case keyboard.KeyArrowUp:
		if currentIndex > 0 {
			currentIndex--
			tv.selectedNode = visibleNodes[currentIndex]
			// Adjust scrollOffset to keep selectedNode in view
			if currentIndex < tv.scrollOffset {
				tv.scrollOffset = currentIndex
			}
		}
	case keyboard.KeyEnter, ' ':
		if currentIndex >= 0 && currentIndex < len(visibleNodes) {
			tn := visibleNodes[currentIndex]
			tv.selectedNode = tn
			if err := tv.handleNodeClick(tn); err != nil {
				tv.logger.Println("Error handling node click:", err)
			}
		}
	default:
		// Handle other keys if needed
	}

	return nil
}

// getSelectedNodeIndex returns the index of the selected node in the visibleNodes list.
func (tv *TreeView) getSelectedNodeIndex(visibleNodes []*TreeNode) int {
	for idx, tn := range visibleNodes {
		if tn.ID == tv.selectedNode.ID {
			return idx
		}
	}
	return -1
}

// drawScrollUp draws the scroll up indicator.
func (tv *TreeView) drawScrollUp(cvs *canvas.Canvas) error {
	if _, err := cvs.SetCell(image.Point{X: 0, Y: 0}, '↑', cell.FgColor(cell.ColorWhite)); err != nil {
		return err
	}
	return nil
}

// drawScrollDown draws the scroll down indicator.
func (tv *TreeView) drawScrollDown(cvs *canvas.Canvas) error {
	if _, err := cvs.SetCell(image.Point{X: 0, Y: cvs.Area().Dy() - 1}, '↓', cell.FgColor(cell.ColorWhite)); err != nil {
		return err
	}
	return nil
}

// drawLabel draws the label of a node at the specified position with given foreground and background colors.
func (tv *TreeView) drawLabel(cvs *canvas.Canvas, label string, x, y int, fgColor, bgColor cell.Color) error {
	tv.logger.Printf("Drawing label: '%s' at X: %d, Y: %d with FG: %v, BG: %v", label, x, y, fgColor, bgColor)
	displayWidth := runewidth.StringWidth(label)
	if x+displayWidth > cvs.Area().Dx() {
		displayWidth = cvs.Area().Dx() - x
	}

	truncatedLabel := truncateString(label, displayWidth)

	for i, r := range truncatedLabel {
		if x+i >= cvs.Area().Dx() || y >= cvs.Area().Dy() {
			// If the x or y position exceeds the canvas dimensions, stop drawing
			break
		}
		if _, err := cvs.SetCell(image.Point{X: x + i, Y: y}, r, cell.FgColor(fgColor), cell.BgColor(bgColor)); err != nil {
			return err
		}
	}
	return nil
}

// Draw renders the treeview widget.
func (tv *TreeView) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	tv.mu.Lock()
	tv.updateVisibleNodes()
	visibleNodes := tv.visibleNodes
	totalHeight := len(visibleNodes)
	width := cvs.Area().Dx()
	tv.canvasWidth = width // Set canvasWidth here
	tv.canvasHeight = cvs.Area().Dy()
	tv.mu.Unlock()

	// Log canvas dimensions
	tv.logger.Printf("Canvas Area: Dx=%d, Dy=%d", tv.canvasWidth, tv.canvasHeight)

	if tv.canvasWidth <= 0 || tv.canvasHeight <= 0 {
		return fmt.Errorf("canvas too small")
	}

	// Calculate the maximum valid scroll offset
	maxScrollOffset := tv.totalContentHeight - tv.canvasHeight
	if maxScrollOffset < 0 {
		maxScrollOffset = 0
	}

	// Clamp scrollOffset to ensure it stays within valid bounds
	if tv.scrollOffset > maxScrollOffset {
		tv.scrollOffset = maxScrollOffset
		tv.logger.Printf("Clamped scrollOffset to maxScrollOffset: %d", tv.scrollOffset)
	}
	if tv.scrollOffset < 0 {
		tv.scrollOffset = 0
		tv.logger.Printf("Clamped scrollOffset to 0")
	}

	tv.logger.Printf("Starting Draw with scrollOffset: %d, totalHeight: %d, canvasHeight: %d", tv.scrollOffset, totalHeight, tv.canvasHeight)

	// Clear the canvas
	if err := cvs.Clear(); err != nil {
		return err
	}

	// Determine the range of nodes to draw
	start := tv.scrollOffset
	end := tv.scrollOffset + tv.canvasHeight
	if end > len(visibleNodes) {
		end = len(visibleNodes)
	}

	// Slice the visibleNodes to only the range to draw
	nodesToDraw := visibleNodes[start:end]

	// Draw nodes
	if err := tv.drawNode(cvs, nodesToDraw); err != nil {
		tv.logger.Printf("Error drawing nodes: %v", err)
		return err
	}

	// Draw scroll indicators if needed
	if tv.scrollOffset > 0 {
		if err := tv.drawScrollUp(cvs); err != nil {
			tv.logger.Printf("Error drawing scroll up indicator: %v", err)
			return err
		}
	}
	if tv.scrollOffset+tv.canvasHeight < totalHeight {
		if err := tv.drawScrollDown(cvs); err != nil {
			tv.logger.Printf("Error drawing scroll down indicator: %v", err)
			return err
		}
	}

	tv.logger.Printf("Finished Draw, final currentY: %d, scrollOffset: %d", end, tv.scrollOffset)
	return nil
}

// Options returns the widget options to satisfy the widgetapi.Widget interface.
func (tv *TreeView) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize:              image.Point{10, 3},
		WantKeyboard:             widgetapi.KeyScopeFocused,
		WantMouse:                widgetapi.MouseScopeWidget,
		ExclusiveKeyboardOnFocus: true,
	}
}

// Select returns the label of the selected node.
func (tv *TreeView) Select() (string, error) {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	if tv.selectedNode != nil {
		return tv.selectedNode.Label, nil
	}
	return "", errors.New("no option selected")
}

// Next moves the selection down.
func (tv *TreeView) Next() {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	visibleNodes := tv.visibleNodes
	currentIndex := tv.getSelectedNodeIndex(visibleNodes)
	if currentIndex >= 0 && currentIndex < len(visibleNodes)-1 {
		currentIndex++
		tv.selectedNode = visibleNodes[currentIndex]
		// Adjust scrollOffset to keep selectedNode in view
		if currentIndex >= tv.scrollOffset+tv.canvasHeight {
			tv.scrollOffset = currentIndex - tv.canvasHeight + 1
		}
	}
}

// Previous moves the selection up.
func (tv *TreeView) Previous() {
	tv.mu.Lock()
	defer tv.mu.Unlock()
	visibleNodes := tv.visibleNodes
	currentIndex := tv.getSelectedNodeIndex(visibleNodes)
	if currentIndex > 0 {
		currentIndex--
		tv.selectedNode = visibleNodes[currentIndex]
		// Adjust scrollOffset to keep selectedNode in view
		if currentIndex < tv.scrollOffset {
			tv.scrollOffset = currentIndex
		}
	}
}

// updateVisibleNodes updates the visibleNodes slice based on scrollOffset and node expansion.
func (tv *TreeView) updateVisibleNodes() {
	var allVisible []*TreeNode
	var traverse func(tn *TreeNode)
	traverse = func(tn *TreeNode) {
		allVisible = append(allVisible, tn)
		if tn.ExpandedState {
			for _, child := range tn.Children {
				traverse(child)
			}
		}
	}
	for _, root := range tv.opts.nodes {
		traverse(root)
	}

	tv.totalContentHeight = len(allVisible)

	// Clamp scrollOffset
	if tv.scrollOffset > tv.totalContentHeight-tv.canvasHeight {
		tv.scrollOffset = tv.totalContentHeight - tv.canvasHeight
		if tv.scrollOffset < 0 {
			tv.scrollOffset = 0
		}
	}

	tv.visibleNodes = allVisible
}

// truncateString truncates a string to fit within a specified width, appending "..." if truncated.
func truncateString(s string, maxWidth int) string {
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}

	ellipsis := "…"
	ellipsisWidth := runewidth.StringWidth(ellipsis)

	if maxWidth <= ellipsisWidth {
		return ellipsis // Return ellipsis if space is too small
	}

	truncatedWidth := 0
	truncatedString := ""

	for _, r := range s {
		charWidth := runewidth.RuneWidth(r)
		if truncatedWidth+charWidth+ellipsisWidth > maxWidth {
			break
		}
		truncatedString += string(r)
		truncatedWidth += charWidth
	}

	return truncatedString + ellipsis
}
