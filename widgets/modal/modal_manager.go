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
	"sync"

	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

// ModalManager tracks the currently visible modal.
type ModalManager struct {
	activeModal *Modal
	mutex       sync.Mutex
}

// ShowModal places the modal into the container whose ID matches modal.ID.
func (mm *ModalManager) ShowModal(modal *Modal, c *container.Container) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.activeModal = modal
	return c.Update(
		modal.ID,
		container.PlaceWidget(modal),
	)
}

// HideModal removes the active modal from its host container.
func (mm *ModalManager) HideModal(c *container.Container) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if mm.activeModal == nil {
		return nil
	}
	err := c.Update(
		mm.activeModal.ID,
		container.Clear(),
	)
	mm.activeModal = nil
	return err
}

// HandleMouse forwards a mouse event to the active modal.
func (mm *ModalManager) HandleMouse(event *terminalapi.Mouse) bool {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if mm.activeModal == nil {
		return false
	}
	mm.activeModal.HandleMouse(event)
	return true
}

// HasActiveModal reports whether a modal is currently visible.
func (mm *ModalManager) HasActiveModal() bool {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	return mm.activeModal != nil
}
