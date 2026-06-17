package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestCalculateRSRankOrdering verifies ranking by ROC and top3 output (ASSIGNMENT.md §4, tool 18).
// ETH ROC = (2400−2000)/2000×100 = 20%, BTC ROC = (110−100)/100×100 = 10%
// ETH should rank #1, BTC #2; top3[0] = "ETH"
func TestCalculateRSRankOrdering(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"period": float64(1),
				"symbols": map[string]any{
					"BTC": []any{100.0, 110.0},
					"ETH": []any{2000.0, 2400.0},
				},
				"benchmark": "BTC",
			},
		},
	}
	result, err := CalculateRelativeStrengthRankHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := result.StructuredContent.(*RSRankResult)
	if !ok {
		t.Fatal("StructuredContent is not *RSRankResult")
	}
	if len(out.Top3) == 0 || out.Top3[0] != "ETH" {
		t.Errorf("top3[0]: want ETH (ROC 20%% > BTC 10%%), got %v", out.Top3)
	}
	if out.ROC != 10.0 {
		t.Errorf("benchmark_roc (BTC): want 10.0, got %v", out.ROC)
	}
	if len(out.Rankings) < 2 || out.Rankings[0].Symbol != "ETH" {
		t.Errorf("rankings[0] should be ETH, got %v", out.Rankings)
	}
}

func TestCalculateRelativeStrengthRankHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"period": float64(2),
				"symbols": map[string]any{
					"BTC": []any{100.0, 105.0, 110.0},
					"ETH": []any{2000.0, 2100.0, 2200.0},
				},
				"benchmark": "BTC",
			},
			wantErr: false,
		},
		{
			name: "Invalid symbols format",
			args: map[string]any{
				"symbols": "invalid",
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
			result, err := CalculateRelativeStrengthRankHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CalculateRelativeStrengthRankHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CalculateRelativeStrengthRankHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CalculateRelativeStrengthRankHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
