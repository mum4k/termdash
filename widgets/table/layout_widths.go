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

package table

// layout_widths.go calculates widths of table columns.

import (
	"errors"
	"fmt"
	"math"
)

// columnWidth is the width of a column in cells.
// This excludes any border, padding or spacing, i.e. this is the data portion
// only.
type columnWidth int

// columnWidths given the content and the available canvas width returns the
// widths of individual columns.
// The argument cvsWidth is assumed to exclude space required for any border,
// padding or spacing.
func columnWidths(content *Content, cvsWidth int) ([]columnWidth, error) {
	if wp := content.opts.columnWidthsPercent; len(wp) > 0 {
		// If the user provided widths for the columns, use those.
		widths, err := splitToPercent(cvsWidth, wp)
		if err != nil {
			return nil, err
		}
		return widths, nil
	}

	columns := int(content.cols)
	// Attempt equal column widths and see if there are any trimmed rows.
	if columns > cvsWidth {
		return nil, fmt.Errorf("the canvas width %d must be at least equal to the number of columns %d, so that each column gets at least one cell", cvsWidth, columns)
	}
	widths := splitEqually(cvsWidth, columns)
	if !widthsTrimCell(content, widths) {
		return widths, nil
	}

	// No widths specified and some rows need to be trimmed.
	// Choose columns widths that optimize for the smallest number of trimmed rows.
	// This is similar to the rod-cutting problem, except instead of maximizing
	// the price, we're minimizing the number of rows that would have their
	// content trimmed.

	inputs := &cutCanvasInputs{
		content:  content,
		cvsWidth: columnWidth(cvsWidth),
		columns:  int(content.cols),
		best:     map[cutState]int{},
	}
	state := &cutState{
		colIdx:   0,
		remWidth: inputs.cvsWidth,
	}

	best := cutCanvas(inputs, state, nil)

	var res []columnWidth
	last := 0
	for _, cut := range best.cuts {
		res = append(res, columnWidth(cut-last))
		last = cut
	}
	res = append(res, columnWidth(cvsWidth-last))
	return res, nil
}

// cutState uniquely identifies a state in the cutting process.
type cutState struct {
	// colIdx is the index of the column whose width is being determined in
	// this execution of cutCanvas.
	colIdx int

	// remWidth is the total remaining width of the canvas for the current
	// column and all the following columns.
	remWidth columnWidth
}

// bestCuts is the best result for a particular cutState.
// Used for memoization.
type bestCuts struct {
	// cost is the smallest achievable cost for the cut state.
	// This is the number of rows that will have to be trimmed.
	cost int
	// cuts are the cuts done so far to get to this state.
	cuts []int
}

// cutCanvasInputs are the inputs to the cutCanvas function.
// These are shared by all the functions in the call stack.
type cutCanvasInputs struct {
	// content is the table content.
	content *Content

	// cvsWidth is the width of the canvas that is available for the data.
	cvsWidth columnWidth

	// columns indicates the total number of columns in the table.
	columns int

	// best is a memoization on top of cutCanvas.
	// It maps cutState to the minimal cost for that state.
	best map[cutState]int
}

// cutCanvas cuts the canvas width to a number of rows optimizing for the
// smallest possible number of trimmed rows.
func cutCanvas(inputs *cutCanvasInputs, state *cutState, cuts []int) *bestCuts {
	minCost := math.MaxInt32
	var minCuts []int

	nextColIdx := state.colIdx + 1
	if nextColIdx > inputs.columns-1 {
		return &bestCuts{
			cost: trimmedRows(inputs.content, state.colIdx, state.remWidth),
			cuts: cuts,
		}
	}

	for colWidth := columnWidth(1); colWidth < state.remWidth; colWidth++ {
		diff := inputs.cvsWidth - state.remWidth
		idxThisCut := diff + colWidth
		costThisCut := trimmedRows(inputs.content, state.colIdx, colWidth)
		nextState := &cutState{
			colIdx:   nextColIdx,
			remWidth: state.remWidth - colWidth,
		}
		nextCuts := append(cuts, int(idxThisCut))

		// Use the memoized result if available.
		var nextBest *bestCuts
		if nextCost, ok := inputs.best[*nextState]; !ok {
			nextBest = cutCanvas(inputs, nextState, nextCuts)
			inputs.best[*nextState] = nextBest.cost // Memoize.
		} else {
			nextBest = &bestCuts{
				cost: nextCost,
				cuts: nextCuts,
			}
		}

		if newMinCost := costThisCut + nextBest.cost; newMinCost < minCost {
			minCost = newMinCost
			minCuts = nextBest.cuts
		}
	}
	return &bestCuts{
		cost: minCost,
		cuts: minCuts,
	}
}

// widthsTrimCell asserts whether there is at least one cell in the content
// whose data gets trimmed when using the specified column widths.
func widthsTrimCell(content *Content, widths []columnWidth) bool {
	for colIdx := 0; colIdx < int(content.cols); colIdx++ {
		if trimmed := trimmedRows(content, colIdx, widths[colIdx]); trimmed > 0 {
			return true
		}
	}
	return false
}

// trimmedRows returns the number of rows that will have data cells with
// trimmed content in column of the specified index if the assigned width of
// the column is colWidth.
func trimmedRows(content *Content, colIdx int, colWidth columnWidth) int {
	trimmed := 0
	for _, row := range content.rows {
		tgtCell := row.cells[colIdx]
		if !tgtCell.trimmable {
			// Cells that have wrapping enabled are never trimmed and so have
			// no influence on the calculated column widths.
			continue
		}
		available := int(colWidth) - 2*tgtCell.hierarchical.getHorizontalPadding()
		if tgtCell.width > available {
			trimmed++
		}
	}
	return trimmed
}

// splitToPercent splits the canvas widths to columns each having the assigned
// percentage of the width.
func splitToPercent(cvsWidth int, widthsPercent []int) ([]columnWidth, error) {
	if cvsWidth == 0 {
		return nil, errors.New("unable to split canvas of zero width to columns")
	}
	if len(widthsPercent) == 0 {
		return nil, errors.New("at least one width percentage must be provided")
	}
	if got := len(widthsPercent); got > cvsWidth {
		return nil, fmt.Errorf("the canvas width %d must be at least equal to the number of columns %d, so that each column gets at least one cell", cvsWidth, got)
	}

	var res []columnWidth
	remaining := cvsWidth
	for i := 0; i < len(widthsPercent)-1; i++ {
		perc := widthsPercent[i]
		var adjPerc float64
		if remaining < cvsWidth {
			ofOrig := float64(cvsWidth) / 100 * float64(perc)
			adjPerc = ofOrig / float64(remaining) * 100
		} else {
			adjPerc = float64(perc)
		}
		cur := int(math.Floor(float64(remaining) / 100 * adjPerc))

		colsAfter := len(widthsPercent) - i - 1
		if remAfter := remaining - cur; remAfter < colsAfter {
			diff := colsAfter - remAfter
			// Leave at least one cell for each remaining column.
			cur -= diff
		}

		if cur < 1 {
			// At least one cell per column.
			cur = 1
		}
		res = append(res, columnWidth(cur))
		remaining -= cur
	}

	res = append(res, columnWidth(remaining))
	return res, nil
}

// splitEqually splits the canvas to equal columns.
func splitEqually(cvsWidth, columns int) []columnWidth {
	each := cvsWidth / columns
	var res []columnWidth
	sum := 0
	for i := 0; i < columns-1; i++ {
		res = append(res, columnWidth(each))
		sum += each
	}
	res = append(res, columnWidth(cvsWidth-sum))
	return res
}
