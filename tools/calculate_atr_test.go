package tools

import (
	"context"
	"testing"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
)

// TestCalculateATRFormula verifies ATR and ATR% calculation (ASSIGNMENT.md §5, §6).
// TR[1]=max(11-9,|11-9|,|9-9|)=2, TR[2]=max(12-10,|12-10|,|10-10|)=2
// ATR(2) = (TR[1]+TR[2])/2 = 2.0, ATR% = 2.0/11*100 = 18.18%
func TestCalculateATRFormula(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"period": float64(2),
				"candles": []any{
					map[string]any{"high": 10.0, "low": 8.0, "close": 9.0},
					map[string]any{"high": 11.0, "low": 9.0, "close": 10.0},
					map[string]any{"high": 12.0, "low": 10.0, "close": 11.0},
				},
			},
		},
	}
	result, err := CalculateATRHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := result.StructuredContent.(*ATRResult)
	if !ok {
		t.Fatal("StructuredContent is not *ATRResult")
	}
	if out.ATR != 2.0 {
		t.Errorf("ATR: want 2.0, got %v", out.ATR)
	}
	if out.ATRPercent != 18.18 {
		t.Errorf("ATR%%: want 18.18, got %v (formula: ATR/close×100)", out.ATRPercent)
	}
}

func TestCalculateATRHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"period": float64(2),
				"candles": []any{
					map[string]any{"high": 10.0, "low": 8.0, "close": 9.0},
					map[string]any{"high": 11.0, "low": 9.0, "close": 10.0},
					map[string]any{"high": 12.0, "low": 10.0, "close": 11.0},
				},
			},
			wantErr: false,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"period": float64(14),
				"candles": []any{
					map[string]any{"high": 10.0, "low": 8.0, "close": 9.0},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid arguments",
			args: map[string]any{
				"period": "invalid",
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
			result, err := CalculateATRHandler(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateATRHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("CalculateATRHandler() returned nil result")
			}
			if !tt.wantErr && result.IsError {
				t.Errorf("CalculateATRHandler() returned error result: %v", result.Content)
			}
		})
	}
}
