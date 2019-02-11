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

// Package clickfsm implements a state machine that tracks mouse clicks.
package clickfsm

import (
	"errors"
	"fmt"
	"image"

	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminalapi"
)

// Callback is called with the filtered mouse events.
//
// Only mouse events that fall inside the area provided to NewClickFSM are
// forwarded.
//
// Doesn't forward the mouse event mouse.ButtonLeft immediately. The mouse.ButtonLeft event
// is only forwarded if it is followed by a mouse.ButtonRelease also inside the area.
// The forwarded event has a position of the location where the mouse.ButtonRelease
// happened.
//
// Forwards the following mouse events transparently:
//   - mouse.ButtonRight
//   - mouse.ButtonMiddle
//   - mouse.ButtonWheelUp
//   - mouse.ButtonWheelDown
//
// Implementations must be thread-safe, events arrive from a separate
// goroutine.
//
// If the callback returns an error, the error is forwarded to the termdash
// infrastructure which panics, unless the user provided a
// termdash.ErrorHandler.
type Callback func(m *terminalapi.Mouse) error

// ClickFSM implements a finite-state machine that tracks mouse clicks within
// an area.
//
// Simplifies tracking of mouse left-button clicks, i.e. when the caller wants
// to perform an action only if both the mouse.ButtonLeft and the mouse.ButtonRelease
// happen within the specified area.
//
// All other mouse events that fall within the area are directly forwarded to
// the callback.
type ClickFSM struct {
	// area is the area provided to NewClickFSM.
	area image.Rectangle

	// callback is the receiver of the forwarded mouse events.
	callback Callback

	// state is the current state of the FSM.
	state mouseStateFn
}

// NewClickFSM creates a new ClickFSM instance that forwards mouse clicks that
// fall within the provided area to the specified callback function.
//
// The area must be at least 1x1 cell large and the callbackFn must not be nil.
func NewClickFSM(area image.Rectangle, callbackFn Callback) (*ClickFSM, error) {
	if min := 1; area.Dx() < min || area.Dy() < min {
		return nil, fmt.Errorf("invalid area %v, must be at least %dx%d cells large", area, min, min)
	}
	if callbackFn == nil {
		return nil, errors.New("the callback function must not be nil")
	}

	return &ClickFSM{
		area:     area,
		callback: callbackFn,
		state:    wantLeftButton,
	}, nil
}

// Event is used to forward terminal events to the state machine.
// Ignores all event types except *terminalapi.Mouse.
func (cf *ClickFSM) Event(ev terminalapi.Event) error {
	switch m := ev.(type) {
	case *terminalapi.Mouse:
		next, err := cf.state(cf, m)
		if err != nil {
			return err
		}
		cf.state = next
		return nil

	default:
		return nil
	}
}

// mouseStateFn is a single state in the state machine.
// Returns the next state.
type mouseStateFn func(fsm *ClickFSM, m *terminalapi.Mouse) (mouseStateFn, error)

// wantLeftButton is the initial state, expecting a left button click inside a
// container. Transparently forwards other mouse events if they fall within the
// specified area.
func wantLeftButton(cf *ClickFSM, m *terminalapi.Mouse) (mouseStateFn, error) {
	if !m.Position.In(cf.area) {
		// Ignore events outside of the specified area.
		return wantLeftButton, nil
	}

	switch m.Button {
	case mouse.ButtonRight, mouse.ButtonMiddle, mouse.ButtonWheelUp, mouse.ButtonWheelDown:
		if err := cf.callback(m); err != nil {
			return nil, err
		}

	case mouse.ButtonLeft:
		return wantRelease, nil
	}
	return wantLeftButton, nil
}

// wantRelease waits for a mouse button release in the same area as
// the click or another left mouse button click.
// Transparently forwards other mouse events if they fall within the specified
// area.
func wantRelease(cf *ClickFSM, m *terminalapi.Mouse) (mouseStateFn, error) {
	if !m.Position.In(cf.area) {
		// Ignore events outside of the specified area.
		return wantRelease, nil
	}

	switch m.Button {
	case mouse.ButtonRight, mouse.ButtonMiddle, mouse.ButtonWheelUp, mouse.ButtonWheelDown:
		if err := cf.callback(m); err != nil {
			return nil, err
		}

	case mouse.ButtonLeft:
		// Stays in the same state still expecting the release.

	case mouse.ButtonRelease:
		if err := cf.callback(&terminalapi.Mouse{
			Position: m.Position,
			Button:   mouse.ButtonLeft,
		}); err != nil {
			return nil, err
		}
		return wantLeftButton, nil
	}
	return wantRelease, nil
}
