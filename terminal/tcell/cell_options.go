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
	"github.com/gdamore/tcell"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// cellColor converts termdash cell color to the tcell format.
func cellColor(c cell.Color) tcell.Color {
	return tcell.Color(c&0x1ff) - 1
}

// fixColor converts the target color for the current color mode
func fixColor(c tcell.Color, colorMode terminalapi.ColorMode) tcell.Color {
	if c == tcell.ColorDefault {
		return c
	}
	switch colorMode {
	case terminalapi.ColorModeNormal:
		c %= tcell.Color(16)
	case terminalapi.ColorMode256:
		c %= tcell.Color(256)
	case terminalapi.ColorMode216:
		c %= tcell.Color(216)
		c += tcell.Color(16)
	case terminalapi.ColorModeGrayscale:
		c %= tcell.Color(24)
		c += tcell.Color(232)
	default:
		c = tcell.ColorDefault
	}
	return c
}

// cellOptsToStyle converts termdash cell color to the tcell format.
func cellOptsToStyle(opts *cell.Options, colorMode terminalapi.ColorMode) tcell.Style {
	st := tcell.StyleDefault

	fg := cellColor(opts.FgColor)
	bg := cellColor(opts.BgColor)

	fg = fixColor(fg, colorMode)
	bg = fixColor(bg, colorMode)

	// FIXME: tcell doesn't have a strikethrough style option
	st = st.Foreground(fg).Background(bg).Bold(opts.Bold).Italic(opts.Italic).Underline(opts.Underline)
	return st
}
