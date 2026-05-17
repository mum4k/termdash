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

package spectrum

import (
	"fmt"
	"image"
	"sync"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/checkbox"
	"github.com/mum4k/termdash/widgets/dropdown"
)

// AlertControl manages a threshold dropdown, an alarm checkbox, and an optional
// warning banner for a spectrum visualization.
//
// The control is intended for overlay-style demos or dashboards that want a
// ready-made alarm UI without reimplementing layout, state, and mouse wiring.
// NewAlertControl remains the single constructor entry point; additional
// methods extend behavior without introducing alternate setup paths.
type AlertControl struct {
	mu sync.RWMutex

	value    int
	values   []int
	toggle   *checkbox.Checkbox
	menu     *dropdown.Dropdown
	active   bool
	onChange func(int) error
}

// alertLayout stores resolved terminal rectangles for the alert control.
type alertLayout struct {
	alarmLabel image.Point
	checkbox   image.Rectangle
	valueLabel image.Point
	menu       image.Rectangle
}

// NewAlertControl returns a new threshold and alarm control.
//
// The generated thresholds cover the inclusive range `[min, max]` in `step`
// increments. The selected threshold snaps to the nearest generated value. The
// onChange callback is the low-level threshold hook and runs after user-driven
// threshold changes.
func NewAlertControl(min, max, step, selected int, onChange func(int) error) (*AlertControl, error) {
	values, err := alertThresholdValues(min, max, step)
	if err != nil {
		return nil, err
	}
	selectedIndex := nearestThresholdIndex(values, selected)

	var control *AlertControl
	toggle, err := checkbox.New("",
		checkbox.Checked(false),
		checkbox.UseIndicatorSet(checkbox.IndicatorSets.Classic),
		checkbox.CellOpts(cell.FgColor(cell.ColorNumber(245))),
		checkbox.FocusedCellOpts(cell.FgColor(cell.ColorNumber(195))),
		checkbox.CheckedCellOpts(cell.FgColor(cell.ColorNumber(118))),
	)
	if err != nil {
		return nil, err
	}

	labels := make([]string, len(values))
	for i, value := range values {
		labels[i] = fmt.Sprintf("%03d", value)
	}

	menu, err := dropdown.New(labels,
		dropdown.Selected(selectedIndex),
		dropdown.Width(7),
		dropdown.GlyphSet(dropdown.GlyphProfiles.Classic),
		dropdown.CellOpts(cell.FgColor(cell.ColorNumber(87)), cell.BgColor(cell.ColorNumber(234))),
		dropdown.FocusedCellOpts(cell.FgColor(cell.ColorNumber(195)), cell.BgColor(cell.ColorNumber(236))),
		dropdown.SelectedCellOpts(cell.FgColor(cell.ColorWhite), cell.BgColor(cell.ColorNumber(60))),
		dropdown.BorderCellOpts(cell.FgColor(cell.ColorNumber(81))),
		dropdown.OnSelect(func(index int, label string) error {
			_ = label
			if control == nil {
				return nil
			}
			value := control.valueForIndex(index)
			control.setValue(value)
			if control.onChange == nil {
				return nil
			}
			return control.onChange(value)
		}),
	)
	if err != nil {
		return nil, err
	}

	control = &AlertControl{
		value:    values[selectedIndex],
		values:   values,
		toggle:   toggle,
		menu:     menu,
		onChange: onChange,
	}
	return control, nil
}

// Threshold returns the currently selected threshold value.
func (ac *AlertControl) Threshold() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.value
}

// Enabled reports whether the alarm checkbox is currently checked.
func (ac *AlertControl) Enabled() bool {
	if ac == nil || ac.toggle == nil {
		return false
	}
	return ac.toggle.Checked()
}

// SetEnabled replaces the checkbox state programmatically.
func (ac *AlertControl) SetEnabled(enabled bool) {
	if ac == nil || ac.toggle == nil {
		return
	}
	ac.toggle.SetChecked(enabled)
}

// UpdateSamples refreshes the alarm state against the latest values.
func (ac *AlertControl) UpdateSamples(values []int) {
	if ac == nil {
		return
	}
	threshold := ac.Threshold()
	active := false
	for _, value := range values {
		if value > threshold {
			active = true
			break
		}
	}
	ac.mu.Lock()
	ac.active = active
	ac.mu.Unlock()
}

// AlertMessage returns the current warning banner message.
func (ac *AlertControl) AlertMessage() string {
	if ac == nil || !ac.Enabled() {
		return ""
	}
	ac.mu.RLock()
	active := ac.active
	ac.mu.RUnlock()
	if !active {
		return ""
	}
	return fmt.Sprintf(" Warning: data exceeds %03d threshold ", ac.Threshold())
}

// SetThreshold snaps value to the nearest configured threshold and updates the
// dropdown selection to match.
func (ac *AlertControl) SetThreshold(value int) {
	if ac == nil {
		return
	}
	ac.setValue(ac.valueForIndex(nearestThresholdIndex(ac.values, value)))
	ac.syncMenu()
}

// Draw renders the alarm checkbox and threshold dropdown within graphArea.
//
// The primary label is used only for horizontal alignment so the control can
// appear just to the right of a graph's leading title.
func (ac *AlertControl) Draw(t terminalapi.Terminal, graphArea image.Rectangle, primaryLabel string) error {
	if ac == nil {
		return nil
	}
	ac.syncMenu()
	layout := ac.layout(graphArea, primaryLabel)
	if layout.menu.Empty() {
		return nil
	}

	drawAlertText(t, layout.alarmLabel, "ALARM", cell.FgColor(cell.ColorNumber(245)))

	checkCanvas, err := canvas.New(layout.checkbox)
	if err == nil && ac.toggle != nil {
		if err := ac.toggle.Draw(checkCanvas, &widgetapi.Meta{Focused: true}); err == nil {
			_ = checkCanvas.Apply(t)
		}
	}

	drawAlertText(t, layout.valueLabel, "Y", cell.FgColor(cell.ColorNumber(245)))

	menuCanvas, err := canvas.New(layout.menu)
	if err != nil {
		return nil
	}
	if err := ac.menu.Draw(menuCanvas, &widgetapi.Meta{Focused: true}); err != nil {
		return err
	}
	return menuCanvas.Apply(t)
}

// HandleMouse routes an absolute mouse click into the alert control.
func (ac *AlertControl) HandleMouse(pos image.Point, graphArea image.Rectangle, primaryLabel string) bool {
	if ac == nil {
		return false
	}
	layout := ac.layout(graphArea, primaryLabel)
	if layout.menu.Empty() {
		return false
	}
	if pos.In(layout.checkbox) {
		rel := pos.Sub(layout.checkbox.Min)
		_ = ac.toggle.Mouse(&terminalapi.Mouse{Position: rel, Button: mouse.ButtonLeft}, &widgetapi.EventMeta{})
		return true
	}
	if !pos.In(layout.menu) {
		ac.menu.Close()
		return false
	}
	rel := pos.Sub(layout.menu.Min)
	_ = ac.menu.Mouse(&terminalapi.Mouse{Position: rel, Button: mouse.ButtonLeft}, &widgetapi.EventMeta{})
	return true
}

// DrawAlert renders the warning banner centered inside pane when focused.
func (ac *AlertControl) DrawAlert(t terminalapi.Terminal, pane image.Rectangle, focused bool) {
	message := ac.AlertMessage()
	if message == "" || !focused {
		return
	}
	width := len([]rune(message))
	if width == 0 || pane.Empty() || width > pane.Dx()-2 {
		return
	}
	innerMinX := pane.Min.X + 1
	innerWidth := pane.Dx() - 2
	x := innerMinX + (innerWidth-width)/2
	y := pane.Min.Y + 1
	if y >= pane.Max.Y {
		return
	}
	drawAlertText(t, image.Point{X: x, Y: y}, message,
		cell.FgColor(cell.ColorYellow),
		cell.BgColor(cell.ColorNumber(52)),
	)
}

// layout resolves terminal positions for the alarm label, checkbox, and menu.
func (ac *AlertControl) layout(graphArea image.Rectangle, primaryLabel string) alertLayout {
	if graphArea.Empty() || graphArea.Dy() < 4 {
		return alertLayout{}
	}

	alarmLabel := "ALARM"
	valueLabel := "Y"
	labelX := graphArea.Min.X + len([]rune(primaryLabel)) + 3
	labelY := graphArea.Min.Y
	checkSize := image.Point{X: ac.toggle.Options().MinimumSize.X, Y: 1}
	checkX := labelX + len([]rune(alarmLabel)) + 1
	valueX := checkX + checkSize.X + 1
	menuX := valueX + len([]rune(valueLabel)) + 1
	maxHeight := graphArea.Max.Y - labelY
	menuSize := ac.menu.CanvasSize(maxHeight)

	if menuX+menuSize.X > graphArea.Max.X {
		menuX = graphArea.Max.X - menuSize.X
		valueX = menuX - len([]rune(valueLabel)) - 1
		checkX = valueX - checkSize.X - 1
		labelX = checkX - len([]rune(alarmLabel)) - 1
	}
	if labelX < graphArea.Min.X {
		labelX = graphArea.Min.X
	}
	if checkX < graphArea.Min.X {
		checkX = graphArea.Min.X
	}
	if valueX < graphArea.Min.X {
		valueX = graphArea.Min.X
	}
	if menuX < graphArea.Min.X {
		menuX = graphArea.Min.X
	}

	return alertLayout{
		alarmLabel: image.Point{X: labelX, Y: labelY},
		checkbox:   image.Rect(checkX, labelY, checkX+checkSize.X, labelY+checkSize.Y),
		valueLabel: image.Point{X: valueX, Y: labelY},
		menu:       image.Rect(menuX, labelY, menuX+menuSize.X, labelY+menuSize.Y),
	}
}

// syncMenu aligns the dropdown selection with the stored threshold value.
func (ac *AlertControl) syncMenu() {
	index := nearestThresholdIndex(ac.values, ac.Threshold())
	if ac.menu.SelectedIndex() != index {
		_ = ac.menu.SetSelected(index)
	}
}

// setValue updates the selected threshold value.
func (ac *AlertControl) setValue(value int) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.value = value
}

// valueForIndex returns the threshold value for a dropdown row.
func (ac *AlertControl) valueForIndex(index int) int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	if index < 0 || index >= len(ac.values) {
		return ac.value
	}
	return ac.values[index]
}

// alertThresholdValues builds the threshold list for the dropdown.
func alertThresholdValues(min, max, step int) ([]int, error) {
	switch {
	case step <= 0:
		return nil, fmt.Errorf("invalid step %d, want 1 <= step", step)
	case max < min:
		return nil, fmt.Errorf("invalid range [%d,%d], want min <= max", min, max)
	}

	var values []int
	for value := min; value <= max; value += step {
		values = append(values, value)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("no threshold values generated")
	}
	return values, nil
}

// nearestThresholdIndex returns the closest threshold index for selected.
func nearestThresholdIndex(values []int, selected int) int {
	if len(values) == 0 {
		return 0
	}
	best := 0
	bestDistance := absInt(values[0] - selected)
	for i, value := range values {
		if distance := absInt(value - selected); distance < bestDistance {
			best = i
			bestDistance = distance
		}
	}
	return best
}

// drawAlertText writes overlay text directly to the terminal.
func drawAlertText(t terminalapi.Terminal, pos image.Point, text string, opts ...cell.Option) {
	cur := pos
	for _, r := range text {
		_ = t.SetCell(cur, r, opts...)
		cur.X++
	}
}

// absInt returns the absolute value of v.
func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
