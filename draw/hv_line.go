package draw

// hv_line.go contains code that draws horizontal and vertical lines.

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
)

// HVLineOption is used to provide options to HVLine().
type HVLineOption interface {
	// set sets the provided option.
	set(*hVLineOptions)
}

// hVLineOptions stores the provided options.
type hVLineOptions struct {
	cellOpts  []cell.Option
	lineStyle LineStyle
}

// hVLineOption implements HVLineOption.
type hVLineOption func(*hVLineOptions)

// set implements HVLineOption.set.
func (o hVLineOption) set(opts *hVLineOptions) {
	o(opts)
}

// DefaultHVLineStyle is the default value for the HVLineStyle option.
const DefaultHVLineStyle = LineStyleLight

// HVLineStyle sets the style of the line.
// Defaults to DefaultHVLineStyle.
func HVLineStyle(ls LineStyle) HVLineOption {
	return hVLineOption(func(opts *hVLineOptions) {
		opts.lineStyle = ls
	})
}

// HVLineCellOpts sets options on the cells that contain the line.
// Where two lines cross, the cell representing the crossing point inherits
// options set on the line that was drawn last.
func HVLineCellOpts(cOpts ...cell.Option) HVLineOption {
	return hVLineOption(func(opts *hVLineOptions) {
		opts.cellOpts = cOpts
	})
}

// HVLine represents one horizontal or vertical line.
type HVLine struct {
	start image.Point
	end   image.Point
}

// HVLines draws horizontal or vertical lines. Handles drawing of the correct
// characters for locations where any two lines cross (e.g. a corner, a T shape
// a cross). Each line must be at least one cell long. Both start and end
// must be on the same horizontal (same X coordinate) or same vertical (same Y
// coordinate) line.
func HVLines(c *canvas.Canvas, lines []HVLine, opts ...HVLineOption) error {
	for _, l := range lines {
		line, err := newHVLine(c, l.start, l.end, opts...)
		if err != nil {
			return err
		}

		switch {
		case line.horizontal():
			for curX := line.start.X; ; curX++ {
				cur := image.Point{curX, line.start.Y}
				if _, err := c.SetCell(cur, line.mainPart, line.opts.cellOpts...); err != nil {
					return err
				}

				if curX == line.end.X {
					break
				}
			}

		case line.vertical():
			for curY := line.start.Y; ; curY++ {
				cur := image.Point{line.start.X, curY}
				if _, err := c.SetCell(cur, line.mainPart, line.opts.cellOpts...); err != nil {
					return err
				}

				if curY == line.end.Y {
					break
				}
			}
		}
	}
	return nil
}

// hVLine represents a line that will be drawn on the canvas.
type hVLine struct {
	// start is the starting point of the line.
	start image.Point

	// end is the ending point of the line.
	end image.Point

	// parts are characters that represent parts of the line of the style
	// chosen in the options.
	parts map[linePart]rune

	// mainPart is either parts[vLine] or parts[hLine] depending on whether
	// this is horizontal or vertical line.
	mainPart rune

	// opts are the options provided in a call to HVLine().
	opts *hVLineOptions
}

// newHVLine creates a new hVLine instance.
// Swaps start and end iof necessary, so that horizontal drawing is always left
// to right and vertical is always top down.
func newHVLine(c *canvas.Canvas, start, end image.Point, opts ...HVLineOption) (*hVLine, error) {
	if ar := c.Area(); !start.In(ar) || !end.In(ar) {
		return nil, fmt.Errorf("both the start%v and the end%v must be in the canvas area: %v", start, end, ar)
	}

	opt := &hVLineOptions{
		lineStyle: DefaultHVLineStyle,
	}
	for _, o := range opts {
		o.set(opt)
	}

	parts, err := lineParts(opt.lineStyle)
	if err != nil {
		return nil, err
	}

	var mainPart rune
	switch {
	case start.X != end.X && start.Y != end.Y:
		return nil, fmt.Errorf("can only draw horizontal (same X coordinates) or vertical (same Y coordinates), got start:%v end:%v", start, end)

	case start.X == end.X && start.Y == end.Y:
		return nil, fmt.Errorf("the line must at least one cell long, got start%v, end%v", start, end)

	case start.X == end.X:
		mainPart = parts[vLine]
		if start.Y > end.Y {
			start, end = end, start
		}

	case start.Y == end.Y:
		mainPart = parts[hLine]
		if start.X > end.X {
			start, end = end, start
		}

	}

	return &hVLine{
		start:    start,
		end:      end,
		parts:    parts,
		mainPart: mainPart,
		opts:     opt,
	}, nil
}

// horizontal determines if this is a horizontal line.
func (hvl *hVLine) horizontal() bool {
	return hvl.mainPart == hvl.parts[hLine]
}

// vertical determines if this is a vertical line.
func (hvl *hVLine) vertical() bool {
	return hvl.mainPart == hvl.parts[vLine]
}
