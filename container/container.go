/*
Package container defines a type that wraps other containers or widgets.

The container supports splitting container into sub containers, defining
container styles and placing widgets.  The container also creates and manages
canvases assigned to the placed widgets.
*/
package container

import (
	"errors"
	"image"

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

// Returns the parent container of this container.
// Returns nil if this container is the root of the container tree.
func (c *Container) Parent(opts ...Option) *Container {
	if c == nil || c.parent == nil {
		return nil
	}

	p := c.parent
	for _, opt := range opts {
		opt.set(p.opts)
	}
	return p
}

// First returns the first sub container of this container.
// This is the left sub container when using SplitVertical() or the top sub
// container when using SplitHorizontal().
// If this container doesn't have the first sub container yet, it will be
// created. Applies the provided options to the first sub container.
// Returns nil if this container contains a widget, containers with widgets
// cannot have sub containers.
func (c *Container) First(opts ...Option) *Container {
	if c == nil || c.opts.widget != nil {
		return nil
	}

	if child := c.first; child != nil {
		for _, opt := range opts {
			opt.set(child.opts)
		}
		return child
	}

	c.first = New(c.term, opts...)
	c.first.parent = c
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
	if c == nil || c.opts.widget != nil {
		return nil
	}

	if child := c.second; child != nil {
		for _, opt := range opts {
			opt.set(child.opts)
		}
		return child
	}

	c.second = New(c.term, opts...)
	c.second.parent = c
	return c.second
}

// Draw requests all widgets in this and all sub containers to draw on their
// respective canvases.
func (c *Container) Draw() error {
	return errors.New("unimplemented")
}
