package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
)

func NewSymbolsResource() *mcp.Resource {
	return &mcp.Resource{
		URI:         "bitkub://symbols",
		Name:        "Trading Symbols",
		Description: "List of all available trading pairs on Bitkub",
		MIMEType:    "application/json",
	}
}

func SymbolsResourceHandler(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	log.Debug().Str("uri", request.Params.URI).Msg("read_resource")

	result, err := market.GetSymbols()
	if err != nil {
		log.Error().Err(err).Msg("GetSymbols failed")
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Error().Err(err).Msg("json marshal failed")
		return nil, fmt.Errorf("failed to marshal symbols: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}

func NewTickerResource() *mcp.ResourceTemplate {
	return &mcp.ResourceTemplate{
		URITemplate: "bitkub://ticker/{symbol}",
		Name:        "Market Ticker",
		Description: "Real-time price and market data for a specific trading pair",
		MIMEType:    "application/json",
	}
}

func TickerResourceHandler(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	log.Debug().Str("uri", request.Params.URI).Msg("read_resource")

	var symbol string
	_, err := fmt.Sscanf(request.Params.URI, "bitkub://ticker/%s", &symbol)
	if err != nil {
		log.Error().Err(err).Str("uri", request.Params.URI).Msg("invalid URI format")
		return nil, fmt.Errorf("invalid URI format: %w", err)
	}

	result, err := market.GetTicker(symbol)
	if err != nil {
		log.Error().Err(err).Str("symbol", symbol).Msg("GetTicker failed")
		return nil, fmt.Errorf("failed to get ticker for %s: %w", symbol, err)
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Error().Err(err).Msg("json marshal failed")
		return nil, fmt.Errorf("failed to marshal ticker: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      request.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		},
	}, nil
}
