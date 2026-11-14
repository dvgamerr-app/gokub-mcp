package tools

import (
	"errors"
	"github.com/dvgamerr-app/go-bitkub/market"
	"strings"
	"testing"
)

func TestRecoverMarketRequestReturnsResult(t *testing.T) {
	result, err := recoverMarketRequest(func() (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("recoverMarketRequest returned error: %v", err)
	}
	if result != "ok" {
		t.Fatalf("result = %q, want ok", result)
	}
}

func TestExchangePairAllowed(t *testing.T) {
	for _, tc := range []struct {
		name, source, quote string
		frozen, want        bool
	}{
		{"exchange THB", "exchange", "THB", false, true},
		{"broker", "broker", "THB", false, false},
		{"non THB", "exchange", "USDT", false, false},
		{"frozen", "exchange", "THB", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := exchangePairAllowed(market.Symbol{Source: tc.source, QuoteAsset: tc.quote, FreezeBuy: tc.frozen})
			if got != tc.want {
				t.Fatalf("allowed=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestRecoverMarketRequestReturnsError(t *testing.T) {
	expected := errors.New("request failed")
	_, err := recoverMarketRequest(func() (string, error) {
		return "", expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("error = %v, want %v", err, expected)
	}
}

func TestRecoverMarketRequestConvertsPanicToError(t *testing.T) {
	_, err := recoverMarketRequest(func() (string, error) {
		panic("nil pointer dereference")
	})
	if err == nil || !strings.Contains(err.Error(), "nil pointer dereference") {
		t.Fatalf("unexpected panic error: %v", err)
	}
}
