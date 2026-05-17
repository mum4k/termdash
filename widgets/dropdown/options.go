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

package dropdown

// options.go contains configurable options for Dropdown.

import (
	"fmt"

	"github.com/mum4k/termdash/cell"
)

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*options)
}

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// SelectFn is called when the user commits a new dropdown selection.
//
// The callback must be thread-safe because it is triggered from the keyboard
// and mouse event handling paths, which run in separate goroutines.
type SelectFn func(index int, label string) error

// options holds the provided options.
type options struct {
	selected         int
	width            int
	glyphs           GlyphProfile
	cellOpts         []cell.Option
	focusedCellOpts  []cell.Option
	selectedCellOpts []cell.Option
	borderCellOpts   []cell.Option
	onSelect         SelectFn
}

// BorderRunes defines the UTF-8 runes used to draw the open dropdown box.
type BorderRunes struct {
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
	Horizontal  rune
	Vertical    rune
}

// GlyphProfile groups reusable dropdown glyph choices.
type GlyphProfile struct {
	ClosedArrow      rune
	OpenArrow        rune
	SelectedPrefix   string
	UnselectedPrefix string
	Border           BorderRunes
}

// validate validates the provided options against the item set.
func (o *options) validate(items []string) error {
	if len(items) == 0 {
		return fmt.Errorf("at least one item must be specified")
	}
	for i, item := range items {
		if item == "" {
			return fmt.Errorf("item[%d] is empty, all items must contain some text", i)
		}
	}
	if o.selected < 0 || o.selected >= len(items) {
		return fmt.Errorf("invalid selected index %d, want 0 <= selected < %d", o.selected, len(items))
	}
	if min := minTriggerWidth; o.width < min {
		return fmt.Errorf("invalid width %d, want %d <= width", o.width, min)
	}
	return nil
}

// newOptions returns options with the default values set.
func newOptions(items []string) *options {
	return &options{
		selected:         0,
		width:            widthFor(items),
		glyphs:           GlyphProfiles.Classic,
		cellOpts:         []cell.Option{cell.FgColor(DefaultTextColor)},
		focusedCellOpts:  []cell.Option{cell.FgColor(DefaultFocusedTextColor)},
		selectedCellOpts: []cell.Option{cell.FgColor(DefaultSelectedTextColor), cell.BgColor(DefaultSelectedFillColor)},
		borderCellOpts:   []cell.Option{cell.FgColor(DefaultBorderColor)},
	}
}

// Default colors used by the dropdown widget.
var (
	DefaultTextColor         = cell.ColorWhite
	DefaultFocusedTextColor  = cell.ColorCyan
	DefaultSelectedTextColor = cell.ColorWhite
	DefaultSelectedFillColor = cell.ColorNumber(60)
	DefaultBorderColor       = cell.ColorCyan
	GlyphProfiles            = struct {
		Classic GlyphProfile
		Minimal GlyphProfile
	}{
		Classic: GlyphProfile{
			ClosedArrow:      '▼',
			OpenArrow:        '▲',
			SelectedPrefix:   ">",
			UnselectedPrefix: " ",
			Border: BorderRunes{
				TopLeft:     '┌',
				TopRight:    '┐',
				BottomLeft:  '└',
				BottomRight: '┘',
				Horizontal:  '─',
				Vertical:    '│',
			},
		},
		Minimal: GlyphProfile{
			ClosedArrow:      '▾',
			OpenArrow:        '▴',
			SelectedPrefix:   "›",
			UnselectedPrefix: " ",
			Border: BorderRunes{
				TopLeft:     '╭',
				TopRight:    '╮',
				BottomLeft:  '╰',
				BottomRight: '╯',
				Horizontal:  '─',
				Vertical:    '│',
			},
		},
	}
)

// Selected sets the initially selected item index.
func Selected(index int) Option {
	return option(func(opts *options) {
		opts.selected = index
	})
}

// Width sets the dropdown width in terminal cells.
//
// Width controls both the closed trigger and the open list box.
func Width(cells int) Option {
	return option(func(opts *options) {
		opts.width = cells
	})
}

// Arrows sets the UTF-8 runes used while the dropdown is closed and open.
func Arrows(closed, open rune) Option {
	return option(func(opts *options) {
		if closed != 0 {
			opts.glyphs.ClosedArrow = closed
		}
		if open != 0 {
			opts.glyphs.OpenArrow = open
		}
	})
}

// RowPrefixes sets the leading text used for selected and unselected rows.
func RowPrefixes(selected, unselected string) Option {
	return option(func(opts *options) {
		if selected != "" {
			opts.glyphs.SelectedPrefix = selected
		}
		if unselected != "" {
			opts.glyphs.UnselectedPrefix = unselected
		}
	})
}

// BorderGlyphs sets the runes used for the open dropdown box border.
func BorderGlyphs(runes BorderRunes) Option {
	return option(func(opts *options) {
		if runes.TopLeft != 0 {
			opts.glyphs.Border.TopLeft = runes.TopLeft
		}
		if runes.TopRight != 0 {
			opts.glyphs.Border.TopRight = runes.TopRight
		}
		if runes.BottomLeft != 0 {
			opts.glyphs.Border.BottomLeft = runes.BottomLeft
		}
		if runes.BottomRight != 0 {
			opts.glyphs.Border.BottomRight = runes.BottomRight
		}
		if runes.Horizontal != 0 {
			opts.glyphs.Border.Horizontal = runes.Horizontal
		}
		if runes.Vertical != 0 {
			opts.glyphs.Border.Vertical = runes.Vertical
		}
	})
}

// GlyphSet sets the dropdown glyphs from a reusable group.
func GlyphSet(set GlyphProfile) Option {
	return option(func(opts *options) {
		if set.ClosedArrow != 0 {
			opts.glyphs.ClosedArrow = set.ClosedArrow
		}
		if set.OpenArrow != 0 {
			opts.glyphs.OpenArrow = set.OpenArrow
		}
		if set.SelectedPrefix != "" {
			opts.glyphs.SelectedPrefix = set.SelectedPrefix
		}
		if set.UnselectedPrefix != "" {
			opts.glyphs.UnselectedPrefix = set.UnselectedPrefix
		}
		opts.glyphs.Border = mergeBorderRunes(opts.glyphs.Border, set.Border)
	})
}

// mergeBorderRunes overlays any non-zero runes from next onto current.
func mergeBorderRunes(current, next BorderRunes) BorderRunes {
	if next.TopLeft != 0 {
		current.TopLeft = next.TopLeft
	}
	if next.TopRight != 0 {
		current.TopRight = next.TopRight
	}
	if next.BottomLeft != 0 {
		current.BottomLeft = next.BottomLeft
	}
	if next.BottomRight != 0 {
		current.BottomRight = next.BottomRight
	}
	if next.Horizontal != 0 {
		current.Horizontal = next.Horizontal
	}
	if next.Vertical != 0 {
		current.Vertical = next.Vertical
	}
	return current
}

// CellOpts sets the default cell styling used for the trigger and unselected
// list rows.
func CellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.cellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// FocusedCellOpts sets the styling used for the closed trigger while the
// widget's container is focused.
func FocusedCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.focusedCellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// SelectedCellOpts sets the styling used for the active row while the list is
// open.
func SelectedCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.selectedCellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// BorderCellOpts sets the styling used for the open dropdown box border.
func BorderCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.borderCellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// OnSelect sets the dropdown's selection hook.
//
// This is the widget's canonical callback surface. Callers that need delayed
// or asynchronous work should build that from this hook so the widget keeps a
// single stable event path.
func OnSelect(fn SelectFn) Option {
	return option(func(opts *options) {
		opts.onSelect = fn
	})
}
