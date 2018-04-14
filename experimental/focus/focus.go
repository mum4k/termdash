// Copyright 2018 Google Inc.
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

// Binary boxes just creates containers with borders.
// Runs as long as there is at least one input (keyboard, mouse or terminal resize) event every 10 seconds.
package main

import (
	"context"
	"image"
	"time"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/fakewidget"
)

// inputEvents sends mouse and keyboard events on the channel.
func inputEvents(ctx context.Context, t terminalapi.Terminal, c *container.Container) <-chan terminalapi.Event {
	ch := make(chan terminalapi.Event)

	go func() {
		for {
			ev := t.Event(ctx)
			switch ev.(type) {
			case *terminalapi.Keyboard, *terminalapi.Mouse:
				ch <- ev
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

	wOpts := widgetapi.Options{
		WantKeyboard: true,
		WantMouse:    true,
	}
	c := container.New(
		t,
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(draw.LineStyleLight),
						container.PlaceWidget(fakewidget.New(widgetapi.Options{
							WantKeyboard: true,
							WantMouse:    true,
							Ratio:        image.Point{5, 1},
						})),
					),
					container.Bottom(
						container.SplitHorizontal(
							container.Top(
								container.Border(draw.LineStyleLight),
								container.PlaceWidget(fakewidget.New(wOpts)),
							),
							container.Bottom(
								container.SplitVertical(
									container.Left(
										container.Border(draw.LineStyleLight),
										container.PlaceWidget(fakewidget.New(wOpts)),
									),
									container.Right(
										container.SplitVertical(
											container.Left(
												container.Border(draw.LineStyleLight),
												container.VerticalAlignMiddle(),
												container.PlaceWidget(fakewidget.New(widgetapi.Options{
													WantKeyboard: true,
													WantMouse:    true,
													Ratio:        image.Point{2, 1},
												})),
											),
											container.Right(
												container.Border(draw.LineStyleLight),
												container.PlaceWidget(fakewidget.New(wOpts)),
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
				container.PlaceWidget(fakewidget.New(wOpts)),
			),
		),
	)

	if err := redraw(t, c); err != nil {
		panic(err)
	}

	events := inputEvents(context.Background(), t, c)
	redrawTimer := time.NewTicker(100 * time.Millisecond)
	defer redrawTimer.Stop()

	const exitTime = 10 * time.Second
	exitTimer := time.NewTicker(exitTime)

	for {
		defer exitTimer.Stop()
		select {
		case ev := <-events:
			switch e := ev.(type) {
			case *terminalapi.Mouse:
				c.Mouse(e)
			case *terminalapi.Keyboard:
				c.Keyboard(e)
			}
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
