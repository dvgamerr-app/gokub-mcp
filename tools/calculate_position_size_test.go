package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCalculatePositionSizeHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"balance":      100000.0,
				"risk_percent": 2.0,
				"entry":        100.0,
				"stop":         95.0,
			},
			wantErr: false,
		},
		{
			name: "Stop higher than entry",
			args: map[string]any{
				"balance":      100000.0,
				"risk_percent": 2.0,
				"entry":        100.0,
				"stop":         105.0,
			},
			wantErr: true,
		},
		{
			name: "Invalid balance",
			args: map[string]any{
				"balance":      -100.0,
				"risk_percent": 2.0,
				"entry":        100.0,
				"stop":         95.0,
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
			result, err := CalculatePositionSizeHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CalculatePositionSizeHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CalculatePositionSizeHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CalculatePositionSizeHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}

// TestCalculatePositionSizeFormula verifies the ASSIGNMENT.md example:
// portfolio 2,000 THB, risk 2% = 40 THB, entry 50, stop 47.5 (5% away)
// → position = 40/0.05 = 800 THB → qty = 800/50 = 16 coins → TP 2R = 55
func TestCalculatePositionSizeFormula(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"balance":      2000.0,
				"risk_percent": 2.0,
				"entry":        50.0,
				"stop":         47.5,
				"maker_fee":    0.0,
				"taker_fee":    0.0,
			},
		},
	}
	result, err := CalculatePositionSizeHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := result.StructuredContent.(PositionSizeOutput)
	if !ok {
		t.Fatal("result.StructuredContent is not PositionSizeOutput")
	}
	if out.RiskTHB != 40 {
		t.Errorf("risk_thb: want 40, got %v", out.RiskTHB)
	}
	if out.PositionValueTHB != 800 {
		t.Errorf("position_value_thb: want 800, got %v", out.PositionValueTHB)
	}
	if out.Qty != 16 {
		t.Errorf("qty: want 16, got %v", out.Qty)
	}
	if out.TakeProfit2R != 55 {
		t.Errorf("take_profit_2R: want 55, got %v", out.TakeProfit2R)
	}
}
