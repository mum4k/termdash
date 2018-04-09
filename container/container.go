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

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminalapi"
)

// Container wraps either sub containers or widgets and positions them on the
// terminal.
// This is thread-safe and must not be copied.
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

	// mu locks the container while it is being drawn or while it receives input events.
	mu sync.Mutex

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
	} else {
		return c.area
	}
}

// split splits the container's usable area into child areas.
// Panics if the container isn't configured for a split.
func (c *Container) split() (image.Rectangle, image.Rectangle) {
	if ar := c.usable(); c.opts.split == splitTypeVertical {
		return area.VSplit(ar)
	} else {
		return area.HSplit(ar)
	}
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
	c.mu.Lock()
	defer c.mu.Unlock()

	return drawTree(c)
}

// Keyboard is used to forward a keyboard event to the container.
// Keyboard events are forwarded to the widget in the currently focused
// container, assuming that the widget registered for keyboard events.
func (c *Container) Keyboard(k *terminalapi.Keyboard) error {
	c.mu.Lock()
	defer c.mu.Unlock()

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
	c.mu.Lock()
	defer c.mu.Unlock()

	c.focusTracker.mouse(m)

	target := pointCont(c, m.Position)
	w := target.opts.widget
	if w == nil || !w.Options().WantMouse {
		return nil
	}

	if !m.Position.In(target.usable()) {
		return nil
	}
	// The sent mouse coordinate is relative to the widget canvas, i.e. zero
	// based, even though the widget might not be in the top left corner on the
	// terminal.
	offset := target.usable().Min
	wm := &terminalapi.Mouse{
		Position: m.Position.Sub(offset),
		Button:   m.Button,
	}
	return w.Mouse(wm)
}
