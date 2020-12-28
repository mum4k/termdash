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

package textinput

// options.go contains configurable options for TextInput.

import (
	"errors"
	"fmt"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/private/wrap"
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
	fillColor        cell.Color
	textColor        cell.Color
	placeHolderColor cell.Color
	highlightedColor cell.Color
	cursorColor      cell.Color
	border           linestyle.LineStyle
	borderColor      cell.Color

	widthPerc     *int
	maxWidthCells *int
	label         string
	labelCellOpts []cell.Option
	labelAlign    align.Horizontal

	placeHolder  string
	hideTextWith rune
	defaultText  string

	filter                   FilterFn
	onSubmit                 SubmitFn
	clearOnSubmit            bool
	exclusiveKeyboardOnFocus bool
}

// validate validates the provided options.
func (o *options) validate() error {
	if min, max, perc := 0, 100, o.widthPerc; perc != nil && (*perc <= min || *perc > max) {
		return fmt.Errorf("invalid WidthPerc(%d), must be value in range %d < value <= %d", *perc, min, max)
	}
	if min, cells := 4, o.maxWidthCells; cells != nil && *cells < min {
		return fmt.Errorf("invalid MaxWidthCells(%d), must be value in range %d <= value", *cells, min)
	}
	if r := o.hideTextWith; r != 0 {
		if err := wrap.ValidText(string(r)); err != nil {
			return fmt.Errorf("invalid HideTextWidth rune %c(%d): %v", r, r, err)
		}
		if got, want := runewidth.RuneWidth(r), 1; got != want {
			return fmt.Errorf("invalid HideTextWidth rune %c(%d), has rune width of %d cells, only runes with width of %d are accepted", r, r, got, want)
		}
	}
	if o.defaultText != "" {
		if err := wrap.ValidText(o.defaultText); err != nil {
			return fmt.Errorf("invalid DefaultText: %v", err)
		}
		for _, r := range o.defaultText {
			if r == '\n' {
				return errors.New("invalid DefaultText: newline characters aren't allowed")
			}
		}
	}
	return nil
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		fillColor:        cell.ColorNumber(DefaultFillColorNumber),
		placeHolderColor: cell.ColorNumber(DefaultPlaceHolderColorNumber),
		highlightedColor: cell.ColorNumber(DefaultHighlightedColorNumber),
		cursorColor:      cell.ColorNumber(DefaultCursorColorNumber),
		labelAlign:       DefaultLabelAlign,
	}
}

// DefaultFillColorNumber is the default color number for the FillColor option.
const DefaultFillColorNumber = 33

// FillColor sets the fill color for the text input field.
// Defaults to DefaultFillColorNumber.
func FillColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.fillColor = c
	})
}

// TextColor sets the color of the text in the input field.
// Defaults to the default terminal color.
func TextColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.textColor = c
	})
}

// DefaultHighlightedColorNumber is the default color number for the
// HighlightedColor option.
const DefaultHighlightedColorNumber = 0

// HighlightedColor sets the color of the text rune directly under the cursor.
// Defaults to the default terminal color.
func HighlightedColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.highlightedColor = c
	})
}

// DefaultCursorColorNumber is the default color number for the CursorColor
// option.
const DefaultCursorColorNumber = 250

// CursorColor sets the color of the cursor.
// Defaults to DefaultCursorColorNumber.
func CursorColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.cursorColor = c
	})
}

// Border adds a border around the text input field.
func Border(ls linestyle.LineStyle) Option {
	return option(func(opts *options) {
		opts.border = ls
	})
}

// BorderColor sets the color of the border.
// Defaults to the default terminal color.
func BorderColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.borderColor = c
	})
}

// WidthPerc sets the width for the text input field as a percentage of the
// container width. Must be a value in the range 0 < perc <= 100.
// Defaults to the width adjusted automatically base on the label length.
func WidthPerc(perc int) Option {
	return option(func(opts *options) {
		opts.widthPerc = &perc
	})
}

// MaxWidthCells sets the maximum width of the text input field as an absolute value
// in cells. Must be a value in the range 4 <= cells.
// This doesn't limit the text that the user can input, if the text overflows
// the width of the input field, it scrolls to the left.
// Defaults to using all available width in the container.
func MaxWidthCells(cells int) Option {
	return option(func(opts *options) {
		opts.maxWidthCells = &cells
	})
}

// Label adds a text label to the left of the input field.
func Label(label string, cOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.label = label
		opts.labelCellOpts = cOpts
	})
}

// DefaultLabelAlign is the default value for the LabelAlign option.
const DefaultLabelAlign = align.HorizontalLeft

// LabelAlign sets the alignment of the label within its area.
// The label is placed to the left of the input field. The width of this area
// can be specified using the LabelWidthPerc option.
// Defaults to DefaultLabelAlign.
func LabelAlign(la align.Horizontal) Option {
	return option(func(opts *options) {
		opts.labelAlign = la
	})
}

// PlaceHolder sets text to be displayed in the input field when it is empty.
// This text disappears when the text input field becomes focused.
func PlaceHolder(text string) Option {
	return option(func(opts *options) {
		opts.placeHolder = text
	})
}

// DefaultPlaceHolderColorNumber is the default color number for the
// PlaceHolderColor option.
const DefaultPlaceHolderColorNumber = 194

// PlaceHolderColor sets the color of the placeholder text.
// Defaults to DefaultPlaceHolderColorNumber.
func PlaceHolderColor(c cell.Color) Option {
	return option(func(opts *options) {
		opts.placeHolderColor = c
	})
}

// HideTextWith sets the rune that should be displayed instead of displaying
// the text. Useful for fields that accept sensitive information like
// passwords.
// The rune must be a printable rune with cell width of one.
func HideTextWith(r rune) Option {
	return option(func(opts *options) {
		opts.hideTextWith = r
	})
}

// FilterFn if provided can be used to filter runes that are allowed in the
// text input field. Any rune for which this function returns false will be
// rejected.
type FilterFn func(rune) bool

// Filter sets a function that will be used to filter characters the user can
// input.
func Filter(fn FilterFn) Option {
	return option(func(opts *options) {
		opts.filter = fn
	})
}

// SubmitFn if provided is called when the user submits the content of the text
// input field, the argument text contains all the text in the field.
// Submitting the input field clears its content.
//
// The callback function must be thread-safe as the keyboard event that
// triggers the submission comes from a separate goroutine.
type SubmitFn func(text string) error

// OnSubmit sets a function that will be called with the text typed by the user
// when they submit the content by pressing the Enter key.
// The SubmitFn must not attempt to read from or modify the TextInput instance
// in any way as while the SubmitFn is executing, the TextInput is mutex
// locked. If the intention is to clear the content on submission, use the
// ClearOnSubmit() option.
func OnSubmit(fn SubmitFn) Option {
	return option(func(opts *options) {
		opts.onSubmit = fn
	})
}

// ClearOnSubmit sets the text input to be cleared when a submit of the content
// is triggered by the user pressing the Enter key.
func ClearOnSubmit() Option {
	return option(func(opts *options) {
		opts.clearOnSubmit = true
	})
}

// ExclusiveKeyboardOnFocus when set ensures that when this widget is focused,
// no other widget receives any keyboard events.
func ExclusiveKeyboardOnFocus() Option {
	return option(func(opts *options) {
		opts.exclusiveKeyboardOnFocus = true
	})
}

// DefaultText sets the text to be present in a newly created input field.
// The text must not contain any control or space characters other than ' '.
// The user can edit this text as normal.
func DefaultText(text string) Option {
	return option(func(opts *options) {
		opts.defaultText = text
	})
}
