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

// Package tab provides functionality for managing tabbed interfaces.
package tab

import (
	"fmt"
	"image"
	"strings"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/runewidth"
	"github.com/mum4k/termdash/widgets/text"
)

// Header displays the tab names and highlights the active tab.
type Header struct {
	mu            sync.Mutex        // Guards the animated frame and click rectangles.
	tm            *Manager          // Reference to the Tab Manager.
	widget        *text.Text        // Text widget for displaying the header.
	tabRectangles []image.Rectangle // Tracks the positions of tabs for mouse interactions.
	height        int               // Stores the height of the Header.
	opts          *Options          // Configuration options for the Header.
	frame         int               // Animation frame for the active tab sweep.
}

// NewHeader creates a new Header.
func NewHeader(tm *Manager, opts *Options) (*Header, error) {
	w, err := text.New(
		text.WrapAtRunes(),      // Use WrapAtRunes for accurate width calculations.
		text.DisableScrolling(), // Disable scrolling to keep the header static.
	)
	if err != nil {
		return nil, err
	}

	// If opts is nil, create default options.
	if opts == nil {
		opts = NewOptions()
	}

	header := &Header{
		tm:     tm,
		widget: w,
		height: 2, // Assuming it occupies 2 lines (tabs and underline).
		opts:   opts,
	}

	if err := header.Update(); err != nil {
		return nil, err
	}

	return header, nil
}

// Update refreshes the tab header to reflect changes.
func (h *Header) Update() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	tabs, activeIndex := h.tm.Snapshot()

	h.widget.Reset()
	h.tabRectangles = make([]image.Rectangle, 0, len(tabs))

	currentX := 0
	var indicator strings.Builder

	for i, tabState := range tabs {
		mainText, notifText := h.tabLabelParts(i == activeIndex, tabState)
		tabWidth := runewidth.StringWidth(mainText) + runewidth.StringWidth(notifText)
		rect := image.Rectangle{
			Min: image.Point{X: currentX, Y: 0},
			Max: image.Point{X: currentX + tabWidth, Y: h.height},
		}
		h.tabRectangles = append(h.tabRectangles, rect)
		if i == activeIndex {
			if err := h.writeActiveLabel(mainText); err != nil {
				return err
			}
			if notifText != "" {
				if err := h.widget.Write(notifText, text.WriteCellOpts(
					cell.FgColor(h.opts.NotificationColor),
					cell.BgColor(h.opts.ActiveTabColor),
					cell.Bold(),
				)); err != nil {
					return err
				}
			}
			indicator.WriteString(h.activeIndicator(tabWidth))
		} else {
			if err := h.widget.Write(mainText,
				text.WriteCellOpts(cell.FgColor(h.opts.InactiveTextColor), cell.BgColor(h.opts.InactiveTabColor)),
			); err != nil {
				return err
			}
			if notifText != "" {
				if err := h.widget.Write(notifText, text.WriteCellOpts(
					cell.FgColor(h.opts.NotificationColor),
					cell.BgColor(h.opts.InactiveTabColor),
					cell.Bold(),
				)); err != nil {
					return err
				}
			}
			if h.opts.AnimatedActiveTab {
				indicator.WriteString(strings.Repeat("─", tabWidth))
			} else {
				indicator.WriteString(strings.Repeat(" ", tabWidth))
			}
		}
		currentX += tabWidth
		if i < len(tabs)-1 {
			if err := h.widget.Write(" ", text.WriteCellOpts(cell.FgColor(h.opts.LabelColor), cell.BgColor(h.opts.InactiveTabColor))); err != nil {
				return err
			}
			indicator.WriteRune(' ')
			currentX++
		}
	}

	if err := h.widget.Write("\n", text.WriteCellOpts(cell.FgColor(h.opts.LabelColor))); err != nil {
		return err
	}
	if h.opts.AnimatedActiveTab && activeIndex >= 0 && activeIndex < len(h.tabRectangles) {
		if err := h.writeAnimatedIndicator(indicator.String(), h.tabRectangles[activeIndex]); err != nil {
			return err
		}
	} else {
		if err := h.widget.Write(indicator.String(), text.WriteCellOpts(cell.FgColor(h.opts.LabelColor))); err != nil {
			return err
		}
	}

	return nil
}

// Advance moves the active-tab sweep forward by one frame and redraws the header.
func (h *Header) Advance() error {
	h.mu.Lock()
	h.frame++
	h.mu.Unlock()
	return h.Update()
}

// tabLabelParts splits a tab label into the main text and the notification suffix.
// The notification suffix is empty when the tab has no active notification.
// Splitting the two parts allows the notification icon to be rendered in a
// distinct alarm color rather than the ordinary label color.
func (h *Header) tabLabelParts(active bool, tabState Snapshot) (main, notif string) {
	icon := h.opts.InactiveIcon
	if active {
		icon = h.opts.ActiveIcon
	}
	main = fmt.Sprintf(" %s %s ", icon, tabState.Name)
	if tabState.Notification {
		notif = h.opts.NotificationIcon + " "
	}
	return main, notif
}

// writeActiveLabel renders the active tab title, optionally greying text under the sweep.
func (h *Header) writeActiveLabel(label string) error {
	if !h.opts.AnimatedActiveTab {
		return h.widget.Write(label, text.WriteCellOpts(cell.Bold(), cell.FgColor(h.opts.ActiveTextColor), cell.BgColor(h.opts.ActiveTabColor)))
	}

	head := h.frame % max(1, runewidth.StringWidth(label))
	cur := 0
	for _, r := range label {
		width := runewidth.RuneWidth(r)
		fg := h.opts.ActiveTextColor
		if h.inSweep(cur, width, head) {
			fg = h.opts.SweepTextColor
		}
		if err := h.widget.Write(string(r), text.WriteCellOpts(cell.Bold(), cell.FgColor(fg), cell.BgColor(h.opts.ActiveTabColor))); err != nil {
			return err
		}
		cur += width
	}
	return nil
}

// activeIndicator renders the underline treatment for the active tab.
func (h *Header) activeIndicator(width int) string {
	if width <= 0 {
		return ""
	}
	if !h.opts.AnimatedActiveTab {
		return strings.Repeat("⎺", width)
	}

	head := h.frame % width
	runes := make([]rune, 0, width)
	for i := 0; i < width; i++ {
		switch {
		case i == head:
			runes = append(runes, '•')
		case i == head-1 || (head == 0 && i == width-1):
			runes = append(runes, '·')
		default:
			runes = append(runes, '⎺')
		}
	}
	return string(runes)
}

// writeAnimatedIndicator recolors the active sweep segment in the underline row.
func (h *Header) writeAnimatedIndicator(indicator string, activeRect image.Rectangle) error {
	cur := 0
	for _, r := range indicator {
		color := h.opts.LabelColor
		if cur >= activeRect.Min.X && cur < activeRect.Max.X {
			color = h.opts.SweepAccentColor
			if r == '⎺' {
				color = h.opts.LabelColor
			}
		}
		if err := h.widget.Write(string(r), text.WriteCellOpts(cell.FgColor(color))); err != nil {
			return err
		}
		cur += runewidth.RuneWidth(r)
	}
	return nil
}

// inSweep reports whether the sweep head overlaps the rune segment.
func (h *Header) inSweep(start, width, head int) bool {
	for i := 0; i < width; i++ {
		if distance(head, start+i) <= 1 {
			return true
		}
	}
	return false
}

// Widget returns the underlying text widget.
func (h *Header) Widget() *text.Text {
	return h.widget
}

// Height returns the height of the Header.
func (h *Header) Height() int {
	return h.height
}

// GetClickedTab returns the index of the tab that was clicked based on the mouse position.
// Returns -1 if no tab was clicked.
func (h *Header) GetClickedTab(p image.Point) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i, rect := range h.tabRectangles {
		if rect.Min.X <= p.X && p.X < rect.Max.X &&
			rect.Min.Y <= p.Y && p.Y < rect.Max.Y {
			return i
		}
	}
	return -1
}

// distance returns the absolute difference between a and b.
func distance(a, b int) int {
	if a < b {
		return b - a
	}
	return a - b
}

// max returns the larger of a or b.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
