package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog/log"
)

func NewSymbolsResource() server.ServerResource {
	return server.ServerResource{
		Resource: mcp.NewResource(
			"bitkub://symbols",
			"Trading Symbols",
			mcp.WithResourceDescription("List of all available trading pairs on Bitkub"),
			mcp.WithMIMEType("application/json"),
		),
		Handler: SymbolsResourceHandler,
	}
}

func SymbolsResourceHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
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

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

func NewTickerResource() server.ServerResourceTemplate {
	return server.ServerResourceTemplate{
		Template: mcp.NewResourceTemplate(
			"bitkub://ticker/{symbol}",
			"Market Ticker",
			mcp.WithTemplateDescription("Real-time price and market data for a specific trading pair"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		Handler: TickerResourceHandler,
	}
}

func TickerResourceHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
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

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}
