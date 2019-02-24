// Copyright 2019 Google Inc.
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

// Package testsegment provides helpers for tests that use the segment package.
package testsegment

import (
	"fmt"
	"image"

	"github.com/mum4k/termdash/draw/segdisp/segment"
	"github.com/mum4k/termdash/internal/canvas/braille"
)

// MustHV draws the segment or panics.
func MustHV(bc *braille.Canvas, ar image.Rectangle, st segment.Type, opts ...segment.Option) {
	if err := segment.HV(bc, ar, st, opts...); err != nil {
		panic(fmt.Sprintf("segment.HV => unexpected error: %v", err))
	}
}

// MustDiagonal draws the segment or panics.
func MustDiagonal(bc *braille.Canvas, ar image.Rectangle, width int, dt segment.DiagonalType, opts ...segment.DiagonalOption) {
	if err := segment.Diagonal(bc, ar, width, dt, opts...); err != nil {
		panic(fmt.Sprintf("segment.Diagonal => unexpected error: %v", err))
	}
}
