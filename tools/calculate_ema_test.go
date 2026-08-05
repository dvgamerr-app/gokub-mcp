package tools

import (
	"context"
	"testing"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
)

// TestCalculateEMAFormula verifies EMA series + current value (ASSIGNMENT.md §3, tool 11).
// period=3, prices=[10,11,12,13,14], k=2/(3+1)=0.5
// SMA(3)=11.0=ema[2], ema[3]=(13-11)*0.5+11=12.0, ema[4]=(14-12)*0.5+12=13.0 (current)
func TestCalculateEMAFormula(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"period": float64(3),
				"prices": []any{10.0, 11.0, 12.0, 13.0, 14.0},
			},
		},
	}
	result, err := CalculateEMAHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := result.StructuredContent.(*EMAResult)
	if !ok {
		t.Fatal("StructuredContent is not *EMAResult")
	}
	if out.Current != 13.0 {
		t.Errorf("current_ema: want 13.0, got %v", out.Current)
	}
	if out.Previous != 12.0 {
		t.Errorf("previous_ema: want 12.0, got %v", out.Previous)
	}
	if out.Trend != "bullish" {
		t.Errorf("trend: want bullish (current > previous), got %s", out.Trend)
	}
}

func TestCalculateEMASinglePrice(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"period": float64(1),
				"prices": []any{42.0},
			},
		},
	}
	result, err := CalculateEMAHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := result.StructuredContent.(*EMAResult)
	if !ok {
		t.Fatal("StructuredContent is not *EMAResult")
	}
	if out.Current != 42 || out.Previous != 42 || out.Trend != "neutral" {
		t.Fatalf("single-price EMA = current %v, previous %v, trend %q", out.Current, out.Previous, out.Trend)
	}
}

func TestCalculateEMAHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"period": float64(3),
				"prices": []any{10.0, 11.0, 12.0, 13.0, 14.0},
			},
			wantErr: false,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"period": float64(10),
				"prices": []any{10.0, 11.0},
			},
			wantErr: true,
		},
		{
			name: "Invalid prices type",
			args: map[string]any{
				"period": float64(3),
				"prices": []any{"10", "11"},
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
			result, err := CalculateEMAHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CalculateEMAHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CalculateEMAHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CalculateEMAHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
