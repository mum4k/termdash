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

// options.go contains configurable options for Indicator.

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
	status      bool
	color       []cell.Option
	maxDiameter int

	labelColor []cell.Option
	labelAlign align.Horizontal
	label      string
}

// validate validates the provided options.
func (o *options) validate() error {
	return nil
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		status: true,
		color: []cell.Option{
			cell.FgColor(cell.ColorRed),
			cell.BgColor(cell.ColorDefault),
		},
		labelAlign:  DefaultLabelAlign,
		maxDiameter: DefaultMaxSize,
	}
}

const DefaultMaxSize = 50

// MaxSize sets maximum size of the indicator widget
func MaxSize(size int) Option {
	return option(func(opts *options) {
		opts.maxDiameter = size
	})
}

// Color sets the color of indicator
func Color(cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.color = cOpts
	})
}

// Label sets a text label to be displayed under the indicator
func Label(text string) Option {
	return option(func(opts *options) {
		opts.label = text
	})
}

// LabelColor sets the color of the labels under the indicator
func LabelColor(cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.labelColor = cOpts
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
