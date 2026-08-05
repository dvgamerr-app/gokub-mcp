package tools

import (
	"context"
	"testing"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
)

func TestCalculateROCHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"period": float64(2),
				"prices": []any{100.0, 105.0, 110.0},
			},
			wantErr: false,
		},
		{
			name: "Non-numeric price",
			args: map[string]any{
				"period": float64(1),
				"prices": []any{100.0, "bad"},
			},
			wantErr: true,
		},
		{
			name: "Not enough data",
			args: map[string]any{
				"period": float64(5),
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
			result, err := CalculateROCHandler(context.Background(), req)

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CalculateROCHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CalculateROCHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CalculateROCHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}

// TestCalculateROCFormula verifies: ROC = ((Close_t - Close_t-n) / Close_t-n) × 100
// ASSIGNMENT.md example: entry 50 → target 55 represents +10% move over the period.
func TestCalculateROCFormula(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"period": float64(1),
				"prices": []any{50.0, 55.0},
			},
		},
	}
	result, err := CalculateROCHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := result.StructuredContent.(*ROCResult)
	if !ok {
		t.Fatal("StructuredContent is not *ROCResult")
	}
	if out.ROC != 10.0 {
		t.Errorf("ROC: want 10.0, got %v (formula: (55-50)/50×100)", out.ROC)
	}
}
