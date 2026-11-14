package tools

import (
	"errors"
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
