// Copyright 2018 Google Inc.
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

package linechart

import "github.com/mum4k/termdash/cell"

// options.go contains configurable options for LineChart.

// Option is used to provide options to New().
type Option interface {
	// set sets the provided option.
	set(*options)
}

// options stores the provided options.
type options struct {
	axesCellOpts   []cell.Option
	xLabelCellOpts []cell.Option
	yLabelCellOpts []cell.Option
}

// newOptions returns a new options instance.
func newOptions(opts ...Option) *options {
	opt := &options{}
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

// AxesCellOpts set the cell options for the X and Y axes.
func AxesCellOpts(co ...cell.Option) Option {
	return option(func(opts *options) {
		opts.axesCellOpts = co
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
