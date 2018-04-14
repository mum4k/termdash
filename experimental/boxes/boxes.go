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
	"log"
	"time"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminalapi"
)

func events(t terminalapi.Terminal, ctx context.Context) <-chan terminalapi.Event {
	ch := make(chan terminalapi.Event)
	go func() {
		for {
			ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
			ev := t.Event(ctx)
			switch ev.(type) {
			case *terminalapi.Error:
				log.Print(ev)
			default:
				ch <- ev
			}

			cancel()
		}
	}()
	return ch
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

	ev := events(t, context.Background())
	for {
		timer := time.NewTicker(10 * time.Second)
		defer timer.Stop()
		select {
		case e := <-ev:
			log.Printf("Event: %v", e)
		case <-timer.C:
			log.Printf("Exiting...")
			return
		}
	}
}
