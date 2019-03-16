// Copyright 2019 Google Inc.
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

package table

// content_row.go defines a type that represents a single row in the table.

import (
	"bytes"
	"fmt"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/wrap"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// RowCallbackFn is a function called when the user activates a Row.
//
// Only Rows that have the RowCallback option can be activated. A Row can be
// activated by a mouse click or by pressing any non-navigation key (see the
// NavigationKeys option) while the row is highlighted.
//
// The called function receives the event that was used to activate the Row.
// The event is either *terminalapi.Keyboard or *terminalapi.Mouse.
//
// The callback function must be light-weight, ideally just storing a value and
// returning, since more button presses might occur.
//
// The callback function must be thread-safe as the mouse or keyboard events
// that activate the row are processed in a separate goroutine.
//
// If the function returns an error, the widget will forward it back to the
// termdash infrastructure which causes a panic, unless the user provided a
// termdash.ErrorHandler.
type RowCallbackFn func(event terminalapi.Event) error

// RowOption is used to provide options to NewRowWithOpts.
type RowOption interface {
	// set sets the provided option.
	set(*Row)
}

// rowOption implements RowOption.
type rowOption func(*Row)

// set implements Option.set.
func (ro rowOption) set(r *Row) {
	ro(r)
}

// RowCallback allows this row to be activated and provides a function that
// should be called upon each row activation by the user.
// This option cannot be used on the header row.
func RowCallback(fn RowCallbackFn) RowOption {
	return rowOption(func(r *Row) {
		r.rowCallback = fn
	})
}

// RowCellOpts sets cell options for the cells that contain the table rows and
// cells.
// This is a hierarchical option, it overrides the one provided at Content
// level and can be overridden when provided at the Cell or Data level.
func RowCellOpts(cellOpts ...cell.Option) RowOption {
	return rowOption(func(r *Row) {
		r.hierarchical.cellOpts = cellOpts
	})
}

// RowHeight sets the height of rows to the provided number of cells.
// The number must be a non-zero positive integer.
// Defaults to row height automatically adjusted to the content.
// This is a hierarchical option, it overrides the one provided at Content
// level and can be overridden when provided at the Cell level.
func RowHeight(height int) RowOption {
	return rowOption(func(r *Row) {
		r.hierarchical.height = &height
	})
}

// RowHorizontalPadding sets the horizontal space between cell wall and its
// content as the number of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
// This is a hierarchical option, it overrides the one provided at Content
// level and can be overridden when provided at the Cell level.
func RowHorizontalPadding(cells int) RowOption {
	return rowOption(func(r *Row) {
		r.hierarchical.horizontalPadding = &cells
	})
}

// RowVerticalPadding sets the vertical space between cell wall and its
// content as the number of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
// This is a hierarchical option, it overrides the one provided at Content
// level and can be overridden when provided at the Cell level.
func RowVerticalPadding(cells int) RowOption {
	return rowOption(func(r *Row) {
		r.hierarchical.verticalPadding = &cells
	})
}

// RowAlignHorizontal sets the horizontal alignment for the content.
// Defaults for left horizontal alignment.
// This is a hierarchical option, it overrides the one provided at Content
// level and can be overridden when provided at the Cell level.
func RowAlignHorizontal(h align.Horizontal) RowOption {
	return rowOption(func(r *Row) {
		r.hierarchical.alignHorizontal = &h
	})
}

// RowAlignVertical sets the vertical alignment for the content.
// Defaults for top vertical alignment.
// This is a hierarchical option, it overrides the one provided at Content
// level and can be overridden when provided at the Cell level.
func RowAlignVertical(v align.Vertical) RowOption {
	return rowOption(func(r *Row) {
		r.hierarchical.alignVertical = &v
	})
}

// RowWrapAtWords sets the content of cells in the row to be wrapped if it
// cannot fit fully.
// Defaults is to not wrap, text that is too long will be trimmed instead.
// This is a hierarchical option, it overrides the one provided at Content
// level and can be overridden when provided at the Cell level.
func RowWrapAtWords() RowOption {
	return rowOption(func(r *Row) {
		wm := wrap.AtWords
		r.hierarchical.wrapMode = &wm
	})
}

// Row is one row in the table.
type Row struct {
	// cells are the cells in this row.
	cells []*Cell

	// isHeader asserts if this row is the header of the table.
	isHeader bool

	// rowCallback is the function to call when this row is activated.
	// Can be nil if the row cannot be activated and is always nil on the
	// header row.
	rowCallback RowCallbackFn
	// hierarchical are the specified hierarchical options at the row level.
	hierarchical *hierarchicalOptions
}

// String implements fmt.Stringer.
func (r *Row) String() string {
	if len(r.cells) == 0 {
		return "Row{}"
	}
	var b bytes.Buffer
	for _, c := range r.cells {
		b.WriteString(c.String())
	}
	return fmt.Sprintf("Row{%v|}", b.String())
}

// NewHeader returns a new Row that will be the header of the table.
// The header remains visible while scrolling and allows for sorting of content
// based on its values. Header row cannot be highlighted.
// Content can only have one header Row.
// If you need to apply options at the Row level, use NewHeaderWithOpts.
func NewHeader(cells ...*Cell) *Row {
	return NewHeaderWithOpts(cells)
}

// NewHeaderWithOpts returns a new Row that will be the header of the table and
// applies the provided options.
// The header remains visible while scrolling and allows for sorting of content
// based on its values. Header row cannot be highlighted.
// Content can only have one header Row.
func NewHeaderWithOpts(cells []*Cell, opts ...RowOption) *Row {
	r := NewRowWithOpts(cells, opts...)
	r.isHeader = true
	return r
}

// NewRow returns a new Row instance with the provided cells.
// If you need to apply options at the Row level, use NewRowWithOpts.
// If you need to add a table header Row, use NewHeader.
func NewRow(cells ...*Cell) *Row {
	return NewRowWithOpts(cells)
}

// NewRowWithOpts returns a new Row instance with the provided cells and applies
// the row options.
func NewRowWithOpts(cells []*Cell, opts ...RowOption) *Row {
	r := &Row{
		cells:        cells,
		hierarchical: &hierarchicalOptions{},
	}
	for _, opt := range opts {
		opt.set(r)
	}
	return r
}

// effectiveColumns returns the number of columns this row effectively occupies.
// This accounts for cells that specify colSpan > 1.
func (r *Row) effectiveColumns() int {
	var cols int
	for _, c := range r.cells {
		cols += c.colSpan
	}
	return cols
}
