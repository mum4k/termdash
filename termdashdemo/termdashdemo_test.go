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
	"testing"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/tab"
)

// TestHandleThreeDArrowKey verifies arrow keys control the ThreeD tab instead
// of falling through to the global tab navigation handler.
func TestHandleThreeDArrowKey(t *testing.T) {
	overview := &tab.Tab{Name: "Overview"}
	threeD := &tab.Tab{Name: "ThreeD"}
	manager := tab.NewManager(overview, threeD)
	stage := &keyboardRecorder{}

	if handled, err := handleThreeDArrowKey(&terminalapi.Keyboard{Key: keyboard.KeyArrowRight}, manager, threeD, stage); handled || err != nil {
		t.Fatalf("inactive ThreeD tab handled = %v, err = %v; want false, nil", handled, err)
	}
	if got := stage.calls; got != 0 {
		t.Fatalf("stage calls while inactive = %d, want 0", got)
	}

	manager.SetActiveTab(1)
	if handled, err := handleThreeDArrowKey(&terminalapi.Keyboard{Key: keyboard.KeyArrowRight}, manager, threeD, stage); !handled || err != nil {
		t.Fatalf("active ThreeD arrow handled = %v, err = %v; want true, nil", handled, err)
	}
	if got, want := stage.lastKey, keyboard.KeyArrowRight; got != want {
		t.Fatalf("stage key = %v, want %v", got, want)
	}
	if handled, err := handleThreeDArrowKey(&terminalapi.Keyboard{Key: keyboard.KeyTab}, manager, threeD, stage); handled || err != nil {
		t.Fatalf("tab key handled = %v, err = %v; want false, nil", handled, err)
	}
}

type keyboardRecorder struct {
	calls   int
	lastKey keyboard.Key
}

func (r *keyboardRecorder) Keyboard(k *terminalapi.Keyboard, _ *widgetapi.EventMeta) error {
	r.calls++
	if k != nil {
		r.lastKey = k.Key
	}
	return nil
}
