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

// Binary radardemo displays a live radar sweep with randomly generated
// contacts and a real-time contact log.
//
// Press Q or Escape to quit.
package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/radar"
	"github.com/mum4k/termdash/widgets/text"
)

// ─── Contact catalogue ─────────────────────────────────────────────────────────

// contactEntry holds a generated contact with display metadata.
type contactEntry struct {
	label     string
	angle     float64
	distance  float64
	elevation float64 // feet ASL
	speed     int     // knots (random, for flavour)
	class     string  // "AIR" | "SEA" | "UNK"
	firstSeen time.Time
}

// catalogue tracks all active contacts and provides thread-safe access.
type catalogue struct {
	mu      sync.Mutex
	entries []*contactEntry
	nextID  int
}

// ─── Contact generator ─────────────────────────────────────────────────────────

// natoCallsigns is the pool of NATO phonetic alphabet designators.
var natoCallsigns = []string{
	"ALPHA", "BRAVO", "CHARLIE", "DELTA", "ECHO", "FOXTROT",
	"GOLF", "HOTEL", "INDIA", "JULIET", "KILO", "LIMA",
	"MIKE", "NOVEMBER", "OSCAR", "PAPA", "QUEBEC", "ROMEO",
	"SIERRA", "TANGO", "UNIFORM", "VICTOR", "WHISKEY", "XRAY",
	"YANKEE", "ZULU",
}

var contactClasses = []string{"AIR", "SEA", "UNK"}
var classWeights = []int{60, 25, 15} // probability weights

func weightedClass() string {
	total := 0
	for _, w := range classWeights {
		total += w
	}
	r := rand.Intn(total)
	cumulative := 0
	for i, w := range classWeights {
		cumulative += w
		if r < cumulative {
			return contactClasses[i]
		}
	}
	return contactClasses[0]
}

// elevationForClass returns a realistic random altitude for the given class.
func elevationForClass(class string) float64 {
	switch class {
	case "AIR":
		// Commercial/military aircraft: 5,000–45,000 ft.
		return float64(5000 + rand.Intn(40000))
	case "SEA":
		// Surface vessels: effectively 0 ft (sea level).
		return float64(rand.Intn(100)) // slight noise for sensor jitter
	default:
		// Unknown: anywhere from sea-level up to high altitude.
		return float64(rand.Intn(30000))
	}
}

// speedForClass returns a plausible speed in knots.
func speedForClass(class string) int {
	switch class {
	case "AIR":
		return 250 + rand.Intn(400) // 250–650 kts
	case "SEA":
		return 5 + rand.Intn(30) // 5–35 kts
	default:
		return rand.Intn(200)
	}
}

// newContact creates a randomised contact entry.
func (cat *catalogue) newContact() *contactEntry {
	cat.nextID++
	idx := cat.nextID % len(natoCallsigns)
	label := fmt.Sprintf("%s-%02d", natoCallsigns[idx], cat.nextID)
	class := weightedClass()
	return &contactEntry{
		label:     label,
		angle:     rand.Float64() * 360.0,
		distance:  0.15 + rand.Float64()*0.80, // keep away from dead centre
		elevation: elevationForClass(class),
		speed:     speedForClass(class),
		class:     class,
		firstSeen: time.Now(),
	}
}

// add appends a new random contact to the catalogue.
func (cat *catalogue) add() {
	cat.mu.Lock()
	defer cat.mu.Unlock()
	cat.entries = append(cat.entries, cat.newContact())
}

// trim removes the oldest contacts, keeping at most max.
func (cat *catalogue) trim(max int) {
	cat.mu.Lock()
	defer cat.mu.Unlock()
	if len(cat.entries) > max {
		cat.entries = cat.entries[len(cat.entries)-max:]
	}
}

// toRadarContacts converts catalogue entries to radar.Contact values.
func (cat *catalogue) toRadarContacts() []*radar.Contact {
	cat.mu.Lock()
	defer cat.mu.Unlock()
	out := make([]*radar.Contact, len(cat.entries))
	for i, e := range cat.entries {
		out[i] = &radar.Contact{
			Angle:     e.angle,
			Distance:  e.distance,
			Elevation: e.elevation,
			Label:     e.label,
			Info:      fmt.Sprintf("%s  ·  %d kt", e.class, e.speed),
		}
	}
	return out
}

// snapshot returns a stable copy of the current entries (sorted by angle).
func (cat *catalogue) snapshot() []*contactEntry {
	cat.mu.Lock()
	defer cat.mu.Unlock()
	cp := make([]*contactEntry, len(cat.entries))
	copy(cp, cat.entries)
	sort.Slice(cp, func(i, j int) bool {
		return cp[i].angle < cp[j].angle
	})
	return cp
}

// ─── Contact log renderer ──────────────────────────────────────────────────────

// bearingStr formats an angle as a zero-padded 3-digit degree string.
func bearingStr(deg float64) string {
	d := int(math.Round(deg)) % 360
	if d < 0 {
		d += 360
	}
	return fmt.Sprintf("%03d°", d)
}

// elevStr formats an elevation in feet with comma-separated thousands.
func elevStr(ft float64) string {
	i := int(math.Round(ft))
	if i < 1000 {
		return fmt.Sprintf("%4d ft", i)
	}
	return fmt.Sprintf("%2d,%03d ft", i/1000, i%1000)
}

// classColor returns a suitable cell color for a contact class.
func classColor(class string) cell.Color {
	switch class {
	case "AIR":
		return cell.ColorRGB24(80, 200, 255) // sky blue
	case "SEA":
		return cell.ColorRGB24(80, 180, 120) // ocean green
	default:
		return cell.ColorRGB24(220, 220, 80) // yellow for unknown
	}
}

// renderContactLog rewrites the text widget with the current contact table.
func renderContactLog(t *text.Text, cat *catalogue) error {
	// ── Header ─────────────────────────────────────────────────────────────
	// WriteReplace() is passed on the first Write so the widget's entire
	// content is atomically replaced. The text widget rejects empty strings,
	// so we cannot use a separate clear call.
	headerStyle := []cell.Option{cell.FgColor(cell.ColorRGB24(0, 220, 60)), cell.Bold()}
	divider := "─────────────────────────────────────────────────────────────────\n"

	if err := t.Write("  RADAR CONTACT LOG\n",
		text.WriteReplace(),
		text.WriteCellOpts(headerStyle...)); err != nil {
		return err
	}
	if err := t.Write(divider, text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(0, 80, 25)))); err != nil {
		return err
	}
	colHeader := fmt.Sprintf("  %-14s  %4s  BRG   RANGE  %-11s  SPD(kt)\n",
		"CALLSIGN", "TYPE", "ELEVATION")
	if err := t.Write(colHeader,
		text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(0, 180, 50)), cell.Bold())); err != nil {
		return err
	}
	if err := t.Write(divider, text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(0, 80, 25)))); err != nil {
		return err
	}

	// ── Rows ───────────────────────────────────────────────────────────────
	entries := cat.snapshot()
	if len(entries) == 0 {
		return t.Write("  (scanning…)\n",
			text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(0, 120, 35))))
	}

	for _, e := range entries {
		// Callsign in contact color.
		if err := t.Write(fmt.Sprintf("  %-14s", e.label),
			text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(255, 80, 80)), cell.Bold())); err != nil {
			return err
		}
		// Class tag.
		if err := t.Write(fmt.Sprintf("  %-4s", e.class),
			text.WriteCellOpts(cell.FgColor(classColor(e.class)))); err != nil {
			return err
		}
		// Bearing, range, elevation, speed — in dim green for the classic mono feel.
		dataStyle := text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(0, 200, 55)))
		row := fmt.Sprintf("  %s  %3.0f%%  %-11s  %d\n",
			bearingStr(e.angle),
			e.distance*100,
			elevStr(e.elevation),
			e.speed,
		)
		if err := t.Write(row, dataStyle); err != nil {
			return err
		}
	}

	if err := t.Write(divider, text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(0, 80, 25)))); err != nil {
		return err
	}
	summary := fmt.Sprintf("  TRACKS ACTIVE: %d        PRESS Q TO QUIT\n", len(entries))
	return t.Write(summary, text.WriteCellOpts(cell.FgColor(cell.ColorRGB24(0, 140, 40))))
}

// ─── Goroutines ────────────────────────────────────────────────────────────────

// generateContacts spawns new random contacts on a random schedule and
// periodically culls the oldest ones to keep the display uncluttered.
func generateContacts(ctx context.Context, cat *catalogue, r *radar.Radar) {
	// Seed a few initial contacts so the scope isn't empty on first render.
	for i := 0; i < 5; i++ {
		cat.add()
	}
	if err := r.SetContacts(cat.toRadarContacts()); err != nil {
		panic(err)
	}

	// Schedule irregular contact arrivals to simulate a realistic traffic picture.
	nextArrival := time.NewTimer(randomInterval(2*time.Second, 5*time.Second))
	defer nextArrival.Stop()

	for {
		select {
		case <-nextArrival.C:
			cat.add()
			cat.trim(12) // Keep at most 12 simultaneous contacts.
			if err := r.SetContacts(cat.toRadarContacts()); err != nil {
				panic(err)
			}
			nextArrival.Reset(randomInterval(1500*time.Millisecond, 4500*time.Millisecond))

		case <-ctx.Done():
			return
		}
	}
}

// refreshContactLog rewrites the text widget at a comfortable read rate.
func refreshContactLog(ctx context.Context, t *text.Text, cat *catalogue) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := renderContactLog(t, cat); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// randomInterval returns a random duration in [lo, hi).
func randomInterval(lo, hi time.Duration) time.Duration {
	spread := int64(hi - lo)
	if spread <= 0 {
		return lo
	}
	return lo + time.Duration(rand.Int63n(spread))
}

// ─── Main ──────────────────────────────────────────────────────────────────────

func main() {
	rand.Seed(time.Now().UnixNano())

	// Open a log file for draw errors. Termdash swallows the terminal output
	// on panic, so errors would be invisible otherwise.
	logF, err := os.OpenFile("/tmp/radardemo.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer logF.Close()
	logger := log.New(logF, "radardemo: ", log.LstdFlags)

	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// ── Radar widget ─────────────────────────────────────────────────────────
	radarWidget, err := radar.New(
		radar.SweepSpeed(45),        // 45 °/s → one rotation every 8 s
		radar.BeamWidth(32),         // 32° phosphor trail
		radar.SweepSpan(360),        // Full rotation
		radar.SweepDirection(radar.DirectionClockwise),
		radar.RangeRings(4),
		radar.BeamColor(0, 255, 70), // Classic neon green
		radar.ContactColor(255, 55, 55),
		radar.ContactChar('◆'),
	)
	if err != nil {
		panic(err)
	}

	// ── Contact log text widget ───────────────────────────────────────────────
	logWidget, err := text.New(
		text.DisableScrolling(),
	)
	if err != nil {
		panic(err)
	}

	// ── Contact catalogue & background workers ────────────────────────────────
	cat := &catalogue{}
	go generateContacts(ctx, cat, radarWidget)
	go refreshContactLog(ctx, logWidget, cat)

	// ── Layout ────────────────────────────────────────────────────────────────
	// The radar scope occupies the top 70% of the terminal; the contact log
	// sits below it in a contrasting bordered panel.
	c, err := container.New(
		t,
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Double),
				container.BorderColor(cell.ColorRGB24(0, 180, 50)),
				container.BorderTitle(" ◈  AIR TRAFFIC CONTROL — SECTOR 7  ◈ "),
				container.PlaceWidget(radarWidget),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderColor(cell.ColorRGB24(0, 120, 35)),
				container.BorderTitle(" CONTACT LOG "),
				container.PlaceWidget(logWidget),
			),
			container.SplitPercent(70),
		),
	)
	if err != nil {
		panic(err)
	}

	// ── Keyboard handler ──────────────────────────────────────────────────────
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' || k.Key == keyboard.KeyEsc {
			cancel()
		}
	}

	// Run at 50 ms intervals for smooth sweep animation (~20 fps).
	// ErrorHandler routes Draw() errors to the log file rather than panicking
	// silently (termdash restores the terminal before the panic, making errors
	// invisible without this handler).
	if err := termdash.Run(ctx, t, c,
		termdash.KeyboardSubscriber(quitter),
		termdash.RedrawInterval(50*time.Millisecond),
		termdash.ErrorHandler(func(e error) {
			logger.Printf("draw error: %v", e)
		}),
	); err != nil {
		logger.Printf("termdash.Run error: %v", err)
		panic(err)
	}
}
