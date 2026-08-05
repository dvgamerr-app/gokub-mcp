package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestCalculateRSIFormula verifies RSI(0-100) and signal (ASSIGNMENT.md §6, tool 14).
// period=2, prices=[100,105,102,108,110]
// avgGain=2.5→4.25→3.125, avgLoss=1.5→0.75→0.375, RS=8.333, RSI=89.29 (overbought)
func TestCalculateRSIFormula(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"period": float64(2),
				"prices": []any{100.0, 105.0, 102.0, 108.0, 110.0},
			},
		},
	}
	result, err := CalculateRSIHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := result.StructuredContent.(*RSIResult)
	if !ok {
		t.Fatal("StructuredContent is not *RSIResult")
	}
	if out.RSI != 89.29 {
		t.Errorf("rsi: want 89.29, got %v", out.RSI)
	}
	if out.Signal != "overbought" {
		t.Errorf("signal: want overbought (RSI ≥70), got %s", out.Signal)
	}
}

func TestCalculateRSIHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"period": float64(2),
				"prices": []any{100.0, 105.0, 102.0, 108.0, 110.0},
			},
			wantErr: false,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"period": float64(14),
				"prices": []any{100.0, 105.0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.args,
				},
			}
			result, err := CalculateRSIHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CalculateRSIHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CalculateRSIHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CalculateRSIHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}

func BenchmarkCalculateRSI(b *testing.B) {
	prices := make([]float64, 512)
	for i := range prices {
		prices[i] = 100 + float64(i%17)
	}

	b.ReportAllocs()
	for b.Loop() {
		calculateRSI(prices, 14)
	}
}
