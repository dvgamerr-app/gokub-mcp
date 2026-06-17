package tools

import "testing"

func TestIsStopTriggered(t *testing.T) {
	cases := []struct {
		side    string
		price   float64
		trigger float64
		want    bool
	}{
		{"sell", 47.0, 47.5, true},  // long stop: price fell below trigger
		{"sell", 48.0, 47.5, false}, // still above stop
		{"sell", 47.5, 47.5, true},  // exactly at trigger fires
		{"buy", 51.0, 50.5, true},   // breakout: price rose above trigger
		{"buy", 50.0, 50.5, false},  // not yet broken out
	}
	for _, c := range cases {
		if got := isStopTriggered(c.side, c.price, c.trigger); got != c.want {
			t.Errorf("isStopTriggered(%s, %.2f, %.2f) = %v, want %v", c.side, c.price, c.trigger, got, c.want)
		}
	}
}
