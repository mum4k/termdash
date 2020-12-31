// Copyright 2021 Google Inc.
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
	"fmt"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/draw"
)

// options.go contains configurable options for HeatMap.

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*options)
}

// options stores the provided options.
type options struct {
	cellChar       rune
	cellWidth      int
	hideXLabels    bool
	hideYLabels    bool
	xLabelCellOpts []cell.Option
	yLabelCellOpts []cell.Option
}

// validate validates the provided options.
func (o *options) validate() error {
	if got, min := o.cellWidth, 0; got < min {
		return fmt.Errorf("invalid CellWidth %d, must be %d <= CellWidth", got, min)
	}
	return nil
}

// newOptions returns a new options instance.
func newOptions(opts ...Option) *options {
	opt := &options{
		cellChar: DefaultChar,
	}
	for _, o := range opts {
		o.set(opt)
	}
	return opt
}

// DefaultChar is the default value for the Char option.
const DefaultChar = draw.DefaultRectChar

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// CellWidth set the width of cells (or grids) in the heat map, not the terminal cell.
func CellWidth(w int) Option {
	return option(func(opts *options) {
		opts.cellWidth = w
	})
}

// ShowXLabels configures the HeatMap so that it displays labels
// on the X axis. This is the default behavior.
func ShowXLabels() Option {
	return option(func(opts *options) {
		opts.hideXLabels = false
	})
}

// ShowYLabels configures the HeatMap so that it displays labels
// on the Y axis. This is the default behavior.
func ShowYLabels() Option {
	return option(func(opts *options) {
		opts.hideYLabels = false
	})
}

// HideXLabels disables the display of labels on the X axis.
func HideXLabels() Option {
	return option(func(opts *options) {
		opts.hideXLabels = true
	})
}

// HideYLabels disables the display of labels on the Y axis.
func HideYLabels() Option {
	return option(func(opts *options) {
		opts.hideYLabels = true
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
