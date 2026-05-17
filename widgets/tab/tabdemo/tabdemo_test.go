package main

import (
	"testing"

	"github.com/mum4k/termdash/widgets/tab"
)

// TestTelemetryHeatmapFrame verifies the denser heatmap frame shape used by the
// unified demo signals page.
func TestTelemetryHeatmapFrame(t *testing.T) {
	xLabels, yLabels, values := telemetryHeatmapFrame(0.5)

	if got, want := len(xLabels), heatmapGridSize; got != want {
		t.Fatalf("len(xLabels) = %d, want %d", got, want)
	}
	if got, want := len(yLabels), heatmapGridSize; got != want {
		t.Fatalf("len(yLabels) = %d, want %d", got, want)
	}
	if got, want := len(values), len(yLabels); got != want {
		t.Fatalf("len(values) = %d, want %d", got, want)
	}
	for row := range values {
		if got, want := len(values[row]), len(xLabels); got != want {
			t.Fatalf("len(values[%d]) = %d, want %d", row, got, want)
		}
	}
}

// TestTelemetryHeatmapFrameB verifies the companion matrix uses the same 26x26 shape.
func TestTelemetryHeatmapFrameB(t *testing.T) {
	xLabels, yLabels, values := telemetryHeatmapFrameB(0.5)

	if got, want := len(xLabels), heatmapGridSize; got != want {
		t.Fatalf("len(xLabels) = %d, want %d", got, want)
	}
	if got, want := len(yLabels), heatmapGridSize; got != want {
		t.Fatalf("len(yLabels) = %d, want %d", got, want)
	}
	if got, want := len(values), len(yLabels); got != want {
		t.Fatalf("len(values) = %d, want %d", got, want)
	}
	for row := range values {
		if got, want := len(values[row]), len(xLabels); got != want {
			t.Fatalf("len(values[%d]) = %d, want %d", row, got, want)
		}
	}
}

// TestNewTelemetryTabs ensures the unified demo can build its telemetry pages.
func TestNewTelemetryTabs(t *testing.T) {
	widgets, signalsTab, thermalTab, alertsTab, err := newTelemetryTabs()
	if err != nil {
		t.Fatalf("newTelemetryTabs() => unexpected error: %v", err)
	}
	if widgets == nil {
		t.Fatal("newTelemetryTabs() => nil widgets")
	}
	if widgets.heatmap == nil || widgets.heatmapB == nil {
		t.Fatal("newTelemetryTabs() => missing heatmap widgets")
	}
	if widgets.alerts == nil {
		t.Fatal("newTelemetryTabs() => missing alert status widget")
	}
	for _, tabPage := range []*tab.Tab{signalsTab, thermalTab, alertsTab} {
		if tabPage == nil || tabPage.Content == nil {
			t.Fatal("newTelemetryTabs() => missing tab content")
		}
	}
}

func TestNewThreeDTab(t *testing.T) {
	widgets, threeDTab, err := newThreeDTab()
	if err != nil {
		t.Fatalf("newThreeDTab() => unexpected error: %v", err)
	}
	if widgets == nil || widgets.stage == nil || widgets.status == nil {
		t.Fatal("newThreeDTab() => missing widgets")
	}
	if threeDTab == nil || threeDTab.Content == nil {
		t.Fatal("newThreeDTab() => missing tab content")
	}
	if got, want := threeDTab.Name, "ThreeD"; got != want {
		t.Fatalf("threeDTab.Name = %q, want %q", got, want)
	}
}

func TestAlertDrillMessageRotates(t *testing.T) {
	tests := []struct {
		count int
		want  string
	}{
		{count: 1, want: "Thermal matrix drift"},
		{count: 2, want: "Channel load crest"},
		{count: 3, want: "Subspace threshold"},
	}
	for _, tc := range tests {
		got, _ := alertDrillMessage(tc.count)
		if got != tc.want {
			t.Fatalf("alertDrillMessage(%d) = %q, want %q", tc.count, got, tc.want)
		}
	}
}
