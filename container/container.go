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
	"fmt"
	"image"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminalapi"
)

// Container wraps either sub containers or widgets and positions them on the
// terminal.
// This is not thread-safe.
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
	area image.Rectangle

	// opts are the options provided to the container.
	opts *options
}

// String represents the container metadata in a human readable format.
// Implements fmt.Stringer.
func (c *Container) String() string {
	return fmt.Sprintf("Container@%p{parent:%p, first:%p, second:%p, area:%+v}", c, c.parent, c.first, c.second, c.area)
}

// New returns a new root container that will use the provided terminal and
// applies the provided options.
func New(t terminalapi.Terminal, opts ...Option) *Container {
	size := t.Size()
	root := &Container{
		term: t,
		// The root container has access to the entire terminal.
		area: image.Rect(0, 0, size.X, size.Y),
		opts: newOptions( /* parent = */ nil),
	}

	// Initially the root is focused.
	root.focusTracker = newFocusTracker(root)
	applyOptions(root, opts...)
	return root
}

// newChild creates a new child container of the given parent.
func newChild(parent *Container, area image.Rectangle) *Container {
	return &Container{
		parent:       parent,
		term:         parent.term,
		focusTracker: parent.focusTracker,
		area:         area,
		opts:         newOptions(parent.opts),
	}
}

// hasBorder determines if this container has a border.
func (c *Container) hasBorder() bool {
	return c.opts.border != draw.LineStyleNone
}

// hasWidget determines if this container has a widget.
func (c *Container) hasWidget() bool {
	return c.opts.widget != nil
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

	adjusted := c.usable()
	wOpts := c.opts.widget.Options()

	if maxX := wOpts.MaximumSize.X; maxX > 0 && adjusted.Dx() > maxX {
		adjusted.Max.X -= adjusted.Dx() - maxX
	}
	if maxY := wOpts.MaximumSize.Y; maxY > 0 && adjusted.Dy() > maxY {
		adjusted.Max.Y -= adjusted.Dy() - maxY
	}

	if wOpts.Ratio.X > 0 && wOpts.Ratio.Y > 0 {
		adjusted = area.WithRatio(adjusted, wOpts.Ratio)
	}
	adjusted, err := align.Rectangle(c.usable(), adjusted, c.opts.hAlign, c.opts.vAlign)
	if err != nil {
		return image.ZR, err
	}
	return adjusted, nil
}

// split splits the container's usable area into child areas.
// Panics if the container isn't configured for a split.
func (c *Container) split() (image.Rectangle, image.Rectangle) {
	ar := c.usable()
	if c.opts.split == splitTypeVertical {
		return area.VSplit(ar)
	}
	return area.HSplit(ar)
}

// createFirst creates and returns the first sub container of this container.
func (c *Container) createFirst() *Container {
	ar, _ := c.split()
	c.first = newChild(c, ar)
	return c.first
}

// createSecond creates and returns the second sub container of this container.
func (c *Container) createSecond() *Container {
	_, ar := c.split()
	c.second = newChild(c, ar)
	return c.second
}

// Draw draws this container and all of its sub containers.
func (c *Container) Draw() error {
	return drawTree(c)
}

// Keyboard is used to forward a keyboard event to the container.
// Keyboard events are forwarded to the widget in the currently focused
// container, assuming that the widget registered for keyboard events.
func (c *Container) Keyboard(k *terminalapi.Keyboard) error {
	w := c.focusTracker.active().opts.widget
	if w == nil || !w.Options().WantKeyboard {
		return nil
	}
	return w.Keyboard(k)
}

// Mouse is used to forward a mouse event to the container.
// Container uses mouse events to track and change which is the active
// (focused) container.
//
// If the container that receives the mouse click contains a widget that
// registered for mouse events, the mouse event is further forwarded to that
// widget. Only mouse events that fall within the widget's canvas are forwarded
// and the coordinates are adjusted relative to the widget's canvas.
func (c *Container) Mouse(m *terminalapi.Mouse) error {
	c.focusTracker.mouse(m)

	target := pointCont(c, m.Position)
	if target == nil { // Ignore mouse clicks where no containers are.
		return nil
	}
	w := target.opts.widget
	if w == nil || !w.Options().WantMouse {
		return nil
	}

	// Ignore clicks falling outside of the container.
	if !m.Position.In(target.usable()) {
		return nil
	}

	// Ignore clicks falling outside of the widget's canvas.
	wa, err := target.widgetArea()
	if err != nil {
		return err
	}
	if !m.Position.In(wa) {
		return nil
	}

	// The sent mouse coordinate is relative to the widget canvas, i.e. zero
	// based, even though the widget might not be in the top left corner on the
	// terminal.
	offset := wa.Min
	wm := &terminalapi.Mouse{
		Position: m.Position.Sub(offset),
		Button:   m.Button,
	}
	return w.Mouse(wm)
}
