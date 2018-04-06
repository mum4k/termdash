package container

// draw.go contains logic to draw containers and the contained widgets.

import (
	"errors"

	"github.com/mum4k/termdash/area"
	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
)

// drawTree draws this container and all of its sub containers.
func drawTree(c *Container) error {
	var errStr string
	preOrder(c, &errStr, visitFunc(func(c *Container) error {
		return drawCont(c)
	}))
	if errStr != "" {
		return errors.New(errStr)
	}
	return nil
}

// drawCont draws the container and its widget.
// TODO(mum4k): Draw the widget.
func drawCont(c *Container) error {
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

		var opts []cell.Option
		if c.focusTracker.isActive(c) {
			opts = append(opts, cell.FgColor(c.opts.inherited.focusedColor))
		} else {
			opts = append(opts, cell.FgColor(c.opts.inherited.borderColor))
		}
		if err := draw.Box(cvs, ar, c.opts.border, opts...); err != nil {
			return err
		}
	}
	return cvs.Apply(c.term)
}
