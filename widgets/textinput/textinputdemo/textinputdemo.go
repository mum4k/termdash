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

// Binary textinputdemo shows the functionality of a text input field.
package main

import (
	"context"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
)

// rotate returns a new slice with inputs rotated by step.
// I.e. for a step of one:
//   inputs[0] -> inputs[len(inputs)-1]
//   inputs[1] -> inputs[0]
// And so on.
func rotate(inputs []rune, step int) []rune {
	return append(inputs[step:], inputs[:step]...)
}

// textState creates a rotated state for the text we are displaying.
func textState(text string, capacity, step int) []rune {
	if capacity == 0 {
		return nil
	}

	var state []rune
	for i := 0; i < capacity; i++ {
		state = append(state, ' ')
	}
	state = append(state, []rune(text)...)
	step = step % len(state)
	return rotate(state, step)
}

// rollText rolls a text across the segment display.
// Exists when the context expires.
func rollText(ctx context.Context, sd *segmentdisplay.SegmentDisplay, updateText <-chan string) {
	colors := []cell.Color{
		cell.ColorNumber(33),
		cell.ColorRed,
		cell.ColorYellow,
		cell.ColorNumber(33),
		cell.ColorGreen,
		cell.ColorRed,
		cell.ColorGreen,
		cell.ColorRed,
	}

	text := "Termdash"
	step := 0
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			state := textState(text, sd.Capacity(), step)
			var chunks []*segmentdisplay.TextChunk
			for i := 0; i < sd.Capacity(); i++ {
				if i >= len(state) {
					break
				}

				color := colors[i%len(colors)]
				chunks = append(chunks, segmentdisplay.NewChunk(
					string(state[i]),
					segmentdisplay.WriteCellOpts(cell.FgColor(color)),
				))
			}
			if len(chunks) == 0 {
				continue
			}
			if err := sd.Write(chunks); err != nil {
				panic(err)
			}
			step++

		case t := <-updateText:
			text = t
			sd.Reset()
			step = 0

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
	rollingSD, err := segmentdisplay.New(
		segmentdisplay.MaximizeSegmentHeight(),
	)
	if err != nil {
		panic(err)
	}
	updateText := make(chan string)
	go rollText(ctx, rollingSD, updateText)

	mirror, err := text.New()
	if err != nil {
		panic(err)
	}

	input, err := textinput.New(
		textinput.Label("New text:", cell.FgColor(cell.ColorNumber(33))),
		textinput.MaxWidthCells(20),
		textinput.Border(linestyle.Light),
		textinput.PlaceHolder("Enter any text"),
		textinput.OnChange(func(data string) {
			mirror.Reset()
			mirror.Write(data)
		}),
	)
	if err != nil {
		panic(err)
	}

	submitB, err := button.New("Submit", func() error {
		updateText <- input.ReadAndClear()
		return nil
	},
		button.GlobalKey(keyboard.KeyEnter),
		button.FillColor(cell.ColorNumber(220)),
	)
	if err != nil {
		panic(err)
	}
	clearB, err := button.New("Clear", func() error {
		input.ReadAndClear()
		updateText <- ""
		return nil
	},
		button.WidthFor("Submit"),
		button.FillColor(cell.ColorNumber(220)),
	)
	if err != nil {
		panic(err)
	}
	quitB, err := button.New("Quit", func() error {
		cancel()
		return nil
	},
		button.WidthFor("Submit"),
		button.FillColor(cell.ColorNumber(196)),
	)
	if err != nil {
		panic(err)
	}

	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(40,
			grid.Widget(
				rollingSD,
			),
		),
	)
	builder.Add(
		grid.RowHeightPerc(40,
			grid.RowHeightPerc(50,
				grid.Widget(
					input,
					container.AlignHorizontal(align.HorizontalCenter),
					container.AlignVertical(align.VerticalBottom),
					container.MarginBottom(1),
				),
			),
			grid.RowHeightPerc(50,
				grid.Widget(
					mirror,
					container.Border(linestyle.Light),
					container.BorderTitle("text"),
					container.AlignHorizontal(align.HorizontalCenter),
					container.AlignHorizontal(align.Horizontal(align.VerticalBottom)),
					container.MarginBottom(1),
				),
			),
		),
	)

	builder.Add(
		grid.RowHeightPerc(20,
			grid.ColWidthPerc(20),
			grid.ColWidthPerc(20,
				grid.Widget(
					submitB,
					container.AlignVertical(align.VerticalTop),
					container.AlignHorizontal(align.HorizontalRight),
				),
			),
			grid.ColWidthPerc(20,
				grid.Widget(
					clearB,
					container.AlignVertical(align.VerticalTop),
					container.AlignHorizontal(align.HorizontalCenter),
				),
			),
			grid.ColWidthPerc(20,
				grid.Widget(
					quitB,
					container.AlignVertical(align.VerticalTop),
					container.AlignHorizontal(align.HorizontalLeft),
				),
			),
			grid.ColWidthPerc(20),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		panic(err)
	}
	c, err := container.New(t, gridOpts...)
	if err != nil {
		panic(err)
	}

	if err := termdash.Run(ctx, t, c, termdash.RedrawInterval(500*time.Millisecond)); err != nil {
		panic(err)
	}
}
