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

package cell

// color.go defines constants for cell colors.

// Color is the color of a cell.
type Color int

// String implements fmt.Stringer()
func (cc Color) String() string {
	if n, ok := colorNames[cc]; ok {
		return n
	}
	return "ColorUnknown"
}

// colorNames maps Color values to human readable names.
var colorNames = map[Color]string{
	ColorDefault: "ColorDefault",
	ColorBlack:   "ColorBlack",
	ColorRed:     "ColorRed",
	ColorGreen:   "ColorGreen",
	ColorYellow:  "ColorYellow",
	ColorBlue:    "ColorBlue",
	ColorMagenta: "ColorMagenta",
	ColorCyan:    "ColorCyan",
	ColorWhite:   "ColorWhite",
}

// The supported terminal colors.
const (
	ColorDefault Color = iota
	ColorBlack
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)
