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

package indicator

// options.go contains configurable options for Speedometer.

import (
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
)

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*options)
}

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// options holds the provided options.
type options struct {
	status       bool
	textCellOpts []cell.Option
	cellOpts     []cell.Option

	labelCellOpts []cell.Option
	labelAlign    align.Horizontal
	label         string
}

// validate validates the provided options.
func (o *options) validate() error {
	return nil
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		status: true,
		textCellOpts: []cell.Option{
			cell.FgColor(cell.ColorRed),
			cell.BgColor(cell.ColorDefault),
		},
		labelAlign: DefaultLabelAlign,
	}
}

// TextCellOpts sets cell options on cells that contain the displayed text
// progress.
func TextCellOpts(cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.textCellOpts = cOpts
	})
}

// CellOpts sets cell options on cells that contain the donut.
func CellOpts(cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.cellOpts = cOpts
	})
}

// Label sets a text label to be displayed under the donut.
func Label(text string, cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.label = text
		opts.labelCellOpts = cOpts
	})
}

// DefaultLabelAlign is the default value for the LabelAlign option.
const DefaultLabelAlign = align.HorizontalCenter

// LabelAlign sets the alignment of the label under the indicator.
func LabelAlign(la align.Horizontal) Option {
	return option(func(opts *options) {
		opts.labelAlign = la
	})
}
