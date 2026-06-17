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
		{0, "Standard", 0.0025, 0.0025},
		{500, "Standard", 0.0025, 0.0025},
		{1000, "Level 1", 0.0025, 0.0025},
		{10000, "Level 2", 0.0024, 0.0024},
		{50000, "Level 3", 0.0023, 0.0023},
		{100000, "Level 4", 0.0020, 0.0023},
		{500000, "Level 5", 0.0015, 0.0023},
		{1000000, "VIP 1", 0.0010, 0.0023},
		{5000000, "VIP 2", 0.0005, 0.0020},
		{10000000, "VIP 3", 0.0000, 0.0015},
		{50000000, "VIP 4", 0.0000, 0.0010},
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
