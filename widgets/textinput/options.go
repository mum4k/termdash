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
	"fmt"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/linestyle"
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

	filter   FilterFn
	onSubmit SubmitFn
}

// validate validates the provided options.
func (o *options) validate() error {
	if min, max, perc := 0, 100, o.widthPerc; perc != nil && (*perc <= min || *perc > max) {
		return fmt.Errorf("invalid WidthPerc(%d), must be value in range %d < value <= %d", *perc, min, max)
	}
	if min, cells := 4, o.maxWidthCells; cells != nil && *cells < min {
		return fmt.Errorf("invalid MaxWidthCells(%d), must be value in range %d <= value", *cells, min)
	}
	return nil
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		fillColor:   cell.ColorNumber(DefaultFillColorNumber),
		cursorColor: cell.ColorNumber(DefaultCursorColorNumber),
		labelAlign:  DefaultLabelAlign,
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

// HideTextWith sets the rune that should be displayed instead of displaying
// the text. Useful for fields that accept sensitive information like
// passwords.
func HideTextWith(r rune) Option {
	return option(func(opts *options) {
		opts.hideTextWith = r
	})
}

// Filter sets a function that will be used to filter characters the user can
// input.
func Filter(fn FilterFn) Option {
	return option(func(opts *options) {
		opts.filter = fn
	})
}

// OnSubmit sets a function that will be called with the text typed by the user
// when they submit the content by pressing the Enter key.
func OnSubmit(fn SubmitFn) Option {
	return option(func(opts *options) {
		opts.onSubmit = fn
	})
}
