package sparkline

// sparks.go contains code that determines which characters should be used to
// represent a value on the SparkLine.

import (
	"fmt"
	"math"

	runewidth "github.com/mattn/go-runewidth"
)

// sparks are the characters used to draw the SparkLine.
var sparks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// visibleMax determines the maximum visible data point given the canvas width.
// Returns a slice that contains only visible data points and the maximum value
// among them.
func visibleMax(data []int, width int) ([]int, int) {
	if width <= 0 || len(data) == 0 {
		return nil, 0
	}

	if width < len(data) {
		data = data[len(data)-width:]
	}

	var max int
	for _, v := range data {
		if v > max {
			max = v
		}
	}
	return data, max
}

// blocks represents the building blocks that display one value on a SparkLine.
// I.e. one vertical bar.
type blocks struct {
	// full is the number of fully populated blocks.
	full int

	// partSpark is the spark character from sparks that should be used in the
	// topmost block. Equals to zero if no partial block should be displayed.
	partSpark rune
}

// toBlocks determines the number of full and partial vertical blocks required
// to represent the provided value given the specified max visible value and
// number of vertical cells available to the SparkLine.
func toBlocks(value, max, vertCells int) blocks {
	if value <= 0 || max <= 0 || vertCells <= 0 {
		return blocks{}
	}

	// How many of the smallest spark elements fit into a cell.
	cellSparks := len(sparks)

	// Scale is how much of the max does one smallest spark element represent,
	// given the vertical cells that will be used to represent the value.
	scale := float64(cellSparks) * float64(vertCells) / float64(max)

	// How many smallest spark elements are needed to represent the value.
	elements := int(round(float64(value) * scale))

	b := blocks{
		full: elements / cellSparks,
	}

	part := elements % cellSparks
	if part > 0 {
		b.partSpark = sparks[part-1]
	}
	return b
}

// round returns the nearest integer, rounding half away from zero.
// Copied from the math package of Go 1.10 for backwards compatibility with Go
// 1.8 where the math.Round function doesn't exist yet.
func round(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}

// init ensures that all spark characters are half-width runes.
// The SparkLine widget assumes that each value can be represented in a column
// that has a width of one cell.
func init() {
	for i, s := range sparks {
		if got := runewidth.RuneWidth(s); got > 1 {
			panic(fmt.Sprintf("all sparks must be half-width runes (width of one), spark[%d] has width %d", i, got))
		}
	}
}
