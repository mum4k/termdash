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

/*
Package container defines a type that wraps other containers or widgets.

The container supports splitting container into sub containers, defining
container styles and placing widgets. The container also creates and manages
canvases assigned to the placed widgets.
*/
package container

import (
	"errors"
	"fmt"
	"image"
	"sync"

	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/private/alignfor"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/event"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Container wraps either sub containers or widgets and positions them on the
// terminal.
// This is thread-safe.
type Container struct {
	// parent is the parent container, nil if this is the root container.
	parent *Container
	// The sub containers, if these aren't nil, the widget must be.
	first  *Container
	second *Container

	// term is the terminal this container is placed on.
	// All containers in the tree share the same terminal.
	term terminalapi.Terminal

	// focusTracker tracks the active (focused) container.
	// All containers in the tree share the same tracker.
	focusTracker *focusTracker

	// area is the area of the terminal this container has access to.
	// Initialized the first time Draw is called.
	area image.Rectangle

	// opts are the options provided to the container.
	opts *options

	// clearNeeded indicates if the terminal needs to be cleared next time we
	// are clearNeeded the container.
	// This is required if the container was updated and thus the layout might
	// have changed.
	clearNeeded bool

	// mu protects the container tree.
	// All containers in the tree share the same lock.
	mu *sync.Mutex
}

// String represents the container metadata in a human readable format.
// Implements fmt.Stringer.
func (c *Container) String() string {
	return fmt.Sprintf("Container@%p{parent:%p, first:%p, second:%p, area:%+v}", c, c.parent, c.first, c.second, c.area)
}

// New returns a new root container that will use the provided terminal and
// applies the provided options.
func New(t terminalapi.Terminal, opts ...Option) (*Container, error) {
	root := &Container{
		term: t,
		opts: newOptions( /* parent = */ nil),
		mu:   &sync.Mutex{},
	}

	// Initially the root is focused.
	root.focusTracker = newFocusTracker(root)
	if err := applyOptions(root, opts...); err != nil {
		return nil, err
	}
	if err := validateOptions(root); err != nil {
		return nil, err
	}
	return root, nil
}

// newChild creates a new child container of the given parent.
func newChild(parent *Container, opts []Option) (*Container, error) {
	child := &Container{
		parent:       parent,
		term:         parent.term,
		focusTracker: parent.focusTracker,
		opts:         newOptions(parent.opts),
		mu:           parent.mu,
	}
	if err := applyOptions(child, opts...); err != nil {
		return nil, err
	}
	return child, nil
}

// hasBorder determines if this container has a border.
func (c *Container) hasBorder() bool {
	return c.opts.border != linestyle.None
}

// hasWidget determines if this container has a widget.
func (c *Container) hasWidget() bool {
	return c.opts.widget != nil
}

// isLeaf determines if this container is a leaf container in the binary tree of containers.
// Only leaf containers are guaranteed to be "visible" on the screen, because
// they are on the top of other non-leaf containers.
func (c *Container) isLeaf() bool {
	return c.first == nil && c.second == nil
}

// usable returns the usable area in this container.
// This depends on whether the container has a border, etc.
func (c *Container) usable() image.Rectangle {
	if c.hasBorder() {
		return area.ExcludeBorder(c.area)
	}
	return c.area
}

// widgetArea returns the area in the container that is available for the
// widget's canvas. Takes the container border, widget's requested maximum size
// and ratio and container's alignment into account.
// Returns a zero area if the container has no widget.
func (c *Container) widgetArea() (image.Rectangle, error) {
	if !c.hasWidget() {
		return image.ZR, nil
	}

	padded, err := c.opts.padding.apply(c.usable())
	if err != nil {
		return image.ZR, err
	}
	wOpts := c.opts.widget.Options()

	adjusted := padded
	if maxX := wOpts.MaximumSize.X; maxX > 0 && adjusted.Dx() > maxX {
		adjusted.Max.X -= adjusted.Dx() - maxX
	}
	if maxY := wOpts.MaximumSize.Y; maxY > 0 && adjusted.Dy() > maxY {
		adjusted.Max.Y -= adjusted.Dy() - maxY
	}

	if wOpts.Ratio.X > 0 && wOpts.Ratio.Y > 0 {
		adjusted = area.WithRatio(adjusted, wOpts.Ratio)
	}
	aligned, err := alignfor.Rectangle(padded, adjusted, c.opts.hAlign, c.opts.vAlign)
	if err != nil {
		return image.ZR, err
	}
	return aligned, nil
}

// split splits the container's usable area into child areas.
// Panics if the container isn't configured for a split.
func (c *Container) split() (image.Rectangle, image.Rectangle, error) {
	ar, err := c.opts.padding.apply(c.usable())
	if err != nil {
		return image.ZR, image.ZR, err
	}
	if c.opts.splitFixed > DefaultSplitFixed {
		if c.opts.split == splitTypeVertical {
			return area.VSplitCells(ar, c.opts.splitFixed)
		}
		return area.HSplitCells(ar, c.opts.splitFixed)
	}

	if c.opts.split == splitTypeVertical {
		return area.VSplit(ar, c.opts.splitPercent)
	}
	return area.HSplit(ar, c.opts.splitPercent)
}

// createFirst creates and returns the first sub container of this container.
func (c *Container) createFirst(opts []Option) error {
	first, err := newChild(c, opts)
	if err != nil {
		return err
	}
	c.first = first
	return nil
}

// createSecond creates and returns the second sub container of this container.
func (c *Container) createSecond(opts []Option) error {
	second, err := newChild(c, opts)
	if err != nil {
		return err
	}
	c.second = second
	return nil
}

// Draw draws this container and all of its sub containers.
func (c *Container) Draw() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.clearNeeded {
		if err := c.term.Clear(); err != nil {
			return fmt.Errorf("term.Clear => error: %v", err)
		}
		c.clearNeeded = false
	}

	// Update the area we are tracking for focus in case the terminal size
	// changed.
	ar, err := area.FromSize(c.term.Size())
	if err != nil {
		return err
	}
	c.focusTracker.updateArea(ar)
	return drawTree(c)
}

// Update updates container with the specified id by setting the provided
// options. This can be used to perform dynamic layout changes, i.e. anything
// between replacing the widget in the container and completely changing the
// layout and splits.
// The argument id must match exactly one container with that was created with
// matching ID() option. The argument id must not be an empty string.
func (c *Container) Update(id string, opts ...Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	target, err := findID(c, id)
	if err != nil {
		return err
	}
	c.clearNeeded = true

	if err := applyOptions(target, opts...); err != nil {
		return err
	}
	if err := validateOptions(c); err != nil {
		return err
	}

	// The currently focused container might not be reachable anymore, because
	// it was under the target. If that is so, move the focus up to the target.
	if !c.focusTracker.reachableFrom(c) {
		c.focusTracker.setActive(target)
	}
	return nil
}

// updateFocusFromMouse processes the mouse event and determines if it changes
// the focused container.
// Caller must hold c.mu.
func (c *Container) updateFocusFromMouse(m *terminalapi.Mouse) {
	target := pointCont(c, m.Position)
	if target == nil { // Ignore mouse clicks where no containers are.
		return
	}
	c.focusTracker.mouse(target, m)
}

// inFocusGroup returns true if this container is in the specified focus group.
func (c *Container) inFocusGroup(fg FocusGroup) bool {
	for _, cg := range c.opts.keyFocusGroups {
		if cg == fg {
			return true
		}
	}
	return false
}

// updateFocusFromKeyboard processes the keyboard event and determines if it
// changes the focused container.
// Caller must hold c.mu.
func (c *Container) updateFocusFromKeyboard(k *terminalapi.Keyboard) {
	active := c.focusTracker.active()
	nextGroupsForKey, isGroupKeyForNext := active.opts.global.keyFocusGroupsNext[k.Key]
	prevGroupsForKey, isGroupKeyForPrev := active.opts.global.keyFocusGroupsPrevious[k.Key]

	nextMatchesContGroup, nextG := nextGroupsForKey.firstMatching(active.opts.keyFocusGroups)
	prevMatchesContGroup, prevG := prevGroupsForKey.firstMatching(active.opts.keyFocusGroups)

	switch {
	case active.opts.global.keyFocusNext != nil && *active.opts.global.keyFocusNext == k.Key:
		c.focusTracker.next( /* group = */ nil)
	case active.opts.global.keyFocusPrevious != nil && *active.opts.global.keyFocusPrevious == k.Key:
		c.focusTracker.previous( /* group = */ nil)
	case isGroupKeyForNext && nextMatchesContGroup:
		c.focusTracker.next(&nextG)
	case isGroupKeyForPrev && prevMatchesContGroup:
		c.focusTracker.previous(&prevG)
	}
}

// processEvent processes events delivered to the container.
func (c *Container) processEvent(ev terminalapi.Event) error {
	// This is done in two stages.
	// 1) under lock we traverse the container and identify all targets
	//    (widgets) that should receive the event.
	// 2) lock is released and events are delivered to the widgets. Widgets
	//    themselves are thread-safe. Lock must be releases when delivering,
	//    because some widgets might try to mutate the container when they
	//    receive the event, like dynamically change the layout.
	c.mu.Lock()
	sendFn, err := c.prepareEvTargets(ev)
	c.mu.Unlock()
	if err != nil {
		return err
	}
	return sendFn()
}

// prepareEvTargets returns a closure, that when called delivers the event to
// widgets that registered for it.
// Also processes the event on behalf of the container (tracks keyboard focus).
// Caller must hold c.mu.
func (c *Container) prepareEvTargets(ev terminalapi.Event) (func() error, error) {
	switch e := ev.(type) {
	case *terminalapi.Mouse:
		c.updateFocusFromMouse(ev.(*terminalapi.Mouse))

		targets, err := c.mouseEvTargets(e)
		if err != nil {
			return nil, err
		}
		return func() error {
			for _, mt := range targets {
				if err := mt.widget.Mouse(mt.ev, mt.meta); err != nil {
					return err
				}
			}
			return nil
		}, nil

	case *terminalapi.Keyboard:
		c.updateFocusFromKeyboard(ev.(*terminalapi.Keyboard))

		targets := c.keyEvTargets()
		return func() error {
			for _, kt := range targets {
				if err := kt.widget.Keyboard(e, kt.meta); err != nil {
					return err
				}
			}
			return nil
		}, nil

	default:
		return nil, fmt.Errorf("container received an unsupported event type %T", ev)
	}
}

// keyEvTarget contains a widget that should receive an event and the metadata
// for the event.
type keyEvTarget struct {
	// widget is the widget that should receive the keyboard event.
	widget widgetapi.Widget
	// meta is the metadata about the event.
	meta *widgetapi.EventMeta
}

// newKeyEvTarget returns a new keyEvTarget.
func newKeyEvTarget(w widgetapi.Widget, meta *widgetapi.EventMeta) *keyEvTarget {
	return &keyEvTarget{
		widget: w,
		meta:   meta,
	}
}

// keyEvTargets returns those widgets found in the container that should
// receive this keyboard event.
// Caller must hold c.mu.
func (c *Container) keyEvTargets() []*keyEvTarget {
	var (
		errStr  string
		targets []*keyEvTarget
		// If the currently focused widget set the ExclusiveKeyboardOnFocus
		// option, this pointer is set to that widget.
		exclusiveWidget widgetapi.Widget
	)

	// All the targets that should receive this event.
	// For now stable ordering (preOrder).
	preOrder(c, &errStr, visitFunc(func(cur *Container) error {
		if !cur.hasWidget() {
			return nil
		}

		focused := cur.focusTracker.isActive(cur)
		meta := &widgetapi.EventMeta{
			Focused: focused,
		}
		wOpt := cur.opts.widget.Options()
		if focused && wOpt.ExclusiveKeyboardOnFocus {
			exclusiveWidget = cur.opts.widget
		}

		switch wOpt.WantKeyboard {
		case widgetapi.KeyScopeNone:
			// Widget doesn't want any keyboard events.
			return nil

		case widgetapi.KeyScopeFocused:
			if focused {
				targets = append(targets, newKeyEvTarget(cur.opts.widget, meta))
			}

		case widgetapi.KeyScopeGlobal:
			targets = append(targets, newKeyEvTarget(cur.opts.widget, meta))
		}
		return nil
	}))

	if exclusiveWidget != nil {
		targets = []*keyEvTarget{
			newKeyEvTarget(exclusiveWidget, &widgetapi.EventMeta{Focused: true}),
		}
	}
	return targets
}

// mouseEvTarget contains a mouse event adjusted relative to the widget's area,
// the widget that should receive it and metadata about the event.
type mouseEvTarget struct {
	// widget is the widget that should receive the mouse event.
	widget widgetapi.Widget
	// ev is the adjusted mouse event.
	ev *terminalapi.Mouse
	// meta is the metadata about the event.
	meta *widgetapi.EventMeta
}

// newMouseEvTarget returns a new mouseEvTarget.
func newMouseEvTarget(w widgetapi.Widget, wArea image.Rectangle, ev *terminalapi.Mouse, meta *widgetapi.EventMeta) *mouseEvTarget {
	return &mouseEvTarget{
		widget: w,
		ev:     adjustMouseEv(ev, wArea),
		meta:   meta,
	}
}

// mouseEvTargets returns those widgets found in the container that should
// receive this mouse event.
// Caller must hold c.mu.
func (c *Container) mouseEvTargets(m *terminalapi.Mouse) ([]*mouseEvTarget, error) {
	var (
		errStr  string
		widgets []*mouseEvTarget
	)

	// All the widgets that should receive this event.
	// For now stable ordering (preOrder).
	preOrder(c, &errStr, visitFunc(func(cur *Container) error {
		if !cur.hasWidget() {
			return nil
		}

		wOpts := cur.opts.widget.Options()
		wa, err := cur.widgetArea()
		if err != nil {
			return err
		}

		meta := &widgetapi.EventMeta{
			Focused: cur.focusTracker.isActive(cur),
		}
		switch wOpts.WantMouse {
		case widgetapi.MouseScopeNone:
			// Widget doesn't want any mouse events.
			return nil

		case widgetapi.MouseScopeWidget:
			// Only if the event falls inside of the widget's canvas.
			if m.Position.In(wa) {
				widgets = append(widgets, newMouseEvTarget(cur.opts.widget, wa, m, meta))
			}

		case widgetapi.MouseScopeContainer:
			// Only if the event falls inside the widget's parent container.
			if m.Position.In(cur.area) {
				widgets = append(widgets, newMouseEvTarget(cur.opts.widget, wa, m, meta))
			}

		case widgetapi.MouseScopeGlobal:
			// Widget wants all mouse events.
			widgets = append(widgets, newMouseEvTarget(cur.opts.widget, wa, m, meta))
		}
		return nil
	}))

	if errStr != "" {
		return nil, errors.New(errStr)
	}
	return widgets, nil
}

// Subscribe tells the container to subscribe itself and widgets to the
// provided event distribution system.
// This method is private to termdash, stability isn't guaranteed and changes
// won't be backward compatible.
func (c *Container) Subscribe(eds *event.DistributionSystem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// maxReps is the maximum number of repetitive events towards widgets
	// before we throttle them.
	const maxReps = 10

	// Subscriber the container itself in order to track keyboard focus.
	want := []terminalapi.Event{
		&terminalapi.Keyboard{},
		&terminalapi.Mouse{},
	}
	eds.Subscribe(want, func(ev terminalapi.Event) {
		if err := c.processEvent(ev); err != nil {
			eds.Event(terminalapi.NewErrorf("failed to process event %v: %v", ev, err))
		}
	}, event.MaxRepetitive(maxReps))
}

// adjustMouseEv adjusts the mouse event relative to the widget area.
func adjustMouseEv(m *terminalapi.Mouse, wArea image.Rectangle) *terminalapi.Mouse {
	// The sent mouse coordinate is relative to the widget canvas, i.e. zero
	// based, even though the widget might not be in the top left corner on the
	// terminal.
	offset := wArea.Min
	if m.Position.In(wArea) {
		return &terminalapi.Mouse{
			Position: m.Position.Sub(offset),
			Button:   m.Button,
		}
	}
	return &terminalapi.Mouse{
		Position: image.Point{-1, -1},
		Button:   m.Button,
	}
}
