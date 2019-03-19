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

// content_cell.go defines a type that represents a single cell in the table.

import (
	"fmt"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/canvas/buffer"
	"github.com/mum4k/termdash/internal/wrap"
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
// Rows still use larger than provided height if wrapping is enabled and the
// content doesn't fit.
// Defaults to cell height automatically adjusted to the content.
// This is a hierarchical option, it overrides the one provided at Content or
// Row level.
func CellHeight(height int) CellOption {
	return cellOption(func(c *Cell) {
		c.hierarchical.height = &height
	})
}

// CellHorizontalPadding sets the horizontal space between cell wall and its
// content as the number of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
// This is a hierarchical option, it overrides the one provided at Content or
// Row level.
func CellHorizontalPadding(cells int) CellOption {
	return cellOption(func(c *Cell) {
		c.hierarchical.horizontalPadding = &cells
	})
}

// CellVerticalPadding sets the vertical space between cell wall and its
// content as the number of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
// This is a hierarchical option, it overrides the one provided at Content or
// Row level.
func CellVerticalPadding(cells int) CellOption {
	return cellOption(func(c *Cell) {
		c.hierarchical.verticalPadding = &cells
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

// CellWrapAtWords sets the content of the cell to be wrapped if it
// cannot fit fully.
// Defaults is to not wrap, text that is too long will be trimmed instead.
// This is a hierarchical option, it overrides the one provided at Content or
// Row level.
func CellWrapAtWords() CellOption {
	return cellOption(func(c *Cell) {
		wm := wrap.AtWords
		c.hierarchical.wrapMode = &wm
	})
}

// Cell is one cell in a Row.
type Cell struct {
	// data are the text data in the cell.
	data *Data
	// width is the width of the data when draws on canvas.
	width int
	// trimmable indicates if the content of this cell would be trimmed if it
	// doesn't fit the columns width.
	trimmable bool

	// wrapped contains the cell data wrapped to lines according to the last
	// known column width.
	wrapped []*Data

	// colSpan specified how many columns does this cell span.
	colSpan int
	// rowSpan specified how many rows does this cell span.
	rowSpan int
	// hierarchical are the specified hierarchical options at the cell level.
	hierarchical *hierarchicalOptions
}

// String implements fmt.Stringer.
func (c *Cell) String() string {
	return fmt.Sprintf("| %v ", c.data.String())
}

// NewCell returns a new Cell with the provided text.
// If you need to apply options at the Cell or Data level use NewCellWithOpts.
// The text contain cannot control characters (unicode.IsControl) or space
// character (unicode.IsSpace) other than:
//   ' ', '\n'
// Any newline ('\n') characters are interpreted as newlines when displaying
// the text.
func NewCell(text string) *Cell {
	return NewCellWithOpts([]*Data{NewData(text)})
}

// NewCellWithOpts returns a new Cell with the provided data and options.
func NewCellWithOpts(data []*Data, opts ...CellOption) *Cell {
	c := &Cell{
		data:         newCombinedData(data),
		width:        dataWidth(data),
		colSpan:      1,
		rowSpan:      1,
		hierarchical: &hierarchicalOptions{},
	}
	for _, opt := range opts {
		opt.set(c)
	}
	return c
}

// wrapToWidth wraps the cell's content to the specified column width.
// This populates c.wrapped.
func (c *Cell) wrapToWidth(cw columnWidth) error {
	c.wrapped = nil
	if buffer.CellsWidth(c.data.cells) == 0 {
		// No content so nothing to wrap.
		return nil
	}

	mode := c.hierarchical.getWrapMode()
	if mode == wrap.Never {
		// No wrapping enabled, all data will be in a single line.
		c.wrapped = []*Data{
			newDataCells(c.data.cells),
		}
		return nil
	}

	w, err := wrap.Cells(c.data.cells, int(cw), mode)
	if err != nil {
		return err
	}

	for _, line := range w {
		c.wrapped = append(c.wrapped, newDataCells(line))
	}
	return nil
}

// dataWidth returns the width of all the runes in this cell when they are printed
// on the terminal.
func dataWidth(data []*Data) int {
	res := 0
	for _, d := range data {
		res += buffer.CellsWidth(d.cells)
	}
	return res
}
