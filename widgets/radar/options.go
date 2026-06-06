// Copyright 2024 Google Inc.
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

package radar

// options.go contains configurable options for Radar.

import (
	"fmt"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/linestyle"
)

// Direction specifies the direction of the radar sweep.
type Direction int

const (
	// DirectionClockwise sweeps the beam in a clockwise direction.
	DirectionClockwise Direction = iota
	// DirectionCounterClockwise sweeps the beam counter-clockwise.
	DirectionCounterClockwise
)

// Option is used to provide options to Radar.
type Option interface {
	// set applies this option to the options struct.
	set(*options)
}

// options holds the provided options.
type options struct {
	// sweepSpeed is the angular velocity of the sweep beam in degrees/second.
	sweepSpeed float64
	// beamWidth is the angular width (in degrees) of the visible fade trail.
	beamWidth float64
	// sweepSpan is the total arc covered in one sweep (360 = full rotation).
	// Values less than 360 cause the beam to oscillate.
	sweepSpan float64
	// startAngle is the initial angle of the beam in degrees from North (clockwise).
	startAngle float64
	// direction is the sweep rotation direction.
	direction Direction

	// rangeRings is the number of concentric distance rings drawn on the display.
	rangeRings int

	// beamR/G/B is the primary beam and trail color (RGB24, 0–255 each).
	beamR, beamG, beamB int
	// contactR/G/B is the color of contact blip runes (RGB24, 0–255 each).
	contactR, contactG, contactB int

	// contactChar is the Unicode rune drawn at each contact's position.
	contactChar rune

	// tooltipEnabled controls whether a tooltip popup is shown when the mouse
	// hovers over a contact blip. Defaults to true.
	tooltipEnabled bool

	// border is the optional border style drawn around the widget.
	border linestyle.LineStyle
	// borderCellOpts are cell options applied to the border cells.
	borderCellOpts []cell.Option
	// borderTitle is the text displayed in the border.
	borderTitle string
}

// newOptions returns options with the default values set.
func newOptions() *options {
	return &options{
		sweepSpeed:     60.0,  // One full rotation every 6 seconds.
		beamWidth:      35.0,  // 35-degree fade trail.
		sweepSpan:      360.0, // Full rotation.
		startAngle:     0.0,   // Start at North.
		direction:      DirectionClockwise,
		rangeRings:     3,
		beamR:          0, // Neon green beam.
		beamG:          255,
		beamB:          70,
		contactR:       255, // Hot red contacts.
		contactG:       50,
		contactB:       50,
		contactChar:    '◆',
		tooltipEnabled: true,
	}
}

// validate validates the provided options.
func (o *options) validate() error {
	if o.sweepSpeed <= 0 {
		return fmt.Errorf("invalid SweepSpeed %v, must be a positive number of degrees per second", o.sweepSpeed)
	}
	if o.beamWidth <= 0 || o.beamWidth > 180 {
		return fmt.Errorf("invalid BeamWidth %v, must be in range (0, 180]", o.beamWidth)
	}
	if o.sweepSpan <= 0 || o.sweepSpan > 360 {
		return fmt.Errorf("invalid SweepSpan %v, must be in range (0, 360]", o.sweepSpan)
	}
	if o.rangeRings < 1 || o.rangeRings > 10 {
		return fmt.Errorf("invalid RangeRings %v, must be in range [1, 10]", o.rangeRings)
	}
	for _, v := range []int{o.beamR, o.beamG, o.beamB, o.contactR, o.contactG, o.contactB} {
		if v < 0 || v > 255 {
			return fmt.Errorf("invalid color component %v, must be in range [0, 255]", v)
		}
	}
	return nil
}

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// beamColorRGB returns the RGB components of the beam color.
func (o *options) beamColorRGB() (int, int, int) {
	return o.beamR, o.beamG, o.beamB
}

// contactColorRGB returns the RGB components of the contact color.
func (o *options) contactColorRGB() (int, int, int) {
	return o.contactR, o.contactG, o.contactB
}

// dimColor returns a very dim version of the beam color for grid lines.
func (o *options) dimColor() cell.Color {
	r := o.beamR * 12 / 100
	g := o.beamG * 12 / 100
	b := o.beamB * 12 / 100
	// Ensure a minimum visible value so the grid is always perceivable.
	if r == 0 && g < 8 && b == 0 {
		g = 8
	}
	return cell.ColorRGB24(r, g, b)
}

// compassColor returns a medium-brightness version of the beam color for N/S/E/W labels.
func (o *options) compassColor() cell.Color {
	return cell.ColorRGB24(o.beamR*55/100, o.beamG*55/100, o.beamB*55/100)
}

// DefaultSweepSpeed is the default sweep speed in degrees per second.
const DefaultSweepSpeed = 60.0

// SweepSpeed sets the angular velocity of the radar sweep beam.
// The value is in degrees per second; e.g. 60 = one full rotation every 6 s.
// Defaults to 60.
func SweepSpeed(degreesPerSecond float64) Option {
	return option(func(o *options) {
		o.sweepSpeed = degreesPerSecond
	})
}

// DefaultBeamWidth is the default angular width of the sweep trail in degrees.
const DefaultBeamWidth = 35.0

// BeamWidth sets the angular width of the visible sweep trail (the phosphor fade).
// The value is in degrees. Larger values create a wider, more gradual fade.
// Defaults to 35.
func BeamWidth(degrees float64) Option {
	return option(func(o *options) {
		o.beamWidth = degrees
	})
}

// DefaultSweepSpan is the default total sweep arc in degrees (full rotation).
const DefaultSweepSpan = 360.0

// SweepSpan sets the total arc swept in one pass. Use 360 for a full rotation.
// Values less than 360 cause the beam to oscillate between 0 and SweepSpan
// rather than continuously rotating.
// Defaults to 360.
func SweepSpan(degrees float64) Option {
	return option(func(o *options) {
		o.sweepSpan = degrees
	})
}

// StartAngle sets the initial beam angle in degrees measured clockwise from
// North (top of the display). Defaults to 0.
func StartAngle(degrees float64) Option {
	return option(func(o *options) {
		o.startAngle = degrees
	})
}

// SweepDirection sets whether the beam sweeps clockwise or counter-clockwise.
// Defaults to DirectionClockwise.
func SweepDirection(d Direction) Option {
	return option(func(o *options) {
		o.direction = d
	})
}

// DefaultRangeRings is the default number of concentric range rings.
const DefaultRangeRings = 3

// RangeRings sets the number of concentric distance markers drawn on the display.
// Must be between 1 and 10. Defaults to 3.
func RangeRings(n int) Option {
	return option(func(o *options) {
		o.rangeRings = n
	})
}

// BeamColor sets the primary color of the radar beam and its phosphor trail.
// r, g, b are RGB24 components in the range [0, 255].
// Defaults to neon green (0, 255, 70).
func BeamColor(r, g, b int) Option {
	return option(func(o *options) {
		o.beamR = r
		o.beamG = g
		o.beamB = b
	})
}

// ContactColor sets the color used to render contact blip runes.
// r, g, b are RGB24 components in the range [0, 255].
// Defaults to red (255, 50, 50).
func ContactColor(r, g, b int) Option {
	return option(func(o *options) {
		o.contactR = r
		o.contactG = g
		o.contactB = b
	})
}

// DefaultContactChar is the default rune used to mark contact positions.
const DefaultContactChar = '◆'

// ContactChar sets the Unicode rune drawn at each contact point on the display.
// Defaults to '◆'.
func ContactChar(ch rune) Option {
	return option(func(o *options) {
		o.contactChar = ch
	})
}

// ShowTooltip enables the mouse-hover tooltip popup. This is the default.
func ShowTooltip() Option {
	return option(func(o *options) {
		o.tooltipEnabled = true
	})
}

// HideTooltip disables the mouse-hover tooltip popup.
func HideTooltip() Option {
	return option(func(o *options) {
		o.tooltipEnabled = false
	})
}

// ── Tooltip colour helpers (derived from beam/contact colours) ────────────────

// tooltipBorderColor returns the colour used for the tooltip box border.
func (o *options) tooltipBorderColor() cell.Color {
	return cell.ColorRGB24(o.beamR, o.beamG, o.beamB)
}

// tooltipBgColor returns a very dark tinted background for the tooltip interior.
func (o *options) tooltipBgColor() cell.Color {
	r := clampRGB(o.beamR * 9 / 100)
	g := clampRGB(o.beamG * 9 / 100)
	b := clampRGB(o.beamB * 9 / 100)
	// Guarantee a faintly distinguishable backdrop.
	if r == 0 && g < 10 && b == 0 {
		g = 10
	}
	return cell.ColorRGB24(r, g, b)
}

// tooltipLabelColor returns the colour used for the contact callsign in the tooltip.
func (o *options) tooltipLabelColor() cell.Color {
	return cell.ColorRGB24(o.contactR, o.contactG, o.contactB)
}

// tooltipDataColor returns the colour for secondary data lines in the tooltip.
func (o *options) tooltipDataColor() cell.Color {
	return cell.ColorRGB24(
		clampRGB(o.beamR*78/100),
		clampRGB(o.beamG*78/100),
		clampRGB(o.beamB*78/100),
	)
}

// Border configures the Radar to have a border of the specified style.
// Optionally accepts cell options to style the border (e.g. color).
func Border(ls linestyle.LineStyle, cOpts ...cell.Option) Option {
	return option(func(o *options) {
		o.border = ls
		o.borderCellOpts = cOpts
	})
}

// BorderTitle sets a text title displayed within the border.
// Has no effect if no border style is configured.
func BorderTitle(title string) Option {
	return option(func(o *options) {
		o.borderTitle = title
	})
}
