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
	"sync"
	"time"

	"github.com/mum4k/termdash/container"
)

// Tab stores one tab label and the container content shown when the tab is active.
//
// The notification state is protected separately so callers can cheaply ask a tab
// whether it still has an active notification without holding the manager lock.
type Tab struct {
	// Name is the label shown in the tab header.
	Name string
	// Content is the container option used to render the tab body.
	Content container.Option

	mu                   sync.RWMutex
	notification         bool
	notificationDeadline time.Time
}

// HasNotification reports whether the tab currently has an active notification.
//
// Expired notifications are cleared lazily when observed so callers do not need a
// dedicated sweeper.
func (t *Tab) HasNotification() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.notification {
		return false
	}
	if !t.notificationDeadline.IsZero() && time.Now().After(t.notificationDeadline) {
		t.notification = false
		t.notificationDeadline = time.Time{}
		return false
	}
	return true
}

// setNotification replaces the current notification state.
func (t *Tab) setNotification(active bool, duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.notification = active
	t.notificationDeadline = time.Time{}
	if active && duration > 0 {
		t.notificationDeadline = time.Now().Add(duration)
	}
}

// Snapshot captures the header-visible state of a tab.
type Snapshot struct {
	Name         string
	Notification bool
}

// Manager owns tab order, active tab state, and notification routing.
type Manager struct {
	mu          sync.RWMutex
	tabs        []*Tab
	activeIndex int
}

// NewManager returns a new Manager.
func NewManager(tabs ...*Tab) *Manager {
	m := &Manager{}
	for _, tab := range tabs {
		m.AddTab(tab)
	}
	return m
}

// AddTab appends one tab to the manager.
func (m *Manager) AddTab(tab *Tab) {
	if tab == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.tabs = append(m.tabs, tab)
	if len(m.tabs) == 1 {
		m.activeIndex = 0
	}
}

// GetTabNum returns the number of tabs currently managed.
func (m *Manager) GetTabNum() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tabs)
}

// GetTabNames returns a copy of the current tab labels.
func (m *Manager) GetTabNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.tabs))
	for _, tab := range m.tabs {
		names = append(names, tab.Name)
	}
	return names
}

// GetActiveIndex returns the current active tab index.
func (m *Manager) GetActiveIndex() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.tabs) == 0 {
		return -1
	}
	return m.activeIndex
}

// GetActiveTab returns the current active tab.
func (m *Manager) GetActiveTab() *Tab {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.tabs) == 0 || m.activeIndex < 0 || m.activeIndex >= len(m.tabs) {
		return nil
	}
	return m.tabs[m.activeIndex]
}

// SetActiveTab makes the specified index active.
//
// Returns true when the requested index became active.
func (m *Manager) SetActiveTab(index int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if index < 0 || index >= len(m.tabs) {
		return false
	}
	m.activeIndex = index
	return true
}

// NextTab activates the next tab, wrapping at the end.
func (m *Manager) NextTab() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.tabs) == 0 {
		return false
	}
	m.activeIndex = (m.activeIndex + 1) % len(m.tabs)
	return true
}

// PreviousTab activates the previous tab, wrapping at the beginning.
func (m *Manager) PreviousTab() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.tabs) == 0 {
		return false
	}
	m.activeIndex--
	if m.activeIndex < 0 {
		m.activeIndex = len(m.tabs) - 1
	}
	return true
}

// SetNotification replaces the notification state for the tab at index.
func (m *Manager) SetNotification(index int, active bool, duration time.Duration) bool {
	m.mu.RLock()
	if index < 0 || index >= len(m.tabs) {
		m.mu.RUnlock()
		return false
	}
	tab := m.tabs[index]
	m.mu.RUnlock()

	tab.setNotification(active, duration)
	return true
}

// GetNotifiedTabs returns all tabs that currently have an active notification.
func (m *Manager) GetNotifiedTabs() []*Tab {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var notified []*Tab
	for _, tab := range m.tabs {
		if tab.HasNotification() {
			notified = append(notified, tab)
		}
	}
	return notified
}

// GetNotifiedIndexes returns the indexes of tabs that currently have an active notification.
func (m *Manager) GetNotifiedIndexes() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var indexes []int
	for i, tab := range m.tabs {
		if tab.HasNotification() {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

// HasNotificationAt reports whether the tab at index currently has an active notification.
func (m *Manager) HasNotificationAt(index int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if index < 0 || index >= len(m.tabs) {
		return false
	}
	return m.tabs[index].HasNotification()
}

// GetTabIndex returns the current index of target.
func (m *Manager) GetTabIndex(target *Tab) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i, tab := range m.tabs {
		if tab == target {
			return i
		}
	}
	return -1
}

// Snapshot returns a stable copy of the header-visible manager state.
func (m *Manager) Snapshot() ([]Snapshot, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := make([]Snapshot, 0, len(m.tabs))
	for _, tab := range m.tabs {
		snapshot = append(snapshot, Snapshot{
			Name:         tab.Name,
			Notification: tab.HasNotification(),
		})
	}
	if len(m.tabs) == 0 {
		return snapshot, -1
	}
	return snapshot, m.activeIndex
}
