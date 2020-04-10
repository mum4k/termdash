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

package button

// options.go contains configurable options for Button.

import (
	"fmt"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/widgetapi"
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
	fillColor   cell.Color
	textColor   cell.Color
	shadowColor cell.Color
	height      int
	width       int
	key         keyboard.Key
	keyScope    widgetapi.KeyScope
	keyUpDelay  time.Duration
}

// validate validates the provided options.
func (o *options) validate() error {
	if min := 1; o.height < min {
		return fmt.Errorf("invalid height %d, must be %d <= height", o.height, min)
	}
	if min := 1; o.width < min {
		return fmt.Errorf("invalid width %d, must be %d <= width", o.width, min)
	}
	if min := time.Duration(0); o.keyUpDelay < min {
		return fmt.Errorf("invalid keyUpDelay %v, must be %v <= keyUpDelay", o.keyUpDelay, min)
	}
	return nil
}

// newOptions returns options with the default values set.
func newOptions(text string) *options {
	return &options{
		fillColor:   cell.ColorNumber(117),
		textColor:   cell.ColorBlack,
		shadowColor: cell.ColorNumber(240),
		height:      DefaultHeight,
		width:       widthFor(text),
		keyUpDelay:  DefaultKeyUpDelay,
	}
}

// FillColor sets the fill color of the button.
func FillColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.fillColor = c
	})
}

// TextColor sets the color of the text label in the button.
func TextColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.textColor = c
	})
}

// ShadowColor sets the color of the shadow under the button.
func ShadowColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.shadowColor = c
	})
}

// DefaultHeight is the default for the Height option.
const DefaultHeight = 3

// Height sets the height of the button in cells.
// Must be a positive non-zero integer.
// Defaults to DefaultHeight.
func Height(cells int) Option {
	return option(func(opts *options) {
		opts.height = cells
	})
}

// Width sets the width of the button in cells.
// Must be a positive non-zero integer.
// Defaults to the auto-width based on the length of the text label.
func Width(cells int) Option {
	return option(func(opts *options) {
		opts.width = cells
	})
}

// WidthFor sets the width of the button as if it was displaying the provided text.
// Useful when displaying multiple buttons with the intention to set all of
// their sizes equal to the one with the longest text.
func WidthFor(text string) Option {
	return option(func(opts *options) {
		opts.width = widthFor(text)
	})
}

// Key configures the keyboard key that presses the button.
// The widget responds to this key only if its container if focused.
// When not provided, the widget ignores all keyboard events.
func Key(k keyboard.Key) Option {
	return option(func(opts *options) {
		opts.key = k
		opts.keyScope = widgetapi.KeyScopeFocused
	})
}

// GlobalKey is like Key, but makes the widget respond to the key even if its
// container isn't focused.
// When not provided, the widget ignores all keyboard events.
func GlobalKey(k keyboard.Key) Option {
	return option(func(opts *options) {
		opts.key = k
		opts.keyScope = widgetapi.KeyScopeGlobal
	})
}

// DefaultKeyUpDelay is the default value for the KeyUpDelay option.
const DefaultKeyUpDelay = 250 * time.Millisecond

// KeyUpDelay is the amount of time the button will remain "pressed down" after
// triggered by the configured key. Termbox doesn't emit events for key
// releases so the button simulates it by timing it.
// This only works if the manual termdash redraw or the periodic redraw
// interval are reasonably close to this delay.
// The duration cannot be negative.
// Defaults to DefaultKeyUpDelay.
func KeyUpDelay(d time.Duration) Option {
	return option(func(opts *options) {
		opts.keyUpDelay = d
	})
}

// widthFor returns the required width for the specified text.
func widthFor(text string) int {
	return runewidth.StringWidth(text) + 2 // One empty cell at each side.
}
