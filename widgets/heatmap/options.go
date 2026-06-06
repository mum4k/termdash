// Copyright 2020 Google Inc.
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

package heatmap

import (
	"errors"

	"github.com/mum4k/termdash/cell"
)

// options.go contains configurable options for HeatMap.

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*options)
}

// options stores the provided options.
type options struct {
	// The default value is 3
	cellWidth      int
	xLabelCellOpts []cell.Option
	yLabelCellOpts []cell.Option
	axisCellOpts   []cell.Option
	palette        []cell.Color
}

// validate validates the provided options.
func (o *options) validate() error {
	if o.cellWidth < 1 {
		return errors.New("cell width must be >= 1")
	}
	return nil
}

// newOptions returns a new options instance.
func newOptions(opts ...Option) *options {
	opt := &options{
		cellWidth: 3,
	}
	for _, o := range opts {
		o.set(opt)
	}
	return opt
}

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// CellWidth set the width of cells (or grids) in the heat map, not the terminal cell.
// The default height of each cell (grid) is 1 and the width is 3.
func CellWidth(w int) Option {
	return option(func(opts *options) {
		opts.cellWidth = w
	})
}

// SquareCells configures each heatmap value to use two terminal columns, which
// reads closer to a square cell on typical terminals where character cells are
// taller than they are wide.
func SquareCells() Option {
	return CellWidth(2)
}

// XLabelCellOpts set the cell options for the labels on the X axis.
func XLabelCellOpts(co ...cell.Option) Option {
	return option(func(opts *options) {
		opts.xLabelCellOpts = co
	})
}

// YLabelCellOpts set the cell options for the labels on the Y axis.
func YLabelCellOpts(co ...cell.Option) Option {
	return option(func(opts *options) {
		opts.yLabelCellOpts = co
	})
}

// AxisCellOpts sets the cell options for the Y-axis rule drawn beside the cells.
func AxisCellOpts(co ...cell.Option) Option {
	return option(func(opts *options) {
		opts.axisCellOpts = co
	})
}

// Palette sets a custom low-to-high color palette for heatmap cells.
//
// When Palette isn't provided, the widget uses the original grayscale mapping.
// When colors are provided, lower values use earlier colors and higher values
// use later colors, preserving backward compatibility while allowing a more
// expressive look.
func Palette(colors ...cell.Color) Option {
	return option(func(opts *options) {
		opts.palette = append([]cell.Color(nil), colors...)
	})
}
