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

package button

import (
	"fmt"
	"image"
	"testing"

	"github.com/mum4k/termdash/internal/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// eventTestCase is one mouse event and the output expectation.
type eventTestCase struct {
	// area if specified, will be provided to UpdateArea *before* processing the event.
	area *image.Rectangle

	// event is the mouse event to send.
	event *terminalapi.Mouse

	// wantClick indicates whether we expect the FSM to recognize a mouse click.
	wantClick bool

	// wantState is the expected button state.
	wantState State
}

func TestFSM(t *testing.T) {
	tests := []struct {
		desc       string
		button     mouse.Button
		area       image.Rectangle
		eventCases []*eventTestCase
	}{
		{
			desc:   "tracks single left button click",
			button: mouse.ButtonLeft,
			area:   image.Rect(0, 0, 1, 1),
			eventCases: []*eventTestCase{
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
					wantClick: true,
					wantState: Up,
				},
			},
		},
		{
			desc:   "updates area so the clicks falls outside",
			button: mouse.ButtonLeft,
			area:   image.Rect(0, 0, 1, 1),
			eventCases: []*eventTestCase{
				{
					area: func() *image.Rectangle {
						ar := image.Rect(1, 1, 2, 2)
						return &ar
					}(),
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Up,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
					wantClick: false,
					wantState: Up,
				},
			},
		},
		{
			desc:   "updates area before release, so the release falls outside",
			button: mouse.ButtonLeft,
			area:   image.Rect(0, 0, 1, 1),
			eventCases: []*eventTestCase{
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					area: func() *image.Rectangle {
						ar := image.Rect(1, 1, 2, 2)
						return &ar
					}(),
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
					wantClick: false,
					wantState: Up,
				},
			},
		},
		{
			desc:   "increased area makes the release count",
			button: mouse.ButtonLeft,
			area:   image.Rect(0, 0, 1, 1),
			eventCases: []*eventTestCase{
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					area: func() *image.Rectangle {
						ar := image.Rect(0, 0, 2, 2)
						return &ar
					}(),
					event:     &terminalapi.Mouse{Position: image.Point{1, 1}, Button: mouse.ButtonRelease},
					wantClick: true,
					wantState: Up,
				},
			},
		},
		{
			desc:   "tracks single right button click",
			button: mouse.ButtonRight,
			area:   image.Rect(0, 0, 1, 1),
			eventCases: []*eventTestCase{
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRight},
					wantClick: false,
					wantState: Down,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
					wantClick: true,
					wantState: Up,
				},
			},
		},
		{
			desc:   "ignores unrelated button in state wantPress",
			button: mouse.ButtonLeft,
			area:   image.Rect(0, 0, 1, 1),
			eventCases: []*eventTestCase{
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRight},
					wantClick: false,
					wantState: Up,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
					wantClick: false,
					wantState: Up,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
					wantClick: true,
					wantState: Up,
				},
			},
		},
		{
			desc:   "reverts to wantPress on unrelated button in state wantRelease",
			button: mouse.ButtonLeft,
			area:   image.Rect(0, 0, 1, 1),
			eventCases: []*eventTestCase{
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRight},
					wantClick: false,
					wantState: Up,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
					wantClick: false,
					wantState: Up,
				},
			},
		},
		{
			desc:   "reports button as down when the tracked button is pressed again in the area",
			button: mouse.ButtonLeft,
			area:   image.Rect(0, 0, 1, 1),
			eventCases: []*eventTestCase{
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
					wantClick: true,
					wantState: Up,
				},
			},
		},
		{
			desc:   "reports button as up when the tracked button is pressed again outside the area",
			button: mouse.ButtonLeft,
			area:   image.Rect(0, 0, 1, 1),
			eventCases: []*eventTestCase{
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{1, 1}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Up,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
					wantClick: true,
					wantState: Up,
				},
			},
		},
		{
			desc:   "ignores clicks outside of area in state wantPress",
			button: mouse.ButtonLeft,
			area:   image.Rect(0, 0, 1, 1),
			eventCases: []*eventTestCase{
				{
					event:     &terminalapi.Mouse{Position: image.Point{1, 1}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Up,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{1, 1}, Button: mouse.ButtonRelease},
					wantClick: false,
					wantState: Up,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
					wantClick: true,
					wantState: Up,
				},
			},
		},
		{
			desc:   "release outside of area releases button too",
			button: mouse.ButtonLeft,
			area:   image.Rect(0, 0, 1, 1),
			eventCases: []*eventTestCase{
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{1, 1}, Button: mouse.ButtonRelease},
					wantClick: false,
					wantState: Up,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonLeft},
					wantClick: false,
					wantState: Down,
				},
				{
					event:     &terminalapi.Mouse{Position: image.Point{0, 0}, Button: mouse.ButtonRelease},
					wantClick: true,
					wantState: Up,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf(tc.desc), func(t *testing.T) {
			fsm := NewFSM(tc.button, tc.area)
			for _, etc := range tc.eventCases {
				if etc.area != nil {
					fsm.UpdateArea(*etc.area)
				}

				gotClick, gotState := fsm.Event(etc.event)
				t.Logf("Called fsm.Event(%v) => %v, %v", etc.event, gotClick, gotState)
				if gotClick != etc.wantClick || gotState != etc.wantState {
					t.Errorf("fsm.Event(%v) => %v, %v, want %v, %v", etc.event, gotClick, gotState, etc.wantClick, etc.wantState)
				}
			}
		})
	}
}
