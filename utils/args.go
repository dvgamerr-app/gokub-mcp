package utils

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/mark3labs/mcp-go/mcp"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func GetStringArg(args map[string]any, key string, defaultValue ...string) string {
	if val, ok := args[key]; ok {
		if fval, ok := val.(string); ok {
			return fval
		}
	}

	if len(defaultValue) == 0 {
		return ""
	}

	return defaultValue[0]
}
func GetFloat64Arg(args map[string]any, key string, defaultValue ...float64) float64 {
	if val, ok := args[key]; ok {
		if fval, ok := val.(float64); ok {
			return fval
		}
	}

	if len(defaultValue) == 0 {
		return 0
	}

	return defaultValue[0]
}

func GetIntArg(args map[string]any, key string, defaultValue ...int) int {
	if val, ok := args[key]; ok {
		if fval, ok := val.(float64); ok {
			return int(fval)
		}
	}
	if len(defaultValue) == 0 {
		return 0
	}

	return defaultValue[0]
}

func GetBoolArg(args map[string]any, key string, defaultValue ...bool) bool {
	if val, ok := args[key]; ok {
		if bval, ok := val.(bool); ok {
			return bval
		}
	}
	if len(defaultValue) == 0 {
		return false
	}

	return defaultValue[0]
}

func ValidateArgs(args any) (map[string]any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments format")
	}
	return argsMap, nil
}

func ErrorResult(err string) (*mcp.CallToolResult, error) {
	return nil, fmt.Errorf("tool: %v", err)
}

func TextResult(message string) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(message), nil
}

func ArtifactsResult(contents string, args any) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: contents},
		},
		StructuredContent: args,
	}, nil
}

func ParseFloatArray(raw []any) ([]float64, error) {
	out := make([]float64, len(raw))
	for i, v := range raw {
		switch n := v.(type) {
		case float64:
			out[i] = n
		case int:
			out[i] = float64(n)
		default:
			return nil, fmt.Errorf("element %d is not a number", i)
		}
	}
	return out, nil
}

func GetFloat64ArrayArg(args map[string]any, key string) ([]float64, error) {
	raw, ok := args[key].([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array", key)
	}

	values, err := ParseFloatArray(raw)
	if err != nil {
		return nil, fmt.Errorf("%s must contain numbers only", key)
	}
	return values, nil
}

func MustJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
