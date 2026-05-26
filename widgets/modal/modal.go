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
	"image"
	"sort"
	"strings"
	"sync"

	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

const (
	minDockedWidth = 8
	maxDockedWidth = 24
)

// Modal renders and manages a set of draggable child widgets.
type Modal struct {
	// ID identifies the container that hosts the modal.
	ID string

	// DraggableItems holds the widgets drawn inside the modal.
	DraggableItems []*DraggableWidget

	// Opts holds the modal configuration.
	Opts *Options

	currentWidth  int
	currentHeight int
	dirty         bool
	mutex         sync.Mutex
}

// DraggableWidget is a widget that can be repositioned with the mouse.
type DraggableWidget struct {
	ID     string
	Title  string
	Widget widgetapi.Widget

	X int
	Y int

	Width  int
	Height int

	ZIndex int
	Border bool

	Minimizable bool

	offsetX int
	offsetY int

	dragging    bool
	minimized   bool
	restoreRect image.Rectangle
	mutex       sync.Mutex
}

// NewModal creates a modal widget.
func NewModal(id string, items []*DraggableWidget, opts *Options) *Modal {
	if opts == nil {
		opts = NewOptions()
	}
	return &Modal{
		ID:             id,
		DraggableItems: items,
		Opts:           opts,
		dirty:          true,
	}
}

// Draw renders the draggable child widgets into the provided canvas.
func (m *Modal) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	area := cvs.Area()

	m.mutex.Lock()
	m.currentWidth = area.Dx()
	m.currentHeight = area.Dy()

	if m.currentWidth < m.Opts.MinimumSize.X || m.currentHeight < m.Opts.MinimumSize.Y {
		m.mutex.Unlock()
		return draw.ResizeNeeded(cvs)
	}

	m.layoutMinimizedLocked()
	items := m.sortedItems()
	m.dirty = false
	m.mutex.Unlock()

	// Draw each item without holding m.mutex to avoid lock-order inversion
	// with dw.mutex inside handleMouse (which acquires m.mutex then dw.mutex).
	for _, item := range items {
		if err := m.drawItem(cvs, meta, item); err != nil {
			return err
		}
	}
	return nil
}

// Keyboard implements widgetapi.Widget.Keyboard.
func (m *Modal) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return nil
}

// Mouse implements widgetapi.Widget.Mouse.
func (m *Modal) Mouse(event *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	return m.handleMouse(event)
}

// Options implements widgetapi.Widget.Options.
func (m *Modal) Options() widgetapi.Options {
	return widgetapi.Options{
		WantMouse:   widgetapi.MouseScopeWidget,
		MinimumSize: m.Opts.MinimumSize,
	}
}

// HandleMouse updates the top-most matching draggable widget.
func (m *Modal) HandleMouse(event *terminalapi.Mouse) {
	_ = m.handleMouse(event)
}

func (m *Modal) handleMouse(event *terminalapi.Mouse) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	items := m.sortedItems()
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		area := item.rect()
		if !item.dragging && !event.Position.In(area) {
			continue
		}
		handled, err := item.handleMouse(event, m.maxZIndex, m.currentWidth, m.currentHeight, m)
		if err != nil {
			return err
		}
		if handled {
			m.dirty = true
			return nil
		}
	}
	return nil
}

// SetDirty marks the modal for redraw.
func (m *Modal) SetDirty(dirty bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.dirty = dirty
}

// NewDraggableWidget creates a draggable child widget configured for the modal.
func NewDraggableWidget(id string, widget widgetapi.Widget, x, y, width, height int, opts *Options) *DraggableWidget {
	if opts == nil {
		opts = NewOptions()
	}
	w := maxInt(width, 1)
	h := maxInt(height, 1)
	return &DraggableWidget{
		ID:          id,
		Title:       defaultTitle(id),
		Widget:      widget,
		X:           x,
		Y:           y,
		Width:       w,
		Height:      h,
		ZIndex:      1,
		Border:      opts.Border,
		Minimizable: true,
	}
}

// HandleMouse updates drag state and position for the draggable widget.
func (dw *DraggableWidget) HandleMouse(event *terminalapi.Mouse, maxZIndexFunc func() int, canvasWidth, canvasHeight int, modal *Modal) bool {
	handled, _ := dw.handleMouse(event, maxZIndexFunc, canvasWidth, canvasHeight, modal)
	return handled
}

func (dw *DraggableWidget) handleMouse(event *terminalapi.Mouse, maxZIndexFunc func() int, canvasWidth, canvasHeight int, modal *Modal) (bool, error) {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	frameArea := dw.rect()
	minimizeArea := dw.minimizeRect()
	titleArea := dw.titleBarRect()
	contentArea := dw.contentRect()

	switch {
	case event.Button == mouse.ButtonLeft && dw.Minimizable && event.Position.In(minimizeArea):
		if dw.minimized {
			dw.restoreLocked(canvasWidth, canvasHeight)
		} else {
			dw.minimizeLocked(frameArea)
			modal.layoutMinimizedLocked()
		}
		return true, nil

	case dw.minimized && event.Button == mouse.ButtonLeft && event.Position.In(frameArea):
		dw.restoreLocked(canvasWidth, canvasHeight)
		modal.layoutMinimizedLocked()
		return true, nil

	case event.Button == mouse.ButtonLeft && !dw.dragging && event.Position.In(titleArea):
		dw.ZIndex = maxZIndexFunc() + 1
		dw.offsetX = event.Position.X - dw.X
		dw.offsetY = event.Position.Y - dw.Y
		dw.dragging = true
		return true, nil

	case dw.dragging && event.Button == mouse.ButtonRelease:
		dw.dragging = false
		return true, nil

	case dw.dragging:
		dw.X = clampInt(event.Position.X-dw.offsetX, 0, maxInt(canvasWidth-dw.Width, 0))
		dw.Y = clampInt(event.Position.Y-dw.offsetY, 0, maxInt(canvasHeight-dw.Height, 0))
		return true, nil

	case !dw.minimized && event.Position.In(contentArea) && dw.Widget.Options().WantMouse != widgetapi.MouseScopeNone:
		childEvent := &terminalapi.Mouse{
			Position: event.Position.Sub(contentArea.Min),
			Button:   event.Button,
		}
		if err := dw.Widget.Mouse(childEvent, &widgetapi.EventMeta{}); err != nil {
			return true, err
		}
		return true, nil
	}

	return false, nil
}

// rect returns the widget rectangle in modal coordinates.
func (dw *DraggableWidget) rect() image.Rectangle {
	return image.Rect(dw.X, dw.Y, dw.X+dw.Width, dw.Y+dw.Height)
}

// titleBarRect returns the draggable title bar rectangle for the widget.
func (dw *DraggableWidget) titleBarRect() image.Rectangle {
	return image.Rect(dw.X, dw.Y, dw.X+dw.Width, dw.Y+1)
}

// minimizeRect returns the clickable rectangle for the minimize or restore control.
func (dw *DraggableWidget) minimizeRect() image.Rectangle {
	return image.Rect(dw.X+maxInt(dw.Width-2, 0), dw.Y, dw.X+dw.Width, dw.Y+1)
}

func (dw *DraggableWidget) contentRect() image.Rectangle {
	frameArea := dw.rect()
	if dw.Border {
		return image.Rect(frameArea.Min.X+1, frameArea.Min.Y+2, frameArea.Max.X-1, frameArea.Max.Y-1)
	}
	return image.Rect(frameArea.Min.X, frameArea.Min.Y+1, frameArea.Max.X, frameArea.Max.Y)
}

// sortedItems returns the draggable items ordered by z-index from back to front.
func (m *Modal) sortedItems() []*DraggableWidget {
	items := append([]*DraggableWidget(nil), m.DraggableItems...)
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].ZIndex < items[j].ZIndex
	})
	return items
}

// drawItem renders a single draggable widget onto the modal canvas.
func (m *Modal) drawItem(dst *canvas.Canvas, meta *widgetapi.Meta, item *DraggableWidget) error {
	visible := item.rect().Intersect(dst.Area())
	if visible.Empty() {
		return nil
	}

	contentArea := visible
	if item.Border {
		if visible.Dx() < 2 || visible.Dy() < 2 {
			return nil
		}
		if err := draw.Border(dst, visible, draw.BorderLineStyle(linestyle.Round)); err != nil {
			return err
		}
		if visible.Dx() <= 2 || visible.Dy() <= 2 {
			if item.minimized {
				return m.drawTitleBar(dst, item, visible)
			}
			return nil
		}
		contentArea = image.Rect(visible.Min.X+1, visible.Min.Y+2, visible.Max.X-1, visible.Max.Y-1)
	} else {
		contentArea = image.Rect(visible.Min.X, visible.Min.Y+1, visible.Max.X, visible.Max.Y)
	}

	if err := m.drawTitleBar(dst, item, visible); err != nil {
		return err
	}

	if item.minimized || contentArea.Empty() || contentArea.Dy() <= 0 || contentArea.Dx() <= 0 {
		return nil
	}

	subCanvas, err := canvas.New(contentArea)
	if err != nil {
		return err
	}
	if !widgetFits(item.Widget.Options(), contentArea) {
		if err := draw.ResizeNeeded(subCanvas); err != nil {
			return err
		}
		return subCanvas.CopyTo(dst)
	}
	if err := item.Widget.Draw(subCanvas, meta); err != nil {
		return err
	}
	return subCanvas.CopyTo(dst)
}

// drawTitleBar renders the window title region and minimize control.
func (m *Modal) drawTitleBar(dst *canvas.Canvas, item *DraggableWidget, visible image.Rectangle) error {
	titleBar := item.titleBarRect().Intersect(dst.Area())
	if titleBar.Empty() {
		return nil
	}

	if err := dst.SetAreaCells(titleBar, ' ', m.Opts.TitleBarCellOpts...); err != nil {
		return err
	}

	maxTitleX := titleBar.Max.X
	if item.Minimizable {
		maxTitleX -= 2
	}
	if maxTitleX > titleBar.Min.X {
		title := " " + item.Title + " "
		overrun := draw.OverrunModeTrim
		if item.minimized {
			title = compactDockTitle(item.Title, maxTitleX-titleBar.Min.X)
			overrun = draw.OverrunModeTrim
		}
		if err := draw.Text(dst, title, titleBar.Min,
			draw.TextCellOpts(m.Opts.TitleCellOpts...),
			draw.TextMaxX(maxTitleX),
			draw.TextOverrunMode(overrun),
		); err != nil {
			return err
		}
	}

	if item.Minimizable && titleBar.Dx() >= 2 {
		controlGlyph := m.Opts.MinimizeGlyph
		if item.minimized {
			controlGlyph = m.Opts.RestoreGlyph
		}
		_, err := dst.SetCell(image.Point{X: titleBar.Max.X - 2, Y: titleBar.Min.Y}, controlGlyph, m.Opts.TitleControlCellOpts...)
		return err
	}
	return nil
}

// maxZIndex returns the highest z-index currently used by any draggable item.
func (m *Modal) maxZIndex() int {
	maxZ := 0
	for _, item := range m.DraggableItems {
		if item.ZIndex > maxZ {
			maxZ = item.ZIndex
		}
	}
	return maxZ
}

// layoutMinimizedLocked docks minimized windows along the bottom edge.
//
// The caller must hold m.mutex.
func (m *Modal) layoutMinimizedLocked() {
	x := 0
	y := maxInt(m.currentHeight-2, 0)
	for _, item := range m.DraggableItems {
		if !item.minimized {
			continue
		}
		remaining := maxInt(m.currentWidth-x, minDockedWidth)
		item.X = x
		item.Y = y
		item.Height = dockedHeight(item.Border)
		item.Width = dockedWidth(item.Title, item.Minimizable, remaining)
		x += item.Width + m.Opts.DockGap
	}
}

// minimizeLocked stores the restore geometry and shrinks the window to dock height.
//
// The caller must hold dw.mutex.
func (dw *DraggableWidget) minimizeLocked(current image.Rectangle) {
	dw.restoreRect = current
	dw.minimized = true
	dw.dragging = false
}

// restoreLocked restores the saved geometry and clamps it to the canvas.
//
// The caller must hold dw.mutex.
func (dw *DraggableWidget) restoreLocked(canvasWidth, canvasHeight int) {
	if !dw.minimized {
		return
	}
	restore := dw.restoreRect
	if restore.Empty() {
		restore = image.Rect(dw.X, dw.Y, dw.X+maxInt(dw.Width, 1), dw.Y+maxInt(dw.Height, 2))
	}
	dw.Width = maxInt(restore.Dx(), 1)
	dw.Height = maxInt(restore.Dy(), dockedHeight(dw.Border))
	dw.X = clampInt(restore.Min.X, 0, maxInt(canvasWidth-dw.Width, 0))
	dw.Y = clampInt(restore.Min.Y, 0, maxInt(canvasHeight-dw.Height, 0))
	dw.minimized = false
}

// dockedWidth returns the docked width for a minimized window title.
func dockedWidth(title string, minimizable bool, available int) int {
	width := len([]rune(" " + title + " "))
	if minimizable {
		width += 2
	}
	width = maxInt(width, minDockedWidth)
	width = minInt(width, maxDockedWidth)
	if available > 0 {
		width = minInt(width, maxInt(available, minDockedWidth))
	}
	return width
}

// compactDockTitle returns a minimized dock label that makes truncation explicit.
func compactDockTitle(title string, width int) string {
	if width <= 0 {
		return ""
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Window"
	}

	full := []rune(" " + title + " ")
	if len(full) <= width {
		return string(full)
	}
	if width <= 4 {
		return string([]rune("...")[:minInt(width, 3)])
	}

	ellipsis := []rune("...")
	availableTitle := width - len(ellipsis) - 1
	if availableTitle < 1 {
		availableTitle = 1
	}
	titleRunes := []rune(title)
	if len(titleRunes) > availableTitle {
		titleRunes = titleRunes[:availableTitle]
	}
	return " " + string(titleRunes) + "..."
}

// dockedHeight returns the window height while docked in minimized form.
func dockedHeight(border bool) int {
	if border {
		return 2
	}
	return 1
}

// widgetFits reports whether the widget's requested minimum size fits inside
// the provided content area.
func widgetFits(opts widgetapi.Options, area image.Rectangle) bool {
	needSize := image.Point{X: 1, Y: 1}
	if opts.MinimumSize.X > 0 && opts.MinimumSize.Y > 0 {
		needSize = opts.MinimumSize
	}
	return area.Dx() >= needSize.X && area.Dy() >= needSize.Y
}

// defaultTitle derives a user-facing title from the draggable widget ID.
func defaultTitle(id string) string {
	if id == "" {
		return "Window"
	}
	parts := strings.FieldsFunc(id, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})
	if len(parts) == 0 {
		return "Window"
	}
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}

// clampInt constrains v to the inclusive range [min, max].
func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// maxInt returns the larger of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// minInt returns the smaller of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
