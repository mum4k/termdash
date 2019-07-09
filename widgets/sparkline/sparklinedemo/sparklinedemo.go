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

	"termdash"
	"termdash/cell"
	"termdash/container"
	"termdash/linestyle"
	"termdash/terminal/termbox"
	"termdash/terminal/terminalapi"
	"termdash/widgets/sparkline"
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

// fillSparkLine continuously fills the SparkLine up to its capacity with
// random values.
func fillSparkLine(ctx context.Context, sl *sparkline.SparkLine, delay time.Duration) {
	const max = 100

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			var values []int
			for i := 0; i < sl.ValueCapacity(); i++ {
				values = append(values, int(rand.Int31n(max+1)))
			}
			if err := sl.Add(values); err != nil {
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
	go fillSparkLine(ctx, yellow, 1*time.Second)

	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(),
					container.Bottom(
						container.Border(linestyle.Light),
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
				container.Border(linestyle.Light),
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
