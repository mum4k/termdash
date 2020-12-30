// Copyright 2020 Google Inc.
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

// text_options.go contains options used for the text displayed by the button.

import "github.com/mum4k/termdash/cell"

// TextOption is used to provide options to NewChunk().
type TextOption interface {
	// set sets the provided option.
	set(*textOptions)
}

// textOptions stores the provided options.
type textOptions struct {
	cellOpts        []cell.Option
	focusedCellOpts []cell.Option
	pressedCellOpts []cell.Option
}

// setDefaultFgColor configures a default color for text if one isn't specified
// in the text options.
func (to *textOptions) setDefaultFgColor(c cell.Color) {
	to.cellOpts = append(
		[]cell.Option{cell.FgColor(c)},
		to.cellOpts...,
	)
}

// newTextOptions returns new textOptions instance.
func newTextOptions(tOpts ...TextOption) *textOptions {
	to := &textOptions{}
	for _, o := range tOpts {
		o.set(to)
	}
	return to
}

// textOption implements TextOption.
type textOption func(*textOptions)

// set implements TextOption.set.
func (to textOption) set(tOpts *textOptions) {
	to(tOpts)
}

// TextCellOpts sets options on the cells that contain the button text.
// If not specified, all cells will just have their foreground color set to the
// value of TextColor().
func TextCellOpts(opts ...cell.Option) TextOption {
	return textOption(func(tOpts *textOptions) {
		tOpts.cellOpts = opts
	})
}

// FocusedTextCellOpts sets options on the cells that contain the button text
// when the widget's container is focused.
// If not specified, TextCellOpts will be used instead.
func FocusedTextCellOpts(opts ...cell.Option) TextOption {
	return textOption(func(tOpts *textOptions) {
		tOpts.focusedCellOpts = opts
	})
}

// PressedTextCellOpts sets options on the cells that contain the button text
// when it is pressed.
// If not specified, TextCellOpts will be used instead.
func PressedTextCellOpts(opts ...cell.Option) TextOption {
	return textOption(func(tOpts *textOptions) {
		tOpts.pressedCellOpts = opts
	})
}
