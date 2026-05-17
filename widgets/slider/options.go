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

package slider

// options.go contains configurable options for Slider.

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

// ChangeFn is called when the user changes the slider value.
//
// The callback must be thread-safe because it is triggered from the keyboard
// and mouse event handling paths, which run in separate goroutines.
type ChangeFn func(value int) error

// options holds the provided options.
type options struct {
	min            int
	max            int
	value          int
	step           int
	width          int
	fillRune       rune
	trackRune      rune
	knobRune       rune
	fillCellOpts   []cell.Option
	trackCellOpts  []cell.Option
	knobCellOpts   []cell.Option
	focusedKnobOps []cell.Option
	onChange       ChangeFn
}

// Default colors used by the slider widget.
var (
	DefaultFillColor        = cell.ColorNumber(221)
	DefaultTrackColor       = cell.ColorNumber(239)
	DefaultKnobColor        = cell.ColorWhite
	DefaultFocusedKnobColor = cell.ColorCyan
)

const (
	DefaultWidth = 18
	DefaultStep  = 1
)

// validate validates the provided options.
func (o *options) validate() error {
	if o.max < o.min {
		return fmt.Errorf("invalid range: max %d must be >= min %d", o.max, o.min)
	}
	if o.step <= 0 {
		return fmt.Errorf("invalid step %d, want step > 0", o.step)
	}
	if o.width <= 0 {
		return fmt.Errorf("invalid width %d, want width > 0", o.width)
	}
	return nil
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		min:            0,
		max:            100,
		value:          0,
		step:           DefaultStep,
		width:          DefaultWidth,
		fillRune:       '█',
		trackRune:      '░',
		knobRune:       '●',
		fillCellOpts:   []cell.Option{cell.FgColor(DefaultFillColor)},
		trackCellOpts:  []cell.Option{cell.FgColor(DefaultTrackColor)},
		knobCellOpts:   []cell.Option{cell.FgColor(DefaultKnobColor)},
		focusedKnobOps: []cell.Option{cell.FgColor(DefaultFocusedKnobColor)},
	}
}

// Min sets the minimum slider value.
func Min(v int) Option {
	return option(func(opts *options) {
		opts.min = v
	})
}

// Max sets the maximum slider value.
func Max(v int) Option {
	return option(func(opts *options) {
		opts.max = v
	})
}

// Value sets the initial slider value.
func Value(v int) Option {
	return option(func(opts *options) {
		opts.value = v
	})
}

// Step sets the increment used for keyboard changes.
func Step(v int) Option {
	return option(func(opts *options) {
		opts.step = v
	})
}

// Width sets the slider width in terminal cells.
func Width(cells int) Option {
	return option(func(opts *options) {
		opts.width = cells
	})
}

// FillRune sets the rune used for the filled portion of the slider.
func FillRune(r rune) Option {
	return option(func(opts *options) {
		opts.fillRune = r
	})
}

// TrackRune sets the rune used for the unfilled portion of the slider.
func TrackRune(r rune) Option {
	return option(func(opts *options) {
		opts.trackRune = r
	})
}

// KnobRune sets the rune used for the slider knob.
func KnobRune(r rune) Option {
	return option(func(opts *options) {
		opts.knobRune = r
	})
}

// FillCellOpts sets the styling used for the filled portion of the slider.
func FillCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.fillCellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// TrackCellOpts sets the styling used for the unfilled portion of the slider.
func TrackCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.trackCellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// KnobCellOpts sets the styling used for the slider knob.
func KnobCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.knobCellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// FocusedKnobCellOpts sets the styling used for the slider knob while the
// widget is focused.
func FocusedKnobCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.focusedKnobOps = append([]cell.Option(nil), cellOpts...)
	})
}

// OnChange sets the slider's value-change hook.
//
// This is the widget's canonical callback surface. Callers that need delayed
// or asynchronous work should build that from this hook so the widget keeps a
// single stable event path.
func OnChange(fn ChangeFn) Option {
	return option(func(opts *options) {
		opts.onChange = fn
	})
}
