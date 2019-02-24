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

package zoom

import (
	"image"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart/internal/axes"
)

// mustNewXDetails creates the XDetails or panics.
func mustNewXDetails(cvsAr image.Rectangle, xp *axes.XProperties) *axes.XDetails {
	xd, err := axes.NewXDetails(cvsAr, xp)
	if err != nil {
		panic(err)
	}
	return xd
}

func TestTracker(t *testing.T) {
	tests := []struct {
		desc    string
		opts    []Option
		xp      *axes.XProperties
		cvsAr   image.Rectangle
		graphAr image.Rectangle
		// mutate if not nil, can mutate the state of the tracker.
		// I.e. send mouse events or update the X scale or canvas areas.
		mutate             func(*Tracker) error
		wantHighlight      bool
		wantHighlightRange *Range
		wantZoom           *axes.XDetails
		wantErr            bool
		wantMutateErr      bool
	}{
		{
			desc: "New fails when graph area doesn't fall inside the canvas",
			xp: &axes.XProperties{
				Min:       0,
				Max:       1,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 10, 10),
			graphAr: image.Rect(20, 20, 30, 30),
			wantErr: true,
		},
		{
			desc: "New fails on ScrollStep too low",
			opts: []Option{
				ScrollStep(0),
			},
			xp: &axes.XProperties{
				Min:       0,
				Max:       1,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 10, 10),
			graphAr: image.Rect(2, 0, 10, 10),
			wantErr: true,
		},
		{
			desc: "New fails on ScrollStep too high",
			opts: []Option{
				ScrollStep(101),
			},
			xp: &axes.XProperties{
				Min:       0,
				Max:       1,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 10, 10),
			graphAr: image.Rect(2, 0, 10, 10),
			wantErr: true,
		},
		{
			desc: "Update fails when graph area doesn't fall inside the canvas",
			xp: &axes.XProperties{
				Min:       0,
				Max:       1,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 10, 10),
			graphAr: image.Rect(1, 1, 9, 9),
			mutate: func(tr *Tracker) error {
				cvsAr := image.Rect(0, 0, 10, 10)
				graphAr := image.Rect(20, 20, 30, 30)
				return tr.Update(tr.baseX, cvsAr, graphAr)
			},
			wantMutateErr: true,
		},
		{
			desc: "no highlight or zoom without mouse events",
			xp: &axes.XProperties{
				Min:       0,
				Max:       1,
				ReqYWidth: 2,
			},
			cvsAr:         image.Rect(0, 0, 10, 10),
			graphAr:       image.Rect(3, 0, 10, 10),
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 10, 10),
				&axes.XProperties{
					Min:       0,
					Max:       1,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights single column",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 6, 6),
			graphAr: image.Rect(2, 0, 6, 6),
			mutate: func(tr *Tracker) error {
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				})
			},
			wantHighlight:      true,
			wantHighlightRange: &Range{Start: 0, End: 1, last: 0},
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 6, 6),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights single column in a new canvas portion after size increase, regression for #148",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 4, 4),
			graphAr: image.Rect(2, 0, 4, 4),
			mutate: func(tr *Tracker) error {
				newX, err := axes.NewXDetails(image.Rect(0, 0, 6, 6), &axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				})
				if err != nil {
					return err
				}

				if err := tr.Update(
					newX,
					image.Rect(0, 0, 6, 6),
					image.Rect(2, 0, 6, 6),
				); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 5},
					Button:   mouse.ButtonLeft,
				})
			},
			wantHighlight:      true,
			wantHighlightRange: &Range{Start: 0, End: 1, last: 0},
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 6, 6),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights multiple columns to the right of start",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 6, 6),
			graphAr: image.Rect(2, 0, 6, 6),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{4, 0},
					Button:   mouse.ButtonLeft,
				})
			},
			wantHighlight:      true,
			wantHighlightRange: &Range{Start: 0, End: 3, last: 2},
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 6, 6),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights multiple columns to the right of start then middle",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 6, 6),
			graphAr: image.Rect(2, 0, 6, 6),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{4, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				})
			},
			wantHighlight:      true,
			wantHighlightRange: &Range{Start: 0, End: 2, last: 1},
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 6, 6),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights multiple columns to the right of start then left of start",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 6, 6),
			graphAr: image.Rect(2, 0, 6, 6),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{4, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				})
			},
			wantHighlight:      true,
			wantHighlightRange: &Range{Start: 0, End: 2, last: 0},
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 6, 6),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights multiple columns to the left of start",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 6, 6),
			graphAr: image.Rect(2, 0, 6, 6),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{4, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				})
			},
			wantHighlight:      true,
			wantHighlightRange: &Range{Start: 0, End: 3, last: 0},
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 6, 6),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights multiple columns to the left of start then middle",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 6, 6),
			graphAr: image.Rect(2, 0, 6, 6),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{4, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				})
			},
			wantHighlight:      true,
			wantHighlightRange: &Range{Start: 1, End: 3, last: 1},
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 6, 6),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights multiple columns to the left of start then right",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 6, 6),
			graphAr: image.Rect(2, 0, 6, 6),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{4, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{5, 0},
					Button:   mouse.ButtonLeft,
				})
			},
			wantHighlight:      true,
			wantHighlightRange: &Range{Start: 2, End: 4, last: 3},
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 6, 6),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights multiple columns in the middle",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 6, 6),
			graphAr: image.Rect(2, 0, 6, 6),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{5, 0},
					Button:   mouse.ButtonLeft,
				})
			},
			wantHighlight:      true,
			wantHighlightRange: &Range{Start: 1, End: 4, last: 3},
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 6, 6),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "does not highlight for clicks outside of graph area",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 6, 6),
			graphAr: image.Rect(2, 0, 6, 6),
			mutate: func(tr *Tracker) error {
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{1, 0},
					Button:   mouse.ButtonLeft,
				})
			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 6, 6),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "doesn't zoom when only one column highlighted",
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonRelease,
				})

			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       0,
					Max:       5,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights and zooms into the X axis once",
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{6, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{6, 0},
					Button:   mouse.ButtonRelease,
				})

			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       1,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights and zooms into the X axis twice",
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				// Zoom into values 1-3.
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{5, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{5, 0},
					Button:   mouse.ButtonRelease,
				}); err != nil {
					return err
				}

				// Zoom into values 2-3.
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{5, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{6, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{6, 0},
					Button:   mouse.ButtonRelease,
				})
			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       2,
					Max:       3,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "doesn't zoom below two values",
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				// Zoom into values 1-3.
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{4, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{5, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{5, 0},
					Button:   mouse.ButtonRelease,
				}); err != nil {
					return err
				}

				// Zoom into values 2-3.
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{4, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{5, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{5, 0},
					Button:   mouse.ButtonRelease,
				}); err != nil {
					return err
				}

				// Doesn't zoom further.
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{4, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{4, 0},
					Button:   mouse.ButtonRelease,
				})
			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       2,
					Max:       3,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "fails to zoom when X coordinate of click too high",
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{7, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{6, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonRelease,
				})
			},
			wantMutateErr: true,
		},
		{
			desc: "cancels highlight and zooms on unrelated mouse button",
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{6, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{6, 0},
					Button:   mouse.ButtonMiddle,
				})

			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       0,
					Max:       5,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "cancels highlight and zooms on button release outside of the graph area",
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{6, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{0, 0},
					Button:   mouse.ButtonRelease,
				})

			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       0,
					Max:       5,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights of single columns doesn't zoom",
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonRelease,
				})

			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       0,
					Max:       5,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "highlights of multiple columns maximizes zoom",
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonRelease,
				})

			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       0,
					Max:       1,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "ignores scroll events outside of graph area",
			opts: []Option{
				ScrollStep(30),
			},
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{1, 0},
					Button:   mouse.ButtonWheelUp,
				})
			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       0,
					Max:       5,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "multiple scroll ups maximize zoom",
			opts: []Option{
				ScrollStep(30),
			},
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonWheelUp,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonWheelUp,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonWheelUp,
				})

			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       0,
					Max:       1,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "multiple scroll downs minimize zoom",
			opts: []Option{
				ScrollStep(30),
			},
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonWheelUp,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonWheelUp,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonWheelUp,
				}); err != nil {
					return err
				}

				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonWheelDown,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonWheelDown,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonWheelDown,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonWheelDown,
				}); err != nil {
					return err
				}
				return tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonWheelDown,
				})

			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "zoom normalized when axis changed (new values)",
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonRelease,
				}); err != nil {
					return err
				}

				newX, err := axes.NewXDetails(image.Rect(0, 0, 8, 8), &axes.XProperties{
					Min:       0,
					Max:       0,
					ReqYWidth: 2,
				})
				if err != nil {
					return err
				}
				return tr.Update(
					newX,
					image.Rect(0, 0, 8, 8),
					image.Rect(2, 0, 8, 8),
				)
			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       0,
					Max:       0,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "fully unzooms when axis changes",
			xp: &axes.XProperties{
				Min:       0,
				Max:       5,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonRelease,
				}); err != nil {
					return err
				}

				newX, err := axes.NewXDetails(image.Rect(0, 0, 8, 8), &axes.XProperties{
					Min:       0,
					Max:       1,
					ReqYWidth: 2,
				})
				if err != nil {
					return err
				}
				return tr.Update(
					newX,
					image.Rect(0, 0, 8, 8),
					image.Rect(2, 0, 8, 8),
				)
			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 8, 8),
				&axes.XProperties{
					Min:       0,
					Max:       1,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "zoom normalized when terminal size changed",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{3, 0},
					Button:   mouse.ButtonRelease,
				}); err != nil {
					return err
				}

				newX, err := axes.NewXDetails(image.Rect(0, 0, 4, 4), &axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				})
				if err != nil {
					return err
				}
				return tr.Update(
					newX,
					image.Rect(0, 0, 4, 4),
					image.Rect(2, 0, 4, 4),
				)
			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 4, 4),
				&axes.XProperties{
					Min:       0,
					Max:       1,
					ReqYWidth: 2,
				},
			),
		},
		{
			desc: "cancels highlight when terminal size changed",
			xp: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			mutate: func(tr *Tracker) error {
				if err := tr.Mouse(&terminalapi.Mouse{
					Position: image.Point{2, 0},
					Button:   mouse.ButtonLeft,
				}); err != nil {
					return err
				}

				newX, err := axes.NewXDetails(image.Rect(0, 0, 4, 4), &axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				})
				if err != nil {
					return err
				}
				return tr.Update(
					newX,
					image.Rect(0, 0, 4, 4),
					image.Rect(2, 0, 4, 4),
				)
			},
			wantHighlight: false,
			wantZoom: mustNewXDetails(
				image.Rect(0, 0, 4, 4),
				&axes.XProperties{
					Min:       0,
					Max:       4,
					ReqYWidth: 2,
				},
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			xd, err := axes.NewXDetails(tc.cvsAr, tc.xp)
			if err != nil {
				t.Fatalf("NewXDetails => unexpected error: %v", err)
			}

			tracker, err := New(xd, tc.cvsAr, tc.graphAr, tc.opts...)
			if (err != nil) != tc.wantErr {
				t.Errorf("New => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if tc.mutate != nil {
				err := tc.mutate(tracker)
				if (err != nil) != tc.wantMutateErr {
					t.Errorf("tc.mutate => unexpected error: %v, wantMutateErr: %v", err, tc.wantMutateErr)
				}
				if err != nil {
					return
				}
			}

			gotHighlight, gotHightlightRange := tracker.Highlight()
			if gotHighlight != tc.wantHighlight {
				t.Errorf("Hightlight => %v, _, want %v, _", gotHighlight, tc.wantHighlight)
			}
			if diff := pretty.Compare(tc.wantHighlightRange, gotHightlightRange); diff != "" {
				t.Errorf("Hightlight => unexpected range, diff (-want, +got):\n%s", diff)
			}

			gotZoom := tracker.Zoom()
			if diff := pretty.Compare(tc.wantZoom, gotZoom); diff != "" {
				t.Errorf("Zoom => unexpected XDetails, diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		desc    string
		baseMin *axes.Value
		baseMax *axes.Value
		min     int
		max     int
		opts    *normalizeOptions
		wantMin int
		wantMax int
	}{
		{
			desc:    "min and max within the base axis",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(3, 0),
			min:     1,
			max:     2,
			wantMin: 1,
			wantMax: 2,
		},
		{
			desc:    "min and max on the edges of base",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(3, 0),
			min:     0,
			max:     3,
			wantMin: 0,
			wantMax: 3,
		},
		{
			desc:    "min and max normalized",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(3, 0),
			min:     -1,
			max:     4,
			wantMin: 0,
			wantMax: 3,
		},
		{
			desc:    "min is below base, max is the first value",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(3, 0),
			min:     -1,
			max:     0,
			wantMin: 0,
			wantMax: 1,
		},
		{
			desc:    "min is below base, max is the first value, no space on the axis",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(0, 0),
			min:     -1,
			max:     0,
			wantMin: 0,
			wantMax: 0,
		},
		{
			desc:    "max is above base, min is the last value",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(3, 0),
			min:     3,
			max:     4,
			wantMin: 2,
			wantMax: 3,
		},
		{
			desc:    "min is below base, max is the first value, no space on the axis",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(0, 0),
			min:     0,
			max:     1,
			wantMin: 0,
			wantMax: 0,
		},
		{
			desc:    "both min and max are below base, min < max",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(3, 0),
			min:     -2,
			max:     -1,
			wantMin: 0,
			wantMax: 1,
		},
		{
			desc:    "both min and max are below base, min > max",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(3, 0),
			min:     -1,
			max:     -2,
			wantMin: 0,
			wantMax: 1,
		},
		{
			desc:    "both min and max are above base, min < max",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(3, 0),
			min:     4,
			max:     5,
			wantMin: 2,
			wantMax: 3,
		},
		{
			desc:    "both min and max are above base, min > max",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(3, 0),
			min:     5,
			max:     4,
			wantMin: 2,
			wantMax: 3,
		},
		{
			desc:    "both min and max are below base, base only has one value",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(0, 0),
			min:     -2,
			max:     -1,
			wantMin: 0,
			wantMax: 0,
		},
		{
			desc:    "max in the middle, min above base",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(4, 0),
			min:     5,
			max:     3,
			wantMin: 3,
			wantMax: 4,
		},
		{
			desc:    "min in the middle, max below base",
			baseMin: axes.NewValue(0, 0),
			baseMax: axes.NewValue(4, 0),
			min:     3,
			max:     -1,
			wantMin: 0,
			wantMax: 3,
		},
		{
			desc: "zoom rolls when base axis rolls to the left",
			opts: &normalizeOptions{
				oldBaseMin: axes.NewValue(10, 0),
				oldBaseMax: axes.NewValue(20, 0),
			},
			baseMin: axes.NewValue(17, 0),
			baseMax: axes.NewValue(27, 0),
			min:     15,
			max:     16,
			wantMin: 22,
			wantMax: 23,
		},
		{
			desc: "zoom rolls when base axis rolls to the right",
			opts: &normalizeOptions{
				oldBaseMin: axes.NewValue(10, 0),
				oldBaseMax: axes.NewValue(20, 0),
			},
			baseMin: axes.NewValue(1, 0),
			baseMax: axes.NewValue(11, 0),
			min:     15,
			max:     16,
			wantMin: 6,
			wantMax: 7,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			gotMin, gotMax := normalize(tc.baseMin, tc.baseMax, tc.min, tc.max, tc.opts)
			if gotMin != tc.wantMin || gotMax != tc.wantMax {
				t.Errorf("normalize => %v, %v, want %v, %v", gotMin, gotMax, tc.wantMin, tc.wantMax)
			}
		})
	}
}

func TestNewZoomedFromBase(t *testing.T) {
	tests := []struct {
		desc    string
		min     int
		max     int
		baseP   *axes.XProperties
		cvsAr   image.Rectangle
		wantP   *axes.XProperties
		wantErr bool
	}{
		{
			desc: "returns zoomed axis",
			min:  1,
			max:  2,
			baseP: &axes.XProperties{
				Min:       0,
				Max:       3,
				ReqYWidth: 2,
				CustomLabels: map[int]string{
					1: "1",
				},
				LO: axes.LabelOrientationVertical,
			},
			cvsAr: image.Rect(0, 0, 10, 10),
			wantP: &axes.XProperties{
				Min:       1,
				Max:       2,
				ReqYWidth: 2,
				CustomLabels: map[int]string{
					1: "1",
				},
				LO: axes.LabelOrientationVertical,
			},
		},
		{
			desc: "fails on negative max",
			min:  1,
			max:  -2,
			baseP: &axes.XProperties{
				Min:       0,
				Max:       3,
				ReqYWidth: 2,
				CustomLabels: map[int]string{
					1: "1",
				},
				LO: axes.LabelOrientationVertical,
			},
			cvsAr:   image.Rect(0, 0, 10, 10),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			base, err := axes.NewXDetails(tc.cvsAr, tc.baseP)
			if err != nil {
				t.Fatalf("NewXDetails => unexpected error: %v", err)
			}

			got, err := newZoomedFromBase(tc.min, tc.max, base, tc.cvsAr)
			if (err != nil) != tc.wantErr {
				t.Errorf("newZoomedFromBase => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			var want *axes.XDetails
			if tc.wantP != nil {
				w, err := axes.NewXDetails(tc.cvsAr, tc.wantP)
				if err != nil {
					t.Fatalf("NewXDetails => unexpected error: %v", err)

				}
				want = w
			}

			if diff := pretty.Compare(want, got); diff != "" {
				t.Errorf("newZoomedFromBase => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestFindCellPair(t *testing.T) {
	tests := []struct {
		desc    string
		cvsAr   image.Rectangle
		baseP   *axes.XProperties
		minCell int
		maxCell int
		wantMin *axes.Value
		wantMax *axes.Value
		wantErr bool
	}{
		{
			desc:  "fails when minCell isn't on the graph",
			cvsAr: image.Rect(0, 0, 4, 4),
			baseP: &axes.XProperties{
				Min: 0,
				Max: 3,
			},
			minCell: -1,
			maxCell: 3,
			wantErr: true,
		},
		{
			desc:  "fails when maxCell isn't on the graph",
			cvsAr: image.Rect(0, 0, 4, 4),
			baseP: &axes.XProperties{
				Min: 0,
				Max: 3,
			},
			minCell: 0,
			maxCell: 4,
			wantErr: true,
		},
		{
			desc:  "nothing to do, cells point at distinct values",
			cvsAr: image.Rect(0, 0, 4, 4),
			baseP: &axes.XProperties{
				Min: 0,
				Max: 2,
			},
			minCell: 0,
			maxCell: 2,
			wantMin: axes.NewValue(0, 2),
			wantMax: axes.NewValue(2, 2),
		},
		{
			desc:  "cells point at the same value, distinct found above max",
			cvsAr: image.Rect(0, 0, 4, 4),
			baseP: &axes.XProperties{
				Min: 0,
				Max: 2,
			},
			minCell: 1,
			maxCell: 2,
			wantMin: axes.NewValue(1, 2),
			wantMax: axes.NewValue(2, 2),
		},
		{
			desc:  "cells point at the same value, distinct found below min",
			cvsAr: image.Rect(0, 0, 4, 4),
			baseP: &axes.XProperties{
				Min: 0,
				Max: 2,
			},
			minCell: 2,
			maxCell: 2,
			wantMin: axes.NewValue(1, 2),
			wantMax: axes.NewValue(2, 2),
		},
		{
			desc:  "cells point at the same value, only distinct are first and last",
			cvsAr: image.Rect(0, 0, 4, 4),
			baseP: &axes.XProperties{
				Min: 0,
				Max: 0,
			},
			minCell: 1,
			maxCell: 2,
			wantMin: axes.NewValue(0, 2),
			wantMax: axes.NewValue(0, 2),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			base, err := axes.NewXDetails(tc.cvsAr, tc.baseP)
			if err != nil {
				t.Fatalf("NewXDetails => unexpected error: %v", err)
			}

			gotMin, gotMax, err := findCellPair(base, tc.minCell, tc.maxCell)
			if (err != nil) != tc.wantErr {
				t.Errorf("findCellPair => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.wantMin, gotMin); diff != "" {
				t.Errorf("findCellPair => unexpected min, diff (-want, +got):\n%s", diff)
			}
			if diff := pretty.Compare(tc.wantMax, gotMax); diff != "" {
				t.Errorf("findCellPair => unexpected max, diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestZoomToHighlight(t *testing.T) {
	tests := []struct {
		desc    string
		baseP   *axes.XProperties
		hRange  *Range
		cvsAr   image.Rectangle
		wantP   *axes.XProperties
		wantErr bool
	}{
		{
			desc:  "fails on impossible range",
			cvsAr: image.Rect(0, 0, 4, 4),
			baseP: &axes.XProperties{
				Min: 0,
				Max: 3,
			},
			hRange:  &Range{Start: -1, End: 2},
			wantErr: true,
		},
		{
			desc:  "zooms to highlighted area",
			cvsAr: image.Rect(0, 0, 4, 4),
			baseP: &axes.XProperties{
				Min: 0,
				Max: 3,
			},
			hRange: &Range{Start: 1, End: 3},
			wantP: &axes.XProperties{
				Min: 1,
				Max: 2,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			base, err := axes.NewXDetails(tc.cvsAr, tc.baseP)
			if err != nil {
				t.Fatalf("NewXDetails => unexpected error: %v", err)
			}

			got, err := zoomToHighlight(base, tc.hRange, tc.cvsAr)
			if (err != nil) != tc.wantErr {
				t.Errorf("zoomToHighlight => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			var want *axes.XDetails
			if tc.wantP != nil {
				w, err := axes.NewXDetails(tc.cvsAr, tc.wantP)
				if err != nil {
					t.Fatalf("NewXDetails => unexpected error: %v", err)
				}
				want = w
			}
			if diff := pretty.Compare(want, got); diff != "" {
				t.Errorf("zoomToHighlight => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestZoomToScroll(t *testing.T) {
	tests := []struct {
		desc    string
		mouse   *terminalapi.Mouse
		cvsAr   image.Rectangle
		graphAr image.Rectangle
		currP   *axes.XProperties
		baseP   *axes.XProperties
		opts    []Option
		wantP   *axes.XProperties
		wantErr bool
	}{
		{
			desc: "scroll up in the middle zooms in evenly",
			opts: []Option{
				ScrollStep(30),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{4, 0},
				Button:   mouse.ButtonWheelUp,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			wantP: &axes.XProperties{
				Min:       1,
				Max:       3,
				ReqYWidth: 2,
			},
		},
		{
			desc: "scroll up at the left edge",
			opts: []Option{
				ScrollStep(30),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{2, 0},
				Button:   mouse.ButtonWheelUp,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			wantP: &axes.XProperties{
				Min:       0,
				Max:       3,
				ReqYWidth: 2,
			},
		},
		{
			desc: "scroll up at the right edge",
			opts: []Option{
				ScrollStep(30),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{6, 0},
				Button:   mouse.ButtonWheelUp,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			wantP: &axes.XProperties{
				Min:       1,
				Max:       4,
				ReqYWidth: 2,
			},
		},
		{
			desc: "zoom in when current is already zoomed",
			opts: []Option{
				ScrollStep(30),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{4, 0},
				Button:   mouse.ButtonWheelUp,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       1,
				Max:       3,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			wantP: &axes.XProperties{
				Min:       2,
				Max:       3,
				ReqYWidth: 2,
			},
		},
		{
			desc: "zoom in moves min over the current max",
			opts: []Option{
				ScrollStep(150),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{6, 0},
				Button:   mouse.ButtonWheelUp,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			wantP: &axes.XProperties{
				Min:       3,
				Max:       4,
				ReqYWidth: 2,
			},
		},
		{
			desc: "zoom in moves max under the current min",
			opts: []Option{
				ScrollStep(150),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{2, 0},
				Button:   mouse.ButtonWheelUp,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			wantP: &axes.XProperties{
				Min:       0,
				Max:       1,
				ReqYWidth: 2,
			},
		},
		{
			desc: "scroll down in the middle zooms out evenly",
			opts: []Option{
				ScrollStep(30),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{4, 0},
				Button:   mouse.ButtonWheelDown,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       2,
				Max:       3,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			wantP: &axes.XProperties{
				Min:       1,
				Max:       4,
				ReqYWidth: 2,
			},
		},
		{
			desc: "scroll down in the middle zooms out completely",
			opts: []Option{
				ScrollStep(30),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{4, 0},
				Button:   mouse.ButtonWheelDown,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       1,
				Max:       3,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
		},
		{
			desc: "scroll down at the left edge",
			opts: []Option{
				ScrollStep(30),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{2, 0},
				Button:   mouse.ButtonWheelDown,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       1,
				Max:       3,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			wantP: &axes.XProperties{
				Min:       1,
				Max:       4,
				ReqYWidth: 2,
			},
		},
		{
			desc: "scroll down at the right edge",
			opts: []Option{
				ScrollStep(30),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{6, 0},
				Button:   mouse.ButtonWheelDown,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       1,
				Max:       3,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
			wantP: &axes.XProperties{
				Min:       0,
				Max:       3,
				ReqYWidth: 2,
			},
		},
		{
			desc: "zoom out moves min below base, zooms out completely",
			opts: []Option{
				ScrollStep(150),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{6, 0},
				Button:   mouse.ButtonWheelDown,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       1,
				Max:       3,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
		},
		{
			desc: "zoom out moves max above base, zooms out completely",
			opts: []Option{
				ScrollStep(150),
			},
			mouse: &terminalapi.Mouse{
				Position: image.Point{2, 0},
				Button:   mouse.ButtonWheelDown,
			},
			cvsAr:   image.Rect(0, 0, 8, 8),
			graphAr: image.Rect(2, 0, 8, 8),
			currP: &axes.XProperties{
				Min:       1,
				Max:       3,
				ReqYWidth: 2,
			},
			baseP: &axes.XProperties{
				Min:       0,
				Max:       4,
				ReqYWidth: 2,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			var curr *axes.XDetails
			if tc.currP != nil {
				c, err := axes.NewXDetails(tc.cvsAr, tc.currP)
				if err != nil {
					t.Fatalf("NewXDetails => unexpected error: %v", err)
				}
				curr = c
			}

			var base *axes.XDetails
			if tc.baseP != nil {
				b, err := axes.NewXDetails(tc.cvsAr, tc.baseP)
				if err != nil {
					t.Fatalf("NewXDetails => unexpected error: %v", err)
				}
				base = b
			}

			got, err := zoomToScroll(tc.mouse, tc.cvsAr, tc.graphAr, curr, base, newOptions(tc.opts...))
			if (err != nil) != tc.wantErr {
				t.Errorf("zoomToScroll => unexpected error: %v, wantErr: %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			var want *axes.XDetails
			if tc.wantP != nil {
				w, err := axes.NewXDetails(tc.cvsAr, tc.wantP)
				if err != nil {
					t.Fatalf("NewXDetails => unexpected error: %v", err)
				}
				want = w
			}
			if diff := pretty.Compare(want, got); diff != "" {
				t.Errorf("zoomToHighlight => unexpected diff (-want, +got):\n%s", diff)
			}

		})
	}
}
