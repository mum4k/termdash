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

package tcell

import (
	tcell "github.com/gdamore/tcell/v2"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// cellColor converts termdash cell color to the tcell format.
func cellColor(c cell.Color) tcell.Color {
	if c == cell.ColorDefault {
		return tcell.ColorDefault
	}
	// Subtract one, because cell.ColorBlack has value one instead of zero.
	// Zero is used for cell.ColorDefault instead.
	return tcell.Color(c-1) + tcell.ColorValid
}

// colorToMode adjusts the color to the color mode.
func colorToMode(c cell.Color, colorMode terminalapi.ColorMode) cell.Color {
	if c == cell.ColorDefault {
		return c
	}
	switch colorMode {
	case terminalapi.ColorModeNormal:
		c %= 16 + 1 // Add one for cell.ColorDefault.
	case terminalapi.ColorMode256:
		c %= 256 + 1 // Add one for cell.ColorDefault.
	case terminalapi.ColorMode216:
		if c <= 216 { // Add one for cell.ColorDefault.
			return c + 16
		}
		c = c%216 + 16
	case terminalapi.ColorModeGrayscale:
		if c <= 24 { // Add one for cell.ColorDefault.
			return c + 232
		}
		c = c%24 + 232
	default:
		c = cell.ColorDefault
	}
	return c
}

// cellOptsToStyle converts termdash cell color to the tcell format.
func cellOptsToStyle(opts *cell.Options, colorMode terminalapi.ColorMode) tcell.Style {
	st := tcell.StyleDefault

	fg := cellColor(colorToMode(opts.FgColor, colorMode))
	bg := cellColor(colorToMode(opts.BgColor, colorMode))

	st = st.Foreground(fg).
		Background(bg).
		Bold(opts.Bold).
		Italic(opts.Italic).
		Underline(opts.Underline).
		StrikeThrough(opts.Strikethrough).
		Reverse(opts.Inverse).
		Blink(opts.Blink).
		Dim(opts.Dim)
	return st
}
