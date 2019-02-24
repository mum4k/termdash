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
	"sync"

	"github.com/mum4k/termdash/internal/alignfor"
	"github.com/mum4k/termdash/internal/area"
	"github.com/mum4k/termdash/internal/event"
	"github.com/mum4k/termdash/internal/widgetapi"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
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
	area image.Rectangle

	// opts are the options provided to the container.
	opts *options

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
	size := t.Size()
	root := &Container{
		term: t,
		// The root container has access to the entire terminal.
		area: image.Rect(0, 0, size.X, size.Y),
		opts: newOptions( /* parent = */ nil),
		mu:   &sync.Mutex{},
	}

	// Initially the root is focused.
	root.focusTracker = newFocusTracker(root)
	if err := applyOptions(root, opts...); err != nil {
		return nil, err
	}
	return root, nil
}

// newChild creates a new child container of the given parent.
func newChild(parent *Container, area image.Rectangle) *Container {
	return &Container{
		parent:       parent,
		term:         parent.term,
		focusTracker: parent.focusTracker,
		area:         area,
		opts:         newOptions(parent.opts),
		mu:           parent.mu,
	}
}

// hasBorder determines if this container has a border.
func (c *Container) hasBorder() bool {
	return c.opts.border != linestyle.None
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
	adjusted, err := alignfor.Rectangle(c.usable(), adjusted, c.opts.hAlign, c.opts.vAlign)
	if err != nil {
		return image.ZR, err
	}
	return adjusted, nil
}

// split splits the container's usable area into child areas.
// Panics if the container isn't configured for a split.
func (c *Container) split() (image.Rectangle, image.Rectangle, error) {
	ar := c.usable()
	if c.opts.split == splitTypeVertical {
		return area.VSplit(ar, c.opts.splitPercent)
	}
	return area.HSplit(ar, c.opts.splitPercent)
}

// createFirst creates and returns the first sub container of this container.
func (c *Container) createFirst() (*Container, error) {
	ar, _, err := c.split()
	if err != nil {
		return nil, err
	}
	c.first = newChild(c, ar)
	return c.first, nil
}

// createSecond creates and returns the second sub container of this container.
func (c *Container) createSecond() (*Container, error) {
	_, ar, err := c.split()
	if err != nil {
		return nil, err
	}
	c.second = newChild(c, ar)
	return c.second, nil
}

// Draw draws this container and all of its sub containers.
func (c *Container) Draw() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return drawTree(c)
}

// updateFocus processes the mouse event and determines if it changes the
// focused container.
func (c *Container) updateFocus(m *terminalapi.Mouse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	target := pointCont(c, m.Position)
	if target == nil { // Ignore mouse clicks where no containers are.
		return
	}
	c.focusTracker.mouse(target, m)
}

// keyboardToWidget forwards the keyboard event to the widget unconditionally.
func (c *Container) keyboardToWidget(k *terminalapi.Keyboard, scope widgetapi.KeyScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if scope == widgetapi.KeyScopeFocused && !c.focusTracker.isActive(c) {
		return nil
	}
	return c.opts.widget.Keyboard(k)
}

// mouseToWidget forwards the mouse event to the widget.
func (c *Container) mouseToWidget(m *terminalapi.Mouse, scope widgetapi.MouseScope) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	target := pointCont(c, m.Position)
	if target == nil { // Ignore mouse clicks where no containers are.
		return nil
	}

	// Ignore clicks falling outside of the container.
	if scope != widgetapi.MouseScopeGlobal && !m.Position.In(c.area) {
		return nil
	}

	// Ignore clicks falling outside of the widget's canvas.
	wa, err := c.widgetArea()
	if err != nil {
		return err
	}
	if scope == widgetapi.MouseScopeWidget && !m.Position.In(wa) {
		return nil
	}

	// The sent mouse coordinate is relative to the widget canvas, i.e. zero
	// based, even though the widget might not be in the top left corner on the
	// terminal.
	offset := wa.Min
	var wm *terminalapi.Mouse
	if m.Position.In(wa) {
		wm = &terminalapi.Mouse{
			Position: m.Position.Sub(offset),
			Button:   m.Button,
		}
	} else {
		wm = &terminalapi.Mouse{
			Position: image.Point{-1, -1},
			Button:   m.Button,
		}
	}
	return c.opts.widget.Mouse(wm)
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

	root := rootCont(c)
	// Subscriber the container itself in order to track keyboard focus.
	eds.Subscribe([]terminalapi.Event{&terminalapi.Mouse{}}, func(ev terminalapi.Event) {
		root.updateFocus(ev.(*terminalapi.Mouse))
	}, event.MaxRepetitive(0)) // One event is enough to change the focus.

	// Subscribe any widgets that specify Keyboard or Mouse in their options.
	var errStr string
	preOrder(root, &errStr, visitFunc(func(c *Container) error {
		if c.hasWidget() {
			wOpt := c.opts.widget.Options()
			switch scope := wOpt.WantKeyboard; scope {
			case widgetapi.KeyScopeNone:
				// Widget doesn't want any keyboard events.

			default:
				eds.Subscribe([]terminalapi.Event{&terminalapi.Keyboard{}}, func(ev terminalapi.Event) {
					if err := c.keyboardToWidget(ev.(*terminalapi.Keyboard), scope); err != nil {
						eds.Event(terminalapi.NewErrorf("failed to send global keyboard event %v to widget %T: %v", ev, c.opts.widget, err))
					}
				}, event.MaxRepetitive(maxReps))
			}

			switch scope := wOpt.WantMouse; scope {
			case widgetapi.MouseScopeNone:
				// Widget doesn't want any mouse events.

			default:
				eds.Subscribe([]terminalapi.Event{&terminalapi.Mouse{}}, func(ev terminalapi.Event) {
					if err := c.mouseToWidget(ev.(*terminalapi.Mouse), scope); err != nil {
						eds.Event(terminalapi.NewErrorf("failed to send mouse event %v to widget %T: %v", ev, c.opts.widget, err))
					}
				}, event.MaxRepetitive(maxReps))
			}
		}
		return nil
	}))
}
