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

// Package dropdown implements an interactive dropdown selection widget.
package dropdown

import (
	"fmt"
	"image"
	"strings"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

const minTriggerWidth = 4

// Dropdown allows users to select one value from a fixed list of items.
//
// When focused, the widget can be opened with Enter, Space, or the arrow keys.
// Once open, Up and Down move the active row, Enter commits the selection, and
// Esc closes the list.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Dropdown struct {
	mu sync.Mutex

	items    []string
	selected int
	cursor   int
	open     bool
	lastSize image.Point

	opts *options
}

// New returns a new Dropdown for the provided item labels.
func New(items []string, opts ...Option) (*Dropdown, error) {
	cloned := cloneItems(items)
	opt := newOptions(cloned)
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(cloned); err != nil {
		return nil, err
	}

	return &Dropdown{
		items:    cloned,
		selected: opt.selected,
		cursor:   opt.selected,
		opts:     opt,
	}, nil
}

// IntRange returns a slice of decimal strings covering the requested range.
//
// The format defaults to "%d" when empty, making it easy to create dropdown
// items such as "1" through "12" or zero-padded variants like "%02d".
func IntRange(start, end, step int, format string) []string {
	if step <= 0 || end < start {
		return nil
	}
	if format == "" {
		format = "%d"
	}

	var out []string
	for v := start; v <= end; v += step {
		out = append(out, fmt.Sprintf(format, v))
	}
	return out
}

// SelectedIndex returns the currently selected item index.
func (d *Dropdown) SelectedIndex() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.selected
}

// SelectedText returns the currently selected item label.
func (d *Dropdown) SelectedText() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.items[d.selected]
}

// SetSelected replaces the current selection programmatically.
func (d *Dropdown) SetSelected(index int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if index < 0 || index >= len(d.items) {
		return fmt.Errorf("invalid selected index %d, want 0 <= selected < %d", index, len(d.items))
	}
	d.selected = index
	d.cursor = index
	return nil
}

// SetItems replaces the dropdown item list.
//
// If the current selection no longer fits into the updated item set, the
// selection resets to the first item.
func (d *Dropdown) SetItems(items []string) error {
	cloned := cloneItems(items)
	if len(cloned) == 0 {
		return fmt.Errorf("at least one item must be specified")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.items = cloned
	if d.selected >= len(d.items) {
		d.selected = 0
	}
	if d.cursor >= len(d.items) {
		d.cursor = d.selected
	}
	if width := widthFor(cloned); d.opts.width < width {
		d.opts.width = width
	}
	return nil
}

// Open expands the dropdown list.
func (d *Dropdown) Open() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.open = true
	d.cursor = d.selected
}

// Close collapses the dropdown list.
func (d *Dropdown) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.open = false
}

// CanvasSize returns the size needed to draw the dropdown in its current
// state within the provided height budget.
//
// When closed, the dropdown only needs a single row for the trigger. When
// open, the returned height includes the trigger, visible rows, and border.
func (d *Dropdown) CanvasSize(maxHeight int) image.Point {
	d.mu.Lock()
	defer d.mu.Unlock()

	if maxHeight < 1 {
		maxHeight = 1
	}
	height := 1
	if d.open {
		if visible := d.visibleRows(maxHeight); visible > 0 {
			height = visible + 3
		}
	}
	return image.Point{X: d.opts.width, Y: height}
}

// Draw draws the Dropdown onto the canvas.
// Implements widgetapi.Widget.Draw.
func (d *Dropdown) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.lastSize = cvs.Area().Size()
	need := image.Point{X: d.opts.width, Y: 1}
	if need.X > cvs.Area().Dx() || need.Y > cvs.Area().Dy() {
		return draw.ResizeNeeded(cvs)
	}

	if err := d.drawTrigger(cvs, meta); err != nil {
		return err
	}
	if !d.open {
		return nil
	}
	return d.drawList(cvs)
}

// TriggerText returns the dropdown's current trigger line text.
func (d *Dropdown) TriggerText() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	arrow := d.opts.glyphs.ClosedArrow
	if d.open {
		arrow = d.opts.glyphs.OpenArrow
	}
	return formatTrigger(d.items[d.selected], arrow, d.opts.width)
}

// TriggerTextFor returns the trigger text for the provided label using the
// dropdown's configured closed-arrow glyphs.
func (d *Dropdown) TriggerTextFor(label string) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return formatTrigger(label, d.opts.glyphs.ClosedArrow, d.opts.width)
}

// drawTrigger renders the closed/open trigger line at the top of the canvas.
func (d *Dropdown) drawTrigger(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	arrow := d.opts.glyphs.ClosedArrow
	cellOpts := d.opts.cellOpts
	if meta.Focused && len(d.opts.focusedCellOpts) > 0 {
		cellOpts = d.opts.focusedCellOpts
	}
	if d.open {
		arrow = d.opts.glyphs.OpenArrow
	}

	trigger := formatTrigger(d.items[d.selected], arrow, d.opts.width)
	return drawLine(cvs, image.Point{}, trigger, cellOpts)
}

// drawList renders the open dropdown menu and its border.
func (d *Dropdown) drawList(cvs *canvas.Canvas) error {
	visible := d.visibleRows(cvs.Area().Dy())
	if visible <= 0 {
		return nil
	}
	start := d.viewStart(visible)
	boxWidth := d.opts.width

	border := d.opts.glyphs.Border
	top := string(border.TopLeft) + strings.Repeat(string(border.Horizontal), boxWidth-2) + string(border.TopRight)
	if err := drawLine(cvs, image.Point{X: 0, Y: 1}, top, d.opts.borderCellOpts); err != nil {
		return err
	}

	for row := 0; row < visible; row++ {
		itemIndex := start + row
		if itemIndex >= len(d.items) {
			break
		}
		y := 2 + row
		if err := drawLine(cvs, image.Point{X: 0, Y: y}, string(border.Vertical), d.opts.borderCellOpts); err != nil {
			return err
		}
		content := formatOption(d.items[itemIndex], itemIndex == d.cursor, boxWidth-2, d.opts.glyphs.SelectedPrefix, d.opts.glyphs.UnselectedPrefix)
		cellOpts := d.opts.cellOpts
		if itemIndex == d.cursor {
			cellOpts = d.opts.selectedCellOpts
		}
		if err := drawLine(cvs, image.Point{X: 1, Y: y}, content, cellOpts); err != nil {
			return err
		}
		if err := drawLine(cvs, image.Point{X: boxWidth - 1, Y: y}, string(border.Vertical), d.opts.borderCellOpts); err != nil {
			return err
		}
	}

	bottomY := 2 + visible
	bottom := string(border.BottomLeft) + strings.Repeat(string(border.Horizontal), boxWidth-2) + string(border.BottomRight)
	return drawLine(cvs, image.Point{X: 0, Y: bottomY}, bottom, d.opts.borderCellOpts)
}

// Keyboard processes keyboard events for the dropdown.
// Implements widgetapi.Widget.Keyboard.
func (d *Dropdown) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	_ = meta

	var callback SelectFn
	var selected int
	var label string

	d.mu.Lock()
	switch k.Key {
	case keyboard.KeyEnter, ' ':
		if !d.open {
			d.open = true
			d.cursor = d.selected
			d.mu.Unlock()
			return nil
		}
		callback, selected, label = d.commitSelectionLocked(d.cursor)
	case keyboard.KeyArrowDown:
		if !d.open {
			d.open = true
			d.cursor = d.selected
		} else if d.cursor < len(d.items)-1 {
			d.cursor++
		}
	case keyboard.KeyArrowUp:
		if !d.open {
			d.open = true
			d.cursor = d.selected
		} else if d.cursor > 0 {
			d.cursor--
		}
	case keyboard.KeyHome:
		if d.open {
			d.cursor = 0
		}
	case keyboard.KeyEnd:
		if d.open {
			d.cursor = len(d.items) - 1
		}
	case keyboard.KeyEsc:
		d.open = false
		d.cursor = d.selected
	}
	d.mu.Unlock()

	if callback != nil {
		return callback(selected, label)
	}
	return nil
}

// Mouse processes mouse events for the dropdown.
// Implements widgetapi.Widget.Mouse.
func (d *Dropdown) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	_ = meta

	if m.Button != mouse.ButtonLeft {
		return nil
	}

	var callback SelectFn
	var selected int
	var label string

	d.mu.Lock()
	optionIndex := -1
	switch {
	case m.Position.Y == 0 && m.Position.X >= 0 && m.Position.X < d.opts.width:
		d.open = !d.open
		if d.open {
			d.cursor = d.selected
		}
	case !d.open:
		// Ignore clicks outside the closed trigger.
	case func() bool {
		optionIndex = d.optionIndexAt(m.Position)
		return optionIndex >= 0
	}():
		callback, selected, label = d.commitSelectionLocked(optionIndex)
	default:
		d.open = false
		d.cursor = d.selected
	}
	d.mu.Unlock()

	if callback != nil {
		return callback(selected, label)
	}
	return nil
}

// Options implements widgetapi.Widget.Options.
func (d *Dropdown) Options() widgetapi.Options {
	d.mu.Lock()
	defer d.mu.Unlock()

	return widgetapi.Options{
		MinimumSize:  image.Point{X: d.opts.width, Y: 1},
		WantKeyboard: widgetapi.KeyScopeFocused,
		WantMouse:    widgetapi.MouseScopeWidget,
	}
}

// commitSelectionLocked stores the selected row and returns the callback data.
func (d *Dropdown) commitSelectionLocked(index int) (SelectFn, int, string) {
	if index < 0 || index >= len(d.items) {
		return nil, 0, ""
	}
	d.selected = index
	d.cursor = index
	d.open = false
	return d.opts.onSelect, index, d.items[index]
}

// optionIndexAt returns the item index under the provided menu-relative point.
func (d *Dropdown) optionIndexAt(pos image.Point) int {
	height := d.lastSize.Y
	if height <= 0 {
		height = len(d.items) + 3
	}
	visible := d.visibleRows(height)
	if !d.open || visible <= 0 {
		return -1
	}
	if pos.Y < 2 || pos.Y >= 2+visible || pos.X <= 0 || pos.X >= d.opts.width-1 {
		return -1
	}
	index := d.viewStart(visible) + (pos.Y - 2)
	if index < 0 || index >= len(d.items) {
		return -1
	}
	return index
}

// visibleRows returns the number of option rows that fit at the given height.
func (d *Dropdown) visibleRows(height int) int {
	if !d.open || height < 4 {
		return 0
	}
	rows := height - 3
	if rows > len(d.items) {
		rows = len(d.items)
	}
	return rows
}

// viewStart returns the first visible item index for the current cursor.
func (d *Dropdown) viewStart(visible int) int {
	if visible <= 0 || len(d.items) <= visible {
		return 0
	}
	start := d.cursor - visible + 1
	if start < 0 {
		start = 0
	}
	maxStart := len(d.items) - visible
	if start > maxStart {
		start = maxStart
	}
	return start
}

// widthFor returns the minimum trigger width required for the provided items.
func widthFor(items []string) int {
	width := minTriggerWidth
	for _, item := range items {
		if itemWidth := runewidth.StringWidth(item) + 4; itemWidth > width {
			width = itemWidth
		}
	}
	return width
}

// cloneItems returns a shallow copy of the dropdown item list.
func cloneItems(items []string) []string {
	return append([]string(nil), items...)
}

// formatTrigger returns the closed/open trigger line with its arrow glyph.
func formatTrigger(label string, arrow rune, width int) string {
	if width < minTriggerWidth {
		width = minTriggerWidth
	}
	labelWidth := width - 4
	return "[" + fitText(label, labelWidth) + " " + string(arrow) + "]"
}

// formatOption returns one rendered option row inside the open list.
func formatOption(label string, selected bool, width int, selectedPrefix, unselectedPrefix string) string {
	if width < 1 {
		return ""
	}
	prefix := unselectedPrefix
	if selected {
		prefix = selectedPrefix
	}
	return prefix + fitText(label, width-runewidth.StringWidth(prefix))
}

// fitText truncates or pads text to the requested terminal-cell width.
func fitText(text string, width int) string {
	if width <= 0 {
		return ""
	}

	var b strings.Builder
	used := 0
	for _, r := range text {
		rw := runewidth.RuneWidth(r)
		if used+rw > width {
			break
		}
		b.WriteRune(r)
		used += rw
	}
	if used < width {
		b.WriteString(strings.Repeat(" ", width-used))
	}
	return b.String()
}

// drawLine writes a single UTF-8 text line while respecting rune widths.
func drawLine(cvs *canvas.Canvas, pos image.Point, text string, opts []cell.Option) error {
	cur := pos
	for _, r := range text {
		cells, err := cvs.SetCell(cur, r, opts...)
		if err != nil {
			return err
		}
		cur.X += cells
	}
	return nil
}
