// Package sparkline is a widget that draws a graph showing a series of values as vertical bars.
package sparkline

import (
	"errors"
	"fmt"
	"image"
	"sync"

	"github.com/mum4k/termdash/canvas"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/draw"
	"github.com/mum4k/termdash/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// SparkLine draws a graph showing a series of values as vertical bars.
//
// Bars can have sub-cell height. The graphs scale adjusts dynamically based on
// the largest displayed value or has a statically set maximum.
//
// Implements widgetapi.Widget. This object is thread-safe.
type SparkLine struct {
	// data are the data points the SparkLine displays.
	data []int

	// mu protects the SparkLine.
	mu sync.Mutex

	// opts are the provided options.
	opts *options
}

// New returns a new SparkLine.
func New(opts ...Option) *SparkLine {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	return &SparkLine{
		opts: opt,
	}
}

// Draw draws the SparkLine widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (sl *SparkLine) Draw(cvs *canvas.Canvas) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	ar := sl.area(cvs)
	visible, max := visibleMax(sl.data, ar.Dx())
	var curX int
	if len(visible) < ar.Dx() {
		curX = ar.Max.X - len(visible)
	} else {
		curX = ar.Min.X
	}

	for _, v := range visible {
		blocks := toBlocks(v, max, ar.Dy())
		curY := ar.Max.Y - 1
		for i := 0; i < blocks.full; i++ {
			cells, err := cvs.SetCell(
				image.Point{curX, curY},
				sparks[len(sparks)-1],
				cell.FgColor(sl.opts.color),
			)
			if err != nil {
				return err
			}

			if cells != 1 {
				panic(fmt.Sprintf("set an unexpected number of cells %d while filling a full block, expected one", cells))
			}
			curY--
		}

		if blocks.partSpark != 0 {
			cells, err := cvs.SetCell(
				image.Point{curX, curY},
				blocks.partSpark,
				cell.FgColor(sl.opts.color),
			)
			if err != nil {
				return err
			}

			if cells != 1 {
				panic(fmt.Sprintf("set an unexpected number of cells %d while filling a partial block, expected one", cells))
			}
		}

		curX++
	}

	if sl.opts.label != "" {
		lStart := image.Point{ar.Min.X, ar.Min.Y - 1}
		if err := draw.Text(cvs, sl.opts.label, lStart,
			draw.TextCellOpts(sl.opts.labelCellOpts...),
			draw.TextOverrunMode(draw.OverrunModeThreeDot),
		); err != nil {
			return err
		}
	}
	return nil
}

// Add adds data points to the SparkLine.
// At least one data point must be provided. All data points must be positive
// integers.
// The last added data point will be the one displayed all the way on the right
// of the SparkLine.
func (sl *SparkLine) Add(data ...int) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	for i, d := range data {
		if d < 0 {
			return fmt.Errorf("data point[%d]: %v must be a positive integer", i, d)
		}
	}
	sl.data = append(sl.data, data...)
	return nil
}

// Clear removes all the data points in the sparkline, effectively returning to
// an empty graph.
func (sl *SparkLine) Clear() {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.data = nil
}

// Keyboard input isn't supported on the SparkLine widget.
func (*SparkLine) Keyboard(k *terminalapi.Keyboard) error {
	return errors.New("the SparkLine widget doesn't support keyboard events")
}

// Mouse input isn't supported on the SparkLine widget.
func (*SparkLine) Mouse(m *terminalapi.Mouse) error {
	return errors.New("the SparkLine widget doesn't support mouse events")
}

// area returns the area of the canvas available to the SparkLine.
func (sl *SparkLine) area(cvs *canvas.Canvas) image.Rectangle {
	cvsAr := cvs.Area()

	maxY := cvsAr.Max.Y
	var minY int
	if sl.opts.height > 0 {
		minY = maxY - sl.opts.height
	} else {
		minY = cvsAr.Min.Y

		if sl.opts.label != "" {
			minY++ // Reserve one line for the label.
		}
	}
	return image.Rect(
		cvsAr.Min.X,
		minY,
		cvsAr.Max.X,
		maxY,
	)
}

// minSize returns the minimum canvas size for the sparkline based on the options.
func (sl *SparkLine) minSize() image.Point {
	// At least one data point.
	const minWidth = 1
	var minHeight int
	if sl.opts.height > 0 {
		minHeight = sl.opts.height
	} else {
		minHeight = 1 // At least one line of characters.
	}

	if sl.opts.label != "" {
		minHeight++ // One line for the text label.
	}
	return image.Point{minWidth, minHeight}
}

// Options implements widgetapi.Widget.Options.
func (sl *SparkLine) Options() widgetapi.Options {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	min := sl.minSize()
	var max image.Point
	if sl.opts.height > 0 {
		max = min
	}

	return widgetapi.Options{
		MinimumSize:  min,
		MaximumSize:  max,
		WantKeyboard: false,
		WantMouse:    false,
	}
}
