// Package main is the ops-dashboard demo for the timeline widget.
//
// # Layout
//
//	┌─ Ops Dashboard ─────────────────────────────────────────────────────────┐
//	│ ┌─ Event Log ─────────────────────────┐ ┌─ Details / Stats ───────────┐ │
//	│ │  ⚡ [15:04:01] db-primary: DOWN     │ │ Severity: CRITICAL          │ │
//	│ │  ▲ [15:04:02] cache-01: pool 80%   │ │ Time:     15:04:01          │ │
//	│ │  ● [15:04:03] api-gateway: ok      │ │ Source:   db-primary        │ │
//	│ │  …                                 │ │ service DOWN - all retries  │ │
//	│ │                                    │ │                             │ │
//	│ │                                    │ │ ── Range stats ─────────    │ │
//	│ │                                    │ │  DEBUG    12                │ │
//	│ │                                    │ │  INFO      8                │ │
//	│ │                                    │ │  WARN      5                │ │
//	│ │                                    │ │  ERROR     3                │ │
//	│ │                                    │ │  CRITICAL  1  ← coloured   │ │
//	│ └────────────────────────────────────┘ └─────────────────────────────┘ │
//	│ ┌─ Timeline Range ──────────────────────────────────────────────────────┐│
//	│ │ 15:00:00           15:02:30           15:05:00                       ││
//	│ │  ● ▲  ⚡ ● ●  ▲ ● ⚡  ✖  ●  ▲  ●  ✖  ⚡  ●  ▲  ●                  ││
//	│ │ ░░░░░░░░░░░░░░░░░░████████████████░░░░░░░░░░░░░░░░░░░░░░░░          ││
//	│ │ Start: 15:01:12 ── 15:03:45  · 34 events in range · [r] reset       ││
//	│ └───────────────────────────────────────────────────────────────────────┘│
//	└──────────────────────────────────────────────────────────────────────────┘
//
// # Controls
//
//   - ↑ / ↓        navigate the event log
//   - Click row    select that event (detail panel updates)
//   - Click scrubber (row 1-2)  set range start / end
//   - r / R / Esc  reset time-range selection
//   - q / Q        quit
//
// # Run
//
//	go run ./widgets/timeline/timelinedemo/
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/borderfx"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/timeline"
)

// ── Synthetic event corpus ───────────────────────────────────────────────────

var components = []string{
	"api-gateway", "db-primary", "db-replica", "auth-svc",
	"cache-01", "worker-01", "worker-02", "lb-01",
	"metrics", "alertmanager", "scheduler", "queue",
}

var corpus = map[timeline.Severity][]string{
	timeline.SeverityDebug: {
		"health-check ok",
		"gc cycle complete",
		"config reload — no changes",
		"connection pool: 4/20 active",
		"keepalive sent",
		"cache warm-up finished",
		"heartbeat received",
		"trace sampled at 1 %",
	},
	timeline.SeverityInfo: {
		"request served in 12 ms",
		"cache hit ratio 94 %",
		"scheduled backup completed",
		"deployment canary started",
		"TLS cert renewed (89 days)",
		"autoscale: +1 replica",
		"circuit breaker closed",
		"leader election: won",
	},
	timeline.SeverityWarn: {
		"latency spike >200 ms",
		"connection pool 80 % full",
		"disk utilisation at 75 %",
		"retry attempt 2/3",
		"rate-limit threshold approaching",
		"slow query >1 s logged",
		"memory pressure: GC thrashing",
		"certificate expires in 7 days",
	},
	timeline.SeverityError: {
		"connection refused",
		"request timed out after 30 s",
		"crash-loop back-off triggered",
		"disk write failed — retrying",
		"upstream returned 503",
		"replication lag >60 s",
		"snapshot restore failed",
		"health-check failed 3×",
	},
	timeline.SeverityCritical: {
		"service DOWN — all retries exhausted",
		"OOM killer triggered",
		"data corruption detected on /dev/sdb",
		"cascading failure across zone-a",
		"split-brain: nodes disagree on leader",
		"certificate expired — TLS failing",
	},
}

// randomSeverity returns a severity with realistic ops-log weights:
// ~30% DEBUG, ~30% INFO, ~20% WARN, ~15% ERROR, ~5% CRITICAL.
func randomSeverity(r *rand.Rand) timeline.Severity {
	n := r.Intn(100)
	switch {
	case n < 30:
		return timeline.SeverityDebug
	case n < 60:
		return timeline.SeverityInfo
	case n < 80:
		return timeline.SeverityWarn
	case n < 95:
		return timeline.SeverityError
	default:
		return timeline.SeverityCritical
	}
}

func syntheticEvent(r *rand.Rand) timeline.Event {
	sev := randomSeverity(r)
	msgs := corpus[sev]
	comp := components[r.Intn(len(components))]
	msg := msgs[r.Intn(len(msgs))]
	now := time.Now()
	return timeline.Event{
		Time:        now.Format("15:04:05"),
		Title:       comp,
		Description: msg,
		Severity:    sev,
		Timestamp:   now,
	}
}

// ── Seed events ───────────────────────────────────────────────────────────────

func seedEvents() []timeline.Event {
	base := time.Now().Add(-30 * time.Second)
	return []timeline.Event{
		{Time: base.Format("15:04:05"), Title: "scheduler", Description: "service started", Severity: timeline.SeverityInfo, Timestamp: base},
		{Time: base.Add(2 * time.Second).Format("15:04:05"), Title: "db-primary", Description: "accepting connections", Severity: timeline.SeverityInfo, Timestamp: base.Add(2 * time.Second)},
		{Time: base.Add(4 * time.Second).Format("15:04:05"), Title: "auth-svc", Description: "token cache warm-up finished", Severity: timeline.SeverityDebug, Timestamp: base.Add(4 * time.Second)},
		{Time: base.Add(6 * time.Second).Format("15:04:05"), Title: "lb-01", Description: "health-check ok", Severity: timeline.SeverityDebug, Timestamp: base.Add(6 * time.Second)},
		{Time: base.Add(8 * time.Second).Format("15:04:05"), Title: "api-gateway", Description: "TLS cert valid for 89 days", Severity: timeline.SeverityInfo, Timestamp: base.Add(8 * time.Second)},
	}
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()
	// Enable motion events so Mouse() fires continuously while a button is
	// held and the cursor moves — required for drag interactions (scrubber
	// ghost pin, scroll drag in the event log).
	t.EnableMouseMotion()

	// Widgets.
	tl, detail, picker := buildWidgets()

	// Container layout.
	cont, err := buildLayout(t, tl, detail, picker)
	if err != nil {
		log.Fatalf("container.New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Animated borders (active panel: gradient arc; inactive: warm gold).
	go setupBorderFX(cont, ctx)

	// Background goroutines.
	go startEventGenerator(ctx, tl, picker)
	go startDetailUpdater(ctx, tl, picker, detail)

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, cont,
		termdash.KeyboardSubscriber(quitter),
		termdash.RedrawInterval(120*time.Millisecond),
	); err != nil && err != context.Canceled {
		log.Fatalf("termdash.Run: %v", err)
	}
}

// ── widget factory ────────────────────────────────────────────────────────────

// buildWidgets creates and wires the three widgets used by the dashboard.
// Seed events are loaded into tl and picker before returning.
func buildWidgets() (*timeline.Timeline, *text.Text, *timeline.TimeRangePicker) {
	tl, err := timeline.New(timeline.FollowTail(), timeline.MaxEvents(2000))
	if err != nil {
		log.Fatalf("timeline.New: %v", err)
	}
	seeds := seedEvents()
	tl.SetEvents(seeds)

	detail, err := text.New(text.WrapAtWords())
	if err != nil {
		log.Fatalf("text.New (detail): %v", err)
	}
	writeDetailPlaceholder(detail)

	picker, err := timeline.NewTimeRangePicker(func(start, end time.Time, hasRange bool) {
		if hasRange {
			tl.SetTimeFilter(start, end)
		} else {
			tl.ClearTimeFilter()
			tl.SetFollowTail(true)
		}
	})
	if err != nil {
		log.Fatalf("NewTimeRangePicker: %v", err)
	}
	for _, e := range seeds {
		picker.AddPickerEvent(e)
	}

	return tl, detail, picker
}

// ── layout factory ────────────────────────────────────────────────────────────

// buildLayout assembles the termdash container tree for the ops dashboard.
func buildLayout(t *tcell.Terminal, tl *timeline.Timeline, detail *text.Text, picker *timeline.TimeRangePicker) (*container.Container, error) {
	return container.New(
		t,
		container.ID("root"),
		container.Border(linestyle.Light),
		container.BorderTitle(" ⚡ Ops Dashboard  ·  live stream  ·  q to quit "),
		container.SplitHorizontal(
			// Top 74 %: event log + detail panel side by side.
			container.Top(
				container.SplitVertical(
					container.Left(
						container.ID("eventlog"),
						container.Border(linestyle.Light),
						container.BorderTitle(" Event Log  (↑/↓ navigate · click to select) "),
						container.PlaceWidget(tl),
					),
					container.Right(
						container.ID("details"),
						container.Border(linestyle.Light),
						container.BorderTitle(" Details & Range Stats "),
						container.PlaceWidget(detail),
					),
					container.SplitPercent(62),
				),
			),
			// Bottom 26 %: time range scrubber.
			container.Bottom(
				container.ID("picker"),
				container.Border(linestyle.Light),
				container.BorderTitle(" Timeline Range  ·  click to select range  ·  [r] reset "),
				container.PlaceWidget(picker),
			),
			container.SplitPercent(74),
		),
	)
}

// ── border effects ────────────────────────────────────────────────────────────

// setupBorderFX wires the GradientArc profile onto every panel.
// Active (focused) panel: animated purple→blue arc.
// Inactive panels: static warm-gold border.
// Blocks until ctx is cancelled — run it in a goroutine.
func setupBorderFX(cont *container.Container, ctx context.Context) {
	fx := borderfx.NewAnimator(cont)
	fx.ApplyProfile(borderfx.Profiles.GradientArc, "root", "eventlog", "details", "picker")
	fx.Run(ctx)
}

// ── background goroutines ─────────────────────────────────────────────────────

// startEventGenerator fires one synthetic event every 800 ms until ctx is done.
func startEventGenerator(ctx context.Context, tl *timeline.Timeline, picker *timeline.TimeRangePicker) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ticker := time.NewTicker(800 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e := syntheticEvent(r)
			tl.AddEvent(e)
			picker.AddPickerEvent(e)
		}
	}
}

// startDetailUpdater polls for selection and range changes every 120 ms and
// rewrites the detail panel whenever something changes.
func startDetailUpdater(ctx context.Context, tl *timeline.Timeline, picker *timeline.TimeRangePicker, detail *text.Text) {
	var lastKey string
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		_, _, hasRange := picker.Selection()
		e := tl.SelectedEvent()
		counts := tl.SeverityCounts()
		total := tl.EventCount()

		// Cheap change-detection key — skip redraw if nothing changed.
		selKey := ""
		if e != nil {
			selKey = e.Time + e.Title
		}
		key := fmt.Sprintf("%s/%d/%v", selKey, total, hasRange)
		if key == lastKey {
			continue
		}
		lastKey = key

		detail.Reset()

		if e == nil {
			writeDetailPlaceholder(detail)
		} else {
			detail.Write("Severity: ", text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			detail.Write(
				fmt.Sprintf("[%s]\n", timeline.SeverityName(e.Severity)),
				text.WriteCellOpts(cell.FgColor(timeline.SeverityColor(e.Severity)), cell.Bold()),
			)
			detail.Write(
				fmt.Sprintf("Time:     %s\nSource:   %s\n\n%s\n",
					e.Time, e.Title, e.Description),
				text.WriteCellOpts(cell.FgColor(cell.ColorWhite)),
			)
		}

		rangeLabel := "All events"
		if hasRange {
			rangeLabel = "In range"
		}
		detail.Write(
			fmt.Sprintf("\n── %s (%d) ───\n", rangeLabel, total),
			text.WriteCellOpts(cell.FgColor(cell.ColorGray)),
		)
		for sev := timeline.SeverityDebug; sev <= timeline.SeverityCritical; sev++ {
			c := counts[sev]
			bar := timeline.FormatMiniBar(c, total, 8)
			detail.Write(
				fmt.Sprintf("  %s %s %3d\n", timeline.SeverityName(sev), bar, c),
				text.WriteCellOpts(cell.FgColor(timeline.SeverityColor(sev))),
			)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeDetailPlaceholder(d *text.Text) {
	d.Write(
		"Select an event with ↑/↓ or click a row to see details here.\n\n"+
			"Use the timeline scrubber below to filter events\nby time range.",
		text.WriteCellOpts(cell.FgColor(cell.ColorGray)),
	)
}
