package prompts

import (
	"context"
	"fmt"
	"gokub/utils"
	"strings"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
)

func NewTradingStrategyPrompt() mcp.Prompt {
	return mcp.NewPrompt("trading_strategy")
}

func TradingStrategyHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	args, _ := utils.ValidateArgs(request.Params.Arguments)

	symbol := "btc_thb"
	if val, ok := args["symbol"].(string); ok {
		symbol = strings.ToLower(val)
	}

	riskTolerance := "medium"
	if val, ok := args["risk_tolerance"].(string); ok {
		riskTolerance = val
	}

	timeframe := "day"
	if val, ok := args["timeframe"].(string); ok {
		timeframe = val
	}

	utils.Logger.Debug().
		Str("symbol", symbol).
		Str("risk", riskTolerance).
		Str("timeframe", timeframe).
		Msg("Generating trading strategy prompt")

	tickers, err := market.GetTicker(symbol)
	if err != nil {
		return nil, err
	}

	var tickerData string
	if len(tickers) > 0 {
		ticker := tickers[0]
		tickerData = fmt.Sprintf(`
Current Market Data for %s:
- Last Price: %.2f THB
- 24h Change: %.2f%%
- 24h High: %.2f THB
- 24h Low: %.2f THB
- 24h Volume: %.2f
`, strings.ToUpper(symbol), ticker.Last, ticker.PercentChange, ticker.High24hr, ticker.Low24hr, ticker.BaseVolume)
	}

	promptText := fmt.Sprintf(`You are a professional cryptocurrency trading advisor. Generate a comprehensive trading strategy for the following scenario:

**Trading Pair:** %s
**Risk Tolerance:** %s
**Trading Timeframe:** %s

%s

Please provide:
1. **Market Analysis:** Analyze the current price action and trend
2. **Entry Strategy:** Specific entry points and conditions
3. **Exit Strategy:** Take profit levels and stop loss placement
4. **Position Sizing:** Recommended position size based on risk tolerance
5. **Risk Management:** Key risk factors to monitor
6. **Trading Rules:** Clear do's and don'ts for this setup

Format your response in a clear, actionable manner suitable for execution.`,
		strings.ToUpper(symbol), riskTolerance, timeframe, tickerData)

	utils.Logger.Info().
		Str("symbol", symbol).
		Msg("Generated trading strategy prompt")

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{
			{
				Role: "user",
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

func NewMarketAnalysisPrompt() mcp.Prompt {
	return mcp.NewPrompt("market_analysis")
}

func MarketAnalysisHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	args, _ := utils.ValidateArgs(request.Params.Arguments)

	symbolsStr := "btc_thb,eth_thb"
	if val, ok := args["symbols"].(string); ok {
		symbolsStr = val
	}

	analysisType := "comprehensive"
	if val, ok := args["analysis_type"].(string); ok {
		analysisType = val
	}

	symbols := strings.Split(symbolsStr, ",")
	utils.Logger.Debug().
		Strs("symbols", symbols).
		Str("analysis_type", analysisType).
		Msg("Generating market analysis prompt")

	var marketData strings.Builder
	marketData.WriteString("Current Market Overview:\n\n")

	for _, symbol := range symbols {
		symbol = strings.TrimSpace(strings.ToLower(symbol))
		tickers, err := market.GetTicker(symbol)
		if err != nil || len(tickers) == 0 {
			continue
		}

		ticker := tickers[0]
		marketData.WriteString(fmt.Sprintf(`**%s:**
- Price: %.2f THB
- 24h Change: %.2f%%
- 24h Volume: %.2f
- High/Low: %.2f / %.2f THB

`, strings.ToUpper(symbol), ticker.Last, ticker.PercentChange, ticker.BaseVolume, ticker.High24hr, ticker.Low24hr))
	}

	promptText := fmt.Sprintf(`You are a cryptocurrency market analyst. Perform a %s analysis of the following markets:

%s

Please provide:
1. **Overall Market Sentiment:** Bullish, Bearish, or Neutral with reasoning
2. **Individual Asset Analysis:** Key observations for each symbol
3. **Correlation Analysis:** How these assets are moving relative to each other
4. **Trading Opportunities:** Potential setups based on current conditions
5. **Risk Factors:** Key risks to watch in current market environment
6. **Timeframe Considerations:** Best timeframes for different trading styles

Provide your analysis in a structured, professional format.`,
		analysisType, marketData.String())

	utils.Logger.Info().
		Strs("symbols", symbols).
		Msg("Generated market analysis prompt")

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{
			{
				Role: "user",
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}
