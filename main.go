package main

import (
	"gokub/prompts"
	"gokub/resources"
	"gokub/tools"
	"gokub/utils"
	"os"

	"github.com/dvgamerr-app/go-bitkub/bitkub"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tmilewski/goenv"
)

func init() {
	goenv.Load()

	utils.InitLogger()

	apiKey := os.Getenv("BTK_APIKEY")
	secretKey := os.Getenv("BTK_SECRET")

	if apiKey == "" || secretKey == "" {
		utils.Logger.Warn().Msg("BTK_APIKEY and BTK_SECRET not set in environment")
		utils.Logger.Info().Msg("Please set them to use Bitkub API features")
	} else {
		if err := bitkub.Initlizer(apiKey, secretKey); err != nil {
			utils.Logger.Fatal().Err(err).Msg("Failed to initialize Bitkub client")
		}
		utils.Logger.Info().Msg("Bitkub client initialized successfully")
	}
}

func main() {
	s := server.NewMCPServer(
		"Bitkub MCP Server 🚀",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
	)

	s.AddTool(tools.NewWalletBalanceTool(), tools.WalletBalanceHandler)
	s.AddTool(tools.NewTickerTool(), tools.TickerHandler)
	s.AddTool(tools.NewMarketDepthTool(), tools.MarketDepthHandler)
	s.AddTool(tools.NewOpenOrdersTool(), tools.OpenOrdersHandler)
	s.AddTool(tools.NewSymbolsTool(), tools.SymbolsHandler)

	s.AddPrompt(prompts.NewTradingStrategyPrompt(), prompts.TradingStrategyHandler)
	s.AddPrompt(prompts.NewMarketAnalysisPrompt(), prompts.MarketAnalysisHandler)

	s.AddResource(resources.NewSymbolsResource().Resource, resources.NewSymbolsResource().Handler)
	s.AddResourceTemplate(resources.NewTickerResource().Template, resources.NewTickerResource().Handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	utils.Logger.Info().Msg("Starting Bitkub MCP Server (HTTP Mode)...")
	utils.Logger.Info().Msg("Available Tools:")
	utils.Logger.Info().Msg("   1. get_wallet_balance - View wallet balances")
	utils.Logger.Info().Msg("   2. get_ticker - Get current market prices")
	utils.Logger.Info().Msg("   3. get_market_depth - View order book")
	utils.Logger.Info().Msg("   4. get_my_open_orders - View your open orders")
	utils.Logger.Info().Msg("   5. get_symbols - List all trading pairs")
	utils.Logger.Info().Msg("Available Prompts:")
	utils.Logger.Info().Msg("   1. trading_strategy - Generate trading strategy recommendations")
	utils.Logger.Info().Msg("   2. market_analysis - Analyze market conditions")
	utils.Logger.Info().Msg("Available Resources:")
	utils.Logger.Info().Msg("   1. bitkub://symbols - List of all trading pairs")
	utils.Logger.Info().Msg("   2. bitkub://ticker/{symbol} - Real-time market data")
	utils.Logger.Info().Str("port", port).Msgf("Server listening on http://localhost:%s", port)
	utils.Logger.Info().Msg("SSE endpoint: /sse")
	utils.Logger.Info().Msg("Message endpoint: /message")

	sseServer := server.NewSSEServer(s,
		server.WithSSEEndpoint("/sse"),
		server.WithMessageEndpoint("/message"),
	)

	addr := ":" + port
	if err := sseServer.Start(addr); err != nil {
		utils.Logger.Fatal().Err(err).Msg("Server error")
	}
}
