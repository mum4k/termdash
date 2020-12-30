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
}

// validate validates the provided options.
func (o *options) validate() error {
	return errors.New("not implemented")
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
