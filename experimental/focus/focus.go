// Binary boxes just creates containers with borders.
// Runs as long as there is at least one input (keyboard, mouse or terminal resize) event every 10 seconds.
package main

import (
	"context"
	"time"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminalapi"
)

// mouseEvents forwards mouse events to the container and redraws it.
func mouseEvents(ctx context.Context, t terminalapi.Terminal, c *container.Container) <-chan *terminalapi.Mouse {
	ch := make(chan *terminalapi.Mouse)

	go func() {
		for {
			ev := t.Event(ctx)
			if m, ok := ev.(*terminalapi.Mouse); ok {
				ch <- m
			}
		}
	}()
	return ch
}

// redraw redraws the containers on the terminal.
func redraw(t terminalapi.Terminal, c *container.Container) error {
	//if err := t.Clear(); err != nil {
	//	return err
	//}
	if err := c.Draw(); err != nil {
		return err
	}

	if err := t.Flush(); err != nil {
		return err
	}
	return nil
}

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
								container.SplitVertical(
									container.Left(
										container.Border(draw.LineStyleLight),
									),
									container.Right(
										container.SplitVertical(
											container.Left(
												container.Border(draw.LineStyleLight),
											),
											container.Right(
												container.Border(draw.LineStyleLight),
											),
										),
									),
								),
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

	if err := redraw(t, c); err != nil {
		panic(err)
	}

	mouse := mouseEvents(context.Background(), t, c)
	redrawTimer := time.NewTicker(100 * time.Millisecond)
	defer redrawTimer.Stop()

	const exitTime = 10 * time.Second
	exitTimer := time.NewTicker(exitTime)

	for {
		defer exitTimer.Stop()
		select {
		case m := <-mouse:
			c.Mouse(m)
			exitTimer.Stop()
			exitTimer = time.NewTicker(exitTime)

		case <-redrawTimer.C:
			if err := redraw(t, c); err != nil {
				panic(err)
			}
		case <-exitTimer.C:
			return
		}
	}
}
