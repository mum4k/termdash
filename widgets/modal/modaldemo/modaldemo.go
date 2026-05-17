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

// Binary modaldemo shows draggable widgets hosted inside a modal.
package main

import (
	"context"
	"image"
	"log"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/modal"
	"github.com/mum4k/termdash/widgets/text"
)

// main boots the modal demo.
func main() {
	term, err := tcell.New()
	if err != nil {
		log.Fatalf("failed to create terminal: %v", err)
	}
	defer term.Close()
	term.EnableMouseMotion()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instructions, err := text.New()
	if err != nil {
		log.Fatalf("failed to create instructions widget: %v", err)
	}
	if err := instructions.Write("Drag windows from their title bars. Click the title control to minimize into the bottom dock. Press Esc to hide the modal and q to quit."); err != nil {
		log.Fatalf("failed to write instructions: %v", err)
	}

	donutWidget, err := donut.New()
	if err != nil {
		log.Fatalf("failed to create donut widget: %v", err)
	}
	donutWidget.Percent(65)

	gaugeWidget, err := gauge.New()
	if err != nil {
		log.Fatalf("failed to create gauge widget: %v", err)
	}
	gaugeWidget.Percent(45)

	textWidget, err := text.New()
	if err != nil {
		log.Fatalf("failed to create text widget: %v", err)
	}
	if err := textWidget.Write("Telemetry uplink\nstable and draggable"); err != nil {
		log.Fatalf("failed to write text widget: %v", err)
	}

	opts := modal.NewOptions(
		modal.Border(true),
		modal.MinimumSize(image.Point{X: 20, Y: 10}),
	)

	item1 := modal.NewDraggableWidget("donut", donutWidget, 2, 2, 18, 10, opts)
	item2 := modal.NewDraggableWidget("gauge", gaugeWidget, 24, 3, 24, 6, opts)
	item3 := modal.NewDraggableWidget("text", textWidget, 14, 12, 24, 6, opts)
	item1.Title = "Donut Gauge"
	item2.Title = "Load Gauge"
	item3.Title = "Telemetry Notes"
	modalWidget := modal.NewModal("modal", []*modal.DraggableWidget{item1, item2, item3}, opts)

	manager := &modal.ModalManager{}

	root, err := container.New(
		term,
		container.ID("root"),
		container.Border(linestyle.Light),
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.PlaceWidget(instructions),
			),
			container.Bottom(
				container.ID("modal"),
			),
			container.SplitPercent(14),
		),
	)
	if err != nil {
		log.Fatalf("failed to create root container: %v", err)
	}

	if err := manager.ShowModal(modalWidget, root); err != nil {
		log.Fatalf("failed to show modal: %v", err)
	}

	eventHandler := modal.NewEventHandler(ctx, cancel, root, manager)

	if err := termdash.Run(
		ctx,
		term,
		root,
		termdash.KeyboardSubscriber(eventHandler.HandleKeyboard),
		termdash.RedrawInterval(50*time.Millisecond),
	); err != nil && err != context.Canceled {
		log.Fatalf("termdash run failed: %v", err)
	}
}
