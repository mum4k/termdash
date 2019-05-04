package linechart

import (
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
)

func TestFormatters(t *testing.T) {
	tests := []struct {
		desc      string
		value     float64
		formatter func(float64) string
		want      string
	}{
		{
			desc:      "Pretty duration formatter handles zero values",
			value:     0,
			formatter: SingleUnitSecondsValueFormatter,
			want:      "0ns",
		},
		{
			desc:      "Pretty duration formatter handles minus minute values",
			value:     -1500,
			formatter: SingleUnitSecondsValueFormatter,
			want:      "-25m",
		},
		{
			desc:      "Pretty duration formatter handles minus minute values",
			value:     -60,
			formatter: SingleUnitSecondsValueFormatter,
			want:      "-1m",
		},
		{
			desc:      "Pretty duration formatter handles nanoseconds",
			value:     1.23e-7,
			formatter: SingleUnitSecondsValueFormatter,
			want:      "123ns",
		},
		{
			desc:      "Pretty duration formatter handles microseconds",
			value:     1.23e-4,
			formatter: SingleUnitSecondsValueFormatter,
			want:      "123µs",
		},
		{
			desc:      "Pretty duration formatter handles milliseconds",
			value:     0.123,
			formatter: SingleUnitSecondsValueFormatter,
			want:      "123ms",
		},
		{
			desc:      "Pretty duration formatter handles seconds",
			value:     12,
			formatter: SingleUnitSecondsValueFormatter,
			want:      "12s",
		},
		{
			desc:      "Pretty duration formatter handles minutes",
			value:     60,
			formatter: SingleUnitSecondsValueFormatter,
			want:      "1m",
		},
		{
			desc:      "Pretty duration formatter handles hours",
			value:     2 * 60 * 60,
			formatter: SingleUnitSecondsValueFormatter,
			want:      "2h",
		},
		{
			desc:      "Pretty duration formatter handles days",
			value:     5 * 24 * 60 * 60,
			formatter: SingleUnitSecondsValueFormatter,
			want:      "5d",
		},
		{
			desc:      "Pretty minus duration formatter handles days",
			value:     -5 * 24 * 60 * 60,
			formatter: SingleUnitSecondsValueFormatter,
			want:      "-5d",
		},
		{
			desc:      "Pretty custom minute formatter with decimals handles days",
			value:     135,
			formatter: SingleUnitDurationValueFormatter(time.Minute, 2),
			want:      "2.25h",
		},
		{
			desc:      "Pretty custom millisecond formatter with decimals handles minutes",
			value:     2525789,
			formatter: SingleUnitDurationValueFormatter(time.Millisecond, 4),
			want:      "42.0965m",
		},
		{
			desc:      "Pretty custom nanosecond formatter with decimals handles days",
			value:     999999999999999,
			formatter: SingleUnitDurationValueFormatter(time.Nanosecond, 8),
			want:      "11.57407407d",
		},
		{
			desc:      "Pretty custom minus nanosecond formatter with decimals handles days",
			value:     -999999999999999,
			formatter: SingleUnitDurationValueFormatter(time.Nanosecond, 8),
			want:      "-11.57407407d",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.formatter(tc.value)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("formatter => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
