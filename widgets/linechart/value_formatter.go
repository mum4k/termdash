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

package linechart

// value_formatter.go provides common implementations of ValueFormatter that can be
// used with the YAxisFormattedValues() LineChart option.

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// durationSingleUnitPrettyFormat returns the pretty format in one single
// unit for a time.Duration, the different returned unit formats
// are: nanoseconds, microseconds, milliseconds, seconds, minutes
// hours, days.
func durationSingleUnitPrettyFormat(d time.Duration, decimals int) string {
	// Check if the duration is less than 0.
	prefix := ""
	if d < 0 {
		prefix = "-"
		d = time.Duration(math.Abs(d.Seconds()) * float64(time.Second))
	}

	switch {
	// Nanoseconds.
	case d.Nanoseconds() < 1000:
		dFmt := prefix + "%dns"
		return fmt.Sprintf(dFmt, d.Nanoseconds())
	// Microseconds.
	case d.Seconds()*1000*1000 < 1000:
		dFmt := prefix + suffixDecimalFormat(decimals, "µs")
		return fmt.Sprintf(dFmt, d.Seconds()*1000*1000)
	// Milliseconds.
	case d.Seconds()*1000 < 1000:
		dFmt := prefix + suffixDecimalFormat(decimals, "ms")
		return fmt.Sprintf(dFmt, d.Seconds()*1000)
	// Seconds.
	case d.Seconds() < 60:
		dFmt := prefix + suffixDecimalFormat(decimals, "s")
		return fmt.Sprintf(dFmt, d.Seconds())
	// Minutes.
	case d.Minutes() < 60:
		dFmt := prefix + suffixDecimalFormat(decimals, "m")
		return fmt.Sprintf(dFmt, d.Minutes())
	// Hours.
	case d.Hours() < 24:
		dFmt := prefix + suffixDecimalFormat(decimals, "h")
		return fmt.Sprintf(dFmt, d.Hours())
	// Days.
	default:
		dFmt := prefix + suffixDecimalFormat(decimals, "d")
		return fmt.Sprintf(dFmt, d.Hours()/24)
	}
}

func suffixDecimalFormat(decimals int, suffix string) string {
	suffix = strings.Replace(suffix, "%", "%%", -1) // Safe `%` character for fmt.
	return fmt.Sprintf("%%.%df%s", decimals, suffix)
}

// ValueFormatterSingleUnitDuration is a factory to create a custom duration
// in a single unit representation formatter based on a unit and the decimals
// to truncate.
// If the received decimal value is negative it will fallback to a 0 decimal
// value.
// The result value formatter handles NaN values, if the value formatter
// receives a NaN float64 it will return an empty string.
func ValueFormatterSingleUnitDuration(unit time.Duration, decimals int) ValueFormatter {
	if decimals < 0 {
		decimals = 0
	}

	return func(v float64) string {
		if math.IsNaN(v) {
			return ""
		}

		d := time.Duration(v * float64(unit))
		return durationSingleUnitPrettyFormat(d, decimals)
	}
}

// ValueFormatterSingleUnitSeconds is a formatter that will receive
// seconds unit in the float64 argument and will return a pretty
// format in one single unit without decimals, it doesn't round,
// it truncates.
// Received seconds that are NaN will be ignored and return an
// empty string.
func ValueFormatterSingleUnitSeconds(seconds float64) string {
	f := ValueFormatterSingleUnitDuration(time.Second, 0)
	return f(seconds)
}

// ValueFormatterRound is a formatter that will receive a float64
// value and will round to the nearest value without decimals.
func ValueFormatterRound(value float64) string {
	f := ValueFormatterRoundWithSuffix("")
	return f(value)
}

// ValueFormatterRoundWithSuffix is a factory that returns a formatter
// that will receive a float64 value and will round to the nearest value
// without decimals adding a suffix to the final value string representation.
func ValueFormatterRoundWithSuffix(suffix string) ValueFormatter {
	return valueFormatterSuffixWithTransformer(0, suffix, math.Round)
}

// ValueFormatterSuffix is a factory that returns a formatter
// that will receive a float64 value and return a string representation with
// the desired number of decimal truncated and a suffix.
func ValueFormatterSuffix(decimals int, suffix string) ValueFormatter {
	return valueFormatterSuffixWithTransformer(decimals, suffix, nil)
}

// valueFormatterSuffixWithTransformer is a factory that returns a formatter
// that will apply a transform function to the received value before
// returning the decimal with suffix representation.
func valueFormatterSuffixWithTransformer(decimals int, suffix string, transformFunc func(float64) float64) ValueFormatter {
	dFmt := suffixDecimalFormat(decimals, suffix)
	return func(value float64) string {
		if math.IsNaN(value) {
			return ""
		}

		if transformFunc != nil {
			value = transformFunc(value)
		}

		return fmt.Sprintf(dFmt, value)
	}
}
