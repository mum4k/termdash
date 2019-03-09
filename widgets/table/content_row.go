package table

// content_row.go defines a type that represents a single row in the table.

import (
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
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

// RowHighlighted sets the row as highlighted, the user can then change which
// row is highlighted using keyboard or mouse input.
func RowHighlighted() RowOption {
	return rowOption(func(r *Row) {
		r.highlighted = true
	})
}

// RowCallback allows this row to be activated and provides a function that
// should be called upon each row activation by the user.
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
		r.hierarchical.height = height
	})
}

// RowHorizontalCellPadding sets the horizontal space between cell wall and its
// content as the number of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
// This is a hierarchical option, it overrides the one provided at Content
// level and can be overridden when provided at the Cell level.
func RowHorizontalCellPadding(cells int) RowOption {
	return rowOption(func(r *Row) {
		r.hierarchical.horizontalCellPadding = &cells
	})
}

// RowVerticalCellPadding sets the vertical space between cell wall and its
// content as the number of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
// This is a hierarchical option, it overrides the one provided at Content
// level and can be overridden when provided at the Cell level.
func RowVerticalCellPadding(cells int) RowOption {
	return rowOption(func(r *Row) {
		r.hierarchical.verticalCellPadding = &cells
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

// Row is one row in the table.
type Row struct {
	cells []*Cell

	rowCallback  RowCallbackFn
	highlighted  bool
	hierarchical *hierarchicalOptions
}

// NewHeader returns a new Row that will be the header of the table.
// The header remains visible while scrolling and allows for sorting of content
// based on its values. Header row cannot be highlighted.
// Content can only have one header Row.
func NewHeader(cells []*Cell, opts ...RowOption) *Row {
	return nil
}

// NewRow returns a new Row instance with the provided cells.
// If you need to apply options at the Row level, use NewRowWithOpts.
// If you need to add a table header Row, use NewHeader.
func NewRow(cells ...*Cell) *Row {
	return nil
}

// NewRowWithOpts returns a new Row instance with the provided cells and applies
// the row options.
func NewRowWithOpts(cells []*Cell, opts ...RowOption) *Row {
	return nil
}
