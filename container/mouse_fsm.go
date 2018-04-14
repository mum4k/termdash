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

// mouse_fsm.go implements a state machine that tracks mouse clicks in regards
// to changing which container is focused.

import (
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminalapi"
)

// mouseStateFn is a single state in the focus tracking state machine.
// Returns the next state.
type mouseStateFn func(ft *focusTracker, m *terminalapi.Mouse) mouseStateFn

// nextForLeftClick determines the next state for a left mouse click.
func nextForLeftClick(ft *focusTracker, m *terminalapi.Mouse) mouseStateFn {
	// The click isn't in any known container.
	if ft.candidate = pointCont(ft.container, m.Position); ft.candidate == nil {
		return mouseWantLeftButton
	}
	return mouseWantRelease
}

// mouseWantLeftButton is the initial state, expecting a left button click inside a container.
func mouseWantLeftButton(ft *focusTracker, m *terminalapi.Mouse) mouseStateFn {
	if m.Button != mouse.ButtonLeft {
		return mouseWantLeftButton
	}
	return nextForLeftClick(ft, m)
}

// mouseWantRelease waits for a mouse button release in the same container as
// the click or a timeout or other left mouse button click.
func mouseWantRelease(ft *focusTracker, m *terminalapi.Mouse) mouseStateFn {
	switch m.Button {
	case mouse.ButtonLeft:
		return nextForLeftClick(ft, m)

	case mouse.ButtonRelease:
		// Process the release.
	default:
		return mouseWantLeftButton
	}

	// The release happened in another container.
	if ft.candidate != pointCont(ft.container, m.Position) {
		return mouseWantLeftButton
	}

	ft.container = ft.candidate
	ft.candidate = nil
	return mouseWantLeftButton
}
