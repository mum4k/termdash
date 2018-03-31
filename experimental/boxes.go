// Binary boxes just creates containers with borders.
package main

import (
	"time"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminal/termbox"
)

func main() {
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	c := container.New(
		t,
		container.SplitVertical(),
	).First(
		container.SplitHorizontal(),
	).First(
		container.Border(draw.LineStyleLight),
	).Parent().Second(
		container.SplitHorizontal(),
	).First(
		container.Border(draw.LineStyleLight),
	).Parent().Second(
		container.Border(draw.LineStyleLight),
	).Root().Second(
		container.Border(draw.LineStyleLight),
	).Root()

	if err := c.Draw(); err != nil {
		panic(err)
	}

	if err := t.Flush(); err != nil {
		panic(err)
	}
	time.Sleep(30 * time.Second)
}
