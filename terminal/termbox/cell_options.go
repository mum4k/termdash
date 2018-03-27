package termbox

// cell_options.go converts termdash cell options to the termbox format.

import (
	"fmt"

	"github.com/mum4k/termdash/cell"
	termbox "github.com/nsf/termbox-go"
)

// cellColor converts termdash cell color to the termbox format.
func cellColor(c cell.Color) (termbox.Attribute, error) {
	switch c {
	case cell.ColorDefault:
		return termbox.ColorDefault, nil
	case cell.ColorBlack:
		return termbox.ColorBlack, nil
	case cell.ColorRed:
		return termbox.ColorRed, nil
	case cell.ColorGreen:
		return termbox.ColorGreen, nil
	case cell.ColorYellow:
		return termbox.ColorYellow, nil
	case cell.ColorBlue:
		return termbox.ColorBlue, nil
	case cell.ColorMagenta:
		return termbox.ColorMagenta, nil
	case cell.ColorCyan:
		return termbox.ColorCyan, nil
	case cell.ColorWhite:
		return termbox.ColorWhite, nil
	default:
		return 0, fmt.Errorf("don't know how to convert cell color %v to the termbox format", c)
	}
}

// cellOptsToFg converts the cell options to the termbox foreground attribute.
func cellOptsToFg(opts *cell.Options) (termbox.Attribute, error) {
	return cellColor(opts.FgColor)
}

// cellOptsToBg converts the cell options to the termbox background attribute.
func cellOptsToBg(opts *cell.Options) (termbox.Attribute, error) {
	return cellColor(opts.BgColor)
}
