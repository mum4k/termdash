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

package sparkline

// options.go contains configurable options for SparkLine.

import "github.com/mum4k/termdash/cell"

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
	label         string
	labelCellOpts []cell.Option
	height        int
	color         cell.Color
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		color: DefaultColor,
	}
}

// Label adds a label above the SparkLine.
func Label(text string, cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.label = text
		opts.labelCellOpts = cOpts
	})
}

// Height sets a fixed height for the SparkLine.
// If not provided, the SparkLine takes all the available vertical space in the
// container.
func Height(h int) Option {
	return option(func(opts *options) {
		opts.height = h
	})
}

// DefaultColor is the default value for the Color option.
const DefaultColor = cell.ColorGreen

// Color sets the color of the SparkLine.
// Defaults to DefaultColor if not set.
func Color(c cell.Color) Option {
	return option(func(opts *options) {
		opts.color = c
	})
}
