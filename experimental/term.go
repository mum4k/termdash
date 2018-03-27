// Binary term just initializes the terminal and sets a few cells.
package main

import (
	"image"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminalapi"
)

func main() {
	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		panic(err)
	}
	defer t.Close()

	if err := t.SetCell(image.Point{0, 0}, 'X', cell.FgColor(cell.ColorMagenta)); err != nil {
		panic(err)
	}

	if err := t.SetCell(t.Size().Sub(image.Point{1, 1}), 'X', cell.FgColor(cell.ColorMagenta)); err != nil {
		panic(err)
	}
	if err := t.Flush(); err != nil {
		panic(err)
	}
	time.Sleep(3 * time.Second)
}
