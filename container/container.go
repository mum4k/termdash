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
	size := t.Size()
	root := &Container{
		term: t,
		// The root container has access to the entire terminal.
		area: image.Rect(0, 0, size.X, size.Y),
		opts: &options{},
	}
	applyOptions(root, opts...)
	return root
}

// newChild creates a new child container of the given parent.
func newChild(parent *Container, area image.Rectangle) *Container {
	return &Container{
		parent: parent,
		term:   parent.term,
		area:   area,
		opts:   &options{},
	}
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
	var errStr string
	drawTree(c, &errStr)
	if errStr != "" {
		return errors.New(errStr)
	}
	return nil
}

// drawTree implements pre-order BST walk through the containers and draws each
// visited container.
func drawTree(c *Container, errStr *string) {
	if c == nil || *errStr != "" {
		return
	}
	if err := c.draw(); err != nil {
		*errStr = err.Error()
		return
	}
	drawTree(c.first, errStr)
	drawTree(c.second, errStr)
}
