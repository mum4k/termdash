// Copyright 2018 Google Inc.
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

package faketerm

// diff.go provides functions that highlight differences between fake terminals.

import (
	"bytes"
	"fmt"
	"image"
	"reflect"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/cell"
)

// optDiff is used to display differences in cell options.
type optDiff struct {
	// point indicates the cell with the differing options.
	point image.Point

	got  *cell.Options
	want *cell.Options
}

// Diff compares the two terminals, returning an empty string if there is not
// difference. If a difference is found, returns a human readable description
// of the differences.
func Diff(want, got *Terminal) string {
	if reflect.DeepEqual(want.BackBuffer(), got.BackBuffer()) {
		return ""
	}

	var b bytes.Buffer
	b.WriteString("found differences between the two fake terminals.\n")
	b.WriteString("   got:\n")
	b.WriteString(got.String())
	b.WriteString("  want:\n")
	b.WriteString(want.String())
	b.WriteString("  diff (unexpected cells highlighted with rune '࿃')\n")
	b.WriteString("  note - this excludes cell options:\n")

	size := got.Size()
	var optDiffs []*optDiff
	cellsDiffer := false
	for row := 0; row < size.Y; row++ {
		for col := 0; col < size.X; col++ {
			p := image.Point{col, row}
			partial, err := got.BackBuffer().IsPartial(p)
			if err != nil {
				panic(fmt.Errorf("unable to determine if point %v is a partial rune: %v", p, err))
			}

			gotCell := got.BackBuffer()[col][row]
			wantCell := want.BackBuffer()[col][row]
			r := gotCell.Rune
			if r != wantCell.Rune {
				r = '࿃'
				cellsDiffer = true
			} else if r == 0 && !partial {
				r = ' '
			}
			b.WriteRune(r)

			if !reflect.DeepEqual(gotCell.Opts, wantCell.Opts) {
				optDiffs = append(optDiffs, &optDiff{
					point: image.Point{col, row},
					got:   gotCell.Opts,
					want:  wantCell.Opts,
				})
			}
		}
		b.WriteRune('\n')
	}

	if len(optDiffs) > 0 {
		b.WriteString("  Found differences in options on some of the cells:\n")
		for _, od := range optDiffs {
			if diff := pretty.Compare(od.want, od.got); diff != "" {
				b.WriteString(fmt.Sprintf("cell %v, diff (-want +got):\n%s\n", od.point, diff))
			}
		}
	}

	if cellsDiffer {
		b.WriteString("  Found differences in some of the cell runes:\n")
		for row := 0; row < size.Y; row++ {
			for col := 0; col < size.X; col++ {
				got := got.BackBuffer()[col][row].Rune
				want := want.BackBuffer()[col][row].Rune
				b.WriteString(fmt.Sprintf("  cell(%v, %v) => got '%c' (rune %d), want '%c' (rune %d)\n", col, row, got, got, want, want))
			}
		}
	}
	return b.String()
}
