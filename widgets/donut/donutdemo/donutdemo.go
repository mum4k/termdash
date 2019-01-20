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

// Binary gaugedemo displays a couple of Gauge widgets.
// Exist when 'q' is pressed.
package main

import (
	"context"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgets/donut"
)

// playType indicates how to play a donut.
type playType int

const (
	playTypePercent playType = iota
	playTypeAbsolute
)

// playDonut continuously changes the displayed percent value on the donut by the
// step once every delay. Exits when the context expires.
func playDonut(ctx context.Context, d *donut.Donut, step int, delay time.Duration, pt playType) {
	progress := 0
	mult := 1

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			switch pt {
			case playTypePercent:
				if err := d.Percent(progress); err != nil {
					panic(err)
				}
			case playTypeAbsolute:
				if err := d.Absolute(progress, 100); err != nil {
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
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	d, err := donut.New()
	if err != nil {
		panic(err)
	}
	go playDonut(ctx, d, 10, 500*time.Millisecond, playTypePercent)

	c, err := container.New(
		t,
		container.Border(draw.LineStyleLight),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.PlaceWidget(d),
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
