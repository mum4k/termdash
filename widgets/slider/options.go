// Copyright 2026 Google Inc.
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

package slider

// options.go contains configurable options for Slider.

import (
	"fmt"

	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
)

// Option is used to provide options.
type Option interface {
	// set sets the provided option.
	set(*options)
}

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// ChangeFn is called when the user changes the slider value.
//
// The callback must be thread-safe because it is triggered from the keyboard
// and mouse event handling paths, which run in separate goroutines.
type ChangeFn func(value int) error

// Direction indicates the axis the slider is drawn on.
type Direction int

// String implements fmt.Stringer.
func (o Direction) String() string {
	if n, ok := orientationNames[o]; ok {
		return n
	}
	return "OrientationUnknown"
}

var orientationNames = map[Direction]string{
	OrientationHorizontal: "OrientationHorizontal",
	OrientationVertical:   "OrientationVertical",
}

const (
	// OrientationHorizontal draws the slider left-to-right.
	OrientationHorizontal Direction = iota
	// OrientationVertical draws the slider bottom-to-top.
	OrientationVertical
)

// Preset is a named visual preset for the slider track.
type Preset int

// String implements fmt.Stringer.
func (s Preset) String() string {
	if n, ok := styleNames[s]; ok {
		return n
	}
	return "StyleUnknown"
}

var styleNames = map[Preset]string{
	StyleBar:              "StyleBar",
	StyleSegmented:        "StyleSegmented",
	StyleSegmentedBlocks:  "StyleSegmentedBlocks",
	StyleDots:             "StyleDots",
	StyleSegmentedDots:    "StyleSegmentedDots",
	StyleSquares:          "StyleSquares",
	StyleSegmentedSquares: "StyleSegmentedSquares",
	StyleStars:            "StyleStars",
}

const (
	// StyleBar is the default solid filled bar with a knob.
	StyleBar Preset = iota
	// StyleSegmented draws a thin segmented line.
	StyleSegmented
	// StyleSegmentedBlocks draws dense rectangular block segments.
	StyleSegmentedBlocks
	// StyleDots draws filled and empty circular dots.
	StyleDots
	// StyleSegmentedDots draws smaller dot segments.
	StyleSegmentedDots
	// StyleSquares draws filled and empty square blocks.
	StyleSquares
	// StyleSegmentedSquares draws smaller square segments.
	StyleSegmentedSquares
	// StyleStars draws filled and empty star segments.
	StyleStars
)

// options holds the provided options.
type options struct {
	min            int
	max            int
	value          int
	step           int
	width          int
	orientation    Direction
	hAlign         align.Horizontal
	vAlign         align.Vertical
	style          Preset
	fillRune       rune
	trackRune      rune
	knobRune       rune
	fillCellOpts   []cell.Option
	trackCellOpts  []cell.Option
	knobCellOpts   []cell.Option
	focusedKnobOps []cell.Option
	onChange       ChangeFn
}

// Default colors used by the slider widget.
var (
	DefaultFillColor        = cell.ColorNumber(221)
	DefaultTrackColor       = cell.ColorNumber(239)
	DefaultKnobColor        = cell.ColorWhite
	DefaultFocusedKnobColor = cell.ColorCyan
)

const (
	// DefaultWidth is the default slider length in terminal cells.
	DefaultWidth = 18
	// DefaultStep is the default increment used for keyboard changes.
	DefaultStep = 1
)

// validate validates the provided options.
func (o *options) validate() error {
	if o.max < o.min {
		return fmt.Errorf("invalid range: max %d must be >= min %d", o.max, o.min)
	}
	if o.step <= 0 {
		return fmt.Errorf("invalid step %d, want step > 0", o.step)
	}
	if o.width <= 0 {
		return fmt.Errorf("invalid width %d, want width > 0", o.width)
	}
	if _, ok := orientationNames[o.orientation]; !ok {
		return fmt.Errorf("invalid orientation %v", o.orientation)
	}
	if _, ok := styleNames[o.style]; !ok {
		return fmt.Errorf("invalid style %v", o.style)
	}
	return nil
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		min:            0,
		max:            100,
		value:          0,
		step:           DefaultStep,
		width:          DefaultWidth,
		orientation:    OrientationHorizontal,
		hAlign:         align.HorizontalLeft,
		vAlign:         align.VerticalTop,
		style:          StyleBar,
		fillRune:       '█',
		trackRune:      '░',
		knobRune:       '●',
		fillCellOpts:   []cell.Option{cell.FgColor(DefaultFillColor)},
		trackCellOpts:  []cell.Option{cell.FgColor(DefaultTrackColor)},
		knobCellOpts:   []cell.Option{cell.FgColor(DefaultKnobColor)},
		focusedKnobOps: []cell.Option{cell.FgColor(DefaultFocusedKnobColor)},
	}
}

// Min sets the minimum slider value.
func Min(v int) Option {
	return option(func(opts *options) {
		opts.min = v
	})
}

// Max sets the maximum slider value.
func Max(v int) Option {
	return option(func(opts *options) {
		opts.max = v
	})
}

// Value sets the initial slider value.
func Value(v int) Option {
	return option(func(opts *options) {
		opts.value = v
	})
}

// Step sets the increment used for keyboard changes.
func Step(v int) Option {
	return option(func(opts *options) {
		opts.step = v
	})
}

// Length sets the number of terminal cells used along the slider axis.
func Length(cells int) Option {
	return option(func(opts *options) {
		opts.width = cells
	})
}

// Width sets the slider length in terminal cells.
//
// Width is kept for compatibility with existing horizontal sliders. For
// vertical sliders, Height or Length is usually clearer.
func Width(cells int) Option {
	return Length(cells)
}

// Height sets the slider length in terminal cells.
//
// This is an alias for Length intended for vertical sliders.
func Height(cells int) Option {
	return Length(cells)
}

// Orientation sets whether the slider is drawn horizontally or vertically.
func Orientation(o Direction) Option {
	return option(func(opts *options) {
		opts.orientation = o
	})
}

// AlignHorizontal positions the slider horizontally inside a larger canvas.
func AlignHorizontal(h align.Horizontal) Option {
	return option(func(opts *options) {
		opts.hAlign = h
	})
}

// AlignVertical positions the slider vertically inside a larger canvas.
func AlignVertical(v align.Vertical) Option {
	return option(func(opts *options) {
		opts.vAlign = v
	})
}

// Style applies one of the built-in visual slider styles.
func Style(s Preset) Option {
	return option(func(opts *options) {
		opts.style = s
		applyStyle(opts, s)
	})
}

// BarStyle applies the default solid bar style.
func BarStyle() Option { return Style(StyleBar) }

// SegmentedStyle applies the thin segmented-line style.
func SegmentedStyle() Option { return Style(StyleSegmented) }

// SegmentedBlocksStyle applies the dense rectangular block style.
func SegmentedBlocksStyle() Option { return Style(StyleSegmentedBlocks) }

// DotsStyle applies the filled and empty circular dot style.
func DotsStyle() Option { return Style(StyleDots) }

// SegmentedDotsStyle applies the smaller dot segment style.
func SegmentedDotsStyle() Option { return Style(StyleSegmentedDots) }

// SquaresStyle applies the filled and empty square style.
func SquaresStyle() Option { return Style(StyleSquares) }

// SegmentedSquaresStyle applies the smaller square segment style.
func SegmentedSquaresStyle() Option { return Style(StyleSegmentedSquares) }

// StarsStyle applies the filled and empty star style.
func StarsStyle() Option { return Style(StyleStars) }

// FillRune sets the rune used for the filled portion of the slider.
func FillRune(r rune) Option {
	return option(func(opts *options) {
		opts.fillRune = r
	})
}

// TrackRune sets the rune used for the unfilled portion of the slider.
func TrackRune(r rune) Option {
	return option(func(opts *options) {
		opts.trackRune = r
	})
}

// KnobRune sets the rune used for the slider knob.
func KnobRune(r rune) Option {
	return option(func(opts *options) {
		opts.knobRune = r
	})
}

// FillCellOpts sets the styling used for the filled portion of the slider.
func FillCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.fillCellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// TrackCellOpts sets the styling used for the unfilled portion of the slider.
func TrackCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.trackCellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// KnobCellOpts sets the styling used for the slider knob.
func KnobCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.knobCellOpts = append([]cell.Option(nil), cellOpts...)
	})
}

// FocusedKnobCellOpts sets the styling used for the slider knob while the
// widget is focused.
func FocusedKnobCellOpts(cellOpts ...cell.Option) Option {
	return option(func(opts *options) {
		opts.focusedKnobOps = append([]cell.Option(nil), cellOpts...)
	})
}

// OnChange sets the slider's value-change hook.
//
// This is the widget's canonical callback surface. Callers that need delayed
// or asynchronous work should build that from this hook so the widget keeps a
// single stable event path.
func OnChange(fn ChangeFn) Option {
	return option(func(opts *options) {
		opts.onChange = fn
	})
}

func applyStyle(opts *options, style Preset) {
	switch style {
	case StyleBar:
		opts.fillRune = '█'
		opts.trackRune = '░'
		opts.knobRune = '●'
	case StyleSegmented:
		opts.fillRune = '─'
		opts.trackRune = '╌'
		opts.knobRune = '●'
	case StyleSegmentedBlocks:
		opts.fillRune = '▌'
		opts.trackRune = '┆'
		opts.knobRune = '█'
	case StyleDots:
		opts.fillRune = '●'
		opts.trackRune = '○'
		opts.knobRune = '●'
	case StyleSegmentedDots:
		opts.fillRune = '•'
		opts.trackRune = '∘'
		opts.knobRune = '•'
	case StyleSquares:
		opts.fillRune = '■'
		opts.trackRune = '□'
		opts.knobRune = '■'
	case StyleSegmentedSquares:
		opts.fillRune = '▪'
		opts.trackRune = '▫'
		opts.knobRune = '▪'
	case StyleStars:
		opts.fillRune = '★'
		opts.trackRune = '☆'
		opts.knobRune = '★'
	}
}
