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

package emojikeyboard

import "github.com/mum4k/termdash/cell"

// OnSelect is the callback signature fired when an emoji is chosen.
type OnSelect func(emoji string)

// Option configures the EmojiKeyboard.
type Option func(*options)

// SelectionGlyphSet groups the UTF-8 runes used to frame a selected emoji.
type SelectionGlyphSet struct {
	Left  rune
	Right rune
}

// PaginationGlyphSet groups the UTF-8 runes used by the page controls.
type PaginationGlyphSet struct {
	Prev   rune
	Next   rune
	Filled rune
	Empty  rune
}

// ColorProfile groups the default colors used by the widget.
type ColorProfile struct {
	NormalFg       cell.Color
	CursorFg       cell.Color
	CursorBg       cell.Color
	SelectedBg     cell.Color
	PageInfoFg     cell.Color
	PageArrowFg    cell.Color
	PageArrowDimFg cell.Color
	PageBarFg      cell.Color
}

// SelectionGlyphSets exposes reusable selection-frame profiles.
var SelectionGlyphSets = struct {
	Block   SelectionGlyphSet
	Bracket SelectionGlyphSet
	Angle   SelectionGlyphSet
}{
	Block:   SelectionGlyphSet{Left: '▐', Right: '▌'},
	Bracket: SelectionGlyphSet{Left: '[', Right: ']'},
	Angle:   SelectionGlyphSet{Left: '⟦', Right: '⟧'},
}

// PaginationGlyphSets exposes reusable pagination profiles.
var PaginationGlyphSets = struct {
	Classic PaginationGlyphSet
	Minimal PaginationGlyphSet
	Line    PaginationGlyphSet
}{
	Classic: PaginationGlyphSet{Prev: '◀', Next: '▶', Filled: '█', Empty: '░'},
	Minimal: PaginationGlyphSet{Prev: '‹', Next: '›', Filled: '■', Empty: '·'},
	Line:    PaginationGlyphSet{Prev: '←', Next: '→', Filled: '━', Empty: '─'},
}

// ColorProfiles exposes reusable color defaults.
var ColorProfiles = struct {
	Default  ColorProfile
	Ice      ColorProfile
	Terminal ColorProfile
}{
	Default: ColorProfile{
		NormalFg:       cell.ColorDefault,
		CursorFg:       cell.ColorDefault,
		CursorBg:       cell.ColorNumber(24),
		SelectedBg:     cell.ColorNumber(22),
		PageInfoFg:     cell.ColorNumber(252),
		PageArrowFg:    cell.ColorNumber(75),
		PageArrowDimFg: cell.ColorNumber(238),
		PageBarFg:      cell.ColorNumber(244),
	},
	Ice: ColorProfile{
		NormalFg:       cell.ColorNumber(252),
		CursorFg:       cell.ColorNumber(252),
		CursorBg:       cell.ColorNumber(24),
		SelectedBg:     cell.ColorNumber(31),
		PageInfoFg:     cell.ColorNumber(153),
		PageArrowFg:    cell.ColorNumber(117),
		PageArrowDimFg: cell.ColorNumber(240),
		PageBarFg:      cell.ColorNumber(81),
	},
	Terminal: ColorProfile{
		NormalFg:       cell.ColorWhite,
		CursorFg:       cell.ColorWhite,
		CursorBg:       cell.ColorGreen,
		SelectedBg:     cell.ColorNumber(28),
		PageInfoFg:     cell.ColorNumber(250),
		PageArrowFg:    cell.ColorGreen,
		PageArrowDimFg: cell.ColorNumber(238),
		PageBarFg:      cell.ColorNumber(34),
	},
}

// options stores the widget configuration.
type options struct {
	onSelect         OnSelect
	emojis           []string
	cellWidth        int
	cellHeight       int
	selectionGlyphs  SelectionGlyphSet
	paginationGlyphs PaginationGlyphSet
	normalFg         cell.Color
	cursorFg         cell.Color
	cursorBg         cell.Color
	selectedBg       cell.Color
	pageInfoFg       cell.Color
	pageArrowFg      cell.Color
	pageArrowDimFg   cell.Color
	pageBarFg        cell.Color
	initialEmoji     string
}

// defaultOptions returns the widget defaults.
func defaultOptions() *options {
	o := &options{
		cellWidth:        5,
		cellHeight:       1,
		selectionGlyphs:  SelectionGlyphSets.Block,
		paginationGlyphs: PaginationGlyphSets.Classic,
	}
	applyColorProfile(o, ColorProfiles.Default)
	return o
}

// applyColorProfile copies a grouped color profile into the options.
func applyColorProfile(o *options, profile ColorProfile) {
	o.normalFg = profile.NormalFg
	o.cursorFg = profile.CursorFg
	o.cursorBg = profile.CursorBg
	o.selectedBg = profile.SelectedBg
	o.pageInfoFg = profile.PageInfoFg
	o.pageArrowFg = profile.PageArrowFg
	o.pageArrowDimFg = profile.PageArrowDimFg
	o.pageBarFg = profile.PageBarFg
}

// normalize clamps and fills option values after user configuration is applied.
func (o *options) normalize() {
	if o.cellWidth < 3 {
		o.cellWidth = 3
	}
	if o.cellHeight < 1 {
		o.cellHeight = 1
	}
	if o.selectionGlyphs.Left == 0 || o.selectionGlyphs.Right == 0 {
		o.selectionGlyphs = SelectionGlyphSets.Block
	}
	if o.paginationGlyphs.Prev == 0 || o.paginationGlyphs.Next == 0 ||
		o.paginationGlyphs.Filled == 0 || o.paginationGlyphs.Empty == 0 {
		o.paginationGlyphs = PaginationGlyphSets.Classic
	}
}

// Emojis replaces the default emoji catalog with a caller-provided list.
func Emojis(emojis ...string) Option {
	return func(o *options) {
		o.emojis = append([]string(nil), emojis...)
	}
}

// CellWidth sets the column width allocated per emoji.
func CellWidth(w int) Option {
	return func(o *options) {
		o.cellWidth = w
	}
}

// CellHeight sets the row height allocated per emoji.
func CellHeight(h int) Option {
	return func(o *options) {
		o.cellHeight = h
	}
}

// OnSelectFunc sets the callback fired when an emoji is selected.
func OnSelectFunc(fn OnSelect) Option {
	return func(o *options) {
		o.onSelect = fn
	}
}

// WithOnSelect sets the callback fired when an emoji is selected.
//
// Deprecated: use OnSelectFunc for new callers.
func WithOnSelect(fn OnSelect) Option {
	return OnSelectFunc(fn)
}

// UseSelectionGlyphSet sets the selection frame from a reusable group.
func UseSelectionGlyphSet(set SelectionGlyphSet) Option {
	return func(o *options) {
		if set.Left != 0 {
			o.selectionGlyphs.Left = set.Left
		}
		if set.Right != 0 {
			o.selectionGlyphs.Right = set.Right
		}
	}
}

// SelectionGlyphs sets the runes used to frame a selected emoji.
func SelectionGlyphs(left, right rune) Option {
	return func(o *options) {
		if left != 0 {
			o.selectionGlyphs.Left = left
		}
		if right != 0 {
			o.selectionGlyphs.Right = right
		}
	}
}

// UsePaginationGlyphSet sets the pagination controls from a reusable group.
func UsePaginationGlyphSet(set PaginationGlyphSet) Option {
	return func(o *options) {
		if set.Prev != 0 {
			o.paginationGlyphs.Prev = set.Prev
		}
		if set.Next != 0 {
			o.paginationGlyphs.Next = set.Next
		}
		if set.Filled != 0 {
			o.paginationGlyphs.Filled = set.Filled
		}
		if set.Empty != 0 {
			o.paginationGlyphs.Empty = set.Empty
		}
	}
}

// PaginationGlyphs sets the pagination runes directly.
func PaginationGlyphs(prev, next, filled, empty rune) Option {
	return func(o *options) {
		if prev != 0 {
			o.paginationGlyphs.Prev = prev
		}
		if next != 0 {
			o.paginationGlyphs.Next = next
		}
		if filled != 0 {
			o.paginationGlyphs.Filled = filled
		}
		if empty != 0 {
			o.paginationGlyphs.Empty = empty
		}
	}
}

// UseColorProfile applies a reusable grouped color profile.
func UseColorProfile(profile ColorProfile) Option {
	return func(o *options) {
		applyColorProfile(o, profile)
	}
}

// NormalFg sets the foreground color for unselected emojis.
func NormalFg(c cell.Color) Option {
	return func(o *options) {
		o.normalFg = c
	}
}

// CursorFg sets the foreground color for the cursor highlight.
func CursorFg(c cell.Color) Option {
	return func(o *options) {
		o.cursorFg = c
	}
}

// CursorBg sets the background color for the cursor highlight.
func CursorBg(c cell.Color) Option {
	return func(o *options) {
		o.cursorBg = c
	}
}

// SelectedBg sets the background color for the confirmed selection highlight.
func SelectedBg(c cell.Color) Option {
	return func(o *options) {
		o.selectedBg = c
	}
}

// PageInfoFg sets the foreground color for the page label.
func PageInfoFg(c cell.Color) Option {
	return func(o *options) {
		o.pageInfoFg = c
	}
}

// PageArrowColors sets the active and inactive pagination arrow colors.
func PageArrowColors(active, inactive cell.Color) Option {
	return func(o *options) {
		o.pageArrowFg = active
		o.pageArrowDimFg = inactive
	}
}

// PageBarFg sets the foreground color for the pagination progress bar.
func PageBarFg(c cell.Color) Option {
	return func(o *options) {
		o.pageBarFg = c
	}
}

// InitialSelection pre-marks the given emoji as selected on first draw.
func InitialSelection(emoji string) Option {
	return func(o *options) {
		o.initialEmoji = emoji
	}
}

// WithInitialSelection pre-marks the given emoji as selected on first draw.
//
// Deprecated: use InitialSelection for new callers.
func WithInitialSelection(emoji string) Option {
	return InitialSelection(emoji)
}
