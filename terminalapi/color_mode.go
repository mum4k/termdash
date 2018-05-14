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

package terminalapi

// color_mode.go defines the terminal color modes.

// ColorMode represents
type ColorMode int

// String implements fmt.Stringer()
func (cm ColorMode) String() string {
	if n, ok := colorModeNames[cm]; ok {
		return n
	}
	return "ColorModeUnknown"
}

// colorModeNames maps ColorMode values to human readable names.
var colorModeNames = map[ColorMode]string{
	ColorMode8:         "ColorMode8",
	ColorMode256:       "ColorMode256",
	ColorMode216:       "ColorMode216",
	ColorModeGrayscale: "ColorModeGrayscale",
}

// Supported color modes.
const (
	ColorMode8 ColorMode = iota
	ColorMode256
	ColorMode216
	ColorModeGrayscale
)
