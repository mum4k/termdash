// Copyright 2018 Google Inc.
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

package container

import (
	"image"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/faketerm"
	"github.com/mum4k/termdash/terminalapi"
)

// pointCase is a test case for the pointCont function.
type pointCase struct {
	desc      string
	point     image.Point
	wantNil   bool
	wantColor cell.Color // expected container identified by its border color
}

func TestPointCont(t *testing.T) {
	tests := []struct {
		desc      string
		termSize  image.Point
		container func(ft *faketerm.Terminal) *Container
		cases     []pointCase
	}{
		{
			desc:     "single container, no border",
			termSize: image.Point{3, 3},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					BorderColor(cell.ColorBlue),
				)
			},
			cases: []pointCase{
				{
					desc:      "inside the container",
					point:     image.Point{1, 1},
					wantColor: cell.ColorBlue,
				},
				{
					desc:      "top left corner",
					point:     image.Point{0, 0},
					wantColor: cell.ColorBlue,
				},
				{
					desc:      "top right corner",
					point:     image.Point{2, 0},
					wantColor: cell.ColorBlue,
				},
				{
					desc:      "bottom left corner",
					point:     image.Point{0, 2},
					wantColor: cell.ColorBlue,
				},
				{
					desc:      "bottom right corner",
					point:     image.Point{2, 2},
					wantColor: cell.ColorBlue,
				},
				{
					desc:    "outside of the container, too large",
					point:   image.Point{3, 3},
					wantNil: true,
				},
				{
					desc:    "outside of the container, too small",
					point:   image.Point{-1, -1},
					wantNil: true,
				},
			},
		},
		{
			desc:     "single container, border",
			termSize: image.Point{3, 3},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					BorderColor(cell.ColorBlue),
				)
			},
			cases: []pointCase{
				{
					desc:      "inside the container",
					point:     image.Point{1, 1},
					wantColor: cell.ColorBlue,
				},
				{
					desc:      "on the border",
					point:     image.Point{0, 1},
					wantColor: cell.ColorBlue,
				},
			},
		},
		{
			desc:     "split containers, parent has no border",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					BorderColor(cell.ColorBlack),
					SplitVertical(
						Left(
							SplitHorizontal(
								Top(
									BorderColor(cell.ColorGreen),
								),
								Bottom(
									BorderColor(cell.ColorWhite),
								),
							),
						),
						Right(
							BorderColor(cell.ColorRed),
						),
					),
				)
			},
			cases: []pointCase{
				{
					desc:      "right sub container, inside corner",
					point:     image.Point{5, 5},
					wantColor: cell.ColorRed,
				},
				{
					desc:      "right sub container, outside corner",
					point:     image.Point{9, 9},
					wantColor: cell.ColorRed,
				},
				{
					desc:      "top left",
					point:     image.Point{0, 0},
					wantColor: cell.ColorGreen,
				},
				{
					desc:      "bottom left",
					point:     image.Point{0, 9},
					wantColor: cell.ColorWhite,
				},
			},
		},
		{
			desc:     "split containers, parent has border",
			termSize: image.Point{10, 10},
			container: func(ft *faketerm.Terminal) *Container {
				return New(
					ft,
					Border(draw.LineStyleLight),
					BorderColor(cell.ColorBlack),
					SplitVertical(
						Left(
							SplitHorizontal(
								Top(
									BorderColor(cell.ColorGreen),
								),
								Bottom(
									BorderColor(cell.ColorWhite),
								),
							),
						),
						Right(
							BorderColor(cell.ColorRed),
						),
					),
				)
			},
			cases: []pointCase{
				{
					desc:      "right sub container, inside corner",
					point:     image.Point{5, 5},
					wantColor: cell.ColorRed,
				},
				{
					desc:      "top right corner focuses parent",
					point:     image.Point{9, 9},
					wantColor: cell.ColorBlack,
				},
				{
					desc:      "right sub container, outside corner",
					point:     image.Point{8, 8},
					wantColor: cell.ColorRed,
				},
				{
					desc:      "top left focuses parent",
					point:     image.Point{0, 0},
					wantColor: cell.ColorBlack,
				},
				{
					desc:      "top left sub container",
					point:     image.Point{1, 1},
					wantColor: cell.ColorGreen,
				},
				{
					desc:      "bottom left focuses parent",
					point:     image.Point{0, 9},
					wantColor: cell.ColorBlack,
				},
				{
					desc:      "bottom left sub container",
					point:     image.Point{1, 8},
					wantColor: cell.ColorWhite,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ft, err := faketerm.New(tc.termSize)
			if err != nil {
				t.Fatalf("faketerm.New => unexpected error: %v", err)
			}

			cont := tc.container(ft)
			for _, pc := range tc.cases {
				gotCont := pointCont(cont, pc.point)
				if (gotCont == nil) != pc.wantNil {
					t.Errorf("%s, pointCont%v => got %v, wantNil: %v", pc.desc, pc.point, gotCont, pc.wantNil)
				}
				if gotCont == nil {
					continue
				}

				gotColor := gotCont.opts.inherited.borderColor
				if gotColor != pc.wantColor {
					t.Errorf("%s, pointCont%v => got container with border color %v, want %v", pc.desc, pc.point, gotColor, pc.wantColor)
				}
			}
		})
	}
}

// contLoc is used in tests to indicate the desired location of a container.
type contLoc int

// String implements fmt.Stringer()
func (cl contLoc) String() string {
	if n, ok := contLocNames[cl]; ok {
		return n
	}
	return "contLocUnknown"
}

// contLocNames maps contLoc values to human readable names.
var contLocNames = map[contLoc]string{
	contLocRoot:  "Root",
	contLocLeft:  "Left",
	contLocRight: "Right",
}

const (
	contLocUnknown contLoc = iota
	contLocRoot
	contLocLeft
	contLocRight
)

func TestFocusTrackerMouse(t *testing.T) {
	ft, err := faketerm.New(image.Point{10, 10})
	if err != nil {
		t.Fatalf("faketerm.New => unexpected error: %v", err)
	}

	var (
		insideLeft  = image.Point{1, 1}
		insideRight = image.Point{6, 6}
	)

	tests := []struct {
		desc string
		// Can be either the mouse event or a time.Duration to pause for.
		events      []*terminalapi.Mouse
		wantFocused contLoc
	}{
		{
			desc:        "initially the root is focused",
			wantFocused: contLocRoot,
		},
		{
			desc: "click and release moves focus to the left",
			events: []*terminalapi.Mouse{
				&terminalapi.Mouse{image.Point{0, 0}, mouse.ButtonLeft},
				&terminalapi.Mouse{image.Point{1, 1}, mouse.ButtonRelease},
			},
			wantFocused: contLocLeft,
		},
		{
			desc: "click and release moves focus to the right",
			events: []*terminalapi.Mouse{
				&terminalapi.Mouse{image.Point{5, 5}, mouse.ButtonLeft},
				&terminalapi.Mouse{image.Point{6, 6}, mouse.ButtonRelease},
			},
			wantFocused: contLocRight,
		},
		{
			desc: "click in the same container is a no-op",
			events: []*terminalapi.Mouse{
				&terminalapi.Mouse{insideRight, mouse.ButtonLeft},
				&terminalapi.Mouse{insideRight, mouse.ButtonRelease},
				&terminalapi.Mouse{insideRight, mouse.ButtonLeft},
				&terminalapi.Mouse{insideRight, mouse.ButtonRelease},
			},
			wantFocused: contLocRight,
		},
		{
			desc: "click in the same container and release never happens",
			events: []*terminalapi.Mouse{
				&terminalapi.Mouse{insideRight, mouse.ButtonLeft},
				&terminalapi.Mouse{insideLeft, mouse.ButtonLeft},
				&terminalapi.Mouse{insideLeft, mouse.ButtonRelease},
			},
			wantFocused: contLocLeft,
		},
		{
			desc: "click in the same container, release elsewhere",
			events: []*terminalapi.Mouse{
				&terminalapi.Mouse{insideRight, mouse.ButtonLeft},
				&terminalapi.Mouse{insideLeft, mouse.ButtonRelease},
			},
			wantFocused: contLocRoot,
		},
		{
			desc: "other buttons are ignored",
			events: []*terminalapi.Mouse{
				&terminalapi.Mouse{insideLeft, mouse.ButtonMiddle},
				&terminalapi.Mouse{insideLeft, mouse.ButtonRelease},
				&terminalapi.Mouse{insideLeft, mouse.ButtonRight},
				&terminalapi.Mouse{insideLeft, mouse.ButtonRelease},
				&terminalapi.Mouse{insideLeft, mouse.ButtonWheelUp},
				&terminalapi.Mouse{insideLeft, mouse.ButtonWheelDown},
			},
			wantFocused: contLocRoot,
		},
		{
			desc: "moving mouse with pressed button and then releasing moves focus",
			events: []*terminalapi.Mouse{
				&terminalapi.Mouse{image.Point{0, 0}, mouse.ButtonLeft},
				&terminalapi.Mouse{image.Point{1, 1}, mouse.ButtonLeft},
				&terminalapi.Mouse{image.Point{2, 2}, mouse.ButtonRelease},
			},
			wantFocused: contLocLeft,
		},
		{
			desc: "click ignored if followed by another click of the same button elsewhere",
			events: []*terminalapi.Mouse{
				&terminalapi.Mouse{insideRight, mouse.ButtonLeft},
				&terminalapi.Mouse{insideLeft, mouse.ButtonLeft},
				&terminalapi.Mouse{insideRight, mouse.ButtonRelease},
				&terminalapi.Mouse{insideRight, mouse.ButtonRelease},
			},
			wantFocused: contLocRoot,
		},
		{
			desc: "click ignored if followed by another click of a different button",
			events: []*terminalapi.Mouse{
				&terminalapi.Mouse{insideRight, mouse.ButtonLeft},
				&terminalapi.Mouse{insideRight, mouse.ButtonMiddle},
				&terminalapi.Mouse{insideRight, mouse.ButtonRelease},
				&terminalapi.Mouse{insideRight, mouse.ButtonRelease},
			},
			wantFocused: contLocRoot,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			root := New(
				ft,
				SplitVertical(
					Left(),
					Right(),
				),
			)

			for _, ev := range tc.events {
				root.Mouse(ev)
			}

			var wantFocused *Container
			switch wf := tc.wantFocused; wf {
			case contLocRoot:
				wantFocused = root
			case contLocLeft:
				wantFocused = root.first
			case contLocRight:
				wantFocused = root.second
			default:
				t.Fatalf("unsupported wantFocused value => %v", wf)
			}

			if !root.focusTracker.isActive(wantFocused) {
				t.Errorf("isActive(%v) => false, want true, status: root(%v):%v, left(%v):%v, right(%v):%v",
					tc.wantFocused,
					contLocRoot, root.focusTracker.isActive(root),
					contLocLeft, root.focusTracker.isActive(root.first),
					contLocRight, root.focusTracker.isActive(root.second),
				)
			}
		})
	}
}
