package utils

import (
	"math"
	"strings"
)

func BaseQuote(symbol string) string {
	parts := strings.Split(strings.ToUpper(strings.TrimSpace(symbol)), "_")
	if len(parts) == 2 && parts[0] == "THB" {
		return parts[1] + "_" + parts[0]
	}
	return strings.Join(parts, "_")
}

func Round(value float64, places ...int) float64 {
	p := 8
	if len(places) > 0 {
		p = places[0]
	}
	factor := math.Pow10(p)
	return math.Round(value*factor) / factor
}
