package termbox

import (
	"fmt"

	"github.com/mum4k/termdash/terminalapi"
	termbox "github.com/nsf/termbox-go"
)

// colorMode converts termdash color modes to the termbox format.
func colorMode(cm terminalapi.ColorMode) (termbox.OutputMode, error) {
	switch cm {
	case terminalapi.ColorMode8:
		return termbox.OutputNormal, nil
	case terminalapi.ColorMode256:
		return termbox.Output256, nil
	case terminalapi.ColorMode216:
		return termbox.Output216, nil
	case terminalapi.ColorModeGrayscale:
		return termbox.OutputGrayscale, nil
	default:
		return -1, fmt.Errorf("don't know how to convert color mode %v to the termbox format", cm)
	}
}
