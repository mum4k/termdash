package axes

// value.go contains code dealing with values on the line chart.

import (
	"fmt"
	"math"

	"github.com/mum4k/termdash/widgets/linechart/numbers"
)

// Value represents one value.
type Value struct {
	// Value is the original unmodified value.
	Value float64
	// Rounded is the value rounded up to the nonZeroPlaces number of non-zero
	// decimal places.
	Rounded float64
	// ZeroDecimals indicates how many decimal places in Rounded have a value
	// of zero.
	ZeroDecimals int
	// NonZeroDecimals indicates the rounding precision used, it is provided on
	// a call to newValue.
	NonZeroDecimals int
}

// NewValue returns a new instance representing the provided value, rounding
// the value up to the specified number of non-zero decimal places.
func NewValue(v float64, nonZeroDecimals int) *Value {
	r, zd := numbers.RoundToNonZeroPlaces(v, nonZeroDecimals)
	return &Value{
		Value:           v,
		Rounded:         r,
		ZeroDecimals:    zd,
		NonZeroDecimals: nonZeroDecimals,
	}
}

// Text returns textual representation of the value.
func (v *Value) Text() string {
	if math.Ceil(v.Rounded) == v.Rounded {
		return fmt.Sprintf("%.0f", v.Rounded)
	}

	format := fmt.Sprintf("%%.%df", v.NonZeroDecimals+v.ZeroDecimals)
	t := fmt.Sprintf(format, v.Rounded)
	if len(t) > 10 {
		t = fmt.Sprintf("%.2e", v.Rounded)
	}
	return t
}
