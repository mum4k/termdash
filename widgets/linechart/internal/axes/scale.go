// Copyright 2019 Google Inc.
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

package axes

// scale.go calculates the scale of the Y axis.

import (
	"fmt"
	"math"

	"github.com/mum4k/termdash/internal/canvas/braille"
)

// YScaleMode determines whether the Y scale is anchored to the zero value.
type YScaleMode int

// String implements fmt.Stringer()
func (ysm YScaleMode) String() string {
	if n, ok := yScaleModeNames[ysm]; ok {
		return n
	}
	return "YScaleModeUnknown"
}

// yScaleModeNames maps YScaleMode values to human readable names.
var yScaleModeNames = map[YScaleMode]string{
	YScaleModeAnchored: "YScaleModeAnchored",
	YScaleModeAdaptive: "YScaleModeAdaptive",
}

const (
	// YScaleModeAnchored is a mode in which the Y scale always starts at value
	// zero regardless of the min and max on the series.
	YScaleModeAnchored YScaleMode = iota

	// YScaleModeAdaptive is a mode where the Y scale adapts its base value
	// according to the min and max on the series.
	// I.e. it starts at min for all-positive series and at max for
	// all-negative series.
	YScaleModeAdaptive
)

// YScale is the scale of the Y axis.
type YScale struct {
	// Min is the minimum value on the axis.
	Min *Value
	// Max is the maximum value on the axis.
	Max *Value
	// Step is the step in the value between pixels.
	Step *Value

	// GraphHeight is the height in cells of the area on the canvas that is
	// dedicated to the graph itself.
	GraphHeight int
	// brailleHeight is the height of the braille canvas based on the GraphHeight.
	brailleHeight int
}

// String implements fmt.Stringer.
func (ys *YScale) String() string {
	return fmt.Sprintf("YScale{Min:%v, Max:%v, Step:%v, GraphHeight:%v}", ys.Min, ys.Max, ys.Step, ys.GraphHeight)
}

// NewYScale calculates the scale of the Y axis, given the boundary values and
// the height of the graph. The nonZeroDecimals dictates rounding of the
// calculated scale, see NewValue for details.
// Max must be greater or equal to min. The graphHeight must be a positive
// number.
func NewYScale(min, max float64, graphHeight, nonZeroDecimals int, mode YScaleMode) (*YScale, error) {
	if max < min {
		return nil, fmt.Errorf("max(%v) cannot be less than min(%v)", max, min)
	}
	if min := 1; graphHeight < min {
		return nil, fmt.Errorf("graphHeight cannot be less than %d, got %d", min, graphHeight)
	}

	brailleHeight := graphHeight * braille.RowMult
	usablePixels := brailleHeight - 1 // One pixel reserved for value zero.

	switch mode {
	case YScaleModeAnchored:
		// Anchor the axis at the zero value.
		if min > 0 {
			min = 0
		}
		if max < 0 {
			max = 0
		}

	case YScaleModeAdaptive:
		// Even in this mode, we still anchor the axis at the zero if all the
		// data points are equal, so we can still draw something.
		if min > 0 && min == max {
			min = 0
		}
		if max < 0 && min == max {
			max = 0
		}
	default:
		return nil, fmt.Errorf("unsupported mode: %v(%d)", mode, mode)
	}
	diff := max - min
	step := NewValue(diff/float64(usablePixels), nonZeroDecimals)
	return &YScale{
		Min:           NewValue(min, nonZeroDecimals),
		Max:           NewValue(max, nonZeroDecimals),
		Step:          step,
		GraphHeight:   graphHeight,
		brailleHeight: brailleHeight,
	}, nil
}

// PixelToValue given a Y coordinate of the pixel, returns its value according
// to the scale. The coordinate must be within bounds of the graph height
// provided to NewYScale. Y coordinates grow down.
func (ys *YScale) PixelToValue(y int) (float64, error) {
	pos, err := yToPosition(y, ys.brailleHeight)
	if err != nil {
		return 0, err
	}

	switch {
	case pos == 0:
		return ys.Min.Rounded, nil
	case pos == ys.brailleHeight-1:
		return ys.Max.Rounded, nil
	default:

		v := float64(pos) * ys.Step.Rounded
		if ys.Min.Value > 0 {
			v += ys.Min.Value
		}
		if ys.Min.Value < 0 {
			diff := -1 * ys.Min.Value
			v -= diff
		}
		return v, nil
	}
}

// ValueToPixel given a value, determines the Y coordinate of the pixel that
// most closely represents the value on the line chart according to the scale.
// The value must be within the bounds provided to NewYScale. Y coordinates
// grow down.
func (ys *YScale) ValueToPixel(v float64) (int, error) {
	if ys.Step.Rounded == 0 {
		return 0, nil
	}

	if ys.Min.Value > 0 {
		v -= ys.Min.Value
	}
	if ys.Min.Value < 0 {
		diff := -1 * ys.Min.Value
		v += diff
	}
	pos := int(math.Round(v / ys.Step.Rounded))
	return positionToY(pos, ys.brailleHeight)
}

// CellLabel given a Y coordinate of a cell on the canvas, determines value of
// the label that should be next to it. The Y coordinate must be within the
// graphHeight provided to NewYScale. Y coordinates grow down.
func (ys *YScale) CellLabel(y int) (*Value, error) {
	pos, err := yToPosition(y, ys.GraphHeight)
	if err != nil {
		return nil, err
	}

	pixelY, err := positionToY(pos*braille.RowMult, ys.brailleHeight)
	if err != nil {
		return nil, err
	}

	v, err := ys.PixelToValue(pixelY)
	if err != nil {
		return nil, err
	}
	return NewValue(v, ys.Min.NonZeroDecimals), nil
}

// XScale is the scale of the X axis.
type XScale struct {
	// Min is the minimum value on the axis.
	Min *Value
	// Max is the maximum value on the axis.
	Max *Value
	// Step is the step in the value between pixels.
	Step *Value

	// GraphWidth is the width in cells of the area on the canvas that is
	// dedicated to the graph.
	GraphWidth int
	// brailleWidth is the width of the braille canvas based on the GraphWidth.
	brailleWidth int
}

// NewXScale calculates the scale of the X axis, given the boundary values and
// the width on the canvas that is available to the X axis.
// The nonZeroDecimals dictates rounding of the calculated scale, see
// NewValue for details.
// The boundary values must be positive or zero and must be min <= max.
// The graphWidth must be a positive number.
func NewXScale(min, max int, graphWidth, nonZeroDecimals int) (*XScale, error) {
	if min < 0 || max < 0 {
		return nil, fmt.Errorf("invalid min:%d or max:%d, the values must not be negative", min, max)
	}
	if min > max {
		return nil, fmt.Errorf("invalid min:%d, max:%d, must be min <= max", min, max)
	}
	if min := 1; graphWidth < min {
		return nil, fmt.Errorf("graphWidth must be at least %d, got %d", min, graphWidth)
	}

	brailleWidth := graphWidth * braille.ColMult
	usablePixels := brailleWidth - 1 // One pixel reserved for value zero.

	minVal := float64(min)
	maxVal := float64(max)
	diff := maxVal - minVal
	step := NewValue(diff/float64(usablePixels), nonZeroDecimals)
	return &XScale{
		Min:          NewValue(minVal, nonZeroDecimals),
		Max:          NewValue(maxVal, nonZeroDecimals),
		Step:         step,
		GraphWidth:   graphWidth,
		brailleWidth: brailleWidth,
	}, nil
}

// String implements fmt.Stringer.
func (xs *XScale) String() string {
	return fmt.Sprintf("XScale{Min:%v, Max:%v, Step:%v, GraphWidth:%v}", xs.Min, xs.Max, xs.Step, xs.GraphWidth)
}

// PixelToValue given a X coordinate of the pixel, returns its value according
// to the scale. The coordinate must be within bounds of the canvas width
// provided to NewXScale. X coordinates grow right.
func (xs *XScale) PixelToValue(x int) (float64, error) {
	if min, max := 0, xs.brailleWidth; x < min || x >= max {
		return 0, fmt.Errorf("invalid x coordinate %d, must be in range %v < x < %v", x, min, max)
	}

	switch {
	case x == 0:
		return xs.Min.Rounded, nil
	case x == xs.brailleWidth-1:
		return xs.Max.Rounded, nil
	default:
		v := float64(x) * xs.Step.Rounded
		if xs.Min.Value > 0 {
			v += xs.Min.Value
		}
		return v, nil
	}
}

// ValueToPixel given a value, determines the X coordinate of the pixel that
// most closely represents the value on the line chart according to the scale.
// The value must be within the bounds provided to NewXScale. X coordinates
// grow right.
func (xs *XScale) ValueToPixel(v int) (int, error) {
	fv := float64(v)
	if min, max := xs.Min.Value, xs.Max.Rounded; fv < min || fv > max {
		return 0, fmt.Errorf("invalid value %v, must be in range %v <= v <= %v", v, min, max)
	}
	if xs.Step.Rounded == 0 {
		return 0, nil
	}
	if xs.Min.Value > 0 {
		fv -= xs.Min.Value
	}
	return int(math.Round(fv / xs.Step.Rounded)), nil
}

// ValueToCell given a value, determines the X coordinate of the cell that
// most closely represents the value on the line chart according to the scale.
// The value must be within the bounds provided to NewXScale. X coordinates
// grow right.
func (xs *XScale) ValueToCell(v int) (int, error) {
	p, err := xs.ValueToPixel(v)
	if err != nil {
		return 0, err
	}
	return p / braille.ColMult, nil
}

// CellLabel given an X coordinate of a cell on the canvas, determines value of the
// label that should be next to it. The X coordinate must be within the
// graphWidth provided to NewXScale. X coordinates grow right.
// The returned value is rounded to the nearest int, rounding half away from zero.
func (xs *XScale) CellLabel(x int) (*Value, error) {
	v, err := xs.PixelToValue(x * braille.ColMult)
	if err != nil {
		return nil, err
	}
	return NewValue(math.Round(v), xs.Min.NonZeroDecimals), nil
}

// positionToY, given a position within the height, returns the Y coordinate of
// the position. Positions grow up, coordinates grow down.
//
// Positions     Y Coordinates
//         2  |  0
//         1  |  1
//         0  |  2
func positionToY(pos int, height int) (int, error) {
	max := height - 1
	if min := 0; pos < min || pos > max {
		return 0, fmt.Errorf("position %d out of bounds %d <= pos <= %d", pos, min, max)
	}
	return max - pos, nil
}

// yToPosition is the reverse of positionToY.
func yToPosition(y int, height int) (int, error) {
	max := height - 1
	if min := 0; y < min || y > max {
		return 0, fmt.Errorf("Y coordinate %d out of bounds %d <= Y <= %d", y, min, max)
	}
	return -1*y + max, nil
}
