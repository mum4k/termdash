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
	"image"
	"testing"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/private/faketerm"
	"github.com/mum4k/termdash/widgets/radio"
	"github.com/mum4k/termdash/widgets/text"
)

func TestRadioDemoLayoutBuilds(t *testing.T) {
	status, err := text.New()
	if err != nil {
		t.Fatalf("text.New => unexpected error: %v", err)
	}
	r, err := radio.New([]radio.Item{
		{Label: "ON"},
		{Label: "OFF"},
	})
	if err != nil {
		t.Fatalf("radio.New => unexpected error: %v", err)
	}

	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(45,
			grid.Widget(r,
				container.Border(linestyle.Round),
				container.BorderTitle("Warp Core"),
				container.PaddingLeft(1),
				container.PaddingTop(1),
				container.Focused(),
			),
		),
		grid.RowHeightPerc(55,
			grid.Widget(status,
				container.Border(linestyle.Round),
				container.BorderTitle("Status"),
				container.PaddingLeft(1),
				container.PaddingTop(1),
			),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		t.Fatalf("grid.Build => unexpected error: %v", err)
	}
	ft := faketerm.MustNew(image.Point{X: 80, Y: 20})
	if _, err := container.New(ft, append(gridOpts,
		container.KeyFocusNext(keyboard.KeyTab),
		container.KeyFocusPrevious(keyboard.KeyBacktab),
	)...); err != nil {
		t.Fatalf("container.New => unexpected error: %v", err)
	}
}
