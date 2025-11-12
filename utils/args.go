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

func MustJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
