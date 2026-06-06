// Copyright 2026 Google Inc.
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

package spectrum

// options.go contains configurable options for Spectrum.

import (
	"fmt"

	"github.com/mum4k/termdash/cell"
)

// Orientation controls whether the spectrum draws as vertical columns or
// horizontal rows.
type Orientation int

const (
	// OrientationVertical draws mirrored columns around a horizontal axis.
	OrientationVertical Orientation = iota
	// OrientationHorizontal draws mirrored rows around a vertical axis.
	OrientationHorizontal
)

// Mode controls whether the widget draws two mirrored channels or a single
// half-duplex channel.
type Mode int

const (
	// ModeStereo draws a mirrored primary and secondary channel.
	ModeStereo Mode = iota
	// ModeHalfDuplex draws only the primary channel against a single axis.
	ModeHalfDuplex
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
	orientation   Orientation
	mode          Mode
	height        int
	maxValue      int
	primaryLabel  string
	secondLabel   string
	labelCellOpts []cell.Option
	axisCellOpts  []cell.Option
	gradient      []cell.Color
	threshold     int
	thresholdLine cell.Color
	alertColor    cell.Color
	primaryPeak   rune
	secondPeak    rune
	halfRune      rune
	primaryRunes  []rune
	secondRunes   []rune
	horizRunes    []rune
}

// clone returns an independent copy of the options.
func (o *options) clone() *options {
	cp := *o
	cp.labelCellOpts = append([]cell.Option(nil), o.labelCellOpts...)
	cp.axisCellOpts = append([]cell.Option(nil), o.axisCellOpts...)
	cp.gradient = append([]cell.Color(nil), o.gradient...)
	cp.primaryRunes = append([]rune(nil), o.primaryRunes...)
	cp.secondRunes = append([]rune(nil), o.secondRunes...)
	cp.horizRunes = append([]rune(nil), o.horizRunes...)
	return &cp
}

// validate validates the provided options.
func (o *options) validate() error {
	if got, min := o.height, 0; got < min {
		return fmt.Errorf("invalid Height %d, must be %d <= Height", got, min)
	}
	if got, min := o.maxValue, 0; got < min {
		return fmt.Errorf("invalid MaxValue %d, must be %d <= MaxValue", got, min)
	}
	if len(o.gradient) == 0 {
		return fmt.Errorf("Gradient must contain at least one color")
	}
	if got, min := o.threshold, 0; got < min {
		return fmt.Errorf("invalid Threshold %d, must be %d <= Threshold", got, min)
	}
	return nil
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		orientation:   OrientationVertical,
		mode:          ModeStereo,
		primaryLabel:  "LEFT",
		secondLabel:   "RIGHT",
		labelCellOpts: []cell.Option{cell.FgColor(cell.ColorWhite)},
		axisCellOpts:  []cell.Option{cell.FgColor(cell.ColorNumber(240))},
		gradient: []cell.Color{
			cell.ColorGreen,
			cell.ColorYellow,
			cell.ColorNumber(214),
			cell.ColorRed,
		},
		thresholdLine: cell.ColorRed,
		alertColor:    cell.ColorRed,
		primaryPeak:   '^',
		secondPeak:    'v',
		halfRune:      '█',
	}
}

// Height sets a fixed height for the Spectrum.
// If not set or set to zero, the widget uses all available vertical space.
func Height(h int) Option {
	return option(func(opts *options) {
		opts.height = h
	})
}

// MaxValue fixes the amplitude scale. Zero keeps the scale adaptive.
func MaxValue(v int) Option {
	return option(func(opts *options) {
		opts.maxValue = v
	})
}

// Vertical configures the widget to draw mirrored vertical columns.
func Vertical() Option {
	return option(func(opts *options) {
		opts.orientation = OrientationVertical
	})
}

// Horizontal configures the widget to draw mirrored horizontal rows.
func Horizontal() Option {
	return option(func(opts *options) {
		opts.orientation = OrientationHorizontal
	})
}

// Stereo configures the widget to draw both channels around the axis.
func Stereo() Option {
	return option(func(opts *options) {
		opts.mode = ModeStereo
	})
}

// HalfDuplex configures the widget to draw only the primary channel.
func HalfDuplex() Option {
	return option(func(opts *options) {
		opts.mode = ModeHalfDuplex
	})
}

// ChannelLabels sets the labels shown around the axis for the primary and
// secondary channels.
func ChannelLabels(primary, secondary string) Option {
	return option(func(opts *options) {
		opts.primaryLabel = primary
		opts.secondLabel = secondary
	})
}

// LabelCellOpts sets the cell options used for the axis labels.
func LabelCellOpts(cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.labelCellOpts = cOpts
	})
}

// AxisCellOpts sets the cell options used for the axis line.
func AxisCellOpts(cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.axisCellOpts = cOpts
	})
}

// Gradient sets the colors used for low-to-high amplitudes.
func Gradient(colors ...cell.Color) Option {
	return option(func(opts *options) {
		opts.gradient = append([]cell.Color(nil), colors...)
	})
}

// Threshold sets the alarm threshold value.
//
// Values at or above the threshold are highlighted with AlertColor, and a
// threshold indicator line is drawn with ThresholdLineColor. Zero disables the
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

// AlertColor sets the color used for rendered values that meet or exceed the
// configured threshold.
func AlertColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.alertColor = c
	})
}

// PeakRunes sets the peak runes for the primary and secondary channels.
func PeakRunes(primary, secondary rune) Option {
	return option(func(opts *options) {
		opts.primaryPeak = primary
		opts.secondPeak = secondary
	})
}

// PrimaryRunes sets the runes used to fill the primary channel body.
// The runes are ordered from low amplitude to high amplitude.
func PrimaryRunes(runes ...rune) Option {
	return option(func(opts *options) {
		opts.primaryRunes = append([]rune(nil), runes...)
	})
}

// SecondaryRunes sets the runes used to fill the secondary channel body.
// The runes are ordered from low amplitude to high amplitude.
func SecondaryRunes(runes ...rune) Option {
	return option(func(opts *options) {
		opts.secondRunes = append([]rune(nil), runes...)
	})
}

// HorizontalRunes sets the runes used to fill horizontally oriented channels.
// The runes are ordered from low amplitude to high amplitude.
func HorizontalRunes(runes ...rune) Option {
	return option(func(opts *options) {
		opts.horizRunes = append([]rune(nil), runes...)
	})
}

// HalfDuplexRune sets the rune used to draw the single-channel half-duplex
// columns or rows.
func HalfDuplexRune(r rune) Option {
	return option(func(opts *options) {
		opts.halfRune = r
	})
}
