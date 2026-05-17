// threed/utils.go

package threed

// clampFloat ensures values stay within specified bounds.
func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
