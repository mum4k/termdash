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

// Package tab provides functionality for managing tabbed interfaces.
package tab

import (
	"github.com/mum4k/termdash/container"
)

// Content displays the content of the active tab.
type Content struct {
	tm *Manager // Reference to the Tab Manager.
}

// NewContent creates a new Content.
func NewContent(tm *Manager) *Content {
	return &Content{
		tm: tm,
	}
}

// Update updates the content based on the active tab.
func (c *Content) Update(cont *container.Container) error {
	activeTab := c.tm.GetActiveTab()
	if activeTab == nil {
		return nil
	}
	return cont.Update("tabContent", activeTab.Content)
}
