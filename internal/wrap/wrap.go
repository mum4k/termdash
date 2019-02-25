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
	"strings"
	"text/scanner"

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
	// width of the canvas, wrapping is performed at rune boundaries.
	AtWords
)

// needed returns true if wrapping is needed for the rune at the horizontal
// position on the canvas that has the specified width.
// This will always return false if no options are provided, since the default
// behavior is to not wrap the text.
func needed(r rune, posX, width int, m Mode) bool {
	if r == '\n' {
		// Don't wrap for newline characters as they aren't printed on the
		// canvas, i.e. they take no horizontal space.
		return false
	}
	rw := runewidth.RuneWidth(r)
	return posX > width-rw && m == AtRunes
}

// Lines finds the starting positions of all lines in the text when the
// text is drawn on a canvas of the provided width and the specified wrapping
// mode.
func Lines(text string, width int, m Mode) []int {
	if width <= 0 || len(text) == 0 {
		return nil
	}

	ls := newLineScanner(text, width, m)
	for state := scanStart; state != nil; state = state(ls) {
	}
	return ls.lines
}

// lineScanner tracks the progress of scanning the input text when finding
// lines. Lines are identified when newline characters are encountered in the
// input text or when the canvas width and configuration requires line
// wrapping.
type lineScanner struct {
	// scanner is a lexer of the input text.
	scanner *scanner.Scanner

	// width is the width of the canvas the text will be drawn on.
	width int

	// posX tracks the horizontal position of the current character on the
	// canvas.
	posX int

	// mode is the wrapping mode.
	mode Mode

	// lines are the starting points of the identified lines.
	lines []int
}

// newLineScanner returns a new line scanner of the provided text.
func newLineScanner(text string, width int, m Mode) *lineScanner {
	var s scanner.Scanner
	s.Init(strings.NewReader(text))
	s.Whitespace = 0 // Don't ignore any whitespace.
	s.Mode = scanner.ScanIdents
	s.IsIdentRune = func(ch rune, i int) bool {
		return i == 0 && ch == '\n'
	}

	return &lineScanner{
		scanner: &s,
		width:   width,
		mode:    m,
	}
}

// scannerState is a state in the FSM that scans the input text and identifies
// newlines.
type scannerState func(*lineScanner) scannerState

// scanStart records the starting location of the current line.
func scanStart(ls *lineScanner) scannerState {
	switch tok := ls.scanner.Peek(); {
	case tok == scanner.EOF:
		return nil

	default:
		ls.lines = append(ls.lines, ls.scanner.Position.Offset)
		return scanLine
	}
}

// scanLine scans a line until it finds its end.
func scanLine(ls *lineScanner) scannerState {
	for {
		switch tok := ls.scanner.Scan(); {
		case tok == scanner.EOF:
			return nil

		case tok == scanner.Ident:
			return scanLineBreak

		case needed(tok, ls.posX, ls.width, ls.mode):
			return scanLineWrap

		default:
			// Move horizontally within the line for each scanned character.
			ls.posX += runewidth.RuneWidth(tok)
		}
	}
}

// scanLineBreak processes a newline character in the input text.
func scanLineBreak(ls *lineScanner) scannerState {
	// Newline characters aren't printed, the following character starts the line.
	if ls.scanner.Peek() != scanner.EOF {
		ls.posX = 0
		ls.lines = append(ls.lines, ls.scanner.Position.Offset+1)
	}
	return scanLine
}

// scanLineWrap processes a line wrap due to canvas width.
func scanLineWrap(ls *lineScanner) scannerState {
	// The character on which we wrapped will be printed and is the start of
	// new line.
	ls.posX = runewidth.StringWidth(ls.scanner.TokenText())
	ls.lines = append(ls.lines, ls.scanner.Position.Offset)
	return scanLine
}
