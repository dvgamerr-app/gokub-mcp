package tools

import (
	"math"
	"testing"
)

func TestNewNetTradePlan(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		stop, balance, risk, target float64
		wantErr                     bool
	}{
		{"one percent", 98, 1000, 2, 1, false},
		{"two percent capped to cash", 99.9, 1000, 2, 2, false},
		{"risk above limit", 98, 1000, 3, 2, true},
		{"invalid stop", 100, 1000, 2, 2, true},
		{"non finite", 98, math.NaN(), 2, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := newNetTradePlan(100, tc.stop, tc.balance, tc.risk, tc.target, .25, .25, .1)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			profit := (p.Target*(1-.0025) - p.Entry*(1+.0025+.001)) * p.Qty
			if math.Abs(profit-p.NetProfit) > 1e-8 {
				t.Fatalf("net profit=%v want %v", profit, p.NetProfit)
			}
			if p.EntryCost > tc.balance+1e-8 || p.NetLoss > tc.balance*tc.risk/100+1e-8 {
				t.Fatal("cash/risk limit exceeded")
			}
		})
	}
}
