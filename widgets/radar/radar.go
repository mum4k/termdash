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

// Package radar is a widget that renders an animated radar sweep display.
//
// The widget draws a rotating beam over a circular grid of range rings and
// crosshairs, reproducing the classic phosphor-persistence look of a
// monochrome PPI radar scope. The sweep trail fades from the beam's leading
// edge back through a configurable angular width.
//
// Contact points are placed on the scope by the caller. Each contact dims
// gradually as the beam sweeps away and brightens again on the next pass,
// matching real radar refresh behaviour.
//
// The sweep speed, beam width, direction (CW / CCW), total arc, and colours
// are all configurable through Option values.
package radar

import (
	"errors"
	"fmt"
	"image"
	"math"
	"sync"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/canvas/braille"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

// Contact represents a single radar return (blip) displayed on the scope.
type Contact struct {
	// Angle is the bearing in degrees, measured clockwise from North
	// (0 = North, 90 = East, 180 = South, 270 = West).
	Angle float64

	// Distance is the range expressed as a fraction of the display radius
	// (0.0 = at the origin, 1.0 = at the outer ring edge).
	Distance float64

	// Elevation is the contact's altitude in feet above sea level.
	Elevation float64

	// Label is a short human-readable identifier shown in the contact log
	// and in the hover tooltip.
	Label string

	// Info is an optional extra line of free-form text appended to the
	// hover tooltip (e.g. "AIR · 350 kt"). Leave empty to omit.
	Info string

	// intensity is the current display brightness of this contact [0.0, 1.0].
	// It is managed automatically by the widget based on beam position.
	intensity float64
}

// Radar renders an animated radar sweep display on the terminal.
//
// Implements widgetapi.Widget. This object is thread-safe.
type Radar struct {
	// contacts are the radar returns currently tracked on the scope.
	contacts []*Contact

	// currentAngle is the leading-edge angle of the sweep beam in degrees.
	currentAngle float64

	// sweepDir is +1.0 for clockwise and −1.0 for counter-clockwise motion.
	sweepDir float64

	// lastDraw records the wall-clock time of the most recent Draw call so
	// that the beam angle can be advanced proportionally to elapsed time.
	lastDraw time.Time

	// mousePos is the last recorded mouse position in absolute terminal coords.
	mousePos image.Point
	// hasMousePos is true once at least one mouse event has been received.
	hasMousePos bool

	// mu protects all mutable state.
	mu sync.Mutex

	// opts are the widget's configuration options.
	opts *options
}

// New returns a new Radar widget configured with the provided options.
// The widget is ready to draw immediately; call SetContacts to populate it.
func New(opts ...Option) (*Radar, error) {
	o := newOptions()
	for _, opt := range opts {
		opt.set(o)
	}
	if err := o.validate(); err != nil {
		return nil, err
	}

	dir := 1.0
	if o.direction == DirectionCounterClockwise {
		dir = -1.0
	}

	return &Radar{
		opts:         o,
		sweepDir:     dir,
		currentAngle: o.startAngle,
	}, nil
}

// SetContacts replaces the set of contact points displayed on the scope.
// Provided options override values set when New was called.
//
// Passing nil or an empty slice clears all contacts.
func (r *Radar) SetContacts(contacts []*Contact, opts ...Option) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, opt := range opts {
		opt.set(r.opts)
	}
	if err := r.opts.validate(); err != nil {
		return err
	}
	r.contacts = contacts
	return nil
}

// ─── Internal geometry helpers ────────────────────────────────────────────────

// radarMidAndRadius returns the braille-pixel centre point and the usable
// display radius (in braille pixels) for the given canvas area.
//
// The braille canvas is ColMult (2) × wider and RowMult (4) × taller than the
// cell canvas, giving sub-character resolution. Because a typical terminal cell
// is ≈ 2 × taller than wide, the braille sub-pixels end up approximately
// square, so equal X and Y radii produce a true circle.
func radarMidAndRadius(ar image.Rectangle) (image.Point, int) {
	bw := ar.Dx() * braille.ColMult
	bh := ar.Dy() * braille.RowMult
	mid := image.Point{
		X: ar.Min.X*braille.ColMult + bw/2,
		Y: ar.Min.Y*braille.RowMult + bh/2,
	}
	// Choose the smaller half-dimension and leave a small margin.
	radius := bw/2 - 3
	if half := bh/2 - 3; half < radius {
		radius = half
	}
	if radius < 1 {
		radius = 1
	}
	return mid, radius
}

// bearingToPoint converts a compass bearing (degrees CW from North) and a
// radius (braille pixels) into an absolute braille-canvas point.
func bearingToPoint(mid image.Point, radius int, deg float64) image.Point {
	rad := deg * math.Pi / 180.0
	return image.Point{
		X: mid.X + int(math.Round(float64(radius)*math.Sin(rad))),
		Y: mid.Y - int(math.Round(float64(radius)*math.Cos(rad))),
	}
}

// normalizeAngle maps an arbitrary angle into the half-open interval [0, 360).
func normalizeAngle(a float64) float64 {
	a = math.Mod(a, 360.0)
	if a < 0 {
		a += 360.0
	}
	return a
}

// angleDiffCW returns the clockwise angular distance from `from` to `to`,
// always in [0, 360).
func angleDiffCW(from, to float64) float64 {
	return normalizeAngle(to - from)
}

// clampRGB clamps an integer to the [0, 255] range required by RGB24 colours.
func clampRGB(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// ─── State updaters ────────────────────────────────────────────────────────────

// advanceSweep moves the beam forward by the angle appropriate for the elapsed
// wall-clock seconds. For partial sweeps (SweepSpan < 360) the beam bounces
// back at the arc's endpoints.
func (r *Radar) advanceSweep(elapsed float64) {
	delta := r.opts.sweepSpeed * elapsed * r.sweepDir
	r.currentAngle += delta

	if r.opts.sweepSpan < 360.0 {
		lo := r.opts.startAngle
		hi := r.opts.startAngle + r.opts.sweepSpan
		if r.currentAngle > hi {
			r.currentAngle = hi - (r.currentAngle - hi)
			r.sweepDir = -r.sweepDir
		} else if r.currentAngle < lo {
			r.currentAngle = lo + (lo - r.currentAngle)
			r.sweepDir = -r.sweepDir
		}
	}

	r.currentAngle = normalizeAngle(r.currentAngle)
}

// updateContactIntensities recalculates the display brightness of every
// contact based on how far the beam has swept past it since the last pass.
//
// A contact is at full brightness (1.0) the instant the beam crosses its
// bearing and fades to zero over one full sweep period. A gentle power curve
// gives the familiar slow initial fade followed by a more abrupt drop-off.
func (r *Radar) updateContactIntensities() {
	period := r.opts.sweepSpan
	for _, c := range r.contacts {
		var age float64
		if r.sweepDir >= 0 {
			// CW: age = how far (degrees) the beam has moved past the contact.
			age = angleDiffCW(c.Angle, r.currentAngle)
		} else {
			// CCW: reverse the "past" direction.
			age = angleDiffCW(r.currentAngle, c.Angle)
		}
		t := age / period
		if t > 1.0 {
			t = 1.0
		}
		// Power-curve fade: stays bright longer, then drops quickly.
		c.intensity = math.Pow(1.0-t, 1.4)
	}
}

// ─── Colour helpers ────────────────────────────────────────────────────────────

// trailColor returns the RGB24 colour for a given position in the fade trail.
// step=0 is the freshest (trailing just behind the beam edge); step=numSteps-1
// is the oldest and dimmest. The fade uses a power curve for a natural
// phosphor-decay look.
func (r *Radar) trailColor(step, numSteps int) cell.Color {
	br, bg, bb := r.opts.beamColorRGB()
	t := 1.0 - float64(step)/float64(numSteps)
	t = math.Pow(t, 2.0) // Quadratic: bright near the beam, fast drop-off.
	return cell.ColorRGB24(
		clampRGB(int(float64(br)*t)),
		clampRGB(int(float64(bg)*t)),
		clampRGB(int(float64(bb)*t)),
	)
}

// contactDisplayColor returns the RGB24 colour for a contact blip scaled by
// the contact's current intensity.
func (r *Radar) contactDisplayColor(intensity float64) cell.Color {
	cr, cg, cb := r.opts.contactColorRGB()
	// Contacts use a softer minimum so they remain faintly visible even when old.
	t := 0.12 + intensity*0.88
	return cell.ColorRGB24(
		clampRGB(int(float64(cr)*t)),
		clampRGB(int(float64(cg)*t)),
		clampRGB(int(float64(cb)*t)),
	)
}

// beamFrontColor returns the bright leading-edge colour (slightly whiter than
// the base beam colour).
func (r *Radar) beamFrontColor() cell.Color {
	br, bg, bb := r.opts.beamColorRGB()
	return cell.ColorRGB24(
		clampRGB(br+45),
		clampRGB(bg+25),
		clampRGB(bb+20),
	)
}

// ─── Tooltip helpers ──────────────────────────────────────────────────────────

// tooltipBearing formats an angle as a zero-padded three-digit bearing string.
func tooltipBearing(deg float64) string {
	d := int(math.Round(deg)) % 360
	if d < 0 {
		d += 360
	}
	return fmt.Sprintf("%03d°", d)
}

// tooltipElevation formats an altitude in feet, with comma-separated thousands.
func tooltipElevation(ft float64) string {
	i := int(math.Round(ft))
	if i < 1000 {
		return fmt.Sprintf("%d ft", i)
	}
	return fmt.Sprintf("%d,%03d ft", i/1000, i%1000)
}

// runeLen returns the number of runes in s (i.e. character count, not byte count).
func runeLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

// drawTooltip renders a popup information box near the contact at cell
// position (contactX, contactY). It is called from Draw while the mutex is held.
func (r *Radar) drawTooltip(cvs *canvas.Canvas, ar image.Rectangle, contactX, contactY int, c *Contact) {
	// ── Build content lines ───────────────────────────────────────────────────
	lines := []string{
		fmt.Sprintf("%c  %s", r.opts.contactChar, c.Label),
		fmt.Sprintf("BRG %s  ·  RNG %d%%", tooltipBearing(c.Angle), int(c.Distance*100)),
		fmt.Sprintf("ALT %s", tooltipElevation(c.Elevation)),
	}
	if c.Info != "" {
		lines = append(lines, c.Info)
	}

	// ── Compute box dimensions ────────────────────────────────────────────────
	maxContent := 0
	for _, l := range lines {
		if n := runeLen(l); n > maxContent {
			maxContent = n
		}
	}
	// innerW: 1 space padding each side + content.
	innerW := maxContent + 2
	boxW := innerW + 2 // +2 for left and right border characters.
	boxH := len(lines) + 2 // +2 for top and bottom border characters.

	// ── Position: right of contact, vertically centred ────────────────────────
	tx := contactX + 2
	ty := contactY - boxH/2

	// Prefer right; fall back to left if it overflows.
	if tx+boxW > ar.Max.X {
		tx = contactX - boxW - 1
	}
	// Hard-clamp to canvas bounds.
	if tx < ar.Min.X {
		tx = ar.Min.X
	}
	if ty < ar.Min.Y {
		ty = ar.Min.Y
	}
	if ty+boxH > ar.Max.Y {
		ty = ar.Max.Y - boxH
	}

	boxRect := image.Rect(tx, ty, tx+boxW, ty+boxH)
	if !boxRect.In(ar) {
		return // Not enough room — skip the tooltip rather than panic.
	}

	// ── Colours ───────────────────────────────────────────────────────────────
	bgColor := r.opts.tooltipBgColor()
	borderColor := r.opts.tooltipBorderColor()
	labelColor := r.opts.tooltipLabelColor()
	dataColor := r.opts.tooltipDataColor()

	bgOpt := cell.BgColor(bgColor)

	// ── Fill background ───────────────────────────────────────────────────────
	// draw.Rectangle fills with spaces; we rely on BgColor to paint the backdrop.
	_ = draw.Rectangle(cvs, boxRect,
		draw.RectChar(' '),
		draw.RectCellOpts(bgOpt))

	// ── Draw rounded border on top ────────────────────────────────────────────
	_ = draw.Border(cvs, boxRect,
		draw.BorderLineStyle(linestyle.Round),
		draw.BorderCellOpts(cell.FgColor(borderColor), bgOpt))

	// ── Draw text lines inside the border ─────────────────────────────────────
	// Text starts one cell inside the left border; maxX stops one cell before
	// the right border.
	maxX := tx + boxW - 1
	for i, line := range lines {
		pt := image.Point{X: tx + 1, Y: ty + 1 + i}
		if !pt.In(ar) {
			continue
		}
		var fgColor cell.Color
		if i == 0 {
			fgColor = labelColor // callsign line stands out in contact colour
		} else {
			fgColor = dataColor
		}
		_ = draw.Text(cvs, " "+line, pt,
			draw.TextCellOpts(cell.FgColor(fgColor), bgOpt),
			draw.TextMaxX(maxX),
			draw.TextOverrunMode(draw.OverrunModeTrim),
		)
	}
}

// ─── Draw ─────────────────────────────────────────────────────────────────────

// Draw renders the radar scope onto the canvas. It advances the sweep beam
// according to the time elapsed since the previous frame and recomputes
// contact intensities before rendering.
//
// Draw implements widgetapi.Widget.Draw.
func (r *Radar) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// ── Advance simulation time ──────────────────────────────────────────────
	now := time.Now()
	if !r.lastDraw.IsZero() {
		r.advanceSweep(now.Sub(r.lastDraw).Seconds())
	}
	r.lastDraw = now
	r.updateContactIntensities()

	ar := cvs.Area()

	// ── Minimum-size guard ───────────────────────────────────────────────────
	if ar.Dx() < 10 || ar.Dy() < 5 {
		return draw.ResizeNeeded(cvs)
	}

	// ── Optional widget-level border ─────────────────────────────────────────
	// When a border style is configured the border is drawn on the outer edge
	// of the cell canvas and the drawing area is inset by one cell on each side
	// so that the scope content never overlaps the border characters.
	if r.opts.border != linestyle.None {
		if err := draw.Border(cvs, ar,
			draw.BorderLineStyle(r.opts.border),
			draw.BorderTitle(r.opts.borderTitle, draw.OverrunModeThreeDot, r.opts.borderCellOpts...),
			draw.BorderCellOpts(r.opts.borderCellOpts...),
		); err != nil {
			return fmt.Errorf("draw.Border => %v", err)
		}
		ar = image.Rect(ar.Min.X+1, ar.Min.Y+1, ar.Max.X-1, ar.Max.Y-1)
	}

	// ── Geometry ─────────────────────────────────────────────────────────────
	mid, maxRadius := radarMidAndRadius(ar)

	// Cell-coordinate centre for placing character-level decorations.
	cellCX := ar.Min.X + ar.Dx()/2
	cellCY := ar.Min.Y + ar.Dy()/2

	// ── Braille canvas ───────────────────────────────────────────────────────
	bc, err := braille.New(ar)
	if err != nil {
		return fmt.Errorf("braille.New => %v", err)
	}

	dimColor := r.opts.dimColor()

	// ── Range rings ──────────────────────────────────────────────────────────
	// Drawn first so that everything else overlays on top of them.
	for i := 1; i <= r.opts.rangeRings; i++ {
		ringR := maxRadius * i / r.opts.rangeRings
		if err := draw.BrailleCircle(bc, mid, ringR,
			draw.BrailleCircleCellOpts(cell.FgColor(dimColor))); err != nil {
			// A non-fatal geometry error (e.g. radius == 0 at tiny sizes).
			// Continue drawing remaining elements.
			_ = err
		}
	}

	// ── Crosshairs ───────────────────────────────────────────────────────────
	north := bearingToPoint(mid, maxRadius, 0)
	south := bearingToPoint(mid, maxRadius, 180)
	east := bearingToPoint(mid, maxRadius, 90)
	west := bearingToPoint(mid, maxRadius, 270)
	_ = draw.BrailleLine(bc, north, south, draw.BrailleLineCellOpts(cell.FgColor(dimColor)))
	_ = draw.BrailleLine(bc, east, west, draw.BrailleLineCellOpts(cell.FgColor(dimColor)))

	// ── Sweep trail (phosphor fade) ───────────────────────────────────────────
	// We render the trail from oldest to newest so that each successive step
	// overrides the previous one where braille cells overlap, ensuring the
	// pixel nearest the beam edge is always the brightest.
	const numTrailSteps = 28
	bw := r.opts.beamWidth

	for step := numTrailSteps - 1; step >= 0; step-- {
		// Angle at this trail step (behind the beam front).
		trailAngle := r.currentAngle - r.sweepDir*bw*(float64(step+1)/float64(numTrailSteps))
		trailAngle = normalizeAngle(trailAngle)
		endPt := bearingToPoint(mid, maxRadius, trailAngle)
		color := r.trailColor(step, numTrailSteps)
		_ = draw.BrailleLine(bc, mid, endPt,
			draw.BrailleLineCellOpts(cell.FgColor(color)))
	}

	// ── Beam leading edge ─────────────────────────────────────────────────────
	beamEnd := bearingToPoint(mid, maxRadius, r.currentAngle)
	_ = draw.BrailleLine(bc, mid, beamEnd,
		draw.BrailleLineCellOpts(cell.FgColor(r.beamFrontColor())))

	// ── Copy braille layer to cell canvas ─────────────────────────────────────
	if err := bc.CopyTo(cvs); err != nil {
		return err
	}

	// ── Compass labels ────────────────────────────────────────────────────────
	// Placed on the cell canvas after the braille copy so they are never
	// overwritten by the sweep trail.
	compassOpt := cell.FgColor(r.opts.compassColor())
	compassPts := []struct {
		pt image.Point
		ch rune
	}{
		{image.Point{cellCX, ar.Min.Y}, 'N'},
		{image.Point{cellCX, ar.Max.Y - 1}, 'S'},
		{image.Point{ar.Max.X - 1, cellCY}, 'E'},
		{image.Point{ar.Min.X, cellCY}, 'W'},
	}
	for _, cp := range compassPts {
		if cp.pt.In(ar) {
			_, _ = cvs.SetCell(cp.pt, cp.ch, compassOpt)
		}
	}

	// ── Contact blips & hover detection ──────────────────────────────────────
	// Contact positions are rendered as character cells (not braille pixels) so
	// that the configured rune is visible at full glyph resolution.
	// We also detect which contact (if any) the mouse cursor is hovering over
	// so we can render a tooltip after all blips are drawn.
	maxCellRX := float64(maxRadius) / float64(braille.ColMult)
	maxCellRY := float64(maxRadius) / float64(braille.RowMult)

	var hoveredContact *Contact
	var hoveredCX, hoveredCY int

	for _, c := range r.contacts {
		if c.intensity < 0.02 {
			continue // Too dim to bother drawing.
		}
		angleRad := c.Angle * math.Pi / 180.0

		// Convert from braille-pixel radius to cell-grid offsets, correcting
		// for the ColMult / RowMult cell-to-pixel ratio.
		cx := cellCX + int(math.Round(c.Distance*maxCellRX*math.Sin(angleRad)))
		cy := cellCY - int(math.Round(c.Distance*maxCellRY*math.Cos(angleRad)))
		pt := image.Point{X: cx, Y: cy}
		if !pt.In(ar) {
			continue
		}
		_, _ = cvs.SetCell(pt, r.opts.contactChar,
			cell.FgColor(r.contactDisplayColor(c.intensity)))

		// Hover check: mouse within a 1-cell manhattan radius of the blip.
		if r.opts.tooltipEnabled && r.hasMousePos {
			dx := r.mousePos.X - cx
			dy := r.mousePos.Y - cy
			if dx >= -1 && dx <= 1 && dy >= -1 && dy <= 1 {
				hoveredContact = c
				hoveredCX = cx
				hoveredCY = cy
			}
		}
	}

	// ── Tooltip ───────────────────────────────────────────────────────────────
	// Drawn last so it always appears on top of all other content.
	if hoveredContact != nil {
		r.drawTooltip(cvs, ar, hoveredCX, hoveredCY, hoveredContact)
	}

	return nil
}

// ─── widgetapi.Widget implementation ──────────────────────────────────────────

// Keyboard implements widgetapi.Widget.Keyboard.
// The Radar widget does not process keyboard events.
func (r *Radar) Keyboard(k *terminalapi.Keyboard, meta *widgetapi.EventMeta) error {
	return errors.New("the Radar widget does not support keyboard events")
}

// Mouse implements widgetapi.Widget.Mouse.
// The widget records the cursor position on every event so that Draw can
// render a contact tooltip when the cursor is over a blip.
func (r *Radar) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mousePos = m.Position
	r.hasMousePos = true
	return nil
}

// Options implements widgetapi.Widget.Options.
func (r *Radar) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize:  image.Point{X: 10, Y: 5},
		WantKeyboard: widgetapi.KeyScopeNone,
		// MouseScopeWidget delivers every mouse event that occurs within the
		// widget's container, which is what we need for hover detection.
		WantMouse: widgetapi.MouseScopeWidget,
	}
}
