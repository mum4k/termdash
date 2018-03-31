/*
Package container defines a type that wraps other containers or widgets.

The container supports splitting container into sub containers, defining
container styles and placing widgets.  The container also creates and manages
canvases assigned to the placed widgets.
*/
package container

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminalapi"
)

// Container wraps either sub containers or widgets and positions them on the
// terminal.
type Container struct {
	// parent is the parent container, nil if this is the root container.
	parent *Container
	// The sub containers, if these aren't nil, the widget must be.
	first  *Container
	second *Container

	// term is the terminal this container is placed on.
	// All containers in the tree share the same terminal.
	term terminalapi.Terminal

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
	o := &options{}
	for _, opt := range opts {
		opt.set(o)
	}

	size := t.Size()
	return &Container{
		term: t,
		// The root container has access to the entire terminal.
		area: image.Rect(0, 0, size.X, size.Y),
		opts: o,
	}
}

// newChild creates a new child container of the given parent.
func newChild(parent *Container, area image.Rectangle, opts ...Option) *Container {
	o := &options{}
	for _, opt := range opts {
		opt.set(o)
	}

	return &Container{
		parent: parent,
		term:   parent.term,
		area:   area,
		opts:   o,
	}
}

// Returns the parent container of this container and applies the provided
// options to the parent container. Returns nil if this container is the root
// of the container tree.
func (c *Container) Parent(opts ...Option) *Container {
	if c.parent == nil {
		return nil
	}

	p := c.parent
	for _, opt := range opts {
		opt.set(p.opts)
	}
	return p
}

// hasBorder determines if this container has a border.
func (c *Container) hasBorder() bool {
	return c.opts.border != draw.LineStyleNone
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
func (c *Container) split() (image.Rectangle, image.Rectangle) {
	if ar := c.usable(); c.opts.split == splitTypeHorizontal {
		return area.HSplit(ar)
	} else {
		return area.VSplit(ar)
	}
}

// First returns the first sub container of this container.
// This is the left sub container when using SplitVertical() or the top sub
// container when using SplitHorizontal().
// If this container doesn't have the first sub container yet, it will be
// created. Applies the provided options to the first sub container.
// Returns nil if this container contains a widget, containers with widgets
// cannot have sub containers.
func (c *Container) First(opts ...Option) *Container {
	if c.opts.widget != nil {
		return nil
	}

	if child := c.first; child != nil {
		for _, opt := range opts {
			opt.set(child.opts)
		}
		return child
	}

	ar, _ := c.split()
	c.first = newChild(c, ar, opts...)
	return c.first
}

// Second returns the second sub container of this container.
// This is the left sub container when using SplitVertical() or the top sub
// container when using SplitHorizontal().
// If this container doesn't have the second sub container yet, it will be
// created. Applies the provided options to the second sub container.
// Returns nil if this container contains a widget, containers with widgets
// cannot have sub containers.
func (c *Container) Second(opts ...Option) *Container {
	if c.opts.widget != nil {
		return nil
	}

	if child := c.second; child != nil {
		for _, opt := range opts {
			opt.set(child.opts)
		}
		return child
	}

	_, ar := c.split()
	c.second = newChild(c, ar, opts...)
	return c.second
}

// Root returns the root container and applies the provided options to the root
// container.
func (c *Container) Root(opts ...Option) *Container {
	for p := c.Parent(); p != nil; p = c.Parent() {
		c = p
	}

	for _, opt := range opts {
		opt.set(c.opts)
	}
	return c
}

// draw draws this container and its widget.
// TODO(mum4k): Draw the widget.
func (c *Container) draw() error {
	// TODO(mum4k): Should be verified against the min size reported by the
	// widget.
	if us := c.usable(); us.Dx() < 1 || us.Dy() < 1 {
		return nil
	}

	cvs, err := canvas.New(c.area)
	if err != nil {
		return err
	}

	if c.hasBorder() {
		ar, err := area.FromSize(cvs.Size())
		if err != nil {
			return err
		}
		if err := draw.Box(cvs, ar, c.opts.border); err != nil {
			return err
		}
	}
	return cvs.Apply(c.term)
}

// Draw draws this container and all of its sub containers.
func (c *Container) Draw() error {
	// TODO(mum4k): Handle resize or split to area too small.
	// TODO(mum4k): Propagate error.
	// TODO(mum4k): Don't require .Root() at the end.
	drawTree(c)
	return nil
}

// drawTree implements pre-order BST walk through the containers and draws each
// visited container.
func drawTree(c *Container) {
	if c == nil {
		return
	}
	if err := c.draw(); err != nil {
		panic(err)
	}
	drawTree(c.first)
	drawTree(c.second)
}
