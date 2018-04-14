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

package termbox

// cell_options.go converts termdash cell options to the termbox format.

import (
	"fmt"

	"github.com/mum4k/termdash/cell"
	tbx "github.com/nsf/termbox-go"
)

// cellColor converts termdash cell color to the termbox format.
func cellColor(c cell.Color) (tbx.Attribute, error) {
	switch c {
	case cell.ColorDefault:
		return tbx.ColorDefault, nil
	case cell.ColorBlack:
		return tbx.ColorBlack, nil
	case cell.ColorRed:
		return tbx.ColorRed, nil
	case cell.ColorGreen:
		return tbx.ColorGreen, nil
	case cell.ColorYellow:
		return tbx.ColorYellow, nil
	case cell.ColorBlue:
		return tbx.ColorBlue, nil
	case cell.ColorMagenta:
		return tbx.ColorMagenta, nil
	case cell.ColorCyan:
		return tbx.ColorCyan, nil
	case cell.ColorWhite:
		return tbx.ColorWhite, nil
	default:
		return 0, fmt.Errorf("don't know how to convert cell color %v to the termbox format", c)
	}
}

// cellOptsToFg converts the cell options to the termbox foreground attribute.
func cellOptsToFg(opts *cell.Options) (tbx.Attribute, error) {
	return cellColor(opts.FgColor)
}

// cellOptsToBg converts the cell options to the termbox background attribute.
func cellOptsToBg(opts *cell.Options) (tbx.Attribute, error) {
	return cellColor(opts.BgColor)
}
