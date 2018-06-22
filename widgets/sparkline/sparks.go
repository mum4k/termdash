package sparkline

import "math"

// sparks.go contains code that determines which characters should be used to
// represent a value on the SparkLine.

// sparks are the characters used to draw the SparkLine.
// Note that the last character representing fully populated cell isn't ever
// used. If we need to fill the cell fully, we use a space character with background
// color set. This ensures we have no gaps between cells.
var sparks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// visibleMax determines the maximum visible data point given the canvas width.
func visibleMax(data []int, width int) int {
	if width <= 0 || len(data) == 0 {
		return 0
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
	return max
}

// blocks represents blocks that display one value on a SparkLine.
type blocks struct {
	// full is the number of fully populated blocks.
	full int

	// partSpark is the spark character from sparks that should be used in the
	// topmost block. Equals to zero if no part blocks should be displayed.
	partSpark rune
}

// toBlocks determines the number of full and partial vertical blocks required
// to represent the provided value given the specified max visible value and
// number of vertical cells available to the SparkLine.
func toBlocks(value, max, vertCells int) blocks {
	if value <= 0 || max <= 0 || vertCells <= 0 {
		return blocks{}
	}

	// How many of the smallesr spark elements fit into a cell.
	cellSparks := len(sparks)

	// Scale is how much of the max does one smallest spark element represent,
	// given the vertical cells that will be used to represent the value.
	scale := float64(cellSparks) * float64(vertCells) / float64(max)

	// How many smallest spark elements are needed to represent the value.
	elements := int(math.Round(float64(value) * scale))

	b := blocks{
		full: elements / cellSparks,
	}

	part := elements % cellSparks
	if part > 0 {
		b.partSpark = sparks[part-1]
	}
	return b
}
