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

package linechart

import (
	"fmt"
	"math"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart/axes"
)

// options.go contains configurable options for LineChart.

// Option is used to provide options to New().
type Option interface {
	// set sets the provided option.
	set(*options)
}

// options stores the provided options.
type options struct {
	axesCellOpts      []cell.Option
	xLabelCellOpts    []cell.Option
	xLabelOrientation axes.LabelOrientation
	yLabelCellOpts    []cell.Option
	yAxisMode         axes.YScaleMode
	yAxisCustomScale  *customScale
}

// validate validates the provided options.
func (o *options) validate() error {
	if o.yAxisCustomScale == nil {
		return nil
	}

	if math.IsNaN(o.yAxisCustomScale.min) || math.IsNaN(o.yAxisCustomScale.max) {
		return fmt.Errorf("both the min(%v) and the max(%v) provided as custom Y scale must be valid numbers", o.yAxisCustomScale.min, o.yAxisCustomScale.max)
	}
	if o.yAxisCustomScale.min >= o.yAxisCustomScale.max {
		return fmt.Errorf("the min(%v) must be less than the max(%v) provided as custom Y scale", o.yAxisCustomScale.min, o.yAxisCustomScale.max)
	}
	return nil
}

// newOptions returns a new options instance.
func newOptions(opts ...Option) *options {
	opt := &options{}
	for _, o := range opts {
		o.set(opt)
	}
	return opt
}

// option implements Option.
type option func(*options)

// set implements Option.set.
func (o option) set(opts *options) {
	o(opts)
}

// AxesCellOpts set the cell options for the X and Y axes.
func AxesCellOpts(co ...cell.Option) Option {
	return option(func(opts *options) {
		opts.axesCellOpts = co
	})
}

// XLabelCellOpts set the cell options for the labels on the X axis.
func XLabelCellOpts(co ...cell.Option) Option {
	return option(func(opts *options) {
		opts.xLabelCellOpts = co
	})
}

// XLabelsVertical makes the labels under the X axis flow vertically.
// Defaults to labels that flow horizontally.
func XLabelsVertical() Option {
	return option(func(opts *options) {
		opts.xLabelOrientation = axes.LabelOrientationVertical
	})
}

// XLabelsHorizontal makes the labels under the X axis flow horizontally.
// This is the default option.
func XLabelsHorizontal() Option {
	return option(func(opts *options) {
		opts.xLabelOrientation = axes.LabelOrientationHorizontal
	})
}

// YLabelCellOpts set the cell options for the labels on the Y axis.
func YLabelCellOpts(co ...cell.Option) Option {
	return option(func(opts *options) {
		opts.yLabelCellOpts = co
	})
}

// YAxisAdaptive makes the Y axis adapt its base value depending on the
// provided series.
// Without this option, the Y axis always starts at the zero value regardless of
// values available in the series.
// When this option is specified and the series don't contain value zero, the Y
// axis will be adapted to the minimum value for all-positive series or the
// maximum value for all-negative series. The Y axis still starts at the zero
// value if the series contain both positive and negative values.
func YAxisAdaptive() Option {
	return option(func(opts *options) {
		opts.yAxisMode = axes.YScaleModeAdaptive
	})
}

// customScale is the custom scale provided via the YAxisCustomScale option.
type customScale struct {
	min, max float64
}

// YAxisCustomScale when provided, the scale of the Y axis will be based on the
// specified minimum and maximum value instead of determining those from the
// LineChart series. Useful to visually stabilize the Y axis for LineChart
// applications that continuously feed values.
// The default behavior is to continuously determine the minimum and maximum
// value from the series before drawing the LineChart.
// Even when this option is provided, the LineChart would still rescale the Y
// axis if a value is encountered that is outside of the range specified here,
// i.e. smaller than the minimum or larger than the maximum.
// Both the minimum and the maximum must be valid numbers and the minimum must
// be smaller than the maximum.
func YAxisCustomScale(min, max float64) Option {
	return option(func(opts *options) {
		opts.yAxisCustomScale = &customScale{
			min: min,
			max: max,
		}
	})
}
