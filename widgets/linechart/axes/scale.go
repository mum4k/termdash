package axes

// scale.go calculates the scale of the Y axis.

import (
	"fmt"
	"math"

	"github.com/mum4k/termdash/canvas/braille"
)

// YScale is the scale of the Y axis.
type YScale struct {
	// Min is the minimum value on the axis.
	Min *Value
	// Max is the maximum value on the axis.
	Max *Value
	// Step is the step in the value between pixels.
	Step *Value

	// cvsHeight is the height of the canvas the scale was calculated for.
	cvsHeight int
	// brailleHeight is the height of the braille canvas based on the cvsHeight.
	brailleHeight int
}

// NewYScale calculates the scale of the Y axis, given the boundary values and
// the height of the canvas. The nonZeroDecimals dictates rounding of the
// calculates scale, see NewValue for details.
func NewYScale(min, max float64, cvsHeight, nonZeroDecimals int) *YScale {
	brailleHeight := cvsHeight * braille.RowMult
	usablePixels := brailleHeight - 1 // One pixel reserved for value zero.

	diff := max - min
	step := NewValue(diff/float64(usablePixels), nonZeroDecimals)
	return &YScale{
		Min:           NewValue(min, nonZeroDecimals),
		Max:           NewValue(max, nonZeroDecimals),
		Step:          step,
		cvsHeight:     cvsHeight,
		brailleHeight: brailleHeight,
	}
}

// PixelToValue given a Y coordinate of the pixel, returns its value according
// to the scale. The coordinate must be within bounds of the canvas height
// provided to NewYScale.
func (ys *YScale) PixelToValue(p int) (float64, error) {
	if min, max := 0, ys.brailleHeight; p < min || p >= max {
		return 0, fmt.Errorf("invalid pixel %d, must be in range %d <= p < %d", p, min, max)
	}

	switch {
	case p == 0:
		return ys.Min.Rounded, nil
	case p == ys.brailleHeight-1:
		return ys.Max.Rounded, nil
	default:
		v := float64(p) * ys.Step.Rounded
		if ys.Min.Value < 0 {
			diff := -1 * ys.Min.Rounded
			v -= diff
		}
		return v, nil
	}
}

// ValueToPixel given a value, determines the Y coordinate of the pixel that
// most closely represents the value on the line chart according to the scale.
// The value must be within the bounds provided to NewYScale.
func (ys *YScale) ValueToPixel(v float64) (int, error) {
	if min, max := ys.Min.Value, ys.Max.Rounded; v < min || v > max {
		return 0, fmt.Errorf("invalid value %v, must be in range %v <= v <= %v", v, min, max)
	}

	if ys.Min.Value < 0 {
		diff := -1 * ys.Min.Rounded
		v += diff
	}
	return int(math.Round(v / ys.Step.Rounded)), nil
}

// CellLabel given a cell position on the canvas, determines value of the label
// that should be next to it. The cell position must be within the cvsHeight
// provided to NewYScale.
func (ys *YScale) CellLabel(cell int) (*Value, error) {
	if min, max := 0, ys.cvsHeight; cell < min || cell >= max {
		return nil, fmt.Errorf("invalid cell %d, must be in range %d <= p < %d", cell, min, max)
	}

	pixel := cell * braille.RowMult
	v, err := ys.PixelToValue(pixel)
	if err != nil {
		return nil, err
	}
	return NewValue(v, ys.Min.NonZeroDecimals), nil
}
