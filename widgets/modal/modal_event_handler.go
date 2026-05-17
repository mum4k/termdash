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

package modal

import (
	"context"
	"sync"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// EventHandler routes mouse and keyboard events used by the modal demo.
type EventHandler struct {
	cancel        context.CancelFunc
	rootContainer *container.Container
	modalManager  *ModalManager
	mutex         sync.Mutex
}

// NewEventHandler creates an event handler for a modal manager and root container.
func NewEventHandler(ctx context.Context, cancel context.CancelFunc, rootContainer *container.Container, modalManager *ModalManager) *EventHandler {
	_ = ctx
	return &EventHandler{
		cancel:        cancel,
		rootContainer: rootContainer,
		modalManager:  modalManager,
	}
}

// HandleMouse intentionally ignores mouse events.
//
// Modal widgets receive mouse input through the standard termdash widget event
// path, which provides coordinates relative to the modal canvas. Forwarding raw
// terminal mouse events from the demo would bypass that translation and make
// dragging inaccurate.
func (eh *EventHandler) HandleMouse(event *terminalapi.Mouse) {
	_ = event
	eh.mutex.Lock()
	defer eh.mutex.Unlock()
}

// HandleKeyboard closes the active modal on Escape and quits on q, Q, or Ctrl+C.
func (eh *EventHandler) HandleKeyboard(k *terminalapi.Keyboard) {
	eh.mutex.Lock()
	defer eh.mutex.Unlock()

	switch k.Key {
	case 'q', 'Q', keyboard.KeyCtrlC:
		eh.cancel()
	case keyboard.KeyEsc:
		_ = eh.modalManager.HideModal(eh.rootContainer)
	}
}
