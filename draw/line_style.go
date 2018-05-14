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

package draw

import "fmt"

// line_style.go contains the Unicode characters used for drawing lines of
// different styles.

// lineStyleChars maps the line styles to the corresponding component characters.
var lineStyleChars = map[LineStyle]map[linePart]rune{
	LineStyleLight: {
		hLine:             '─',
		vLine:             '│',
		topLeftCorner:     '┌',
		topRightCorner:    '┐',
		bottomLeftCorner:  '└',
		bottomRightCorner: '┘',
	},
}

// lineParts returns the line component characters for the provided line style.
func lineParts(ls LineStyle) (map[linePart]rune, error) {
	parts, ok := lineStyleChars[ls]
	if !ok {
		return nil, fmt.Errorf("unsupported line style %v", ls)
	}
	return parts, nil
}

// LineStyle defines the supported line styles.Q
type LineStyle int

// String implements fmt.Stringer()
func (ls LineStyle) String() string {
	if n, ok := lineStyleNames[ls]; ok {
		return n
	}
	return "LineStyleUnknown"
}

// lineStyleNames maps LineStyle values to human readable names.
var lineStyleNames = map[LineStyle]string{
	LineStyleLight: "LineStyleLight",
}

// Supported line styles.
const (
	LineStyleNone LineStyle = iota
	LineStyleLight
)

// linePart identifies individual line parts.
type linePart int

// String implements fmt.Stringer()
func (lp linePart) String() string {
	if n, ok := linePartNames[lp]; ok {
		return n
	}
	return "linePartUnknown"
}

// linePartNames maps linePart values to human readable names.
var linePartNames = map[linePart]string{
	vLine:             "linePartVLine",
	topLeftCorner:     "linePartTopLeftCorner",
	topRightCorner:    "linePartTopRightCorner",
	bottomLeftCorner:  "linePartBottomLeftCorner",
	bottomRightCorner: "linePartBottomRightCorner",
}

const (
	hLine linePart = iota
	vLine
	topLeftCorner
	topRightCorner
	bottomLeftCorner
	bottomRightCorner
)
