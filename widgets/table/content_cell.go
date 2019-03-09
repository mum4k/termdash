package table

// content_cell.go defines a type that represents a single cell in the table.

import (
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
)

// CellOption is used to provide options to NewCellWithOpts.
type CellOption interface {
	// set sets the provided option.
	set(*Cell)
}

// cellOption implements CellOption.
type cellOption func(*Cell)

// set implements Option.set.
func (co cellOption) set(c *Cell) {
	co(c)
}

// CellColSpan configures the number of columns this cell spans.
// The number must be a non-zero positive integer.
// Defaults to a cell spanning just one column.
func CellColSpan(cols int) CellOption {
	return cellOption(func(c *Cell) {
		c.colSpan = cols
	})
}

// CellRowSpan configures the number of rows this cell spans.
// The number must be a non-zero positive integer.
// Defaults to a cell spanning just one row.
func CellRowSpan(rows int) CellOption {
	return cellOption(func(c *Cell) {
		c.rowSpan = rows
	})
}

// CellOpts sets cell options for the cells that contain the table cells and
// cells.
// This is a hierarchical option, it overrides the one provided at Content or
// Row level and can be overridden at the Data level.
func CellOpts(cellOpts ...cell.Option) CellOption {
	return cellOption(func(c *Cell) {
		c.hierarchical.cellOpts = cellOpts
	})
}

// CellHeight sets the height of cells to the provided number of cells.
// The number must be a non-zero positive integer.
// Defaults to cell height automatically adjusted to the content.
// This is a hierarchical option, it overrides the one provided at Content or
// Row level.
func CellHeight(height int) CellOption {
	return cellOption(func(c *Cell) {
		c.hierarchical.height = height
	})
}

// CellHorizontalCellPadding sets the horizontal space between cell wall and its
// content as the number of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
// This is a hierarchical option, it overrides the one provided at Content or
// Row level.
func CellHorizontalCellPadding(cells int) CellOption {
	return cellOption(func(c *Cell) {
		c.hierarchical.horizontalCellPadding = &cells
	})
}

// CellVerticalCellPadding sets the vertical space between cell wall and its
// content as the number of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
// This is a hierarchical option, it overrides the one provided at Content or
// Row level.
func CellVerticalCellPadding(cells int) CellOption {
	return cellOption(func(c *Cell) {
		c.hierarchical.verticalCellPadding = &cells
	})
}

// CellAlignHorizontal sets the horizontal alignment for the content.
// Defaults for left horizontal alignment.
// This is a hierarchical option, it overrides the one provided at Content or
// Row level.
func CellAlignHorizontal(h align.Horizontal) CellOption {
	return cellOption(func(c *Cell) {
		c.hierarchical.alignHorizontal = &h
	})
}

// CellAlignVertical sets the vertical alignment for the content.
// Defaults for top vertical alignment.
// This is a hierarchical option, it overrides the one provided at Content or
// Row level.
func CellAlignVertical(v align.Vertical) CellOption {
	return cellOption(func(c *Cell) {
		c.hierarchical.alignVertical = &v
	})
}

// Cell is one cell in a Row.
type Cell struct {
	data []*Data

	colSpan      int
	rowSpan      int
	hierarchical *hierarchicalOptions
}

// NewCell returns a new Cell with the provided text.
// If you need to apply options at the Cell or Data level use NewCellWithOpts.
func NewCell(text string) *Cell {
	return nil
}

// NewCellWithOpts returns a new Cell with the provided data and options.
func NewCellWithOpts(data []*Data, opts ...CellOption) *Cell {
	return nil
}
