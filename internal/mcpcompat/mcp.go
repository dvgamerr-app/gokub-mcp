package mcpcompat

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Tool = sdkmcp.Tool
type CallToolResult = sdkmcp.CallToolResult
type Content = sdkmcp.Content
type TextContent = sdkmcp.TextContent

type CallToolParams struct {
	Name      string
	Arguments any
}

type CallToolRequest struct {
	Params CallToolParams
}

type ToolHandler func(context.Context, CallToolRequest) (*CallToolResult, error)

type ToolOption func(*Tool)

type propertyOption func(*property)

type property struct {
	schema   map[string]any
	required bool
}

func NewTool(name string, options ...ToolOption) Tool {
	tool := Tool{
		Name: name,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}
	for _, option := range options {
		option(&tool)
	}
	return tool
}

func WithDescription(description string) ToolOption {
	return func(tool *Tool) {
		tool.Description = description
	}
}

func WithString(name string, options ...propertyOption) ToolOption {
	return withProperty(name, "string", options...)
}

func WithNumber(name string, options ...propertyOption) ToolOption {
	return withProperty(name, "number", options...)
}

func WithBoolean(name string, options ...propertyOption) ToolOption {
	return withProperty(name, "boolean", options...)
}

func WithArray(name string, options ...propertyOption) ToolOption {
	return withProperty(name, "array", options...)
}

func WithObject(name string, options ...propertyOption) ToolOption {
	return withProperty(name, "object", options...)
}

func withProperty(name, kind string, options ...propertyOption) ToolOption {
	return func(tool *Tool) {
		value := property{schema: map[string]any{"type": kind}}
		for _, option := range options {
			option(&value)
		}

		schema := tool.InputSchema.(map[string]any)
		properties := schema["properties"].(map[string]any)
		properties[name] = value.schema
		if value.required {
			required, _ := schema["required"].([]string)
			schema["required"] = append(required, name)
		}
	}
}

func Required() propertyOption {
	return func(value *property) {
		value.required = true
	}
}

func Description(description string) propertyOption {
	return func(value *property) {
		value.schema["description"] = description
	}
}

func DefaultNumber(number float64) propertyOption {
	return func(value *property) {
		value.schema["default"] = number
	}
}

func NewToolResultText(text string) *CallToolResult {
	return &CallToolResult{
		Content: []Content{&TextContent{Text: text}},
	}
}

func AddTool(server *sdkmcp.Server, tool Tool, handler ToolHandler) {
	sdkmcp.AddTool(server, &tool, func(ctx context.Context, request *sdkmcp.CallToolRequest, arguments map[string]any) (*sdkmcp.CallToolResult, any, error) {
		result, err := handler(ctx, CallToolRequest{
			Params: CallToolParams{
				Name:      request.Params.Name,
				Arguments: arguments,
			},
		})
		return result, nil, err
	})
}
