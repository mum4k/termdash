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

package text

// write_options.go contains options used when writing content to the Text widget.

import (
	"sort"

	"github.com/mum4k/termdash/cell"
)

// WriteOption is used to provide options to Write().
type WriteOption interface {
	// set sets the provided option.
	set(*writeOptions)
}

// writeOptions stores the provided options.
type writeOptions struct {
	cellOpts *cell.Options
}

// newWriteOptions returns new writeOptions instance.
func newWriteOptions(wOpts ...WriteOption) *writeOptions {
	wo := &writeOptions{
		cellOpts: cell.NewOptions(),
	}
	for _, o := range wOpts {
		o.set(wo)
	}
	return wo
}

// writeOption implements WriteOption.
type writeOption func(*writeOptions)

// set implements WriteOption.set.
func (wo writeOption) set(wOpts *writeOptions) {
	wo(wOpts)
}

// WriteCellOpts sets options on the cells that contain the text.
func WriteCellOpts(opts ...cell.Option) WriteOption {
	return writeOption(func(wOpts *writeOptions) {
		wOpts.cellOpts = cell.NewOptions(opts...)
	})
}

// optsRange are write options that apply to a range of bytes in the text.
type optsRange struct {
	// low is the first byte where these options apply.
	low int

	// high is the end of the range. The opts apply to all bytes in range low
	// <= b < high.
	high int

	// opts are the options provided at a call to Write().
	opts *writeOptions
}

// newOptsRange returns a new optsRange.
func newOptsRange(low, high int, opts *writeOptions) *optsRange {
	return &optsRange{
		low:  low,
		high: high,
		opts: opts,
	}
}

// givenWOpts stores the write options provided on all the calls to Write().
// The options are keyed by their low indices.
type givenWOpts map[int]*optsRange

// newGivenWOpts returns a new givenWOpts instance.
func newGivenWOpts() givenWOpts {
	return givenWOpts{}
}

// forPosition returns write options that apply to character at the specified
// byte position.
func (g givenWOpts) forPosition(pos int) *optsRange {
	if or, ok := g[pos]; ok {
		return or
	}

	var keys []int
	for k := range g {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	res := newOptsRange(0, 0, newWriteOptions())
	for _, k := range keys {
		or := g[k]
		if or.low > pos {
			break
		}
		if or.high > pos {
			res = or
		}
	}
	return res
}
