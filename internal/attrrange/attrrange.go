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

// Package attrrange simplifies tracking of attributes that apply to a range of
// items.
// Refer to the examples in the test file for details on usage.
package attrrange

import (
	"fmt"
	"sort"
)

// AttrRange is a range of items that share the same attributes.
type AttrRange struct {
	// Low is the first position where these attributes apply.
	Low int

	// High is the end of the range. The attributes apply to all items in range
	// Low <= b < high.
	High int

	// AttrIdx is the index of the attributes that apply to this range.
	AttrIdx int
}

// newAttrRange returns a new AttrRange instance.
func newAttrRange(low, high, attrIdx int) *AttrRange {
	return &AttrRange{
		Low:     low,
		High:    high,
		AttrIdx: attrIdx,
	}
}

// Tracker tracks attributes that apply to a range of items.
// This object is not thread safe.
type Tracker struct {
	// ranges maps low indices of ranges to the attribute ranges.
	ranges map[int]*AttrRange
}

// NewTracker returns a new tracker of ranges that share the same attributes.
func NewTracker() *Tracker {
	return &Tracker{
		ranges: map[int]*AttrRange{},
	}
}

// Add adds a new range of items that share attributes with the specified
// index.
// The low position of the range must not overlap with low position of any
// existing range.
func (t *Tracker) Add(low, high, attrIdx int) error {
	ar := newAttrRange(low, high, attrIdx)
	if ar, ok := t.ranges[low]; ok {
		return fmt.Errorf("already have range starting on low:%d, existing:%+v", low, ar)
	}
	t.ranges[low] = ar
	return nil
}

// ForPosition returns attribute index that apply to the specified position.
// Returns ErrNotFound when the requested position wasn't found in any of the
// known ranges.
func (t *Tracker) ForPosition(pos int) (*AttrRange, error) {
	if ar, ok := t.ranges[pos]; ok {
		return ar, nil
	}

	var keys []int
	for k := range t.ranges {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	var res *AttrRange
	for _, k := range keys {
		ar := t.ranges[k]
		if ar.Low > pos {
			break
		}
		if ar.High > pos {
			res = ar
		}
	}

	if res == nil {
		return nil, fmt.Errorf("did not find attribute range for position %d", pos)
	}
	return res, nil
}
