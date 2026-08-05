package tools

import (
	"context"
	"testing"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
)

// TestCalculateCAPMFormula verifies CAPM formula (ASSIGNMENT.md Extra E3).
// E(Ri) = Rf + β·(Rm−Rf) = 0.02 + 1.2*(0.08−0.02) = 0.02 + 0.072 = 0.092
func TestCalculateCAPMFormula(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"risk_free_rate": 0.02,
				"beta":           1.2,
				"market_return":  0.08,
			},
		},
	}
	result, err := CalculateCAPMHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := result.StructuredContent.(*CAPMResult)
	if !ok {
		t.Fatal("StructuredContent is not *CAPMResult")
	}
	if out.ExpectedReturn != 0.092 {
		t.Errorf("expected_return: want 0.092, got %v (formula: E(Ri)=Rf+β·(Rm−Rf))", out.ExpectedReturn)
	}
}

func TestCalculateCAPMHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
	}{
		{
			name: "Valid input",
			args: map[string]any{
				"risk_free_rate": 0.02,
				"beta":           1.2,
				"market_return":  0.08,
			},
			wantErr: false,
		},
		{
			name: "Negative risk free rate",
			args: map[string]any{
				"risk_free_rate": -0.01,
				"beta":           1.0,
				"market_return":  0.08,
			},
			wantErr: true, // The handler returns ErrorResult which is not an error in Go return value, but result.IsError is true.
			// Wait, utils.ErrorResult returns (*mcp.CallToolResult, error).
			// In mcp-go, usually we return a result with IsError=true, and nil error, OR we return error.
			// Let's check utils.ErrorResult implementation if possible, but assuming standard pattern.
			// The handler checks: if riskFreeRate < 0 { return utils.ErrorResult(...) }
			// So I should check result.IsError.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.args,
				},
			}
			result, err := CalculateCAPMHandler(context.Background(), req)

			// If the handler returns an actual error (err != nil), that's a failure of the handler execution usually.
			// But utils.ErrorResult might return (res, nil) where res.IsError is true.

			if err != nil {
				// If we expected an error and got one, that's fine?
				// Usually handlers return nil error but set IsError=true for validation errors.
				// Let's assume we are checking for "logical success".
				// If wantErr is true, we expect either err != nil OR result.IsError == true
			}

			if tt.wantErr {
				if err == nil && (result == nil || !result.IsError) {
					t.Errorf("CalculateCAPMHandler() expected error, got success")
				}
			} else {
				if err != nil {
					t.Errorf("CalculateCAPMHandler() error = %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("CalculateCAPMHandler() returned error result: %v", result.Content)
				}
			}
		})
	}
}
