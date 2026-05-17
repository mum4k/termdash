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

// Package spectrum implements an audio-style spectrum analyzer widget.
package spectrum

import (
	"errors"
	"fmt"
	"image"
	"math"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Spectrum draws mirrored or half-duplex activity bars for fast-moving audio
// or telemetry signals.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Spectrum struct {
	mu sync.Mutex

	primary   []int
	secondary []int
	mode      Mode
	lastSpan  int
	opts      *options
}

// Sample identifies a visible data point under a rendered spectrum cell.
type Sample struct {
	// X is the one-based visible sample index.
	X int
	// Y is the sample value at X.
	Y int
}

// Update replaces channel samples using automatic mode selection.
// Passing a nil secondary slice selects half-duplex mode.
func (s *Spectrum) Update(primary, secondary []int) error {
	if secondary == nil {
		return s.SetHalfDuplex(primary)
	}
	return s.SetStereo(primary, secondary)
}

// New returns a new Spectrum widget.
func New(opts ...Option) (*Spectrum, error) {
	opt := newOptions()
	for _, o := range opts {
		o.set(opt)
	}
	if err := opt.validate(); err != nil {
		return nil, err
	}
	return &Spectrum{
		mode: opt.mode,
		opts: opt,
	}, nil
}

// Configure updates rendering options while preserving the current samples.
func (s *Spectrum) Configure(opts ...Option) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.opts.clone()
	for _, o := range opts {
		o.set(next)
	}
	if err := next.validate(); err != nil {
		return err
	}

	s.opts = next
	s.mode = next.mode
	return nil
}

// SetStereo replaces the mirrored primary and secondary channels.
func (s *Spectrum) SetStereo(primary, secondary []int) error {
	if err := validateSamples(primary); err != nil {
		return fmt.Errorf("primary: %w", err)
	}
	if err := validateSamples(secondary); err != nil {
		return fmt.Errorf("secondary: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.primary = copyInts(s.primary, primary)
	s.secondary = copyInts(s.secondary, secondary)
	s.mode = ModeStereo
	return nil
}

// SetHalfDuplex replaces the primary channel and switches the widget to
// half-duplex mode.
func (s *Spectrum) SetHalfDuplex(primary []int) error {
	if err := validateSamples(primary); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.primary = copyInts(s.primary, primary)
	s.secondary = nil
	s.mode = ModeHalfDuplex
	return nil
}

// ValueCapacity returns the number of visible values the widget could display
// along its active sampling axis on the last draw.
func (s *Spectrum) ValueCapacity() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastSpan
}

// ValueAt returns the visible sample under pos for a canvas of size.
//
// The provided position must be relative to the widget canvas, not the
// terminal. The returned X value is one-based and follows the samples currently
// visible after fitting to the rendered span.
func (s *Spectrum) ValueAt(size, pos image.Point) (Sample, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if size.X <= 0 || size.Y <= 0 {
		return Sample{}, false
	}
	return s.valueAt(image.Rect(0, 0, size.X, size.Y), pos)
}

// Draw draws the Spectrum widget onto the canvas.
// Implements widgetapi.Widget.Draw.
func (s *Spectrum) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	_ = meta

	s.mu.Lock()
	defer s.mu.Unlock()

	needAr, err := area.FromSize(s.minSize())
	if err != nil {
		return err
	}
	if !needAr.In(cvs.Area()) {
		return draw.ResizeNeeded(cvs)
	}

	if s.opts.orientation == OrientationHorizontal {
		return s.drawHorizontal(cvs)
	}
	return s.drawVertical(cvs)
}

// Keyboard input isn't supported on the Spectrum widget.
func (*Spectrum) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return errors.New("the Spectrum widget doesn't support keyboard events")
}

// Mouse input isn't supported on the Spectrum widget.
func (*Spectrum) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	return errors.New("the Spectrum widget doesn't support mouse events")
}

// Options implements widgetapi.Widget.Options.
func (s *Spectrum) Options() widgetapi.Options {
	s.mu.Lock()
	defer s.mu.Unlock()

	min := s.minSize()
	var max image.Point
	if s.opts.height > 0 {
		max = image.Point{X: 0, Y: min.Y}
	}
	return widgetapi.Options{
		MinimumSize:  min,
		MaximumSize:  max,
		WantKeyboard: widgetapi.KeyScopeNone,
		WantMouse:    widgetapi.MouseScopeNone,
	}
}

// drawVertical renders the spectrum around a horizontal center axis.
func (s *Spectrum) drawVertical(cvs *canvas.Canvas) error {
	ar := s.area(cvs)
	labelTop := ar.Min.Y
	labelBottom := ar.Max.Y - 1
	body := image.Rect(ar.Min.X, ar.Min.Y+1, ar.Max.X, ar.Max.Y-1)
	if body.Dy() < 2 {
		return draw.ResizeNeeded(cvs)
	}

	if s.opts.primaryLabel != "" {
		if err := draw.Text(cvs, s.opts.primaryLabel, image.Point{X: ar.Min.X, Y: labelTop}, draw.TextCellOpts(s.opts.labelCellOpts...)); err != nil {
			return err
		}
	}
	if s.opts.secondLabel != "" {
		if err := draw.Text(cvs, s.opts.secondLabel, image.Point{X: ar.Min.X, Y: labelBottom}, draw.TextCellOpts(s.opts.labelCellOpts...)); err != nil {
			return err
		}
	}

	// Half-duplex mode uses a single channel rising above the axis.
	if s.mode == ModeHalfDuplex {
		axisY := body.Max.Y - 1
		s.lastSpan = body.Dx()
		if err := s.drawHorizontalAxis(cvs, body.Min.X, body.Max.X, axisY); err != nil {
			return err
		}
		maxValue := maxSample(s.opts.maxValue, s.primary)
		visible := fitSamples(s.primary, body.Dx())
		startX := body.Min.X
		height := axisY - body.Min.Y
		thresholdCells := s.thresholdCells(maxValue, height)
		for i, sample := range visible {
			cells := scaledCells(sample, maxValue, height)
			for level := 0; level < cells; level++ {
				y := axisY - 1 - level
				if y < body.Min.Y {
					break
				}
				color := gradientColor(s.opts.gradient, cellRatio(level+1, height))
				color = s.thresholdColor(color, level+1, thresholdCells)
				r := s.opts.halfRune
				if level == cells-1 {
					r = s.opts.primaryPeak
				}
				if _, err := cvs.SetCell(image.Point{X: startX + i, Y: y}, r, cell.FgColor(color)); err != nil {
					return err
				}
			}
		}
		if err := s.drawUpperThreshold(cvs, body.Min.X, body.Max.X, axisY, thresholdCells); err != nil {
			return err
		}
		return nil
	}

	// Stereo mode mirrors primary and secondary channels around the axis.
	upperHeight := (body.Dy() - 1) / 2
	lowerHeight := body.Dy() - 1 - upperHeight
	if upperHeight < 1 || lowerHeight < 1 {
		return draw.ResizeNeeded(cvs)
	}
	axisY := body.Min.Y + upperHeight
	s.lastSpan = body.Dx()
	if err := s.drawHorizontalAxis(cvs, body.Min.X, body.Max.X, axisY); err != nil {
		return err
	}

	maxValue := maxSample(s.opts.maxValue, s.primary, s.secondary)
	primary := fitSamples(s.primary, body.Dx())
	secondary := fitSamples(s.secondary, body.Dx())
	upperThreshold := s.thresholdCells(maxValue, upperHeight)
	lowerThreshold := s.thresholdCells(maxValue, lowerHeight)
	for i := 0; i < body.Dx(); i++ {
		x := body.Min.X + i
		if i < len(primary) {
			cells := scaledCells(primary[i], maxValue, upperHeight)
			for level := 0; level < cells; level++ {
				y := axisY - 1 - level
				if y < body.Min.Y {
					break
				}
				color := gradientColor(s.opts.gradient, cellRatio(level+1, upperHeight))
				color = s.thresholdColor(color, level+1, upperThreshold)
				r := channelRune(level, cells, s.opts.primaryRunes, upperRune)
				if level == cells-1 {
					r = s.opts.primaryPeak
				}
				if _, err := cvs.SetCell(image.Point{X: x, Y: y}, r, cell.FgColor(color)); err != nil {
					return err
				}
			}
		}
		if i < len(secondary) {
			cells := scaledCells(secondary[i], maxValue, lowerHeight)
			for level := 0; level < cells; level++ {
				y := axisY + 1 + level
				if y >= body.Max.Y {
					break
				}
				color := gradientColor(s.opts.gradient, cellRatio(level+1, lowerHeight))
				color = s.thresholdColor(color, level+1, lowerThreshold)
				r := channelRune(level, cells, s.opts.secondRunes, lowerRune)
				if level == cells-1 {
					r = s.opts.secondPeak
				}
				if _, err := cvs.SetCell(image.Point{X: x, Y: y}, r, cell.FgColor(color)); err != nil {
					return err
				}
			}
		}
	}
	if err := s.drawUpperThreshold(cvs, body.Min.X, body.Max.X, axisY, upperThreshold); err != nil {
		return err
	}
	if err := s.drawLowerThreshold(cvs, body.Min.X, body.Max.X, axisY, lowerThreshold); err != nil {
		return err
	}
	return nil
}

// drawHorizontal renders the spectrum around a vertical center axis.
func (s *Spectrum) drawHorizontal(cvs *canvas.Canvas) error {
	ar := s.area(cvs)
	labelRow := ar.Min.Y
	body := image.Rect(ar.Min.X, ar.Min.Y+1, ar.Max.X, ar.Max.Y)
	if body.Dx() < 3 || body.Dy() < 1 {
		return draw.ResizeNeeded(cvs)
	}

	if s.opts.primaryLabel != "" {
		if err := draw.Text(cvs, s.opts.primaryLabel, image.Point{X: ar.Min.X, Y: labelRow}, draw.TextCellOpts(s.opts.labelCellOpts...)); err != nil {
			return err
		}
	}
	if s.opts.secondLabel != "" {
		start := ar.Max.X - runewidth.StringWidth(s.opts.secondLabel)
		if start < ar.Min.X {
			start = ar.Min.X
		}
		if err := draw.Text(cvs, s.opts.secondLabel, image.Point{X: start, Y: labelRow}, draw.TextCellOpts(s.opts.labelCellOpts...)); err != nil {
			return err
		}
	}

	// Half-duplex mode expands a single channel to the right of the axis.
	if s.mode == ModeHalfDuplex {
		axisX := body.Min.X
		s.lastSpan = body.Dy()
		if err := s.drawVerticalAxis(cvs, axisX, body.Min.Y, body.Max.Y); err != nil {
			return err
		}
		maxValue := maxSample(s.opts.maxValue, s.primary)
		visible := fitSamples(s.primary, body.Dy())
		startY := body.Min.Y
		width := body.Max.X - axisX - 1
		thresholdCells := s.thresholdCells(maxValue, width)
		for i, sample := range visible {
			cells := scaledCells(sample, maxValue, width)
			y := startY + i
			for level := 0; level < cells; level++ {
				x := axisX + 1 + level
				if x >= body.Max.X {
					break
				}
				color := gradientColor(s.opts.gradient, cellRatio(level+1, width))
				color = s.thresholdColor(color, level+1, thresholdCells)
				r := s.opts.halfRune
				if level == cells-1 {
					r = s.opts.primaryPeak
				}
				if _, err := cvs.SetCell(image.Point{X: x, Y: y}, r, cell.FgColor(color)); err != nil {
					return err
				}
			}
		}
		if err := s.drawRightThreshold(cvs, axisX, body.Min.Y, body.Max.Y, thresholdCells); err != nil {
			return err
		}
		return nil
	}

	// Stereo mode mirrors primary and secondary channels around the axis.
	leftWidth := (body.Dx() - 1) / 2
	rightWidth := body.Dx() - 1 - leftWidth
	if leftWidth < 1 || rightWidth < 1 {
		return draw.ResizeNeeded(cvs)
	}
	axisX := body.Min.X + leftWidth
	s.lastSpan = body.Dy()
	if err := s.drawVerticalAxis(cvs, axisX, body.Min.Y, body.Max.Y); err != nil {
		return err
	}

	maxValue := maxSample(s.opts.maxValue, s.primary, s.secondary)
	primary := fitSamples(s.primary, body.Dy())
	secondary := fitSamples(s.secondary, body.Dy())
	leftThreshold := s.thresholdCells(maxValue, leftWidth)
	rightThreshold := s.thresholdCells(maxValue, rightWidth)
	for i := 0; i < body.Dy(); i++ {
		y := body.Min.Y + i
		if i < len(primary) {
			cells := scaledCells(primary[i], maxValue, leftWidth)
			for level := 0; level < cells; level++ {
				x := axisX - 1 - level
				if x < body.Min.X {
					break
				}
				color := gradientColor(s.opts.gradient, cellRatio(level+1, leftWidth))
				color = s.thresholdColor(color, level+1, leftThreshold)
				r := channelRune(level, cells, s.opts.horizRunes, horizontalBodyRune)
				if level == cells-1 {
					r = s.opts.primaryPeak
				}
				if _, err := cvs.SetCell(image.Point{X: x, Y: y}, r, cell.FgColor(color)); err != nil {
					return err
				}
			}
		}
		if i < len(secondary) {
			cells := scaledCells(secondary[i], maxValue, rightWidth)
			for level := 0; level < cells; level++ {
				x := axisX + 1 + level
				if x >= body.Max.X {
					break
				}
				color := gradientColor(s.opts.gradient, cellRatio(level+1, rightWidth))
				color = s.thresholdColor(color, level+1, rightThreshold)
				r := channelRune(level, cells, s.opts.horizRunes, horizontalBodyRune)
				if level == cells-1 {
					r = s.opts.secondPeak
				}
				if _, err := cvs.SetCell(image.Point{X: x, Y: y}, r, cell.FgColor(color)); err != nil {
					return err
				}
			}
		}
	}
	if err := s.drawLeftThreshold(cvs, axisX, body.Min.Y, body.Max.Y, leftThreshold); err != nil {
		return err
	}
	if err := s.drawRightThreshold(cvs, axisX, body.Min.Y, body.Max.Y, rightThreshold); err != nil {
		return err
	}
	return nil
}

// drawHorizontalAxis draws a one-cell-thick horizontal reference axis.
func (s *Spectrum) drawHorizontalAxis(cvs *canvas.Canvas, minX, maxX, y int) error {
	for x := minX; x < maxX; x++ {
		if _, err := cvs.SetCell(image.Point{X: x, Y: y}, '─', s.opts.axisCellOpts...); err != nil {
			return err
		}
	}
	return nil
}

// drawVerticalAxis draws a one-cell-thick vertical reference axis.
func (s *Spectrum) drawVerticalAxis(cvs *canvas.Canvas, x, minY, maxY int) error {
	for y := minY; y < maxY; y++ {
		if _, err := cvs.SetCell(image.Point{X: x, Y: y}, '│', s.opts.axisCellOpts...); err != nil {
			return err
		}
	}
	return nil
}

// drawUpperThreshold draws a threshold marker above a horizontal axis.
func (s *Spectrum) drawUpperThreshold(cvs *canvas.Canvas, minX, maxX, axisY, thresholdCells int) error {
	if thresholdCells <= 0 {
		return nil
	}
	y := axisY - thresholdCells
	if y < cvs.Area().Min.Y || y >= axisY {
		return nil
	}
	return s.drawThresholdRow(cvs, minX, maxX, y)
}

// drawLowerThreshold draws a threshold marker below a horizontal axis.
func (s *Spectrum) drawLowerThreshold(cvs *canvas.Canvas, minX, maxX, axisY, thresholdCells int) error {
	if thresholdCells <= 0 {
		return nil
	}
	y := axisY + thresholdCells
	if y <= axisY || y >= cvs.Area().Max.Y {
		return nil
	}
	return s.drawThresholdRow(cvs, minX, maxX, y)
}

// drawLeftThreshold draws a threshold marker to the left of a vertical axis.
func (s *Spectrum) drawLeftThreshold(cvs *canvas.Canvas, axisX, minY, maxY, thresholdCells int) error {
	if thresholdCells <= 0 {
		return nil
	}
	x := axisX - thresholdCells
	if x < cvs.Area().Min.X || x >= axisX {
		return nil
	}
	return s.drawThresholdColumn(cvs, x, minY, maxY)
}

// drawRightThreshold draws a threshold marker to the right of a vertical axis.
func (s *Spectrum) drawRightThreshold(cvs *canvas.Canvas, axisX, minY, maxY, thresholdCells int) error {
	if thresholdCells <= 0 {
		return nil
	}
	x := axisX + thresholdCells
	if x <= axisX || x >= cvs.Area().Max.X {
		return nil
	}
	return s.drawThresholdColumn(cvs, x, minY, maxY)
}

// drawThresholdRow draws a red horizontal threshold indicator.
func (s *Spectrum) drawThresholdRow(cvs *canvas.Canvas, minX, maxX, y int) error {
	for x := minX; x < maxX; x++ {
		if _, err := cvs.SetCell(image.Point{X: x, Y: y}, '─', cell.FgColor(s.opts.thresholdLine)); err != nil {
			return err
		}
	}
	return nil
}

// drawThresholdColumn draws a red vertical threshold indicator.
func (s *Spectrum) drawThresholdColumn(cvs *canvas.Canvas, x, minY, maxY int) error {
	for y := minY; y < maxY; y++ {
		if _, err := cvs.SetCell(image.Point{X: x, Y: y}, '│', cell.FgColor(s.opts.thresholdLine)); err != nil {
			return err
		}
	}
	return nil
}

// thresholdCells maps the configured threshold onto rendered cells.
func (s *Spectrum) thresholdCells(max, span int) int {
	if s.opts.threshold <= 0 || max <= 0 || span <= 0 {
		return 0
	}
	return scaledCells(s.opts.threshold, max, span)
}

// thresholdColor returns the alert color when level is at or above threshold.
func (s *Spectrum) thresholdColor(defaultColor cell.Color, level, thresholdCells int) cell.Color {
	if thresholdCells > 0 && level >= thresholdCells {
		return s.opts.alertColor
	}
	return defaultColor
}

// area resolves the drawable region, honoring a fixed height when configured.
func (s *Spectrum) area(cvs *canvas.Canvas) image.Rectangle {
	return s.areaForRect(cvs.Area())
}

// areaForRect resolves the drawable region within ar.
func (s *Spectrum) areaForRect(ar image.Rectangle) image.Rectangle {
	if s.opts.height <= 0 {
		return ar
	}
	maxY := ar.Max.Y
	return image.Rect(ar.Min.X, maxY-s.opts.height, ar.Max.X, maxY)
}

// valueAt maps a canvas-relative point to the nearest visible sample.
func (s *Spectrum) valueAt(ar image.Rectangle, pos image.Point) (Sample, bool) {
	ar = s.areaForRect(ar)
	if !pos.In(ar) {
		return Sample{}, false
	}
	if s.opts.orientation == OrientationHorizontal {
		return s.horizontalValueAt(ar, pos)
	}
	return s.verticalValueAt(ar, pos)
}

// verticalValueAt maps a vertical spectrum point to a sample.
func (s *Spectrum) verticalValueAt(ar image.Rectangle, pos image.Point) (Sample, bool) {
	body := image.Rect(ar.Min.X, ar.Min.Y+1, ar.Max.X, ar.Max.Y-1)
	if body.Dy() < 2 || !pos.In(body) {
		return Sample{}, false
	}

	if s.mode == ModeHalfDuplex {
		return s.verticalHalfDuplexValueAt(body, pos)
	}

	upperHeight := (body.Dy() - 1) / 2
	lowerHeight := body.Dy() - 1 - upperHeight
	if upperHeight < 1 || lowerHeight < 1 {
		return Sample{}, false
	}
	axisY := body.Min.Y + upperHeight
	switch {
	case pos.Y < axisY:
		return s.verticalPrimaryValueAt(body, pos, axisY, upperHeight)
	case pos.Y > axisY:
		return s.verticalSecondaryValueAt(body, pos, axisY, lowerHeight)
	default:
		return Sample{}, false
	}
}

// horizontalValueAt maps a horizontal spectrum point to a sample.
func (s *Spectrum) horizontalValueAt(ar image.Rectangle, pos image.Point) (Sample, bool) {
	body := image.Rect(ar.Min.X, ar.Min.Y+1, ar.Max.X, ar.Max.Y)
	if body.Dx() < 3 || body.Dy() < 1 || !pos.In(body) {
		return Sample{}, false
	}

	if s.mode == ModeHalfDuplex {
		return s.horizontalHalfDuplexValueAt(body, pos)
	}

	leftWidth := (body.Dx() - 1) / 2
	rightWidth := body.Dx() - 1 - leftWidth
	if leftWidth < 1 || rightWidth < 1 {
		return Sample{}, false
	}
	axisX := body.Min.X + leftWidth
	switch {
	case pos.X < axisX:
		return s.horizontalPrimaryValueAt(body, pos, axisX, leftWidth)
	case pos.X > axisX:
		return s.horizontalSecondaryValueAt(body, pos, axisX, rightWidth)
	default:
		return Sample{}, false
	}
}

// verticalHalfDuplexValueAt returns a visible sample only when pos overlaps a
// rendered half-duplex column.
func (s *Spectrum) verticalHalfDuplexValueAt(body image.Rectangle, pos image.Point) (Sample, bool) {
	axisY := body.Max.Y - 1
	height := axisY - body.Min.Y
	sample, ok := sampleAtColumn(s.primary, body.Dx(), pos.X-body.Min.X)
	if !ok {
		return Sample{}, false
	}
	cells := scaledCells(sample.Y, maxSample(s.opts.maxValue, s.primary), height)
	if cells == 0 {
		return Sample{}, false
	}
	minY := axisY - cells
	return sample, pos.Y >= minY && pos.Y < axisY
}

// verticalPrimaryValueAt returns a visible primary-channel sample only when pos
// overlaps the rendered upper signal.
func (s *Spectrum) verticalPrimaryValueAt(body image.Rectangle, pos image.Point, axisY, upperHeight int) (Sample, bool) {
	sample, ok := sampleAtColumn(s.primary, body.Dx(), pos.X-body.Min.X)
	if !ok {
		return Sample{}, false
	}
	cells := scaledCells(sample.Y, maxSample(s.opts.maxValue, s.primary, s.secondary), upperHeight)
	if cells == 0 {
		return Sample{}, false
	}
	minY := axisY - cells
	return sample, pos.Y >= minY && pos.Y < axisY
}

// verticalSecondaryValueAt returns a visible secondary-channel sample only when
// pos overlaps the rendered lower signal.
func (s *Spectrum) verticalSecondaryValueAt(body image.Rectangle, pos image.Point, axisY, lowerHeight int) (Sample, bool) {
	sample, ok := sampleAtColumn(s.secondary, body.Dx(), pos.X-body.Min.X)
	if !ok {
		return Sample{}, false
	}
	cells := scaledCells(sample.Y, maxSample(s.opts.maxValue, s.primary, s.secondary), lowerHeight)
	if cells == 0 {
		return Sample{}, false
	}
	maxY := axisY + cells
	return sample, pos.Y > axisY && pos.Y <= maxY
}

// horizontalHalfDuplexValueAt returns a visible sample only when pos overlaps
// a rendered half-duplex row.
func (s *Spectrum) horizontalHalfDuplexValueAt(body image.Rectangle, pos image.Point) (Sample, bool) {
	axisX := body.Min.X
	width := body.Max.X - axisX - 1
	sample, ok := sampleAtColumn(s.primary, body.Dy(), pos.Y-body.Min.Y)
	if !ok {
		return Sample{}, false
	}
	cells := scaledCells(sample.Y, maxSample(s.opts.maxValue, s.primary), width)
	if cells == 0 {
		return Sample{}, false
	}
	maxX := axisX + cells
	return sample, pos.X > axisX && pos.X <= maxX
}

// horizontalPrimaryValueAt returns a visible primary-channel sample only when
// pos overlaps the rendered left-side signal.
func (s *Spectrum) horizontalPrimaryValueAt(body image.Rectangle, pos image.Point, axisX, leftWidth int) (Sample, bool) {
	sample, ok := sampleAtColumn(s.primary, body.Dy(), pos.Y-body.Min.Y)
	if !ok {
		return Sample{}, false
	}
	cells := scaledCells(sample.Y, maxSample(s.opts.maxValue, s.primary, s.secondary), leftWidth)
	if cells == 0 {
		return Sample{}, false
	}
	minX := axisX - cells
	return sample, pos.X >= minX && pos.X < axisX
}

// horizontalSecondaryValueAt returns a visible secondary-channel sample only
// when pos overlaps the rendered right-side signal.
func (s *Spectrum) horizontalSecondaryValueAt(body image.Rectangle, pos image.Point, axisX, rightWidth int) (Sample, bool) {
	sample, ok := sampleAtColumn(s.secondary, body.Dy(), pos.Y-body.Min.Y)
	if !ok {
		return Sample{}, false
	}
	cells := scaledCells(sample.Y, maxSample(s.opts.maxValue, s.primary, s.secondary), rightWidth)
	if cells == 0 {
		return Sample{}, false
	}
	maxX := axisX + cells
	return sample, pos.X > axisX && pos.X <= maxX
}

// minSize returns the minimum canvas size needed for the current configuration.
func (s *Spectrum) minSize() image.Point {
	minHeight := 4
	if s.opts.orientation == OrientationVertical && s.mode == ModeStereo {
		minHeight = 5
	}
	if s.opts.height > 0 && s.opts.height > minHeight {
		minHeight = s.opts.height
	}
	return image.Point{X: 3, Y: minHeight}
}

// validateSamples rejects negative samples before they reach the render path.
func validateSamples(values []int) error {
	for i, v := range values {
		if v < 0 {
			return fmt.Errorf("sample[%d]: %d must be >= 0", i, v)
		}
	}
	return nil
}

// tailInts returns the last max values to fit the currently visible span.
func tailInts(values []int, max int) []int {
	if max <= 0 || len(values) == 0 {
		return nil
	}
	if len(values) <= max {
		return values
	}
	return values[len(values)-max:]
}

// fitSamples adapts the provided samples to exactly span cells by interpolation.
func fitSamples(values []int, span int) []int {
	if span <= 0 || len(values) == 0 {
		return nil
	}
	if len(values) >= span {
		return tailInts(values, span)
	}
	if span == 1 {
		return []int{values[len(values)-1]}
	}
	out := make([]int, span)
	last := len(values) - 1
	for i := range out {
		idx := i * last / (span - 1)
		out[i] = values[idx]
	}
	return out
}

// sampleAtColumn returns the fitted sample at a visible column or row.
func sampleAtColumn(values []int, span, index int) (Sample, bool) {
	if index < 0 || index >= span {
		return Sample{}, false
	}
	visible := fitSamples(values, span)
	if index >= len(visible) {
		return Sample{}, false
	}
	base := 0
	if len(values) > span {
		base = len(values) - span
	}
	return Sample{X: base + index + 1, Y: visible[index]}, true
}

// copyInts reuses dst capacity when possible to avoid per-frame allocations.
func copyInts(dst, src []int) []int {
	if cap(dst) < len(src) {
		dst = make([]int, len(src))
	} else {
		dst = dst[:len(src)]
	}
	copy(dst, src)
	return dst
}

// maxSample returns either a fixed configured max or the observed max sample.
func maxSample(fixed int, series ...[]int) int {
	if fixed > 0 {
		return fixed
	}
	maximum := 1
	for _, values := range series {
		for _, v := range values {
			if v > maximum {
				maximum = v
			}
		}
	}
	return maximum
}

// scaledCells maps a sample value onto a number of rendered cells.
func scaledCells(v, max, span int) int {
	if v <= 0 || max <= 0 || span <= 0 {
		return 0
	}
	if v >= max {
		return span
	}
	return int(math.Ceil(float64(v) * float64(span) / float64(max)))
}

// gradientColor picks a color from a low-to-high gradient by normalized ratio.
func gradientColor(colors []cell.Color, ratio float64) cell.Color {
	if len(colors) == 0 {
		return cell.ColorGreen
	}
	if len(colors) == 1 {
		return colors[0]
	}
	if ratio <= 0 {
		return colors[0]
	}
	if ratio >= 1 {
		return colors[len(colors)-1]
	}
	index := int(ratio * float64(len(colors)-1))
	if index >= len(colors) {
		index = len(colors) - 1
	}
	return colors[index]
}

// cellRatio converts a one-based level into a normalized [0,1] ratio.
func cellRatio(level, span int) float64 {
	if span <= 1 {
		return 1
	}
	return float64(level-1) / float64(span-1)
}

// upperRune selects the visual glyph for the primary vertical channel.
func upperRune(level, cells int) rune {
	switch {
	case cells <= 1:
		return 'i'
	case level == 0:
		return 'i'
	case level >= cells-2:
		return '!'
	default:
		return '|'
	}
}

// lowerRune selects the visual glyph for the secondary vertical channel.
func lowerRune(level, cells int) rune {
	switch {
	case cells <= 1:
		return 'i'
	case level == 0:
		return 'i'
	case level >= cells-2:
		return '¡'
	default:
		return '|'
	}
}

// horizontalBodyRune selects the body glyph for horizontal channels.
func horizontalBodyRune(level, cells int) rune {
	switch {
	case cells <= 1:
		return '-'
	case level >= cells-2:
		return '='
	default:
		return '-'
	}
}

// channelRune selects a body rune from a custom amplitude scale or fallback.
func channelRune(level, cells int, runes []rune, fallback func(int, int) rune) rune {
	if len(runes) == 0 {
		return fallback(level, cells)
	}
	if len(runes) == 1 || cells <= 1 {
		return runes[0]
	}
	ratio := cellRatio(level+1, cells)
	idx := int(math.Ceil(ratio * float64(len(runes)-1)))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(runes) {
		idx = len(runes) - 1
	}
	return runes[idx]
}

// maxInt returns the larger of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
