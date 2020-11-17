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
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/heatmap"
)

func main() {
	xLabels := []string{
		"12:00",
		"12:05",
		"12:10",
		"12:15",
		"12:20",
	}
	yLabels := []string{
		"10",
		"20",
		"30",
		"40",
		"50",
		"60",
		"70",
		"80",
		"90",
		"100",
	}
	values := map[string][]int64{
		"12:00": {10, 20, 30, 40, 50, 50, 40, 30, 20, 10},
		"12:05": {50, 40, 30, 20, 10, 10, 20, 30, 40, 50},
		"12:10": {10, 20, 30, 40, 50, 50, 40, 30, 20, 10},
		"12:15": {50, 40, 30, 20, 10, 10, 20, 30, 40, 50},
		"12:20": {10, 20, 30, 40, 50, 50, 40, 30, 0, 0},
	}

	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	hp, err := heatmap.NewHeatMap()
	if err != nil {
		panic(err)
	}
	if err := hp.SetColumns(xLabels, values); err != nil {
		panic(err)
	}
	hp.SetYLabels(yLabels)

	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.PlaceWidget(hp),
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}
}
