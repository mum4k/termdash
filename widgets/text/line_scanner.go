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

package text

// line_scanner.go contains code that finds lines within text.

import (
	"strings"
	"text/scanner"

	"github.com/mum4k/termdash/internal/runewidth"
)

// wrapNeeded returns true if wrapping is needed for the rune at the horizontal
// position on the canvas.
func wrapNeeded(r rune, cvsPosX, cvsWidth int, opts *options) bool {
	if r == '\n' {
		// Don't wrap for newline characters as they aren't printed on the
		// canvas, i.e. they take no horizontal space.
		return false
	}
	rw := runewidth.RuneWidth(r)
	return cvsPosX > cvsWidth-rw && opts.wrapAtRunes
}

// findLines finds the starting positions of all lines in the text when the
// text is drawn on a canvas of the provided width with the specified options.
func findLines(text string, cvsWidth int, opts *options) []int {
	if cvsWidth <= 0 || text == "" {
		return nil
	}

	ls := newLineScanner(text, cvsWidth, opts)
	for state := scanStart; state != nil; state = state(ls) {
	}
	return ls.lines
}

// lineScanner tracks the progress of scanning the input text when finding
// lines. Lines are identified when newline characters are encountered in the
// input text or when the canvas width and configuration requires line
// wrapping.
type lineScanner struct {
	// scanner lexes the input text.
	scanner *scanner.Scanner

	// cvsWidth is the width of the canvas the text will be drawn on.
	cvsWidth int

	// cvsPosX tracks the horizontal position of the current character on the
	// canvas.
	cvsPosX int

	// opts are the widget options.
	opts *options

	// lines are the starting points of the identified lines.
	lines []int
}

// newLineScanner returns a new line scanner of the provided text.
func newLineScanner(text string, cvsWidth int, opts *options) *lineScanner {
	var s scanner.Scanner
	s.Init(strings.NewReader(text))
	s.Whitespace = 0 // Don't ignore any whitespace.
	s.Mode = scanner.ScanIdents
	s.IsIdentRune = func(ch rune, i int) bool {
		return i == 0 && ch == '\n'
	}

	return &lineScanner{
		scanner:  &s,
		cvsWidth: cvsWidth,
		opts:     opts,
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

		case wrapNeeded(tok, ls.cvsPosX, ls.cvsWidth, ls.opts):
			return scanLineWrap

		default:
			// Move horizontally within the line for each scanned character.
			ls.cvsPosX += runewidth.RuneWidth(tok)
		}
	}
}

// scanLineBreak processes a newline character in the input text.
func scanLineBreak(ls *lineScanner) scannerState {
	// Newline characters aren't printed, the following character starts the line.
	if ls.scanner.Peek() != scanner.EOF {
		ls.cvsPosX = 0
		ls.lines = append(ls.lines, ls.scanner.Position.Offset+1)
	}
	return scanLine
}

// scanLineWrap processes a line wrap due to canvas width.
func scanLineWrap(ls *lineScanner) scannerState {
	// The character on which we wrapped will be printed and is the start of
	// new line.
	ls.cvsPosX = runewidth.StringWidth(ls.scanner.TokenText())
	ls.lines = append(ls.lines, ls.scanner.Position.Offset)
	return scanLine
}
