package utils

import "math"

func Round(value float64, places ...int) float64 {
	p := 8
	if len(places) > 0 {
		p = places[0]
	}
	factor := math.Pow10(p)
	return math.Round(value*factor) / factor
}
