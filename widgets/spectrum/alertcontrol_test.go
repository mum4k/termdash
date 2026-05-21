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
	"image"
	"strings"
	"testing"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/faketerm"
)

// TestNewAlertControl validates constructor argument handling and defaults.
func TestNewAlertControl(t *testing.T) {
	tests := []struct {
		name     string
		min      int
		max      int
		step     int
		selected int
		wantErr  bool
	}{
		{name: "defaults", min: 200, max: 600, step: 50, selected: 500},
		{name: "rejects zero step", min: 200, max: 600, step: 0, selected: 500, wantErr: true},
		{name: "rejects inverted range", min: 600, max: 200, step: 50, selected: 500, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			control, err := NewAlertControl(tc.min, tc.max, tc.step, tc.selected, nil)
			if tc.wantErr {
				if err == nil {
					t.Fatal("NewAlertControl => nil error, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewAlertControl => unexpected error: %v", err)
			}
			if control == nil {
				t.Fatal("NewAlertControl => nil control")
			}
			if got, want := control.Threshold(), tc.selected; got != want {
				t.Fatalf("Threshold = %d, want %d", got, want)
			}
		})
	}
}

// TestAlertThresholdHelpers covers threshold list generation and nearest lookup.
func TestAlertThresholdHelpers(t *testing.T) {
	values, err := alertThresholdValues(200, 600, 50)
	if err != nil {
		t.Fatalf("alertThresholdValues => unexpected error: %v", err)
	}
	if got, want := values, []int{200, 250, 300, 350, 400, 450, 500, 550, 600}; len(got) != len(want) {
		t.Fatalf("alertThresholdValues length = %d, want %d", len(got), len(want))
	}
	if got := nearestThresholdIndex([]int{200, 250, 300}, 260); got != 1 {
		t.Fatalf("nearestThresholdIndex = %d, want 1", got)
	}
}

// TestAlertControlDrawAndMouse verifies the control renders and handles clicks.
func TestAlertControlDrawAndMouse(t *testing.T) {
	var selected int
	control, err := NewAlertControl(200, 600, 50, 500, func(value int) error {
		selected = value
		return nil
	})
	if err != nil {
		t.Fatalf("NewAlertControl => unexpected error: %v", err)
	}

	graphArea := image.Rect(2, 2, 98, 18)
	primaryLabel := "PING LATTICE"
	ft := faketerm.MustNew(image.Point{X: 100, Y: 32})
	if err := control.Draw(ft, graphArea, primaryLabel); err != nil {
		t.Fatalf("Draw => unexpected error: %v", err)
	}
	rendered := ft.String()
	if !strings.Contains(rendered, "ALARM") || !strings.Contains(rendered, "[ ]") || !strings.Contains(rendered, "500") {
		t.Fatalf("Draw output = %q, want label, checkbox, and default threshold", rendered)
	}

	layout := control.layout(graphArea, primaryLabel)
	if layout.menu.Empty() {
		t.Fatal("layout.menu is empty, want usable dropdown area")
	}

	if !control.HandleMouse(layout.checkbox.Min, graphArea, primaryLabel) {
		t.Fatal("HandleMouse checkbox => false, want true")
	}
	if !control.Enabled() {
		t.Fatal("Enabled = false, want true")
	}

	if !control.HandleMouse(layout.menu.Min.Add(image.Point{X: 1, Y: 0}), graphArea, primaryLabel) {
		t.Fatal("HandleMouse trigger => false, want true")
	}
	if err := control.Draw(ft, graphArea, primaryLabel); err != nil {
		t.Fatalf("Draw after opening => unexpected error: %v", err)
	}
	layout = control.layout(graphArea, primaryLabel)
	targetIndex := nearestThresholdIndex(control.values, 350)
	optionPos := layout.menu.Min.Add(image.Point{X: 2, Y: 2 + targetIndex})
	if !control.HandleMouse(optionPos, graphArea, primaryLabel) {
		t.Fatal("HandleMouse option => false, want true")
	}
	if got, want := control.Threshold(), 350; got != want {
		t.Fatalf("Threshold after selection = %d, want %d", got, want)
	}
	if got, want := selected, 350; got != want {
		t.Fatalf("onChange value = %d, want %d", got, want)
	}

	if control.HandleMouse(image.Point{}, image.Rect(0, 0, 1, 1), primaryLabel) {
		t.Fatal("HandleMouse outside => true, want false")
	}
}

// TestAlertControlAlertState verifies banner visibility follows samples and focus.
func TestAlertControlAlertState(t *testing.T) {
	control, err := NewAlertControl(200, 600, 50, 500, nil)
	if err != nil {
		t.Fatalf("NewAlertControl => unexpected error: %v", err)
	}

	layout := control.layout(image.Rect(2, 2, 98, 18), "PING LATTICE")
	if !control.HandleMouse(layout.checkbox.Min, image.Rect(2, 2, 98, 18), "PING LATTICE") {
		t.Fatal("HandleMouse checkbox => false, want true")
	}

	control.UpdateSamples([]int{100, 200, 550})
	if got := control.AlertMessage(); !strings.Contains(got, "Warning: data exceeds 500 threshold") {
		t.Fatalf("AlertMessage = %q, want warning banner", got)
	}

	ft := faketerm.MustNew(image.Point{X: 100, Y: 32})
	pane := image.Rect(0, 10, 100, 32)
	control.DrawAlert(ft, pane, true)
	if rendered := ft.String(); !strings.Contains(rendered, "Warning: data exceeds 500 threshold") {
		t.Fatalf("DrawAlert output = %q, want centered warning", rendered)
	}

	ft = faketerm.MustNew(image.Point{X: 100, Y: 32})
	control.DrawAlert(ft, pane, false)
	if rendered := ft.String(); strings.Contains(rendered, "Warning: data exceeds 500 threshold") {
		t.Fatalf("DrawAlert unfocused output = %q, want no warning", rendered)
	}

	control.UpdateSamples([]int{100, 200, 300})
	if got := control.AlertMessage(); got != "" {
		t.Fatalf("AlertMessage below threshold = %q, want empty", got)
	}
}

func TestAlertControlSetters(t *testing.T) {
	control, err := NewAlertControl(200, 600, 50, 500, nil)
	if err != nil {
		t.Fatalf("NewAlertControl => unexpected error: %v", err)
	}

	control.SetEnabled(true)
	if !control.Enabled() {
		t.Fatal("Enabled = false after SetEnabled(true), want true")
	}

	control.SetThreshold(340)
	if got, want := control.Threshold(), 350; got != want {
		t.Fatalf("Threshold = %d, want %d after snapping", got, want)
	}
}

// TestDrawAlertText verifies overlay text helper writes directly to terminals.
func TestDrawAlertText(t *testing.T) {
	ft := faketerm.MustNew(image.Point{X: 12, Y: 4})
	drawAlertText(ft, image.Point{X: 1, Y: 1}, "TEST", cell.FgColor(cell.ColorGreen))
	if rendered := ft.String(); !strings.Contains(rendered, "TEST") {
		t.Fatalf("drawAlertText output = %q, want TEST", rendered)
	}
}
