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

// Binary sparklinedemo displays a couple of SparkLine widgets.
// Exist when 'q' is pressed.
package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/internal/cell"
	"github.com/mum4k/termdash/internal/draw"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgets/sparkline"
)

// playSparkLine continuously adds values to the SparkLine, once every delay.
// Exits when the context expires.
func playSparkLine(ctx context.Context, sl *sparkline.SparkLine, delay time.Duration) {
	const max = 100

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			v := int(rand.Int31n(max + 1))
			if err := sl.Add([]int{v}); err != nil {
				panic(err)
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
	green, err := sparkline.New(
		sparkline.Label("Green SparkLine", cell.FgColor(cell.ColorBlue)),
		sparkline.Color(cell.ColorGreen),
	)
	if err != nil {
		panic(err)
	}
	go playSparkLine(ctx, green, 250*time.Millisecond)
	red, err := sparkline.New(
		sparkline.Label("Red SparkLine", cell.FgColor(cell.ColorBlue)),
		sparkline.Color(cell.ColorRed),
	)
	if err != nil {
		panic(err)
	}
	go playSparkLine(ctx, red, 500*time.Millisecond)
	yellow, err := sparkline.New(
		sparkline.Label("Yellow SparkLine", cell.FgColor(cell.ColorGreen)),
		sparkline.Color(cell.ColorYellow),
	)
	if err != nil {
		panic(err)
	}
	go playSparkLine(ctx, yellow, 1*time.Second)

	c, err := container.New(
		t,
		container.Border(draw.LineStyleLight),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(),
					container.Bottom(
						container.Border(draw.LineStyleLight),
						container.BorderTitle("SparkLine group"),
						container.SplitHorizontal(
							container.Top(
								container.PlaceWidget(green),
							),
							container.Bottom(
								container.PlaceWidget(red),
							),
						),
					),
				),
			),
			container.Right(
				container.Border(draw.LineStyleLight),
				container.PlaceWidget(yellow),
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
