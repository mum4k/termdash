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

// Binary gaugedemo displays a couple of Gauge widgets.
// Exist when 'q' is pressed.
package main

import (
	"context"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/gauge"
)

// playType indicates how to play a gauge.
type playType int

const (
	playTypePercent playType = iota
	playTypeAbsolute
)

// playGauge continuously changes the displayed percent value on the gauge by the
// step once every delay. Exits when the context expires.
func playGauge(ctx context.Context, g *gauge.Gauge, step int, delay time.Duration, pt playType) {
	progress := 0
	mult := 1

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			switch pt {
			case playTypePercent:
				if err := g.Percent(progress); err != nil {
					panic(err)
				}
			case playTypeAbsolute:
				if err := g.Absolute(progress, 100); err != nil {
					panic(err)
				}
			}

			progress += step * mult
			if progress > 100 || 100-progress < step {
				progress = 100
			} else if progress < 0 || progress < step {
				progress = 0
			}

			if progress == 100 {
				mult = -1
			} else if progress == 0 {
				mult = 1
			}

		case <-ctx.Done():
			return
		}
	}
}

func main() {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	slim, err := gauge.New(
		gauge.Height(1),
		gauge.Border(linestyle.Light),
		gauge.BorderTitle("Percentage progress"),
	)
	if err != nil {
		panic(err)
	}
	go playGauge(ctx, slim, 10, 500*time.Millisecond, playTypePercent)

	absolute, err := gauge.New(
		gauge.Height(1),
		gauge.Color(cell.ColorBlue),
		gauge.Border(linestyle.Light),
		gauge.BorderTitle("Absolute progress"),
	)
	if err != nil {
		panic(err)
	}
	go playGauge(ctx, absolute, 17, 500*time.Millisecond, playTypeAbsolute)

	noProgress, err := gauge.New(
		gauge.Height(1),
		gauge.Border(linestyle.Light, cell.FgColor(cell.ColorMagenta)),
		gauge.BorderTitle("Without progress text"),
		gauge.HideTextProgress(),
	)
	if err != nil {
		panic(err)
	}
	go playGauge(ctx, noProgress, 5, 250*time.Millisecond, playTypePercent)

	withLabel, err := gauge.New(
		gauge.Height(3),
		gauge.TextLabel("你好，世界! text label and no border"),
		gauge.Color(cell.ColorRed),
		gauge.FilledTextColor(cell.ColorBlack),
		gauge.EmptyTextColor(cell.ColorYellow),
	)
	if err != nil {
		panic(err)
	}
	go playGauge(ctx, withLabel, 3, 500*time.Millisecond, playTypePercent)

	c, err := container.New(
		t,
		container.SplitVertical(
			container.Left(
				container.Border(linestyle.Light),
				container.BorderTitle("PRESS Q TO QUIT"),
				container.SplitHorizontal(
					container.Top(
						container.PlaceWidget(slim),
					),
					container.Bottom(
						container.SplitVertical(
							container.Left(
								container.PlaceWidget(absolute),
							),
							container.Right(),
						),
					),
				),
			),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.PlaceWidget(noProgress),
					),
					container.Bottom(
						container.PlaceWidget(withLabel),
					),
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

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}
}
