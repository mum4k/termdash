package table

// content_data.go defines a type that represents data within a table cell.

import (
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/internal/canvas/buffer"
)

// DataOption is used to provide options to NewDataWithOpts.
type DataOption interface {
	// set sets the provided option.
	set(*dataOptions)
}

// dataOptions stores the provided data options.
type dataOptions struct {
	cellOpts []cell.Option
}

// dataOption implements DataOption.
type dataOption func(*dataOptions)

// set implements Option.set.
func (do dataOption) set(dOpts *dataOptions) {
	do(dOpts)
}

// DataCellOpts sets options on the cells that contain the data.
func DataCellOpts(cellOpts ...cell.Option) DataOption {
	return dataOption(func(dOpts *dataOptions) {
		dOpts.cellOpts = cellOpts
	})
}

// Data is part of (or the full) the data that is displayed inside one Cell.
type Data struct {
	cells []*buffer.Cell
}

// NewData creates new Data with the provided text and applies the options.
func NewData(text string, opts ...DataOption) *Data {
	dOpts := &dataOptions{}
	for _, opt := range opts {
		opt.set(dOpts)
	}

	var cells []*buffer.Cell
	for _, r := range text {
		cells = append(cells, buffer.NewCell(r, dOpts.cellOpts...))
	}
	return &Data{
		cells: cells,
	}
}
