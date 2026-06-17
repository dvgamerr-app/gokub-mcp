package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// TestExtractClosePricesValues verifies extracted prices match input candles (ASSIGNMENT.md Extra E2).
func TestExtractClosePricesValues(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"candles": []any{
					map[string]any{"close": 100.0},
					map[string]any{"close": 101.0},
				},
			},
		},
	}
	result, err := ExtractClosePricesHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := result.StructuredContent.(*ExtractedPrices)
	if !ok {
		t.Fatal("StructuredContent is not *ExtractedPrices")
	}
	if out.DataPoints != 2 {
		t.Errorf("data_points: want 2, got %d", out.DataPoints)
	}
	if len(out.Prices) != 2 || out.Prices[0] != 100.0 || out.Prices[1] != 101.0 {
		t.Errorf("prices: want [100.0 101.0], got %v", out.Prices)
	}
}

func TestExtractClosePricesHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"candles": []any{
					map[string]any{"close": 100.0},
					map[string]any{"close": 101.0},
				},
			},
			wantErr: false,
		},
		{
			name: "Empty candles",
			args: map[string]any{
				"candles": []any{},
			},
			wantErr: true,
		},
		{
			name: "No close price",
			args: map[string]any{
				"candles": []any{
					map[string]any{"high": 100.0},
				},
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
			result, err := ExtractClosePricesHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("ExtractClosePricesHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("ExtractClosePricesHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("ExtractClosePricesHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
