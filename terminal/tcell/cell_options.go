package tcell

import (
	"github.com/gdamore/tcell"
	"github.com/mum4k/termdash/cell"
)

// cellColor converts termdash cell color to the tcell format.
func cellColor(c cell.Color) tcell.Color {
	return tcell.Color(int(c)&0x1ff) - 1
}

// cellOptsToStyle converts termdash cell color to the tcell format.
func cellOptsToStyle(opts *cell.Options) tcell.Style {
	fg := cellColor(opts.FgColor)
	bg := cellColor(opts.BgColor)

	st := tcell.StyleDefault
	st = st.Foreground(fg).Background(bg)
	return st
}
