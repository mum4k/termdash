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
	if colorMode == DefaultColorMode {
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

	st = st.Foreground(fg).Background(bg)
	return st
}
