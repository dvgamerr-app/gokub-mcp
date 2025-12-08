package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

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
