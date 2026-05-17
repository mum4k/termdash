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

// Package emojikeyboard implements a paginated emoji grid widget.
package emojikeyboard

import (
	"fmt"
	"image"
	"slices"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// EmojiKeyboard displays a paginated grid of emojis.
type EmojiKeyboard struct {
	mu sync.Mutex

	emojis []string
	opts   *options

	page      int
	cursorIdx int
	cols      int
	rows      int
	pageSize  int
	canvasW   int
	canvasH   int

	selectedEmoji string
}

// New creates a paginated emoji keyboard widget.
func New(opts ...Option) *EmojiKeyboard {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}
	o.normalize()

	emojis := append([]string(nil), o.emojis...)
	if len(emojis) == 0 {
		emojis = allEmojis()
	}

	selected := ""
	if containsEmoji(emojis, o.initialEmoji) {
		selected = o.initialEmoji
	}

	return &EmojiKeyboard{
		emojis:        emojis,
		opts:          o,
		selectedEmoji: selected,
	}
}

// Draw renders the emoji grid and pagination controls.
func (ek *EmojiKeyboard) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	ek.mu.Lock()
	defer ek.mu.Unlock()

	width := cvs.Area().Dx()
	height := cvs.Area().Dy()
	ek.canvasW = width
	ek.canvasH = height

	if width < ek.opts.cellWidth || height < ek.opts.cellHeight+1 {
		return draw.ResizeNeeded(cvs)
	}

	cols := maxInt(width/ek.opts.cellWidth, 1)
	rows := maxInt((height-1)/ek.opts.cellHeight, 1)
	pageSize := maxInt(cols*rows, 1)

	ek.cols = cols
	ek.rows = rows
	ek.pageSize = pageSize

	totalPages := ek.totalPages()
	if ek.page >= totalPages {
		ek.page = totalPages - 1
	}
	if ek.page < 0 {
		ek.page = 0
	}
	if ek.cursorIdx >= pageSize {
		ek.cursorIdx = pageSize - 1
	}
	if ek.cursorIdx < 0 {
		ek.cursorIdx = 0
	}

	start, end := ek.pageBounds()
	for i, emoji := range ek.emojis[start:end] {
		col := i % cols
		row := i / cols
		x := col * ek.opts.cellWidth
		y := row * ek.opts.cellHeight
		if y >= height-1 {
			continue
		}
		ek.drawEmojiCell(cvs, image.Point{X: x, Y: y}, emoji, i == ek.cursorIdx && meta.Focused, emoji == ek.selectedEmoji)
	}

	ek.drawPaginationRow(cvs, width, height, totalPages)
	return nil
}

// Keyboard handles keyboard navigation and selection.
func (ek *EmojiKeyboard) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	ek.mu.Lock()
	defer ek.mu.Unlock()
	ek.ensureGeometry()

	switch k.Key {
	case keyboard.KeyArrowRight:
		ek.moveCursor(1)
	case keyboard.KeyArrowLeft:
		ek.moveCursor(-1)
	case keyboard.KeyArrowDown:
		ek.moveCursor(ek.cols)
	case keyboard.KeyArrowUp:
		ek.moveCursor(-ek.cols)
	case keyboard.KeyPgDn:
		ek.nextPage()
	case keyboard.KeyPgUp:
		ek.prevPage()
	case keyboard.KeyEnter:
		ek.selectCurrent()
	}
	return nil
}

// Mouse handles mouse-based pagination and selection.
func (ek *EmojiKeyboard) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	if m.Button != mouse.ButtonLeft {
		return nil
	}

	ek.mu.Lock()
	defer ek.mu.Unlock()

	if ek.canvasH > 0 && m.Position.Y == ek.canvasH-1 {
		switch m.Position.X {
		case 0:
			ek.prevPage()
		case ek.canvasW - 1:
			if ek.canvasW > 1 {
				ek.nextPage()
			}
		}
		return nil
	}

	if ek.cols == 0 || ek.opts.cellWidth == 0 || ek.opts.cellHeight == 0 {
		return nil
	}

	col := m.Position.X / ek.opts.cellWidth
	row := m.Position.Y / ek.opts.cellHeight
	idx := row*ek.cols + col
	if idx < 0 || idx >= ek.pageSize {
		return nil
	}

	absIdx := ek.page*ek.pageSize + idx
	if absIdx < 0 || absIdx >= len(ek.emojis) {
		return nil
	}

	ek.cursorIdx = idx
	ek.fireSelect(ek.emojis[absIdx])
	return nil
}

// ensureGeometry seeds keyboard navigation with a reasonable default geometry
// before the first draw computes the real canvas-based values.
func (ek *EmojiKeyboard) ensureGeometry() {
	if ek.cols > 0 && ek.rows > 0 && ek.pageSize > 0 {
		return
	}
	ek.cols = maxInt(len(ek.emojis), 1)
	ek.rows = 1
	ek.pageSize = maxInt(len(ek.emojis), 1)
}

// Options implements widgetapi.Widget.Options.
func (ek *EmojiKeyboard) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize:  image.Point{X: ek.opts.cellWidth, Y: ek.opts.cellHeight + 1},
		WantKeyboard: widgetapi.KeyScopeFocused,
		WantMouse:    widgetapi.MouseScopeWidget,
	}
}

// SelectedEmoji returns the currently confirmed emoji selection.
func (ek *EmojiKeyboard) SelectedEmoji() string {
	ek.mu.Lock()
	defer ek.mu.Unlock()
	return ek.selectedEmoji
}

// SetEmojis replaces the emoji catalog shown by the widget.
func (ek *EmojiKeyboard) SetEmojis(emojis []string) {
	ek.mu.Lock()
	defer ek.mu.Unlock()

	ek.emojis = append([]string(nil), emojis...)
	if len(ek.emojis) == 0 {
		ek.emojis = allEmojis()
	}
	if !containsEmoji(ek.emojis, ek.selectedEmoji) {
		ek.selectedEmoji = ""
	}
	ek.page = 0
	ek.cursorIdx = 0
}

// drawEmojiCell renders one emoji cell with its focused or selected treatment.
func (ek *EmojiKeyboard) drawEmojiCell(cvs *canvas.Canvas, origin image.Point, emoji string, isCursor, isSelected bool) {
	opts := ek.cellOpts(isCursor, isSelected)
	if isSelected {
		ek.drawSelectedEmoji(cvs, origin, emoji, opts, isCursor)
		return
	}
	ek.drawEmojiRunes(cvs, origin, emoji, opts)
}

// cellOpts returns the color options for a single emoji cell.
func (ek *EmojiKeyboard) cellOpts(isCursor, isSelected bool) []cell.Option {
	switch {
	case isCursor:
		return []cell.Option{
			cell.FgColor(ek.opts.cursorFg),
			cell.BgColor(ek.opts.cursorBg),
		}
	case isSelected:
		return []cell.Option{
			cell.FgColor(ek.opts.normalFg),
			cell.BgColor(ek.opts.selectedBg),
		}
	default:
		return []cell.Option{cell.FgColor(ek.opts.normalFg)}
	}
}

// drawSelectedEmoji renders the framed selected state for a single emoji.
func (ek *EmojiKeyboard) drawSelectedEmoji(cvs *canvas.Canvas, origin image.Point, emoji string, emojiOpts []cell.Option, isCursor bool) {
	bracketOpts := emojiOpts
	if isCursor {
		bracketOpts = []cell.Option{
			cell.FgColor(ek.opts.selectedBg),
			cell.BgColor(ek.opts.cursorBg),
		}
	}

	if origin.X < ek.canvasW {
		_, _ = cvs.SetCell(origin, ek.opts.selectionGlyphs.Left, bracketOpts...)
	}
	ek.drawEmojiRunes(cvs, image.Point{X: origin.X + 1, Y: origin.Y}, emoji, emojiOpts)

	closeX := origin.X + ek.opts.cellWidth - 2
	if closeX < origin.X+2 {
		closeX = origin.X + 2
	}
	if closeX < ek.canvasW {
		_, _ = cvs.SetCell(image.Point{X: closeX, Y: origin.Y}, ek.opts.selectionGlyphs.Right, bracketOpts...)
	}
}

// drawEmojiRunes writes the emoji runes to the canvas starting at the provided point.
func (ek *EmojiKeyboard) drawEmojiRunes(cvs *canvas.Canvas, start image.Point, emoji string, opts []cell.Option) {
	point := start
	for _, r := range emoji {
		if point.X >= ek.canvasW {
			break
		}
		_, _ = cvs.SetCell(point, r, opts...)
		point.X++
	}
}

// drawPaginationRow renders the pagination strip on the last row.
func (ek *EmojiKeyboard) drawPaginationRow(cvs *canvas.Canvas, width, height, totalPages int) {
	paginationY := height - 1

	prevArrowColor := ek.opts.pageArrowDimFg
	if ek.page > 0 {
		prevArrowColor = ek.opts.pageArrowFg
	}
	nextArrowColor := ek.opts.pageArrowDimFg
	if ek.page < totalPages-1 {
		nextArrowColor = ek.opts.pageArrowFg
	}

	_, _ = cvs.SetCell(image.Point{X: 0, Y: paginationY}, ek.opts.paginationGlyphs.Prev, cell.FgColor(prevArrowColor))

	label := fmt.Sprintf(" pg %d/%d ", ek.page+1, totalPages)
	labelRunes := []rune(label)
	labelEnd := 1 + len(labelRunes)
	for i, r := range labelRunes {
		x := 1 + i
		if x >= width-1 {
			break
		}
		_, _ = cvs.SetCell(image.Point{X: x, Y: paginationY}, r, cell.FgColor(ek.opts.pageInfoFg))
	}

	barStart := labelEnd
	barEnd := width - 1
	barWidth := barEnd - barStart
	if barWidth > 0 && totalPages > 1 {
		filled := barWidth * (ek.page + 1) / totalPages
		for i := 0; i < barWidth; i++ {
			x := barStart + i
			if x >= width-1 {
				break
			}
			glyph := ek.opts.paginationGlyphs.Empty
			if i < filled {
				glyph = ek.opts.paginationGlyphs.Filled
			}
			_, _ = cvs.SetCell(image.Point{X: x, Y: paginationY}, glyph, cell.FgColor(ek.opts.pageBarFg))
		}
	}

	if width > 1 {
		_, _ = cvs.SetCell(image.Point{X: width - 1, Y: paginationY}, ek.opts.paginationGlyphs.Next, cell.FgColor(nextArrowColor))
	}
}

// totalPages returns the page count for the current geometry.
func (ek *EmojiKeyboard) totalPages() int {
	if ek.pageSize <= 0 {
		return 1
	}
	totalPages := (len(ek.emojis) + ek.pageSize - 1) / ek.pageSize
	return maxInt(totalPages, 1)
}

// pageBounds returns the emoji slice bounds for the current page.
func (ek *EmojiKeyboard) pageBounds() (int, int) {
	start := ek.page * ek.pageSize
	end := minInt(start+ek.pageSize, len(ek.emojis))
	return start, end
}

// moveCursor moves the in-page cursor by the provided delta.
func (ek *EmojiKeyboard) moveCursor(delta int) {
	newIdx := ek.cursorIdx + delta
	if newIdx < 0 {
		ek.prevPage()
		return
	}
	if newIdx >= ek.pageSize {
		ek.nextPage()
		return
	}

	absIdx := ek.page*ek.pageSize + newIdx
	if absIdx >= len(ek.emojis) {
		return
	}
	ek.cursorIdx = newIdx
}

// nextPage advances to the next page and resets the cursor.
func (ek *EmojiKeyboard) nextPage() {
	if ek.page < ek.totalPages()-1 {
		ek.page++
		ek.cursorIdx = 0
	}
}

// prevPage moves back one page and resets the cursor.
func (ek *EmojiKeyboard) prevPage() {
	if ek.page > 0 {
		ek.page--
		ek.cursorIdx = 0
	}
}

// selectCurrent confirms the emoji under the current cursor.
func (ek *EmojiKeyboard) selectCurrent() {
	absIdx := ek.page*ek.pageSize + ek.cursorIdx
	if absIdx >= 0 && absIdx < len(ek.emojis) {
		ek.fireSelect(ek.emojis[absIdx])
	}
}

// fireSelect stores the selected emoji and triggers the callback asynchronously.
func (ek *EmojiKeyboard) fireSelect(emoji string) {
	ek.selectedEmoji = emoji
	if ek.opts.onSelect != nil {
		go ek.opts.onSelect(emoji)
	}
}

// containsEmoji reports whether the emoji list contains the provided entry.
func containsEmoji(emojis []string, want string) bool {
	return want != "" && slices.Contains(emojis, want)
}

// minInt returns the smaller of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt returns the larger of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
