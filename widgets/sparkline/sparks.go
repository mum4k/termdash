// Copyright 2018 Google Inc.
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

package sparkline

// sparks.go contains code that determines which characters should be used to
// represent a value on the SparkLine.

import (
	"fmt"
	"math"

	"github.com/mum4k/termdash/private/runewidth"
)

// sparks are the default characters used to draw the SparkLine.
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
	return toBlocksWithSparks(value, max, vertCells, sparks)
}

// toBlocksWithSparks determines the number of full and partial vertical
// blocks required to represent the provided value using the provided spark set.
func toBlocksWithSparks(value, max, vertCells int, sparkSet []rune) blocks {
	if value <= 0 || max <= 0 || vertCells <= 0 {
		return blocks{}
	}
	if len(sparkSet) == 0 {
		return blocks{}
	}

	// How many of the smallest spark elements fit into a cell.
	cellSparks := len(sparkSet)

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
		b.partSpark = sparkSet[part-1]
	}
	return b
}

// validateSparkRunes ensures all runes in the spark set occupy exactly one
// terminal cell.
func validateSparkRunes(sparkSet []rune) error {
	if len(sparkSet) == 0 {
		return fmt.Errorf("SparkRunes must contain at least one rune")
	}
	for i, s := range sparkSet {
		if got := runewidth.RuneWidth(s); got != 1 {
			return fmt.Errorf("SparkRunes[%d]=%q has width %d, expected 1", i, s, got)
		}
	}
	return nil
}

// init ensures that all spark characters are half-width runes.
// The SparkLine widget assumes that each value can be represented in a column
// that has a width of one cell.
func init() {
	if err := validateSparkRunes(sparks); err != nil {
		panic(fmt.Sprintf("invalid default spark rune set: %v", err))
	}
}
