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

package table

import (
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
)

// options.go contains configurable options for Table.

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
	keyUp             keyboard.Key
	keyDown           keyboard.Key
	keyPgUp           keyboard.Key
	keyPgDown         keyboard.Key
	disableHighlight  bool
	disableSorting    bool
	highlightDelay    time.Duration
	highlightCellOpts []cell.Option
}

// validate validates the provided options.
func (o *options) validate() error {
	return nil
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		keyUp:     DefaultKeyUp,
		keyDown:   DefaultKeyDown,
		keyPgUp:   DefaultKeyPageUp,
		keyPgDown: DefaultKeyPageDown,
	}
}

// The default keys for navigating the content.
const (
	DefaultKeyUp       = keyboard.KeyArrowUp
	DefaultKeyDown     = keyboard.KeyArrowDown
	DefaultKeyPageUp   = keyboard.KeyPgUp
	DefaultKeyPageDown = keyboard.KeyPgDn
)

// NavigationKeys configures the keyboard keys that allow the user to navigate
// the table content.
// The provided keys must be unique, e.g. the same key cannot be both up and
// down.
func NavigationKeys(up, down, pageUp, pageDown keyboard.Key) Option {
	return option(func(opts *options) {
		opts.keyUp = up
		opts.keyDown = down
		opts.keyPgUp = pageUp
		opts.keyPgDown = pageDown
	})
}

// DisableHighlight disables the ability to highlight a row using keyboard or
// mouse.
// Navigation and scrolling still works, but it no longer moves the highlighted
// row, it just moves up or down a row at a time.
// The highlight functionality is enabled by default.
func DisableHighlight() Option {
	return option(func(opts *options) {
		opts.disableHighlight = true
	})
}

// DisableSorting disables the ability to sort the content by column values.
// The sorting functionality is enabled by default.
func DisableSorting() Option {
	return option(func(opts *options) {
		opts.disableSorting = true
	})
}

// HighlightDelay configures the time after which a highlighted row loses the
// highlight.
// This only works if the manual termdash redraw or the periodic redraw
// interval are reasonably close to this delay.
// The duration cannot be negative.
// Defaults to zero which means a highlighted row remains highlighted forever.
func HighlightDelay(d time.Duration) Option {
	return option(func(opts *options) {
		opts.highlightDelay = d
	})
}

// HighlightCellOpts sets the cell options on cells that are part of a
// highlighted row.
// Defaults to DefaultHighlightColor.
func HighlightColor(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.highlightCellOpts = cellOpts
	})
}
