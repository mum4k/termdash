package termbox

// cell_options.go converts termdash cell options to the termbox format.

import (
	"fmt"

	"github.com/mum4k/termdash/cell"
	tbx "github.com/nsf/termbox-go"
)

// cellColor converts termdash cell color to the termbox format.
func cellColor(c cell.Color) (tbx.Attribute, error) {
	switch c {
	case cell.ColorDefault:
		return tbx.ColorDefault, nil
	case cell.ColorBlack:
		return tbx.ColorBlack, nil
	case cell.ColorRed:
		return tbx.ColorRed, nil
	case cell.ColorGreen:
		return tbx.ColorGreen, nil
	case cell.ColorYellow:
		return tbx.ColorYellow, nil
	case cell.ColorBlue:
		return tbx.ColorBlue, nil
	case cell.ColorMagenta:
		return tbx.ColorMagenta, nil
	case cell.ColorCyan:
		return tbx.ColorCyan, nil
	case cell.ColorWhite:
		return tbx.ColorWhite, nil
	default:
		return 0, fmt.Errorf("don't know how to convert cell color %v to the termbox format", c)
	}
}

// cellOptsToFg converts the cell options to the termbox foreground attribute.
func cellOptsToFg(opts *cell.Options) (tbx.Attribute, error) {
	return cellColor(opts.FgColor)
}

// cellOptsToBg converts the cell options to the termbox background attribute.
func cellOptsToBg(opts *cell.Options) (tbx.Attribute, error) {
	return cellColor(opts.BgColor)
}
