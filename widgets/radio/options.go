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

package radio

// options.go contains configurable options for Radio.

import (
	"fmt"

	"github.com/mum4k/termdash/cell"
)

// Item represents one selectable radio option.
type Item struct {
	Label            string
	SelectedText     string
	UnselectedText   string
	SelectedRune     rune
	UnselectedRune   rune
	CellOpts         []cell.Option
	SelectedCellOpts []cell.Option
}

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

// ChangeFn is called when the user selects a radio option.
//
// The callback must be thread-safe because it is triggered from the keyboard
// and mouse event handling paths, which run in separate goroutines.
type ChangeFn func(index int, label string) error

// options holds the provided options.
type options struct {
	selected     int
	gap          int
	indicatorGap int
	indicators   IndicatorSet
	onChange     ChangeFn
}

// IndicatorSet defines the UTF-8 strings used to render selected and
// unselected radio indicators.
type IndicatorSet struct {
	Selected   string
	Unselected string
}

// Default styling used by radio items when the caller doesn't provide item
// specific values.
var (
	DefaultCellOpts         = []cell.Option{cell.FgColor(cell.ColorWhite)}
	DefaultSelectedCellOpts = []cell.Option{cell.FgColor(cell.ColorCyan)}
	IndicatorSets           = struct {
		Circle  IndicatorSet
		Square  IndicatorSet
		Diamond IndicatorSet
		Target  IndicatorSet
	}{
		Circle:  IndicatorSet{Selected: "◉", Unselected: "○"},
		Square:  IndicatorSet{Selected: "■", Unselected: "□"},
		Diamond: IndicatorSet{Selected: "◆", Unselected: "◇"},
		Target:  IndicatorSet{Selected: "◎", Unselected: "·"},
	}
)

const (
	DefaultSelectedRune   = '◉'
	DefaultUnselectedRune = '○'
	DefaultGap            = 3
	DefaultIndicatorGap   = 1
)

// validate validates the provided options.
func (o *options) validate(items []Item) error {
	if len(items) == 0 {
		return fmt.Errorf("at least one item must be specified")
	}
	if o.selected < 0 || o.selected >= len(items) {
		return fmt.Errorf("invalid selected index %d, want 0 <= selected < %d", o.selected, len(items))
	}
	if o.gap < 0 {
		return fmt.Errorf("invalid gap %d, want gap >= 0", o.gap)
	}
	for i, item := range items {
		if item.Label == "" {
			return fmt.Errorf("item[%d] has an empty label", i)
		}
	}
	return nil
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		selected:     0,
		gap:          DefaultGap,
		indicatorGap: DefaultIndicatorGap,
		indicators:   IndicatorSets.Circle,
	}
}

// Selected sets the initially selected item index.
func Selected(index int) Option {
	return option(func(opts *options) {
		opts.selected = index
	})
}

// Gap sets the number of blank cells between radio items.
func Gap(cells int) Option {
	return option(func(opts *options) {
		opts.gap = cells
	})
}

// IndicatorGap sets the number of blank cells between the indicator and label.
func IndicatorGap(cells int) Option {
	return option(func(opts *options) {
		if cells < 0 {
			cells = 0
		}
		opts.indicatorGap = cells
	})
}

// Indicators sets the default UTF-8 strings used for radio indicators.
func Indicators(selected, unselected string) Option {
	return option(func(opts *options) {
		if selected != "" {
			opts.indicators.Selected = selected
		}
		if unselected != "" {
			opts.indicators.Unselected = unselected
		}
	})
}

// UseIndicatorSet sets the default indicators from a reusable group.
func UseIndicatorSet(set IndicatorSet) Option {
	return option(func(opts *options) {
		if set.Selected != "" {
			opts.indicators.Selected = set.Selected
		}
		if set.Unselected != "" {
			opts.indicators.Unselected = set.Unselected
		}
	})
}

// OnChange sets the radio widget's selection hook.
//
// This is the widget's canonical callback surface. Callers that need delayed
// or asynchronous work should build that from this hook so the widget keeps a
// single stable event path.
func OnChange(fn ChangeFn) Option {
	return option(func(opts *options) {
		opts.onChange = fn
	})
}
