package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/bitkub"
	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/tmilewski/goenv"
)

func init() {
	// โหลด environment variables
	goenv.Load()

	// ดึง API keys จาก environment
	apiKey := os.Getenv("BTK_APIKEY")
	secretKey := os.Getenv("BTK_SECRETKEY")

	if apiKey == "" || secretKey == "" {
		fmt.Println("⚠️  Warning: BTK_APIKEY and BTK_SECRETKEY not set in environment")
		fmt.Println("Please set them to use Bitkub API features")
	} else {
		// Initialize Bitkub SDK
		if err := bitkub.Initlizer(apiKey, secretKey); err != nil {
			fmt.Printf("❌ Failed to initialize Bitkub client: %v\n", err)
		}
	}
}

func main() {
	// สร้าง MCP server
	s := server.NewMCPServer(
		"Bitkub MCP Server 🚀",
		"1.0.0",
		server.WithLogging(),
	)

	// Tool 1: Get Wallet Balance
	walletTool := mcp.NewTool("get_wallet_balance",
		mcp.WithDescription("Get wallet balance from Bitkub account - returns available and reserved balance for all currencies"),
	)
	s.AddTool(walletTool, getWalletBalanceHandler)

	// Tool 2: Get Ticker (Market Price)
	tickerTool := mcp.NewTool("get_ticker",
		mcp.WithDescription("Get current market ticker/price for a cryptocurrency symbol (e.g., btc_thb, eth_thb)"),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb, ada_thb). Use lowercase with underscore."),
		),
	)
	s.AddTool(tickerTool, getTickerHandler)

	// Tool 3: Get Market Depth (Order Book)
	depthTool := mcp.NewTool("get_market_depth",
		mcp.WithDescription("Get market depth (order book) showing bids and asks for a symbol"),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of orders to return (default: 10, max: 100)"),
		),
	)
	s.AddTool(depthTool, getMarketDepthHandler)

	// Tool 4: Get My Open Orders
	openOrdersTool := mcp.NewTool("get_my_open_orders",
		mcp.WithDescription("Get your currently open orders for a trading pair"),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb)"),
		),
	)
	s.AddTool(openOrdersTool, getMyOpenOrdersHandler)

	// Tool 5: Get Symbols (Available Trading Pairs)
	symbolsTool := mcp.NewTool("get_symbols",
		mcp.WithDescription("Get list of all available trading pairs and their info"),
	)
	s.AddTool(symbolsTool, getSymbolsHandler)

	// เริ่ม server
	fmt.Println("🚀 Starting Bitkub MCP Server...")
	fmt.Println("📋 Available Tools:")
	fmt.Println("   1. get_wallet_balance - View wallet balances")
	fmt.Println("   2. get_ticker - Get current market prices")
	fmt.Println("   3. get_market_depth - View order book")
	fmt.Println("   4. get_my_open_orders - View your open orders")
	fmt.Println("   5. get_symbols - List all trading pairs")

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
		os.Exit(1)
	}
}

// Handler สำหรับ get wallet balance
func getWalletBalanceHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// เรียก Bitkub API เพื่อดึงข้อมูล balance
	balances, err := market.GetBalances()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get wallet balance: %v", err)), nil
	}

	// แปลง balance เป็น text
	result := "📊 Wallet Balance:\n\n"
	totalTHB := 0.0

	for currency, balance := range balances {
		if balance.Available > 0 || balance.Reserved > 0 {
			result += fmt.Sprintf("💰 %s:\n", strings.ToUpper(currency))
			result += fmt.Sprintf("   Available: %.8f\n", balance.Available)
			result += fmt.Sprintf("   Reserved:  %.8f\n", balance.Reserved)
			result += fmt.Sprintf("   Total:     %.8f\n\n", balance.Available+balance.Reserved)

			// คำนวณมูลค่ารวม (สำหรับ THB)
			if strings.ToUpper(currency) == "THB" {
				totalTHB = balance.Available + balance.Reserved
			}
		}
	}

	if totalTHB > 0 {
		result += fmt.Sprintf("💵 Total THB: %.2f THB\n", totalTHB)
	}

	return mcp.NewToolResultText(result), nil
}

// Handler สำหรับ get ticker
func getTickerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get symbol from arguments
	symbolArg, ok := request.Params.Arguments["symbol"]
	if !ok {
		return mcp.NewToolResultError("symbol parameter is required"), nil
	}

	symbol, ok := symbolArg.(string)
	if !ok {
		return mcp.NewToolResultError("symbol must be a string"), nil
	}

	// Convert to lowercase for API
	symbol = strings.ToLower(symbol)

	tickers, err := market.GetTicker(symbol)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get ticker: %v", err)), nil
	}

	if len(tickers) == 0 {
		return mcp.NewToolResultError("No ticker data found for symbol: " + symbol), nil
	}

	ticker := tickers[0]
	result := fmt.Sprintf("📈 %s Market Ticker:\n\n", strings.ToUpper(symbol))
	result += fmt.Sprintf("💰 Last Price:   %s THB\n", ticker.Last)
	result += fmt.Sprintf("📊 24h Volume:   %s\n", ticker.BaseVolume)
	result += fmt.Sprintf("📈 24h High:     %s THB\n", ticker.High24hr)
	result += fmt.Sprintf("📉 24h Low:      %s THB\n", ticker.Low24hr)
	result += fmt.Sprintf("🔄 24h Change:   %s%%\n", ticker.PercentChange)
	result += fmt.Sprintf("💵 Best Bid:     %s THB\n", ticker.HighestBid)
	result += fmt.Sprintf("💸 Best Ask:     %s THB\n", ticker.LowestAsk)

	return mcp.NewToolResultText(result), nil
}

// Handler สำหรับ get market depth
func getMarketDepthHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get symbol from arguments
	symbolArg, ok := request.Params.Arguments["symbol"]
	if !ok {
		return mcp.NewToolResultError("symbol parameter is required"), nil
	}

	symbol, ok := symbolArg.(string)
	if !ok {
		return mcp.NewToolResultError("symbol must be a string"), nil
	}

	limit := 10
	if limitArg, ok := request.Params.Arguments["limit"]; ok {
		if limitVal, ok := limitArg.(float64); ok {
			limit = int(limitVal)
			if limit > 100 {
				limit = 100
			}
		}
	}

	symbol = strings.ToLower(symbol)

	depth, err := market.GetDepth(symbol, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get market depth: %v", err)), nil
	}

	result := fmt.Sprintf("📊 Market Depth for %s:\n\n", strings.ToUpper(symbol))
	result += "📉 ASKS (Sell Orders):\n"
	for i := len(depth.Asks) - 1; i >= 0 && i >= len(depth.Asks)-5; i-- {
		result += fmt.Sprintf("   %.2f THB | %.8f\n", depth.Asks[i][0], depth.Asks[i][1])
	}

	result += "\n━━━━━━━━━━━━━━━━━━━━━━\n\n"

	result += "📈 BIDS (Buy Orders):\n"
	for i := 0; i < len(depth.Bids) && i < 5; i++ {
		result += fmt.Sprintf("   %.2f THB | %.8f\n", depth.Bids[i][0], depth.Bids[i][1])
	}

	return mcp.NewToolResultText(result), nil
}

// Handler สำหรับ get my open orders
func getMyOpenOrdersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get symbol from arguments
	symbolArg, ok := request.Params.Arguments["symbol"]
	if !ok {
		return mcp.NewToolResultError("symbol parameter is required"), nil
	}

	symbol, ok := symbolArg.(string)
	if !ok {
		return mcp.NewToolResultError("symbol must be a string"), nil
	}

	symbol = strings.ToLower(symbol)

	orders, err := market.GetMyOpenOrders(symbol)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get open orders: %v", err)), nil
	}

	if len(orders) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No open orders for %s", strings.ToUpper(symbol))), nil
	}

	result := fmt.Sprintf("📋 Open Orders for %s:\n\n", strings.ToUpper(symbol))
	for i, order := range orders {
		result += fmt.Sprintf("%d. Order ID: %s\n", i+1, order.ID)
		result += fmt.Sprintf("   Side: %s\n", strings.ToUpper(order.Side))
		result += fmt.Sprintf("   Type: %s\n", order.Type)

		// Parse string values
		if rate, err := strconv.ParseFloat(order.Rate, 64); err == nil {
			result += fmt.Sprintf("   Rate: %.2f THB\n", rate)
		}
		if amount, err := strconv.ParseFloat(order.Amount, 64); err == nil {
			result += fmt.Sprintf("   Amount: %.8f\n", amount)
		}

		result += fmt.Sprintf("   Timestamp: %d\n\n", order.Timestamp)
	}

	return mcp.NewToolResultText(result), nil
}

// Handler สำหรับ get symbols
func getSymbolsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symbols, err := market.GetSymbols()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get symbols: %v", err)), nil
	}

	result := "📋 Available Trading Pairs:\n\n"
	activeCount := 0
	for _, sym := range symbols {
		if sym.Status == "active" {
			result += fmt.Sprintf("• %s\n", strings.ToUpper(sym.Symbol))
			activeCount++
		}
	}

	result += fmt.Sprintf("\nTotal: %d active trading pairs\n", activeCount)

	return mcp.NewToolResultText(result), nil
}
