package tools

import (
	"fmt"
	"github.com/dvgamerr-app/go-bitkub/market"
	"strings"
	"testing"
)

func TestFetchOpenOrders(t *testing.T) {
	for _, tc := range []struct {
		name, source, input, wantErr string
		requestErr                   bool
	}{
		{name: "exchange empty orders", source: "exchange", input: "BTC_THB"},
		{name: "broker unsupported", source: "broker", input: "btc_thb", wantErr: "order status is unknown"},
		{name: "unknown symbol", source: "exchange", input: "missing_thb", wantErr: "symbol not found"},
		{name: "safe SDK error", source: "exchange", input: "btc_thb", requestErr: true, wantErr: "400 Bad Request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			orders, err := fetchOpenOrders(tc.input, func() ([]market.Symbol, error) { return []market.Symbol{{Symbol: "BTC_THB", Source: tc.source}}, nil }, func(symbol string) ([]market.OpenOrder, error) {
				called = true
				if symbol != "btc_thb" {
					t.Fatalf("symbol=%s", symbol)
				}
				if tc.requestErr {
					return nil, fmt.Errorf("400 Bad Request : request headers SECRET")
				}
				return []market.OpenOrder{}, nil
			})
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error=%v", err)
				}
				if strings.Contains(err.Error(), "SECRET") {
					t.Fatal("request header leaked")
				}
			} else if err != nil || len(orders) != 0 {
				t.Fatalf("orders=%v error=%v", orders, err)
			}
			if (tc.source == "broker" || tc.input == "missing_thb") && called {
				t.Fatal("unsupported symbol sent to private endpoint")
			}
		})
	}
}
