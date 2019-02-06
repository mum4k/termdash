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

// Binary segmentdisplaydemo shows the functionality of a segment display.
package main

import (
	"context"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
)

// clock displays the current time on the segment display.
// Exists when the context expires.
func clock(ctx context.Context, sd *segmentdisplay.SegmentDisplay) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := segmentdisplay.NewChunk(time.Now().Format("150405"))
			if err := sd.Write([]*segmentdisplay.TextChunk{now}); err != nil {
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
	sd, err := segmentdisplay.New()
	if err != nil {
		panic(err)
	}
	go clock(ctx, sd)

	c, err := container.New(
		t,
		container.Border(draw.LineStyleLight),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.PlaceWidget(sd),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(1*time.Second)); err != nil {
		panic(err)
	}
}
