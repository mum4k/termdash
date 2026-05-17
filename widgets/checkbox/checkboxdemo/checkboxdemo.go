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

// Binary checkboxdemo shows the functionality of a checkbox widget.
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
	"github.com/mum4k/termdash/widgets/checkbox"
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
	writeStatus := func(cloak, shields bool) error {
		status.Reset()
		if err := status.Write("  Cloak: ", text.WriteReplace(), text.WriteCellOpts(cell.FgColor(cell.ColorCyan))); err != nil {
			return err
		}
		if err := status.Write(onOff(cloak)+"\n", text.WriteCellOpts(cell.FgColor(stateColor(cloak)))); err != nil {
			return err
		}
		if err := status.Write("  Shields: ", text.WriteCellOpts(cell.FgColor(cell.ColorYellow))); err != nil {
			return err
		}
		return status.Write(onOff(shields), text.WriteCellOpts(cell.FgColor(stateColor(shields))))
	}

	cloakOn := false
	shieldsOn := true
	if err := writeStatus(cloakOn, shieldsOn); err != nil {
		panic(err)
	}

	cloak, err := checkbox.New("Enable Cloak",
		checkbox.UseIndicatorSet(checkbox.IndicatorSets.Heavy),
		checkbox.OnChange(func(checked bool) error {
		cloakOn = checked
		return writeStatus(cloakOn, shieldsOn)
	}))
	if err != nil {
		panic(err)
	}

	shields, err := checkbox.New("Raise Shields",
		checkbox.Checked(true),
		checkbox.UseIndicatorSet(checkbox.IndicatorSets.Rounded),
		checkbox.OnChange(func(checked bool) error {
		shieldsOn = checked
		return writeStatus(cloakOn, shieldsOn)
	}))
	if err != nil {
		panic(err)
	}

	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(40,
			grid.Widget(cloak,
				container.Border(linestyle.Round),
				container.BorderTitle("Cloak"),
				container.PaddingLeft(1),
				container.PaddingTop(1),
				container.Focused(),
			),
		),
		grid.RowHeightPerc(30,
			grid.Widget(shields,
				container.Border(linestyle.Round),
				container.BorderTitle("Shields"),
				container.PaddingLeft(1),
				container.PaddingTop(1),
			),
		),
		grid.RowHeightPerc(30,
			grid.Widget(status,
				container.Border(linestyle.Round),
				container.BorderTitle("State"),
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

func onOff(v bool) string {
	if v {
		return "ONLINE"
	}
	return "OFFLINE"
}

func stateColor(v bool) cell.Color {
	if v {
		return cell.ColorGreen
	}
	return cell.ColorRed
}
