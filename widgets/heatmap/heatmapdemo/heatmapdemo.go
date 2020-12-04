// Copyright 2020 Google Inc.
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

// Binary heatmapdemo displays a heatmap widget.
// Exist when 'q' is pressed.
package main

import (
	"context"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/heatmap"
	"math/rand"
	"time"
)

// playHeatMap continuously changes the displayed values on the heat map once every delay.
// Exits when the context expires.
func playHeatMap(ctx context.Context, hp *heatmap.HeatMap, delay time.Duration) {
	const max = 100

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			maxRows, maxCols := hp.ValueCapacity()
			rows := rand.Intn(maxRows) + 1
			cols := rand.Intn(maxCols) + 1

			var values [][]float64
			for i := 0; i < rows; i++ {
				rv := make([]float64, cols)
				for j := 0; j < cols; j++ {
					rv = append(rv, float64(rand.Int31n(max+1)))
				}
				values = append(values, rv)
			}

			if err := hp.Values(nil, nil, values); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func getData() ([]string, []string, [][]float64) {
	var xl = []string{
		"one",
		"two",
		"three",
		"four",
		"five",
		"six",
	}
	var yl = []string{
		"one",
		"two",
		"three",
		"four",
		"five",
	}
	var values = [][]float64{
		{1, 2, 3, 4, 5, 6},
		{-1, -2, -3, -4, -5, -6},
		{7, 8, 9, 10, 11, 12},
		{12, 11, 10, 9, 8, 7},
		{6, 5, 4, 3, 2, 1},
	}
	return xl, yl, values
}

func main() {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	hp, err := heatmap.New()
	if err != nil {
		panic(err)
	}

	if err := hp.Values(getData()); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go playHeatMap(ctx, hp, 1*time.Second)

	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.PlaceWidget(hp),
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
