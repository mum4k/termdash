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

// Binary buttondemo shows the functionality of a button widget.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
)

// buttonChunks creates the text chunks for a button from the provided text.
func buttonChunks(text string) []*button.TextChunk {
	if len(text) == 0 {
		return nil
	}
	first := string(text[0])
	rest := string(text[1:])

	return []*button.TextChunk{
		button.NewChunk(
			"<",
			button.TextCellOpts(cell.FgColor(cell.ColorWhite)),
			button.FocusedTextCellOpts(cell.FgColor(cell.ColorBlack)),
			button.PressedTextCellOpts(cell.FgColor(cell.ColorBlack)),
		),
		button.NewChunk(
			first,
			button.TextCellOpts(cell.FgColor(cell.ColorRed)),
		),
		button.NewChunk(
			rest,
			button.TextCellOpts(cell.FgColor(cell.ColorWhite)),
			button.FocusedTextCellOpts(cell.FgColor(cell.ColorBlack)),
			button.PressedTextCellOpts(cell.FgColor(cell.ColorBlack)),
		),
		button.NewChunk(
			">",
			button.TextCellOpts(cell.FgColor(cell.ColorWhite)),
			button.FocusedTextCellOpts(cell.FgColor(cell.ColorBlack)),
			button.PressedTextCellOpts(cell.FgColor(cell.ColorBlack)),
		),
	}
}

func main() {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())

	val := 0
	display, err := segmentdisplay.New()
	if err != nil {
		panic(err)
	}
	if err := display.Write([]*segmentdisplay.TextChunk{
		segmentdisplay.NewChunk(fmt.Sprintf("%d", val)),
	}); err != nil {
		panic(err)
	}

	addB, err := button.NewFromChunks(buttonChunks("Submit"), func() error {
		val++
		return display.Write([]*segmentdisplay.TextChunk{
			segmentdisplay.NewChunk(fmt.Sprintf("%d", val)),
		})
	},
		button.GlobalKeys('s', 'S'),
		button.DisableShadow(),
		button.Height(1),
		button.TextHorizontalPadding(0),
		button.FillColor(cell.ColorBlack),
		button.FocusedFillColor(cell.ColorNumber(117)),
		button.PressedFillColor(cell.ColorNumber(220)),
	)
	if err != nil {
		panic(err)
	}

	subB, err := button.NewFromChunks(buttonChunks("Cancel"), func() error {
		val--
		return display.Write([]*segmentdisplay.TextChunk{
			segmentdisplay.NewChunk(fmt.Sprintf("%d", val)),
		})
	},
		button.FillColor(cell.ColorNumber(220)),
		button.GlobalKeys('c', 'C'),
		button.DisableShadow(),
		button.Height(1),
		button.TextHorizontalPadding(0),
		button.FillColor(cell.ColorBlack),
		button.FocusedFillColor(cell.ColorNumber(117)),
		button.PressedFillColor(cell.ColorNumber(220)),
	)
	if err != nil {
		panic(err)
	}

	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.SplitHorizontal(
			container.Top(
				container.PlaceWidget(display),
			),
			container.Bottom(
				container.SplitVertical(
					container.Left(
						container.PlaceWidget(addB),
						container.AlignHorizontal(align.HorizontalRight),
						container.PaddingRight(5),
					),
					container.Right(
						container.PlaceWidget(subB),
						container.AlignHorizontal(align.HorizontalLeft),
						container.PaddingLeft(5),
					),
				),
			),
			container.SplitPercent(60),
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

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(100*time.Millisecond)); err != nil {
		panic(err)
	}
}
