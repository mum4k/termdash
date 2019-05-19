// Copyright 2019 Google Inc.
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

// Binary indicatordemo displays a few of Indicator widgets.
// Exist when 'q' is pressed.
package main

import (
	"context"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/indicator"
)

// playType indicates how to play a indicator.
type playType int

const (
	flash playType = iota
	pulse
)

// playIndicator continuously changes the displayed percent value on the Indicator by the
// step once every delay. Exits when the context expires.
func playIndicator(ctx context.Context, i *indicator.Indicator, delay time.Duration, pt playType) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			switch pt {
			case flash:
				if err := i.Flash(); err != nil {
					panic(err)
				}
			case pulse:
				if err := i.Pulse(); err != nil {
					panic(err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	green, err := indicator.New(indicator.TextCellOpts(cell.FgColor(cell.ColorGreen)),
		indicator.Label("label", cell.FgColor(cell.ColorBlue)))
	if err != nil {
		panic(err)
	}
	go playIndicator(ctx, green, 200*time.Millisecond, pulse)

	blue, err := indicator.New(indicator.TextCellOpts(cell.FgColor(cell.ColorBlue)),
		indicator.Label("long text label", cell.FgColor(cell.ColorGreen)))
	if err != nil {
		panic(err)
	}
	go playIndicator(ctx, blue, 50*time.Millisecond, flash)

	yellow, err := indicator.New(indicator.TextCellOpts(cell.FgColor(cell.ColorYellow)))
	if err != nil {
		panic(err)
	}
	go playIndicator(ctx, yellow, 500*time.Millisecond, pulse)

	red, err := indicator.New(indicator.TextCellOpts(cell.FgColor(cell.ColorRed)))
	if err != nil {
		panic(err)
	}
	go playIndicator(ctx, red, 250*time.Millisecond, flash)

	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.SplitVertical(
			container.Left(
				container.SplitVertical(
					container.Left(container.PlaceWidget(green)),
					container.Right(container.PlaceWidget(blue)),
				),
			),
			container.Right(
				container.SplitVertical(
					container.Left(container.PlaceWidget(yellow)),
					container.Right(container.PlaceWidget(red)),
				),
			),
		),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(10*time.Millisecond)); err != nil {
		panic(err)
	}
}
