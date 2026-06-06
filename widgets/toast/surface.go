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

package toast

import (
	"sync"

	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Surface is a reusable overlay widget that hosts toast managers at multiple
// placements.
//
// Use Manager directly when an application needs one stack. Use Surface when a
// single widget should support multiple corners or a custom overlay point.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Surface struct {
	mu       sync.Mutex
	managers map[Placement]*Manager
	order    []Placement
	opts     *surfaceOptions
}

// NewSurface returns a toast surface with one default placement registered.
func NewSurface(opts ...SurfaceOption) (*Surface, error) {
	so := newSurfaceOptions(opts...)
	if err := so.validate(); err != nil {
		return nil, err
	}
	s := &Surface{
		managers: map[Placement]*Manager{},
		opts:     so,
	}
	if err := s.Register(so.defaultPlacement); err != nil {
		return nil, err
	}
	for _, placement := range so.placements {
		if err := s.Register(placement.placement, placement.opts...); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Register creates or replaces the manager for a placement.
func (s *Surface) Register(p Placement, opts ...Option) error {
	managerOpts := append([]Option(nil), s.opts.defaultToastOpts...)
	if p != PlacementCustom {
		managerOpts = append(managerOpts, Anchor(p))
	}
	managerOpts = append(managerOpts, opts...)

	manager, err := New(managerOpts...)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.managers[p]; !ok {
		s.order = append(s.order, p)
	}
	s.managers[p] = manager
	return nil
}

// Notify adds a notification to the surface's default placement.
func (s *Surface) Notify(title, message string, opts ...NotificationOption) string {
	return s.NotifyAt(s.opts.defaultPlacement, title, message, opts...)
}

// NotifyAt adds a notification to the requested placement.
//
// Built-in placements are created lazily. PlacementCustom must be registered
// first with Surface.Register or SurfacePlacement so it can provide a
// CustomPosition option.
func (s *Surface) NotifyAt(p Placement, title, message string, opts ...NotificationOption) string {
	manager := s.manager(p)
	if manager == nil {
		if p == PlacementCustom {
			return ""
		}
		if err := s.Register(p); err != nil {
			return ""
		}
		manager = s.manager(p)
		if manager == nil {
			return ""
		}
	}
	return manager.Notify(title, message, opts...)
}

// Manager returns the manager for a placement when one exists.
func (s *Surface) Manager(p Placement) (*Manager, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	manager, ok := s.managers[p]
	return manager, ok
}

// Clear removes all notifications from every registered placement.
func (s *Surface) Clear() {
	s.mu.Lock()
	managers := s.managerSnapshotLocked()
	s.mu.Unlock()
	for _, manager := range managers {
		manager.Clear()
	}
}

// ClearAt removes all notifications from one placement.
func (s *Surface) ClearAt(p Placement) {
	if manager := s.manager(p); manager != nil {
		manager.Clear()
	}
}

// Count returns the total number of notifications held by the surface.
func (s *Surface) Count() int {
	s.mu.Lock()
	managers := s.managerSnapshotLocked()
	s.mu.Unlock()

	total := 0
	for _, manager := range managers {
		total += manager.Count()
	}
	return total
}

// Draw implements widgetapi.Widget.
func (s *Surface) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	if ar := cvs.Area(); ar.Dx() < s.opts.minimumSize.X || ar.Dy() < s.opts.minimumSize.Y {
		return draw.ResizeNeeded(cvs)
	}

	s.mu.Lock()
	managers := s.managerSnapshotLocked()
	s.mu.Unlock()
	for _, manager := range managers {
		if err := manager.Draw(cvs, meta); err != nil {
			return err
		}
	}
	return nil
}

// Keyboard implements widgetapi.Widget.
func (s *Surface) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	_, _ = k, meta
	return nil
}

// Mouse implements widgetapi.Widget.
func (s *Surface) Mouse(event *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	s.mu.Lock()
	managers := s.managerSnapshotLocked()
	s.mu.Unlock()
	for i := len(managers) - 1; i >= 0; i-- {
		if err := managers[i].Mouse(event, meta); err != nil {
			return err
		}
	}
	return nil
}

// Options implements widgetapi.Widget.
func (s *Surface) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize: s.opts.minimumSize,
		WantMouse:   widgetapi.MouseScopeWidget,
	}
}

func (s *Surface) manager(p Placement) *Manager {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.managers[p]
}

func (s *Surface) managerSnapshotLocked() []*Manager {
	managers := make([]*Manager, 0, len(s.order))
	for _, p := range s.order {
		if manager := s.managers[p]; manager != nil {
			managers = append(managers, manager)
		}
	}
	return managers
}
