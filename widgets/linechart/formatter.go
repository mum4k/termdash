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
		dfmt := prefix + "%dns"
		return fmt.Sprintf(dfmt, d.Nanoseconds())
	// Microseconds.
	case d.Seconds()*1000*1000 < 1000:
		dfmt := prefix + decimalFormat(decimals, "µs")
		return fmt.Sprintf(dfmt, d.Seconds()*1000*1000)
	// Milliseconds.
	case d.Seconds()*1000 < 1000:
		dfmt := prefix + decimalFormat(decimals, "ms")
		return fmt.Sprintf(dfmt, d.Seconds()*1000)
	// Seconds.
	case d.Seconds() < 60:
		dfmt := prefix + decimalFormat(decimals, "s")
		return fmt.Sprintf(dfmt, d.Seconds())
	// Minutes.
	case d.Minutes() < 60:
		dfmt := prefix + decimalFormat(decimals, "m")
		return fmt.Sprintf(dfmt, d.Minutes())
	// Hours.
	case d.Hours() < 24:
		dfmt := prefix + decimalFormat(decimals, "h")
		return fmt.Sprintf(dfmt, d.Hours())
	// Days.
	default:
		dfmt := prefix + decimalFormat(decimals, "d")
		return fmt.Sprintf(dfmt, d.Hours()/24)
	}
}

func decimalFormat(decimals int, unit string) string {
	return fmt.Sprintf("%%.%df%s", decimals, unit)
}

// SingleUnitDurationValueFormatter is a factory to create a custom duration
// in a single unit representation formatter based on a unit and the decimals
// to truncate.
func SingleUnitDurationValueFormatter(unit time.Duration, decimals int) ValueFormatter {
	return func(v float64) string {
		d := time.Duration(v * float64(unit))
		return durationSingleUnitPrettyFormat(d, decimals)
	}
}

// SingleUnitSecondsValueFormatter is a formatter that will receive
// seconds unit in the float64 argument and will return a pretty
// format in one single unit without decimals, it doesn't round,
// it truncates.
var SingleUnitSecondsValueFormatter = SingleUnitDurationValueFormatter(time.Second, 0)
