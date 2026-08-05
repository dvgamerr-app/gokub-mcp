package main

import (
	"context"
	"flag"
	"net/http"
	"os"

	"github.com/dvgamerr-app/go-bitkub/bitkub"
	mcpcompat "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/dvgamerr-app/gokub-mcp/prompts"
	"github.com/dvgamerr-app/gokub-mcp/resources"
	"github.com/dvgamerr-app/gokub-mcp/tools"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog/log"
	"github.com/tmilewski/goenv"
)

func init() {
	goenv.Load()

	utils.InitLogger()

	apiKey := os.Getenv("BTK_APIKEY")
	secretKey := os.Getenv("BTK_SECRET")

	if apiKey == "" || secretKey == "" {
		log.Warn().Msg("BTK_APIKEY and BTK_SECRET not set in environment")
		log.Info().Msg("Please set them to use Bitkub API features")
	} else {
		if err := bitkub.Initlizer(apiKey, secretKey); err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize Bitkub client")
		}
		log.Info().Msg("Bitkub client initialized successfully")
	}
}

type toolRegistration struct {
	tool    mcpcompat.Tool
	handler mcpcompat.ToolHandler
}

func toolRegistrations() []toolRegistration {
	return []toolRegistration{
		{tools.NewWalletBalanceTool(), tools.WalletBalanceHandler},
		{tools.NewTickerTool(), tools.TickerHandler},
		{tools.NewMarketDepthTool(), tools.MarketDepthHandler},
		{tools.NewOpenOrdersTool(), tools.OpenOrdersHandler},
		{tools.NewSymbolsTool(), tools.SymbolsHandler},
		{tools.NewFeeScheduleTool(), tools.FeeScheduleHandler},
		{tools.NewCalculatePositionSizeTool(), tools.CalculatePositionSizeHandler},
		{tools.NewCalculateSpreadTool(), tools.CalculateSpreadHandler},
		{tools.NewCalculateLiquidityDepthTool(), tools.CalculateLiquidityDepthHandler},
		{tools.NewGetMarketScreenerTool(), tools.GetMarketScreenerHandler},
		{tools.NewHistoricalCandlesTool(), tools.HistoricalCandlesHandler},
		{tools.NewExtractClosePricesTool(), tools.ExtractClosePricesHandler},
		{tools.NewCalculateEMATool(), tools.CalculateEMAHandler},
		{tools.NewCalculateROCTool(), tools.CalculateROCHandler},
		{tools.NewCalculateATRTool(), tools.CalculateATRHandler},
		{tools.NewCalculateRSITool(), tools.CalculateRSIHandler},
		{tools.NewCalculateRelativeStrengthRankTool(), tools.CalculateRelativeStrengthRankHandler},
		{tools.NewDetectBreakoutSignalTool(), tools.DetectBreakoutSignalHandler},
		{tools.NewDetectPullbackSignalTool(), tools.DetectPullbackSignalHandler},
		{tools.NewCheckMarketRegimeTool(), tools.CheckMarketRegimeHandler},
		{tools.NewCalculateCAPMTool(), tools.CalculateCAPMHandler},
		{tools.NewSymbolRulesTool(), tools.SymbolRulesHandler},
		{tools.NewRoundToExchangeRulesTool(), tools.RoundToExchangeRulesHandler},
		{tools.NewValidateTradeSetupTool(), tools.ValidateTradeSetupHandler},
		{tools.NewPlaceLimitOrderTool(), tools.PlaceLimitOrderHandler},
		{tools.NewPlaceStopLimitOrderTool(), tools.PlaceStopLimitOrderHandler},
		{tools.NewClientSideStopWorkerTool(), tools.ClientSideStopWorkerHandler},
		{tools.NewGetOrderStatusTool(), tools.GetOrderStatusHandler},
		{tools.NewCancelOrderTool(), tools.CancelOrderHandler},
		{tools.NewCheckTradePnLTool(), tools.CheckTradePnLHandler},
		{tools.NewCalculateTrailingStopTool(), tools.CalculateTrailingStopHandler},
		{tools.NewCheckExitSignalsTool(), tools.CheckExitSignalsHandler},
		{tools.NewLogTradeEntryTool(), tools.LogTradeEntryHandler},
		{tools.NewLogTradeExitTool(), tools.LogTradeExitHandler},
		{tools.NewCalculateExpectancyTool(), tools.CalculateExpectancyHandler},
		{tools.NewGetTradeHistoryTool(), tools.GetTradeHistoryHandler},
		{tools.NewGetMarketOverviewTool(), tools.GetMarketOverviewHandler},
		{tools.NewSimulateTradeTool(), tools.SimulateTradeHandler},
		{tools.NewPnLWithFeesTool(), tools.PnLWithFeesHandler},
	}
}

func logServerInfo(registrations []toolRegistration, mode string) {
	log.Info().Msgf("Starting Bitkub MCP Server (%s Mode)...", mode)
	if len(registrations) == 0 {
		return
	}

	log.Info().Msg("Available Tools:")
	for index, registration := range registrations {
		log.Info().Msgf("   %d. %s - %s", index+1, registration.tool.Name, registration.tool.Description)
	}
}

var (
	name    = "Bitkub MCP Server 🚀"
	version = "dev"
)

func main() {
	serveHTTP := flag.Bool("serv", false, "Run server in Streamable HTTP mode instead of stdio")
	flag.BoolVar(serveHTTP, "s", false, "Run server in Streamable HTTP mode instead of stdio (shorthand)")
	flag.Parse()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Version: version,
	}, nil)

	registrations := toolRegistrations()
	for _, registration := range registrations {
		mcpcompat.AddTool(server, registration.tool, registration.handler)
	}

	server.AddPrompt(prompts.NewTradingStrategyPrompt(), prompts.TradingStrategyHandler)
	server.AddPrompt(prompts.NewMarketAnalysisPrompt(), prompts.MarketAnalysisHandler)
	server.AddResource(resources.NewSymbolsResource(), resources.SymbolsResourceHandler)
	server.AddResourceTemplate(resources.NewTickerResource(), resources.TickerResourceHandler)

	if *serveHTTP {
		logServerInfo(registrations, "Streamable HTTP")

		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}

		log.Info().Str("port", port).Msgf("Server listening on http://localhost:%s/mcp", port)
		log.Info().Msg("MCP endpoint: /mcp (stateless, protocol 2026-07-28)")

		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, &mcp.StreamableHTTPOptions{
			Stateless:                    true,
			JSONResponse:                 true,
			PropagateRequestCancellation: true,
		})
		mux := http.NewServeMux()
		mux.Handle("/mcp", handler)

		addr := ":" + port
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal().Err(err).Msg("Server error")
		}
	} else {
		logServerInfo(registrations, "stdio")

		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatal().Err(err).Msg("Server error")
		}
	}
}
