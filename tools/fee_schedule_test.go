package tools

import "testing"

// TestDetermineFeeSchedule verifies all fee tiers (Standard → VIP 4) (ASSIGNMENT.md §1, §6, tool 6).
func TestDetermineFeeSchedule(t *testing.T) {
	cases := []struct {
		credits float64
		level   string
		maker   float64
		taker   float64
	}{
		{0, "Standard", 0.25, 0.25},
		{500, "Standard", 0.25, 0.25},
		{1000, "Level 1", 0.25, 0.25},
		{10000, "Level 2", 0.24, 0.24},
		{50000, "Level 3", 0.23, 0.23},
		{100000, "Level 4", 0.20, 0.23},
		{500000, "Level 5", 0.15, 0.23},
		{1000000, "VIP 1", 0.10, 0.23},
		{5000000, "VIP 2", 0.05, 0.20},
		{10000000, "VIP 3", 0.00, 0.15},
		{50000000, "VIP 4", 0.00, 0.10},
	}
	for _, c := range cases {
		got := determineFeeSchedule(c.credits)
		if got.Level != c.level {
			t.Errorf("credits=%.0f: level want %s, got %s", c.credits, c.level, got.Level)
		}
		if got.MakerFee != c.maker {
			t.Errorf("credits=%.0f: maker_fee want %v, got %v", c.credits, c.maker, got.MakerFee)
		}
		if got.TakerFee != c.taker {
			t.Errorf("credits=%.0f: taker_fee want %v, got %v", c.credits, c.taker, got.TakerFee)
		}
	}
}
