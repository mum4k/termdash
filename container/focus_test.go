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
	"fmt"
	"image"
	"testing"
	"time"

	"github.com/mum4k/termdash/internal/cell"
	"github.com/mum4k/termdash/internal/draw"
	"github.com/mum4k/termdash/internal/event"
	"github.com/mum4k/termdash/internal/event/testevent"
	"github.com/mum4k/termdash/internal/mouse"
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
		container func(ft *faketerm.Terminal) (*Container, error)
		cases     []pointCase
	}{
		{
			desc:     "single container, no border",
			termSize: image.Point{3, 3},
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
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
			container: func(ft *faketerm.Terminal) (*Container, error) {
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

			cont, err := tc.container(ft)
			if err != nil {
				t.Fatalf("tc.container => unexpected error: %v", err)
			}
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
		events        []*terminalapi.Mouse
		wantFocused   contLoc
		wantProcessed int
	}{
		{
			desc:        "initially the root is focused",
			wantFocused: contLocRoot,
		},
		{
			desc: "click and release moves focus to the left",
			events: []*terminalapi.Mouse{
				{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				{Position: image.Point{1, 1}, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocLeft,
			wantProcessed: 2,
		},
		{
			desc: "click and release moves focus to the right",
			events: []*terminalapi.Mouse{
				{Position: image.Point{5, 5}, Button: mouse.ButtonLeft},
				{Position: image.Point{6, 6}, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocRight,
			wantProcessed: 2,
		},
		{
			desc: "click in the same container is a no-op",
			events: []*terminalapi.Mouse{
				{Position: insideRight, Button: mouse.ButtonLeft},
				{Position: insideRight, Button: mouse.ButtonRelease},
				{Position: insideRight, Button: mouse.ButtonLeft},
				{Position: insideRight, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocRight,
			wantProcessed: 4,
		},
		{
			desc: "click in the same container and release never happens",
			events: []*terminalapi.Mouse{
				{Position: insideRight, Button: mouse.ButtonLeft},
				{Position: insideLeft, Button: mouse.ButtonLeft},
				{Position: insideLeft, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocLeft,
			wantProcessed: 3,
		},
		{
			desc: "click in the same container, release elsewhere",
			events: []*terminalapi.Mouse{
				{Position: insideRight, Button: mouse.ButtonLeft},
				{Position: insideLeft, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocRoot,
			wantProcessed: 2,
		},
		{
			desc: "other buttons are ignored",
			events: []*terminalapi.Mouse{
				{Position: insideLeft, Button: mouse.ButtonMiddle},
				{Position: insideLeft, Button: mouse.ButtonRelease},
				{Position: insideLeft, Button: mouse.ButtonRight},
				{Position: insideLeft, Button: mouse.ButtonRelease},
				{Position: insideLeft, Button: mouse.ButtonWheelUp},
				{Position: insideLeft, Button: mouse.ButtonWheelDown},
			},
			wantFocused:   contLocRoot,
			wantProcessed: 6,
		},
		{
			desc: "moving mouse with pressed button and then releasing moves focus",
			events: []*terminalapi.Mouse{
				{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
				{Position: image.Point{1, 1}, Button: mouse.ButtonLeft},
				{Position: image.Point{2, 2}, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocLeft,
			wantProcessed: 3,
		},
		{
			desc: "click ignored if followed by another click of the same button elsewhere",
			events: []*terminalapi.Mouse{
				{Position: insideRight, Button: mouse.ButtonLeft},
				{Position: insideLeft, Button: mouse.ButtonLeft},
				{Position: insideRight, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocRoot,
			wantProcessed: 3,
		},
		{
			desc: "click ignored if followed by another click of a different button",
			events: []*terminalapi.Mouse{
				{Position: insideRight, Button: mouse.ButtonLeft},
				{Position: insideRight, Button: mouse.ButtonMiddle},
				{Position: insideRight, Button: mouse.ButtonRelease},
			},
			wantFocused:   contLocRoot,
			wantProcessed: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			root, err := New(
				ft,
				SplitVertical(
					Left(),
					Right(),
				),
			)
			if err != nil {
				t.Fatalf("New => unexpected error: %v", err)
			}

			eds := event.NewDistributionSystem()
			root.Subscribe(eds)
			for _, ev := range tc.events {
				eds.Event(ev)
			}
			if err := testevent.WaitFor(5*time.Second, func() error {
				if got, want := eds.Processed(), tc.wantProcessed; got != want {
					return fmt.Errorf("the event distribution system processed %d events, want %d", got, want)
				}
				return nil
			}); err != nil {
				t.Fatalf("testevent.WaitFor => %v", err)
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
