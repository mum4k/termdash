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

// content.go defines a type that allow callers to populate the table with
// content.

import (
	"fmt"
	"sync"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/wrap"
	"github.com/mum4k/termdash/linestyle"
)

// ContentOption is used to provide options to NewContent.
type ContentOption interface {
	// set sets the provided option.
	set(*contentOptions)
}

// contentOptions stores options that apply to the content level.
type contentOptions struct {
	border              linestyle.LineStyle
	borderCellOpts      []cell.Option
	columnWidthsPercent []int
	horizontalSpacing   *int
	verticalSpacing     *int

	// hierarchical are the specified hierarchical options at the content
	// level.
	hierarchical *hierarchicalOptions
}

// newContentOptions returns a new contentOptions instance with the options
// applied.
func newContentOptions(opts ...ContentOption) *contentOptions {
	co := &contentOptions{
		hierarchical: &hierarchicalOptions{},
	}
	for _, opt := range opts {
		opt.set(co)
	}
	return co
}

// getHorizontalSpacing returns the user provided horizontal spacing value or
// zero if unset.
func (co *contentOptions) getHorizontalSpacing() int {
	if co.horizontalSpacing != nil {
		return *co.horizontalSpacing
	}
	return 0
}

// getVerticalSpacing returns the user provided vertical spacing value or
// zero if unset.
func (co *contentOptions) getVerticalSpacing() int {
	if co.verticalSpacing != nil {
		return *co.verticalSpacing
	}
	return 0
}

// contentOption implements ContentOption.
type contentOption func(*contentOptions)

// set implements Option.set.
func (co contentOption) set(cOpts *contentOptions) {
	co(cOpts)
}

// Border configures the table to have a border of the specified line style.
// Defaults to linestyle.None which means no border.
func Border(ls linestyle.LineStyle) ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		cOpts.border = ls
	})
}

// BorderCellOpts sets cell options for the cells that contain the border.
func BorderCellOpts(cellOpts ...cell.Option) ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		cOpts.borderCellOpts = cellOpts
	})
}

// ColumnWidthsPercent sets the widths of columns to the provided percentage.
// The number of values must match the number of Columns specified on the call
// to NewContent. All the values must be in the range 0 < v <= 100 and the sum
// of the values must be 100.
// If content wrapping isn't enabled (see WrapContent), defaults to column
// width automatically adjusted to the content. When wrapping is enabled, all
// columns will have equal width.
func ColumnWidthsPercent(widths ...int) ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		cOpts.columnWidthsPercent = widths
	})
}

// HorizontalSpacing sets the horizontal space between cells as the number
// of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
func HorizontalSpacing(cells int) ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		cOpts.horizontalSpacing = &cells
	})
}

// VerticalSpacing sets the vertical space between cells as the number
// of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
func VerticalSpacing(cells int) ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		cOpts.verticalSpacing = &cells
	})
}

// ContentCellOpts sets cell options for the cells that contain the table rows
// and cells.
// This is a hierarchical option and can be overridden when provided at Row,
// Cell or Data level.
func ContentCellOpts(cellOpts ...cell.Option) ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		cOpts.hierarchical.cellOpts = cellOpts
	})
}

// ContentRowHeight sets the height of rows to the provided number of cells.
// The number must be a non-zero positive integer.
// Rows still use larger than provided height if wrapping is enabled and the
// content doesn't fit.
// Defaults to row height automatically adjusted to the content.
// This is a hierarchical option and can be overridden when provided at Row
// level.
func ContentRowHeight(height int) ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		cOpts.hierarchical.height = &height
	})
}

// HorizontalPadding sets the horizontal space between cell wall and its
// content as the number of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
// This is a hierarchical option and can be overridden when provided at Row
// or Cell level.
func HorizontalPadding(cells int) ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		cOpts.hierarchical.horizontalPadding = &cells
	})
}

// VerticalPadding sets the vertical space between cell wall and its
// content as the number of cells on the terminal that are left empty.
// The value must be a non-zero positive integer.
// Defaults to zero cells.
// This is a hierarchical option and can be overridden when provided at Row
// or Cell level.
func VerticalPadding(cells int) ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		cOpts.hierarchical.verticalPadding = &cells
	})
}

// AlignHorizontal sets the horizontal alignment for the content.
// Defaults for left horizontal alignment.
// This is a hierarchical option and can be overridden when provided at Row
// or Cell level.
func AlignHorizontal(h align.Horizontal) ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		cOpts.hierarchical.alignHorizontal = &h
	})
}

// AlignVertical sets the vertical alignment for the content.
// Defaults for top vertical alignment.
// This is a hierarchical option and can be overridden when provided at Row
// or Cell level.
func AlignVertical(v align.Vertical) ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		cOpts.hierarchical.alignVertical = &v
	})
}

// WrapAtWords sets the content of individual cells to be wrapped if it
// cannot fit fully.
// Defaults is to not wrap, text that is too long will be trimmed instead.
// This is a hierarchical option and can be overridden when provided at Row
// or Cell level.
func WrapAtWords() ContentOption {
	return contentOption(func(cOpts *contentOptions) {
		wm := wrap.AtWords
		cOpts.hierarchical.wrapMode = &wm
	})
}

// Columns specifies the number of columns in the table.
type Columns int

// Content is the content displayed in the table.
//
// Content is organized into rows of cells. Each cell can zero, one or multiple
// instances of text data with their own cell options.
//
// Certain options are applied hierarchically, the values provided at the
// Content level apply to all child rows and cells. Specifying a different
// value at a lower level overrides the values provided above.
//
// This object is thread-safe.
type Content struct {
	// cols is the number of columns in the content.
	cols Columns
	// header is the header row, or nil if one wasn't provided.
	header *Row
	// rows are the rows in the table.
	rows []*Row

	// layout describes the layout of this table on a canvas.
	layout *contentLayout

	// opts are the options provided to NewContent.
	opts *contentOptions

	// mu protects the Content
	mu sync.Mutex
}

// NewContent returns a new Content instance.
//
// The number of columns must be a non-zero positive integer.
// All rows must contain the same number of columns (the same number of cells)
// allowing for the CellColSpan option.
func NewContent(cols Columns, rows []*Row, opts ...ContentOption) (*Content, error) {
	c := &Content{
		cols:   cols,
		layout: &contentLayout{},
		opts:   newContentOptions(opts...),
	}
	for _, r := range rows {
		if err := c.addRow(r); err != nil {
			return nil, err
		}
	}

	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// validate validates the content.
func (c *Content) validate() error {
	return validateContent(c)
}

// AddRow adds a row to the content.
// If you need to apply options at the Row level, use AddRowWithOpts.
func (c *Content) AddRow(cells ...*Cell) error {
	return c.AddRowWithOpts(cells)
}

// addRow adds the row to the content.
// Maps hierarchical options to parents and applies inherited cell options to
// data cells.
func (c *Content) addRow(row *Row) error {
	row.hierarchical.parent = c.opts.hierarchical
	for _, tableCell := range row.cells {
		tableCell.hierarchical.parent = row.hierarchical
		tableCell.trimmable = tableCell.hierarchical.getWrapMode() == wrap.Never
		for _, dataCell := range tableCell.data.cells {
			if !dataCell.Opts.IsDefault() {
				continue
			}
			tableCell.hierarchical.getCellOpts().Set(dataCell.Opts)
		}
	}

	if row.isHeader {
		if c.header != nil {
			return fmt.Errorf("the content can only have one header row, already have: %v", c.header)
		}
		c.header = row
	} else {
		c.rows = append(c.rows, row)
	}
	return nil
}

// AddRowWithOpts adds a row to the content and applies the options.
func (c *Content) AddRowWithOpts(cells []*Cell, opts ...RowOption) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.addRow(NewRowWithOpts(cells, opts...)); err != nil {
		return err
	}
	return c.validate()
}
