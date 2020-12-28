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
	"os/user"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/textinput"
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

	var username string
	u, err := user.Current()
	if err != nil {
		username = "mum4k"
	} else {
		username = u.Username
	}

	userInput, err := textinput.New(
		textinput.Label("Username: ", cell.FgColor(cell.ColorNumber(33))),
		textinput.DefaultText(username),
		textinput.MaxWidthCells(20),
		textinput.ExclusiveKeyboardOnFocus(),
	)
	uidInput, err := textinput.New(
		textinput.Label("UID:      ", cell.FgColor(cell.ColorNumber(33))),
		textinput.DefaultText("1000"),
		textinput.MaxWidthCells(20),
		textinput.ExclusiveKeyboardOnFocus(),
	)
	gidInput, err := textinput.New(
		textinput.Label("GID:      ", cell.FgColor(cell.ColorNumber(33))),
		textinput.DefaultText("1000"),
		textinput.MaxWidthCells(20),
		textinput.ExclusiveKeyboardOnFocus(),
	)
	homeInput, err := textinput.New(
		textinput.Label("Home:     ", cell.FgColor(cell.ColorNumber(33))),
		textinput.DefaultText(fmt.Sprintf("/home/%s", username)),
		textinput.MaxWidthCells(20),
		textinput.ExclusiveKeyboardOnFocus(),
	)

	submitB, err := button.NewFromChunks(buttonChunks("Submit"), func() error {
		return nil
	},
		button.Key(keyboard.KeyEnter),
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

	cancelB, err := button.NewFromChunks(buttonChunks("Cancel"), func() error {
		cancel()
		return nil
	},
		button.FillColor(cell.ColorNumber(220)),
		button.Key(keyboard.KeyEnter),
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
		container.KeyFocusNext(keyboard.KeyTab),
		container.KeyFocusGroupsNext(keyboard.KeyArrowDown, 1),
		container.KeyFocusGroupsPrevious(keyboard.KeyArrowUp, 1),
		container.KeyFocusGroupsNext(keyboard.KeyArrowRight, 2),
		container.KeyFocusGroupsPrevious(keyboard.KeyArrowLeft, 2),
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.SplitHorizontal(
					container.Top(
						container.SplitHorizontal(
							container.Top(
								container.Focused(),
								container.KeyFocusGroups(1),
								container.PlaceWidget(userInput),
							),
							container.Bottom(
								container.KeyFocusGroups(1),
								container.KeyFocusSkip(),
								container.PlaceWidget(uidInput),
							),
						),
					),
					container.Bottom(
						container.SplitHorizontal(
							container.Top(
								container.KeyFocusGroups(1),
								container.KeyFocusSkip(),
								container.PlaceWidget(gidInput),
							),
							container.Bottom(
								container.KeyFocusGroups(1),
								container.KeyFocusSkip(),
								container.PlaceWidget(homeInput),
							),
						),
					),
				),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.SplitVertical(
							container.Left(
								container.KeyFocusGroups(1, 2),
								container.PlaceWidget(submitB),
								container.AlignHorizontal(align.HorizontalRight),
								container.PaddingRight(5),
							),
							container.Right(
								container.KeyFocusGroups(1, 2),
								container.PlaceWidget(cancelB),
								container.AlignHorizontal(align.HorizontalLeft),
								container.PaddingLeft(5),
							),
						),
					),
					container.Bottom(
						container.KeyFocusSkip(),
					),
					container.SplitFixed(3),
				),
			),
			container.SplitFixed(6),
		),
	)
	if err != nil {
		panic(err)
	}

	if err := termdash.Run(ctx, t, c, termdash.RedrawInterval(100*time.Millisecond)); err != nil {
		panic(err)
	}
}
