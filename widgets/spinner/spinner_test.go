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

package spinner

import "testing"

func TestLookupAndNames(t *testing.T) {
	if got, want := len(Names()), 38; got != want {
		t.Fatalf("len(Names()) = %d, want %d", got, want)
	}
	if first := Names()[0]; first != "arc" {
		t.Fatalf("Names()[0] = %q, want arc", first)
	}
	if _, ok := Lookup("pulse"); !ok {
		t.Fatal("Lookup(pulse) = missing spinner")
	}
	if _, ok := Lookup("missing"); ok {
		t.Fatal("Lookup(missing) unexpectedly succeeded")
	}
}

func TestFramesAndDecoration(t *testing.T) {
	star := Must("star")
	if got := star.Name(); got != "star" {
		t.Fatalf("Name() = %q, want star", got)
	}
	if got := star.Interval(); got != 10 {
		t.Fatalf("Interval() = %d, want 10", got)
	}
	if got := star.Frame(0); got != "✶" {
		t.Fatalf("Frame(0) = %q, want ✶", got)
	}
	if got := star.Frame(7); got != "✸" {
		t.Fatalf("Frame(7) = %q, want ✸", got)
	}
	if got := star.DecorateRight(" Warp Core Flux ", 0); got != " Warp Core Flux ✶" {
		t.Fatalf("DecorateRight() = %q", got)
	}
	if got := Must("pulse").DecorateLeft(" LCARS ", 3); got != "⎽ LCARS " {
		t.Fatalf("DecorateLeft() = %q", got)
	}
}

func TestSingleCellHelpers(t *testing.T) {
	pulse := Must("pulse")
	if !pulse.SingleCell() {
		t.Fatal("pulse should be single-cell")
	}
	runes, ok := pulse.RuneFrames()
	if !ok {
		t.Fatal("RuneFrames() = not single-cell for pulse")
	}
	if got := string(runes); got != "⎺⎻⎼⎽⎼⎻" {
		t.Fatalf("RuneFrames() = %q", got)
	}
	r, ok := Must("dots_10").Rune(6)
	if !ok || r != '⡠' {
		t.Fatalf("Rune(6) = %q, %v; want ⡠, true", r, ok)
	}
	if Must("shark").SingleCell() {
		t.Fatal("shark should not be single-cell")
	}
	if _, ok := Must("shark").RuneFrames(); ok {
		t.Fatal("RuneFrames() unexpectedly succeeded for shark")
	}
}

func TestMustPanicsOnUnknownName(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Must(missing) did not panic")
		}
	}()
	_ = Must("missing")
}
