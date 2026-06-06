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

// Binary dropdowndemo shows the functionality of a dropdown widget.
package main

import (
	"context"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/dropdown"
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

	writeStatus := func(alarm, torpedoes string) error {
		status.Reset()
		if err := status.Write("  Alarm Threshold: ", text.WriteReplace(), text.WriteCellOpts(cell.FgColor(cell.ColorYellow))); err != nil {
			return err
		}
		if err := status.Write(alarm+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorWhite))); err != nil {
			return err
		}
		if err := status.Write("  Torpedoes To Load: ", text.WriteCellOpts(cell.FgColor(cell.ColorCyan))); err != nil {
			return err
		}
		return status.Write(torpedoes, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	}

	alarmChoices := dropdown.IntRange(200, 600, 50, "%03d")
	torpedoChoices := dropdown.IntRange(1, 12, 1, "%02d")

	alarmText := alarmChoices[0]
	torpedoText := torpedoChoices[0]
	if err := writeStatus(alarmText, torpedoText); err != nil {
		panic(err)
	}

	alarmDD, err := dropdown.New(alarmChoices,
		dropdown.Selected(6),
		dropdown.GlyphSet(dropdown.GlyphProfiles.Minimal),
		dropdown.OnSelect(func(_ int, label string) error {
			alarmText = label
			return writeStatus(alarmText, torpedoText)
		}),
	)
	if err != nil {
		panic(err)
	}

	torpedoDD, err := dropdown.New(torpedoChoices,
		dropdown.GlyphSet(dropdown.GlyphProfiles.Minimal),
		dropdown.OnSelect(func(_ int, label string) error {
			torpedoText = label
			return writeStatus(alarmText, torpedoText)
		}),
	)
	if err != nil {
		panic(err)
	}

	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(35,
			grid.ColWidthPerc(50,
				grid.Widget(alarmDD,
					container.Border(linestyle.Round),
					container.BorderTitle("Alarm"),
					container.PaddingLeft(1),
					container.PaddingTop(1),
					container.Focused(),
				),
			),
			grid.ColWidthPerc(50,
				grid.Widget(torpedoDD,
					container.Border(linestyle.Round),
					container.BorderTitle("Torpedoes"),
					container.PaddingLeft(1),
					container.PaddingTop(1),
				),
			),
		),
		grid.RowHeightPerc(65,
			grid.Widget(status,
				container.Border(linestyle.Round),
				container.BorderTitle("Selections"),
				container.PaddingLeft(1),
				container.PaddingTop(1),
			),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		panic(err)
	}

	c, err := container.New(
		t,
		append(gridOpts,
			container.KeyFocusNext(keyboard.KeyTab),
			container.KeyFocusPrevious(keyboard.KeyBacktab),
		)...,
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
