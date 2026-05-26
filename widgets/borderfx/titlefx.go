// Copyright 2026 Google Inc.
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

package borderfx

import (
	"strings"

	spin "github.com/mum4k/termdash/widgets/spinner"
)

// TitleSpinnerPlacement controls where a spinner is rendered relative to a title.
type TitleSpinnerPlacement int

const (
	// TitleSpinnerRight appends the spinner to the title.
	TitleSpinnerRight TitleSpinnerPlacement = iota + 1
	// TitleSpinnerLeft prefixes the spinner to the title.
	TitleSpinnerLeft
)

// TitleSpec describes one focus-aware border title.
//
// Base and Charset cover the common reveal effect. Spinner, LeftSpin, and
// RightSpin are lower-level hooks for title ornamentation so callers can build
// outward from a single stable title representation rather than choosing
// between competing title APIs.
type TitleSpec struct {
	Base             string
	Charset          string
	Spinner          spin.Spinner
	LeftSpin         spin.Spinner
	RightSpin        spin.Spinner
	SpinnerPlacement TitleSpinnerPlacement
}

// DecryptCharsetGroup groups the common scramble/decrypt character sets used
// by borderfx title reveal effects.
type DecryptCharsetGroup struct {
	WarpCoreFlux   string
	LCARSTelemetry string
	Comms          string
	Controls       string
	Default        string
	Sensors        string
}

// DecryptCharsets exposes reusable title-reveal character groups for the
// borderfx user API. The intended call site is:
//
//	borderfx.DecryptingTitle(" Comms ", borderfx.DecryptCharsets.Comms)
//
// so callers can choose a named reveal set without hand-copying character
// sequences into their own widget setup code.
var DecryptCharsets = DecryptCharsetGroup{
	WarpCoreFlux:   "01$#*+-",
	LCARSTelemetry: "><=|/-",
	Comms:          "[]{}:;?",
	Controls:       ".,:*~^",
	Default:        "⠿⣾⣶⣤⣀⠒",
	Sensors:        "⠿⣾⣶⣤⣀⠒",
}

// TitleCharsets is a compatibility alias for older code that started using
// the earlier grouped charset name before DecryptCharsets became the preferred
// public entry point.
var TitleCharsets = DecryptCharsets

// DecryptingTitle returns a title spec that reveals itself from the provided
// charset when the pane becomes focused.
//
// This is the simplest public entry point for the "decrypt on focus" title
// effect used by the borderfx demo. Callers can keep building on the returned
// spec with the existing spinner fields or the chainable helpers below.
func DecryptingTitle(base, charset string) TitleSpec {
	return TitleSpec{
		Base:    base,
		Charset: charset,
	}
}

// WithLeftSpinner returns a copy of the title spec with a left-side spinner.
func (s TitleSpec) WithLeftSpinner(spinner spin.Spinner) TitleSpec {
	s.LeftSpin = spinner
	return s
}

// WithRightSpinner returns a copy of the title spec with a right-side spinner.
func (s TitleSpec) WithRightSpinner(spinner spin.Spinner) TitleSpec {
	s.RightSpin = spinner
	return s
}

// Plain returns the resting version of the title.
func (s TitleSpec) Plain() string {
	return s.Base
}

// Decorated returns the title with its spinner frame, if configured.
func (s TitleSpec) Decorated(step int) string {
	left, right := s.resolvedSpinners()
	title := s.Base
	if len(left.Frames()) > 0 {
		title = left.DecorateLeft(title, step)
	}
	if len(right.Frames()) > 0 {
		title = right.DecorateRight(title, step)
	}
	return title
}

// Scrambled returns the decode/reveal version of the title.
func (s TitleSpec) Scrambled(reveal int) string {
	if s.Charset == "" {
		return s.Base
	}
	if reveal >= len([]rune(s.Base)) {
		return s.Base
	}
	set := []rune(s.Charset)
	title := []rune(s.Base)
	var b strings.Builder
	for i, r := range title {
		switch {
		case i < reveal || r == ' ':
			b.WriteRune(r)
		default:
			b.WriteRune(set[(i+reveal)%len(set)])
		}
	}
	return b.String()
}

func (s TitleSpec) resolvedSpinners() (spin.Spinner, spin.Spinner) {
	left := s.LeftSpin
	right := s.RightSpin
	if len(s.Spinner.Frames()) == 0 {
		return left, right
	}
	if s.SpinnerPlacement == TitleSpinnerLeft {
		left = s.Spinner
		return left, right
	}
	right = s.Spinner
	return left, right
}

// HasSpinners reports whether the spec has any spinner configured.
func (s TitleSpec) HasSpinners() bool {
	left, right := s.resolvedSpinners()
	return len(left.Frames()) > 0 || len(right.Frames()) > 0
}
