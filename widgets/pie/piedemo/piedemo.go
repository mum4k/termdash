// Copyright 2026 Google Inc.
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

package main

import (
	"context"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/pie"
)

func playPie(ctx context.Context, p *pie.Pie, values [][]int, delay time.Duration) {
	idx := 0
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = p.Values(values[idx])
			idx = (idx + 1) % len(values)
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

	pie1, _ := pie.New()
	_ = pie1.Values([]int{30, 70})
	go playPie(ctx, pie1, [][]int{{30, 70}, {50, 50}, {80, 20}}, 1*time.Second)

	pie2, _ := pie.New()
	_ = pie2.Values([]int{10, 20, 30, 40})
	go playPie(ctx, pie2, [][]int{{10, 20, 30, 40}, {40, 30, 20, 10}, {25, 25, 25, 25}}, 2*time.Second)

	c, _ := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.SplitVertical(
			container.Left(container.PlaceWidget(pie1)),
			container.Right(container.PlaceWidget(pie2)),
		),
	)

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(1*time.Second)); err != nil {
		panic(err)
	}
}
