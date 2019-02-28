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

// Package wrap implements line wrapping at character or word boundaries.
package wrap

import (
	"github.com/mum4k/termdash/internal/canvas/buffer"
	"github.com/mum4k/termdash/internal/runewidth"
)

// Mode sets the wrapping mode.
type Mode int

// String implements fmt.Stringer()
func (m Mode) String() string {
	if n, ok := modeNames[m]; ok {
		return n
	}
	return "ModeUnknown"
}

// modeNames maps Mode values to human readable names.
var modeNames = map[Mode]string{
	Never:   "WrapModeNever",
	AtRunes: "WrapModeAtRunes",
	AtWords: "WrapModeAtWords",
}

const (
	// Never is the default wrapping mode, which disables line wrapping.
	Never Mode = iota

	// AtRunes is a wrapping mode where if the width of the text crosses the
	// width of the canvas, wrapping is performed at rune boundaries.
	AtRunes

	// AtWords is a wrapping mode where if the width of the text crosses the
	// width of the canvas, wrapping is performed at word boundaries. The
	// wrapping still switches back to the AtRunes mode for any words that are
	// longer than the width.
	AtWords
)

// needed returns true if wrapping is needed for the rune at the horizontal
// position on the canvas that has the specified width.
// This will always return false if no options are provided, since the default
// behavior is to not wrap the text.
func needed(r rune, posX, width int, m Mode) bool {
	rw := runewidth.RuneWidth(r)
	return posX > width-rw && m == AtRunes
}

// Cells returns the cells wrapped into individual lines according to the
// specified width and wrapping mode.
//
// This function consumes any cells that contain newline characters and uses
// them to start new lines.
//
// If the mode is AtWords, this function also drops cells with leading space
// character before a word at which the wrap occurs.
func Cells(cells []*buffer.Cell, width int, m Mode) [][]*buffer.Cell {
	if width <= 0 || len(cells) == 0 {
		return nil
	}

	cs := newCellScanner(cells, width, m)
	for state := scanCellLine; state != nil; state = state(cs) {
	}
	return cs.lines
}

// cellScannerState is a state in the FSM that scans the input text and identifies
// newlines.
type cellScannerState func(*cellScanner) cellScannerState

// cellScanner tracks the progress of scanning the input cells when finding
// lines.
type cellScanner struct {
	// cells are the cells being scanned.
	cells []*buffer.Cell

	// nextIdx is the index of the cell that will be returned by next.
	nextIdx int

	// width is the width of the canvas the text will be drawn on.
	width int

	// posX tracks the horizontal position of the current cell on the canvas.
	posX int

	// mode is the wrapping mode.
	mode Mode

	// lines are the identified lines.
	lines [][]*buffer.Cell

	// line is the current line.
	line []*buffer.Cell
}

// newCellScanner returns a scanner of the provided cells.
func newCellScanner(cells []*buffer.Cell, width int, m Mode) *cellScanner {
	return &cellScanner{
		cells: cells,
		width: width,
		mode:  m,
	}
}

// next returns the next cell and advances the scanner.
// Returns nil when there are no more cells to scan.
func (cs *cellScanner) next() *buffer.Cell {
	c := cs.peek()
	if c != nil {
		cs.nextIdx++
	}
	return c
}

// peek returns the next cell without advancing the scanner's position.
// Returns nil when there are no more cells to peek at.
func (cs *cellScanner) peek() *buffer.Cell {
	if cs.nextIdx >= len(cs.cells) {
		return nil
	}
	return cs.cells[cs.nextIdx]
}

// peekPrev returns the previous cell without changing the scanner's position.
// Returns nil if the scanner is at the first cell.
func (cs *cellScanner) peekPrev() *buffer.Cell {
	if cs.nextIdx == 0 {
		return nil
	}
	return cs.cells[cs.nextIdx-1]
}

// scanCellLine scans a line until it finds its end due to a newline character
// or the specified width.
func scanCellLine(cs *cellScanner) cellScannerState {
	for {

		cell := cs.next()
		if cell == nil {
			if len(cs.line) > 0 || cs.peekPrev().Rune == '\n' {
				cs.lines = append(cs.lines, cs.line)
			}
			return nil
		}

		switch r := cell.Rune; {
		case r == '\n':
			return scanCellLineBreak

		case needed(r, cs.posX, cs.width, cs.mode):
			return scanCellLineWrap

		default:
			// Move horizontally within the line for each scanned cell.
			cs.posX += runewidth.RuneWidth(r)

			// Copy the cell into the current line.
			cs.line = append(cs.line, cell)
		}
	}
}

// scanCellLineBreak processes a newline character cell.
func scanCellLineBreak(cs *cellScanner) cellScannerState {
	cs.lines = append(cs.lines, cs.line)
	cs.posX = 0
	cs.line = nil
	return scanCellLine
}

// scanCellLineWrap processes a line wrap due to canvas width.
func scanCellLineWrap(cs *cellScanner) cellScannerState {
	// The character on which we wrapped will be printed and is the start of
	// new line.
	cs.lines = append(cs.lines, cs.line)
	cs.posX = runewidth.RuneWidth(cs.peekPrev().Rune)
	cs.line = []*buffer.Cell{cs.peekPrev()}
	return scanCellLine
}
