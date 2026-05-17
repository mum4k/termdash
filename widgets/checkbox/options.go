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

package checkbox

// options.go contains configurable options for Checkbox.

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

// ChangeFn is called when the user toggles the checkbox state.
//
// The callback must be thread-safe because it is triggered from the keyboard
// and mouse event handling paths, which run in separate goroutines.
type ChangeFn func(checked bool) error

// options holds the provided options.
type options struct {
	checked         bool
	indicator       IndicatorSet
	labelGap        int
	cellOpts        []cell.Option
	focusedCellOpts []cell.Option
	checkedCellOpts []cell.Option
	onChange        ChangeFn
}

// IndicatorSet defines the UTF-8 strings used to represent unchecked and
// checked states.
type IndicatorSet struct {
	Unchecked string
	Checked   string
}

// IndicatorSets groups reusable checkbox indicators.
var IndicatorSets = struct {
	Classic IndicatorSet
	Heavy   IndicatorSet
	Rounded IndicatorSet
}{
	Classic: IndicatorSet{Unchecked: "[ ]", Checked: "[x]"},
	Heavy:   IndicatorSet{Unchecked: "☐", Checked: "☑"},
	Rounded: IndicatorSet{Unchecked: "◯", Checked: "◉"},
}

// Default colors used by the checkbox widget.
var (
	DefaultTextColor        = cell.ColorWhite
	DefaultFocusedTextColor = cell.ColorCyan
	DefaultCheckedTextColor = cell.ColorGreen
)

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		indicator:       IndicatorSets.Classic,
		labelGap:        1,
		cellOpts:        []cell.Option{cell.FgColor(DefaultTextColor)},
		focusedCellOpts: []cell.Option{cell.FgColor(DefaultFocusedTextColor)},
		checkedCellOpts: []cell.Option{cell.FgColor(DefaultCheckedTextColor)},
	}
}

// Checked configures the initial checked state.
func Checked(checked bool) Option {
	return option(func(opts *options) {
		opts.checked = checked
	})
}

// Indicators sets the UTF-8 strings used for unchecked and checked states.
func Indicators(unchecked, checked string) Option {
	return option(func(opts *options) {
		if unchecked != "" {
			opts.indicator.Unchecked = unchecked
		}
		if checked != "" {
			opts.indicator.Checked = checked
		}
	})
}

// UseIndicatorSet sets both checkbox indicator strings from a reusable group.
func UseIndicatorSet(set IndicatorSet) Option {
	return option(func(opts *options) {
		if set.Unchecked != "" {
			opts.indicator.Unchecked = set.Unchecked
		}
		if set.Checked != "" {
			opts.indicator.Checked = set.Checked
		}
	})
}

// LabelGap sets the number of blank cells between the indicator and label.
func LabelGap(cells int) Option {
	return option(func(opts *options) {
		if cells < 0 {
			cells = 0
		}
		opts.labelGap = cells
	})
}

// CellOpts sets the default cell styling used while the checkbox is unchecked.
func CellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.cellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// FocusedCellOpts sets the styling used while the checkbox is focused and
// unchecked.
func FocusedCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.focusedCellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// CheckedCellOpts sets the styling used while the checkbox is checked.
func CheckedCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.checkedCellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// OnChange sets the checkbox's toggle hook.
//
// This is the widget's canonical callback surface. Callers that need delayed
// or asynchronous behavior should build that from this hook so the widget keeps
// a single stable event path.
func OnChange(fn ChangeFn) Option {
	return option(func(opts *options) {
		opts.onChange = fn
	})
}
