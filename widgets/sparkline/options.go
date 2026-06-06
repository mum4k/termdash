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

import (
	"fmt"

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
	label         string
	labelCellOpts []cell.Option
	height        int
	color         cell.Color
	threshold     int
	thresholdLine cell.Color
	alertColor    cell.Color
	sparkRunes    []rune
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		color:         DefaultColor,
		thresholdLine: cell.ColorRed,
		alertColor:    cell.ColorRed,
		sparkRunes:    append([]rune(nil), sparks...),
	}
}

// validate validates the provided options.
func (o *options) validate() error {
	if got, min := o.height, 0; got < min {
		return fmt.Errorf("invalid Height %d, must be %d <= Height", got, min)
	}
	if err := validateSparkRunes(o.sparkRunes); err != nil {
		return err
	}
	return nil
}

// Label adds a label above the SparkLine.
func Label(text string, cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.label = text
		opts.labelCellOpts = cOpts
	})
}

// Height sets a fixed height for the SparkLine.
// If not provided or set to zero, the SparkLine takes all the available
// vertical space in the container. Must be a positive or zero integer.
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

// Threshold sets the alarm threshold value for the SparkLine. Values at or
// above the threshold can be highlighted during drawing. Zero disables the
// threshold overlay.
func Threshold(v int) Option {
	return option(func(opts *options) {
		opts.threshold = v
	})
}

// ThresholdLineColor sets the color used to draw the threshold indicator line.
func ThresholdLineColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.thresholdLine = c
	})
}

// AlertColor sets the color used to draw portions of bars that exceed the
// configured threshold.
func AlertColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.alertColor = c
	})
}

// SparkRunes sets the runes used to render bar heights from low to high.
// All runes must occupy exactly one cell.
//
// Calling SparkRunes with no arguments restores the default spark rune set.
func SparkRunes(runes ...rune) Option {
	return option(func(opts *options) {
		if len(runes) == 0 {
			opts.sparkRunes = append([]rune(nil), sparks...)
			return
		}
		opts.sparkRunes = append([]rune(nil), runes...)
	})
}
