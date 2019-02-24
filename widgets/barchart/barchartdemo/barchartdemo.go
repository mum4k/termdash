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

// Binary barchartdemo displays a couple of BarChart widgets.
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
	"github.com/mum4k/termdash/internal/terminal/termbox"
	"github.com/mum4k/termdash/internal/terminalapi"
	"github.com/mum4k/termdash/widgets/barchart"
)

// playBarChart continuously changes the displayed values on the bar chart once every delay.
// Exits when the context expires.
func playBarChart(ctx context.Context, bc *barchart.BarChart, delay time.Duration) {
	const (
		bars = 6
		max  = 100
	)

	values := make([]int, 6)

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			for i := range values {
				values[i] = int(rand.Int31n(max + 1))
			}

			if err := bc.Values(values, max); err != nil {
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
	bc, err := barchart.New(
		barchart.BarColors([]cell.Color{
			cell.ColorBlue,
			cell.ColorRed,
			cell.ColorYellow,
			cell.ColorBlue,
			cell.ColorGreen,
			cell.ColorRed,
		}),
		barchart.ValueColors([]cell.Color{
			cell.ColorRed,
			cell.ColorYellow,
			cell.ColorBlue,
			cell.ColorGreen,
			cell.ColorRed,
			cell.ColorBlue,
		}),
		barchart.ShowValues(),
		barchart.Labels([]string{
			"CPU1",
			"",
			"CPU3",
		}),
	)
	if err != nil {
		panic(err)
	}
	go playBarChart(ctx, bc, 1*time.Second)

	c, err := container.New(
		t,
		container.Border(draw.LineStyleLight),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.PlaceWidget(bc),
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
