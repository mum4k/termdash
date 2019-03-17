package table

// content_layout.go stores layout calculated for a canvas size.

import (
	"errors"
	"image"
	"math"

	"github.com/mum4k/termdash/internal/wrap"
)

// columnWidth is the width of a column in cells.
// This excludes any border, padding or spacing, i.e. this is the data portion
// only.
type columnWidth int

// contentLayout determines how the content gets placed onto the canvas.
type contentLayout struct {
	// lastCvsAr is the are of the last canvas the content was drawn on.
	// This is image.ZR if the content hasn't been drawn yet.
	lastCvsAr image.Rectangle

	// columnWidths are the widths of individual columns in the table.
	columnWidths []columnWidth

	// Details about HV lines that are the borders.
}

// newContentLayout calculates new layout for the content when drawn on a
// canvas represented with the provided area.
func newContentLayout(content *Content, cvsAr image.Rectangle) (*contentLayout, error) {
	return nil, errors.New("unimplemented")
}

// columnWidths given the content and the available canvas width returns the
// widths of individual columns.
// The argument cvsWidth is assumed to exclude space required for any border,
// padding or spacing.
func columnWidths(content *Content, cvsWidth int) []columnWidth {
	// This is similar to the rod-cutting problem, except instead of maximizing
	// the price, we're minimizing the number of rows that would have their
	// content trimmed.

	idxColumnCosts := columnCosts(content, cvsWidth)
	inputs := &cutCanvasInputs{
		idxColumnCosts: idxColumnCosts,
		cvsWidth:       cvsWidth,
		columns:        int(content.cols),
		best:           map[cutState]int{},
	}
	state := cutState{
		colIdx:   0,
		remWidth: cvsWidth,
	}
	//log.Printf("idxColumnCosts: %v", idxColumnCosts)

	best := cutCanvas(inputs, state, nil)
	//log.Printf("minCost: %+v", best)

	var res []columnWidth
	last := 0
	for _, cut := range best.cuts {
		res = append(res, columnWidth(cut-last))
		last = cut
	}
	res = append(res, columnWidth(cvsWidth-last))
	return res
}

// cutState uniqiuely identifies a state in the cutting process.
type cutState struct {
	// colIdx is the index of the column whose width is being determined in
	// this execution of cutCanvas.
	colIdx int
	// remWidth is the total remaining width of the canvas for the current
	// column and all the following columns.
	remWidth int
}

// bestCuts is the best result for a particular cutState.
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
	// idxColumnCosts are the pre-calculated costs of individual cuts for each column.
	// Map the column index to the costs.
	idxColumnCosts map[int]widthCost

	// cvsWidth is the width of the canvas that is available for the data.
	cvsWidth int

	// columns indicates the total number of columns in the table.
	columns int

	// best is a memoization on top of cutCanvas.
	// It maps cutState to the minimal cost for that state.
	best map[cutState]int
}

func cutCanvas(inputs *cutCanvasInputs, state cutState, cuts []int) *bestCuts {
	//log.Printf("cutCanvas state:%+v, cuts:%v", state, cuts)

	minCost := math.MaxInt32
	var minCuts []int

	widthCosts := inputs.idxColumnCosts[state.colIdx]
	nextColIdx := state.colIdx + 1
	if nextColIdx > inputs.columns-1 {
		//log.Printf("cutCanvas state:%+v -> no more cuts cost:%d cuts:%v", state, widthCosts[state.remWidth], cuts)
		return &bestCuts{
			cost: widthCosts[state.remWidth],
			cuts: cuts,
		}
	}

	for colWidth := 1; colWidth < state.remWidth; colWidth++ {
		diff := inputs.cvsWidth - state.remWidth
		idxThisCut := diff + colWidth
		costThisCut := widthCosts[colWidth]
		nextState := cutState{
			colIdx:   nextColIdx,
			remWidth: state.remWidth - colWidth,
		}
		nextCuts := append(cuts, idxThisCut)

		// Memoized?
		var nextBest *bestCuts
		if nextCost, ok := inputs.best[nextState]; !ok {
			nextBest = cutCanvas(inputs, nextState, nextCuts)
			inputs.best[nextState] = nextBest.cost
		} else {
			nextBest = &bestCuts{
				cost: nextCost,
				cuts: nextCuts,
			}
		}

		if newMinCost := costThisCut + nextBest.cost; newMinCost < minCost {
			//log.Printf("cutCanvas state:%+v, at cuts %v, costThisCut from widthCosts:%v, at width %d:%d, nextBest:%+v, minCost:%d", state, cuts, widthCosts, colWidth, costThisCut, nextBest, minCost)
			minCost = newMinCost
			minCuts = nextBest.cuts
			//log.Printf("cutCanvas state:%+v, new minCost:%d minCuts:%v", state, minCost, minCuts)
		}
	}
	//log.Printf("cutCanvas state:%+v, cuts:%v -> minCost:%d, minCuts:%d", state, cuts, minCost, minCuts)
	return &bestCuts{
		cost: minCost,
		cuts: minCuts,
	}
}

// widthCost maps column widths to the number of rows that would be trimmed on
// that width because data in the column are longer than the width.
type widthCost map[int]int

// columnCosts calculates the costs of cutting the columns to various widths.
// Returns a map of column indexes to their cut costs.
func columnCosts(content *Content, cvsWidth int) map[int]widthCost {
	idxColumnCosts := map[int]widthCost{}
	maxColWidth := cvsWidth - (int(content.cols) - 1)
	for colIdx := 0; colIdx < int(content.cols); colIdx++ {
		for colWidth := 1; colWidth <= maxColWidth; colWidth++ {
			wc, ok := idxColumnCosts[colIdx]
			if !ok {
				wc = widthCost{}
				idxColumnCosts[colIdx] = wc
			}
			wc[colWidth] = trimmedRows(content, colIdx, colWidth)
		}
	}
	return idxColumnCosts
}

// trimmedRows returns the number of rows that will have data cells with
// trimmed content in column of the specified index if the assigned width of
// the column is colWidth.
func trimmedRows(content *Content, colIdx int, colWidth int) int {
	trimmed := 0
	for _, row := range content.rows {
		tgtCell := row.cells[colIdx]
		if tgtCell.hierarchical.getWrapMode() != wrap.Never {
			// Cells that have wrapping enabled are never trimmed.
			continue
		}
		if tgtCell.width() > colWidth {
			trimmed++
		}
	}
	return trimmed
}
