package main

import (
	"flag"
	"gokub/prompts"
	"gokub/resources"
	"gokub/tools"
	"gokub/utils"
	"os"

	"github.com/dvgamerr-app/go-bitkub/bitkub"
	"github.com/mark3labs/mcp-go/server"
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

func logServerInfo(s *server.MCPServer, mode string) {
	log.Info().Msgf("Starting Bitkub MCP Server (%s Mode)...", mode)

	if tools := s.ListTools(); len(tools) > 0 {
		log.Info().Msg("Available Tools:")
		i := 1
		for _, tool := range tools {
			log.Info().Msgf("   %d. %s - %s", i, tool.Tool.Name, tool.Tool.Description)
			i++
		}
	}
}

var (
	name    = "Bitkub MCP Server 🚀"
	version = "dev"
)

func main() {
	serveHTTP := flag.Bool("serv", false, "Run server in HTTP mode instead of stdio")
	flag.BoolVar(serveHTTP, "s", false, "Run server in HTTP mode instead of stdio (shorthand)")
	flag.Parse()

	s := server.NewMCPServer(
		name,
		version,
		server.WithResourceCapabilities(true, true),
	)

	s.AddTool(tools.NewWalletBalanceTool(), tools.WalletBalanceHandler)
	s.AddTool(tools.NewTickerTool(), tools.TickerHandler)
	s.AddTool(tools.NewMarketDepthTool(), tools.MarketDepthHandler)
	s.AddTool(tools.NewOpenOrdersTool(), tools.OpenOrdersHandler)
	s.AddTool(tools.NewSymbolsTool(), tools.SymbolsHandler)
	s.AddTool(tools.NewFeeScheduleTool(), tools.FeeScheduleHandler)
	s.AddTool(tools.NewCalculatePositionSizeTool(), tools.CalculatePositionSizeHandler)
	s.AddTool(tools.NewCalculateSpreadTool(), tools.CalculateSpreadHandler)
	s.AddTool(tools.NewCalculateLiquidityDepthTool(), tools.CalculateLiquidityDepthHandler)
	s.AddTool(tools.NewGetMarketScreenerTool(), tools.GetMarketScreenerHandler)
	s.AddTool(tools.NewHistoricalCandlesTool(), tools.HistoricalCandlesHandler)
	s.AddTool(tools.NewExtractClosePricesTool(), tools.ExtractClosePricesHandler)
	s.AddTool(tools.NewCalculateEMATool(), tools.CalculateEMAHandler)
	s.AddTool(tools.NewCalculateROCTool(), tools.CalculateROCHandler)
	s.AddTool(tools.NewCalculateATRTool(), tools.CalculateATRHandler)
	s.AddTool(tools.NewCalculateRSITool(), tools.CalculateRSIHandler)
	s.AddTool(tools.NewCalculateRelativeStrengthRankTool(), tools.CalculateRelativeStrengthRankHandler)
	s.AddTool(tools.NewDetectBreakoutSignalTool(), tools.DetectBreakoutSignalHandler)
	s.AddTool(tools.NewDetectPullbackSignalTool(), tools.DetectPullbackSignalHandler)
	s.AddTool(tools.NewCheckMarketRegimeTool(), tools.CheckMarketRegimeHandler)
	s.AddTool(tools.NewCalculateCAPMTool(), tools.CalculateCAPMHandler)
	s.AddTool(tools.NewSymbolRulesTool(), tools.SymbolRulesHandler)
	s.AddTool(tools.NewRoundToExchangeRulesTool(), tools.RoundToExchangeRulesHandler)
	s.AddTool(tools.NewValidateTradeSetupTool(), tools.ValidateTradeSetupHandler)
	s.AddTool(tools.NewPlaceLimitOrderTool(), tools.PlaceLimitOrderHandler)
	s.AddTool(tools.NewPlaceStopLimitOrderTool(), tools.PlaceStopLimitOrderHandler)
	s.AddTool(tools.NewClientSideStopWorkerTool(), tools.ClientSideStopWorkerHandler)
	s.AddTool(tools.NewGetOrderStatusTool(), tools.GetOrderStatusHandler)
	s.AddTool(tools.NewCancelOrderTool(), tools.CancelOrderHandler)
	s.AddTool(tools.NewCheckTradePnLTool(), tools.CheckTradePnLHandler)
	s.AddTool(tools.NewCalculateTrailingStopTool(), tools.CalculateTrailingStopHandler)
	s.AddTool(tools.NewCheckExitSignalsTool(), tools.CheckExitSignalsHandler)
	s.AddTool(tools.NewLogTradeEntryTool(), tools.LogTradeEntryHandler)
	s.AddTool(tools.NewLogTradeExitTool(), tools.LogTradeExitHandler)
	s.AddTool(tools.NewCalculateExpectancyTool(), tools.CalculateExpectancyHandler)
	s.AddTool(tools.NewGetTradeHistoryTool(), tools.GetTradeHistoryHandler)
	s.AddTool(tools.NewGetMarketOverviewTool(), tools.GetMarketOverviewHandler)
	s.AddTool(tools.NewSimulateTradeTool(), tools.SimulateTradeHandler)
	s.AddTool(tools.NewPnLWithFeesTool(), tools.PnLWithFeesHandler)

	s.AddPrompt(prompts.NewTradingStrategyPrompt(), prompts.TradingStrategyHandler)
	s.AddPrompt(prompts.NewMarketAnalysisPrompt(), prompts.MarketAnalysisHandler)

	s.AddResource(resources.NewSymbolsResource().Resource, resources.NewSymbolsResource().Handler)
	s.AddResourceTemplate(resources.NewTickerResource().Template, resources.NewTickerResource().Handler)

	if *serveHTTP {
		logServerInfo(s, "HTTP")

		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}

		log.Info().Str("port", port).Msgf("Server listening on http://localhost:%s", port)
		log.Info().Msg("Endpoint SSE: /sse, Message: /msg")

		sseServer := server.NewSSEServer(s,
			server.WithSSEEndpoint("/sse"),
			server.WithMessageEndpoint("/msg"),
		)

		addr := ":" + port
		if err := sseServer.Start(addr); err != nil {
			log.Fatal().Err(err).Msg("Server error")
		}
	} else {
		logServerInfo(s, "stdio")

		if err := server.ServeStdio(s); err != nil {
			log.Fatal().Err(err).Msg("Server error")
		}
	}
}
