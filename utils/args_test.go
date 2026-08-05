package utils

import (
	"slices"
	"testing"
)

func TestGetFloat64ArrayArg(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		want    []float64
		wantErr string
	}{
		{
			name: "numbers",
			args: map[string]any{"prices": []any{1.5, 2}},
			want: []float64{1.5, 2},
		},
		{
			name:    "missing",
			args:    map[string]any{},
			wantErr: "prices must be an array",
		},
		{
			name:    "wrong container",
			args:    map[string]any{"prices": "1,2"},
			wantErr: "prices must be an array",
		},
		{
			name:    "non-number",
			args:    map[string]any{"prices": []any{1.5, "2"}},
			wantErr: "prices must contain numbers only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFloat64ArrayArg(tt.args, "prices")
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("GetFloat64ArrayArg() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetFloat64ArrayArg() unexpected error: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("GetFloat64ArrayArg() = %v, want %v", got, tt.want)
			}
		})
	}
}
