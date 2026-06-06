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

// Package radio implements an interactive radio button group widget.
package radio

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

// Radio lets the user choose exactly one item from a fixed list of options.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Radio struct {
	mu sync.Mutex

	items    []Item
	selected int

	opts *options
}

// New returns a new Radio for the provided items.
func New(items []Item, opts ...Option) (*Radio, error) {
	cloned := cloneItems(items)
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(cloned); err != nil {
		return nil, err
	}
	normalizeItems(cloned, opt)

	return &Radio{
		items:    cloned,
		selected: opt.selected,
		opts:     opt,
	}, nil
}

// SelectedIndex returns the current selected item index.
func (r *Radio) SelectedIndex() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.selected
}

// SelectedText returns the current selected item label.
func (r *Radio) SelectedText() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.items[r.selected].Label
}

// SetSelected replaces the current selected item.
func (r *Radio) SetSelected(index int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if index < 0 || index >= len(r.items) {
		return fmtInvalidSelected(index, len(r.items))
	}
	r.selected = index
	return nil
}

// Draw draws the Radio widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (r *Radio) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	need := image.Point{X: widthFor(r.items, r.opts.gap, r.opts.indicatorGap), Y: 1}
	if need.X > cvs.Area().Dx() || need.Y > cvs.Area().Dy() {
		return draw.ResizeNeeded(cvs)
	}

	curX := 0
	for i, item := range r.items {
		text := renderItem(item, i == r.selected, r.opts.indicatorGap)
		cellOpts := item.CellOpts
		if i == r.selected {
			cellOpts = item.SelectedCellOpts
		}
		if err := draw.Text(cvs, text, image.Point{X: curX, Y: 0}, draw.TextCellOpts(cellOpts...)); err != nil {
			return err
		}
		curX += runewidth.StringWidth(text)
		if i < len(r.items)-1 {
			curX += r.opts.gap
		}
	}
	return nil
}

// Keyboard processes keyboard events for the radio group.
// Implements widgetapi.Widget.Keyboard.
func (r *Radio) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	_ = meta
	switch k.Key {
	case keyboard.KeyArrowLeft:
		return r.selectItem(r.SelectedIndex() - 1)
	case keyboard.KeyArrowRight:
		return r.selectItem(r.SelectedIndex() + 1)
	case keyboard.KeyHome:
		return r.selectItem(0)
	case keyboard.KeyEnd:
		return r.selectItem(len(r.items) - 1)
	default:
		return nil
	}
}

// Mouse processes mouse events for the radio group.
// Implements widgetapi.Widget.Mouse.
func (r *Radio) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	_ = meta
	if m.Button != mouse.ButtonLeft {
		return nil
	}
	if m.Position.Y != 0 || m.Position.X < 0 {
		return nil
	}
	index := r.itemAtX(m.Position.X)
	if index < 0 {
		return nil
	}
	return r.selectItem(index)
}

// Options implements widgetapi.Widget.Options.
func (r *Radio) Options() widgetapi.Options {
	r.mu.Lock()
	defer r.mu.Unlock()
	return widgetapi.Options{
		MinimumSize:  image.Point{X: widthFor(r.items, r.opts.gap, r.opts.indicatorGap), Y: 1},
		WantKeyboard: widgetapi.KeyScopeFocused,
		WantMouse:    widgetapi.MouseScopeWidget,
	}
}

// Text returns the radio group's current rendered one-line text.
func (r *Radio) Text() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var b strings.Builder
	for i, item := range r.items {
		b.WriteString(renderItem(item, i == r.selected, r.opts.indicatorGap))
		if i < len(r.items)-1 {
			b.WriteString(strings.Repeat(" ", r.opts.gap))
		}
	}
	return b.String()
}

// ItemText returns the rendered text for the requested item.
func (r *Radio) ItemText(index int) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if index < 0 || index >= len(r.items) {
		return ""
	}
	return renderItem(r.items[index], index == r.selected, r.opts.indicatorGap)
}

// selectItem updates the active item and calls the selection hook when needed.
func (r *Radio) selectItem(index int) error {
	r.mu.Lock()
	if index < 0 || index >= len(r.items) || index == r.selected {
		r.mu.Unlock()
		return nil
	}
	r.selected = index
	callback := r.opts.onChange
	label := r.items[index].Label
	r.mu.Unlock()

	if callback != nil {
		return callback(index, label)
	}
	return nil
}

// itemAtX returns the item under the provided horizontal cell coordinate.
func (r *Radio) itemAtX(x int) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	curX := 0
	for i, item := range r.items {
		width := runewidth.StringWidth(renderItem(item, i == r.selected, r.opts.indicatorGap))
		if x >= curX && x < curX+width {
			return i
		}
		curX += width + r.opts.gap
	}
	return -1
}

// cloneItems copies the item slice and nested cell option slices.
func cloneItems(items []Item) []Item {
	out := append([]Item(nil), items...)
	for i := range out {
		out[i].CellOpts = append([]cell.Option(nil), out[i].CellOpts...)
		out[i].SelectedCellOpts = append([]cell.Option(nil), out[i].SelectedCellOpts...)
	}
	return out
}

// normalizeItems fills in default indicator strings and cell styling.
func normalizeItems(items []Item, opts *options) {
	for i := range items {
		normalizeIndicatorText(&items[i], opts.indicators)
		if len(items[i].CellOpts) == 0 {
			items[i].CellOpts = append([]cell.Option(nil), DefaultCellOpts...)
		}
		if len(items[i].SelectedCellOpts) == 0 {
			items[i].SelectedCellOpts = append([]cell.Option(nil), DefaultSelectedCellOpts...)
		}
	}
}

// renderItem returns the one-line text for a single radio item.
func renderItem(item Item, selected bool, indicatorGap int) string {
	indicator := item.UnselectedText
	if selected {
		indicator = item.SelectedText
	}
	return indicator + strings.Repeat(" ", indicatorGap) + item.Label
}

// widthFor returns the maximum width needed to render the full radio group.
func widthFor(items []Item, gap, indicatorGap int) int {
	width := 0
	for i, item := range items {
		itemWidth := runewidth.StringWidth(renderItem(item, false, indicatorGap))
		if selectedWidth := runewidth.StringWidth(renderItem(item, true, indicatorGap)); selectedWidth > itemWidth {
			itemWidth = selectedWidth
		}
		width += itemWidth
		if i < len(items)-1 {
			width += gap
		}
	}
	return width
}

// normalizeIndicatorText resolves per-item indicator strings from defaults.
func normalizeIndicatorText(item *Item, defaults IndicatorSet) {
	if item.SelectedText == "" {
		if item.SelectedRune != 0 {
			item.SelectedText = string(item.SelectedRune)
		} else if defaults.Selected != "" {
			item.SelectedText = defaults.Selected
		} else {
			item.SelectedText = string(DefaultSelectedRune)
		}
	}
	if item.UnselectedText == "" {
		if item.UnselectedRune != 0 {
			item.UnselectedText = string(item.UnselectedRune)
		} else if defaults.Unselected != "" {
			item.UnselectedText = defaults.Unselected
		} else {
			item.UnselectedText = string(DefaultUnselectedRune)
		}
	}
}

// fmtInvalidSelected returns a standard out-of-range selection error.
func fmtInvalidSelected(index, count int) error {
	return fmt.Errorf("invalid selected index %d, want 0 <= selected < %d", index, count)
}
