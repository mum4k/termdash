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

// Binary radiodemo shows the functionality of a radio widget.
package main

import (
	"context"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/radio"
	"github.com/mum4k/termdash/widgets/text"
)

func main() {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	status, err := text.New()
	if err != nil {
		panic(err)
	}
	writeStatus := func(label string) error {
		status.Reset()
		if err := status.Write("  Warp Core: ", text.WriteReplace(), text.WriteCellOpts(cell.FgColor(cell.ColorCyan))); err != nil {
			return err
		}
		return status.Write(label, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	}
	if err := writeStatus("ON"); err != nil {
		panic(err)
	}

	r, err := radio.New([]radio.Item{
		{
			Label:            "ON",
			CellOpts:         []cell.Option{cell.FgColor(cell.ColorGreen)},
			SelectedCellOpts: []cell.Option{cell.FgColor(cell.ColorGreen)},
		},
		{
			Label:            "OFF",
			CellOpts:         []cell.Option{cell.FgColor(cell.ColorRed)},
			SelectedCellOpts: []cell.Option{cell.FgColor(cell.ColorRed)},
		},
	}, radio.UseIndicatorSet(radio.IndicatorSets.Circle), radio.OnChange(func(index int, label string) error {
		_ = index
		return writeStatus(label)
	}))
	if err != nil {
		panic(err)
	}

	c, err := container.New(
		t,
		container.Border(linestyle.Round),
		container.BorderTitle("Radio Demo"),
		container.PaddingLeft(2),
		container.PaddingTop(1),
		container.SplitHorizontal(
			container.Top(
				container.PlaceWidget(r),
				container.Focused(),
			),
			container.Bottom(
				container.PlaceWidget(status),
			),
			container.SplitPercent(45),
		),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' || k.Key == keyboard.KeyEsc {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(50*time.Millisecond)); err != nil {
		panic(err)
	}
}
