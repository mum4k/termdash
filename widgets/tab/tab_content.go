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
