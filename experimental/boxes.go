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
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(draw.LineStyleLight),
					),
					container.Bottom(
						container.SplitHorizontal(
							container.Top(
								container.Border(draw.LineStyleLight),
							),
							container.Bottom(
								container.Border(draw.LineStyleLight),
							),
						),
					),
				),
			),
			container.Right(
				container.Border(draw.LineStyleLight),
			),
		),
	)

	if err := c.Draw(); err != nil {
		panic(err)
	}

	if err := t.Flush(); err != nil {
		panic(err)
	}
	time.Sleep(3 * time.Second)
}
