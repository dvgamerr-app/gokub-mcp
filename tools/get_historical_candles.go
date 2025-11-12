package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strings"
	"time"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

type Candle struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

var validResolutions map[int]string = map[int]string{
	1:    "1",
	5:    "5",
	15:   "15",
	60:   "60",
	240:  "240",
	1440: "1D",
}

func NewHistoricalCandlesTool() mcp.Tool {
	return mcp.NewTool("get_historical_candles",
		mcp.WithDescription(`Get historical candlestick/OHLCV data for a symbol with specified timeframe and limit`),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("Trading pair symbol (e.g., btc_thb, eth_thb). Use lowercase with underscore."),
		),
		mcp.WithNumber("resolution",
			mcp.Description("Timeframe resolution in minutes (1, 5, 15, 60, 240, 1440). Default: 60"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of candles to retrieve (1-1000). Default: 100"),
		),
	)
}

func HistoricalCandlesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for historical candles")
		return utils.ErrorResult("invalid arguments")
	}

	symbol := strings.ToUpper(utils.GetStringArg(args, "symbol"))
	resolution := utils.GetIntArg(args, "resolution", 60)
	limit := utils.GetIntArg(args, "limit", 100)

	if limit < 1 || limit > 1000 {
		return utils.ErrorResult("limit must be between 1 and 1000")
	}

	resolutionStr, ok := validResolutions[resolution]
	if !ok {
		return utils.ErrorResult("invalid resolution. Use: 1, 5, 15, 60, 240, or 1440")
	}

	now := time.Now().Unix()
	from := now - int64(limit*resolution*60)

	candles, err := market.GetHistory(market.HistoryRequest{
		Symbol:     symbol,
		Resolution: resolutionStr,
		From:       from,
		To:         now,
	})

	if err != nil {
		log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get historical candles")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	if len(candles.Close) == 0 {
		log.Warn().Str("symbol", symbol).Msg("No candle data found")
		return utils.ErrorResult("no data: " + symbol)
	}

	dataLen := min(limit, len(candles.Close))

	result := make([]*Candle, dataLen)
	for i := range dataLen {
		result[i] = &Candle{
			Timestamp: candles.Time[i],
			Open:      candles.Open[i],
			High:      candles.High[i],
			Low:       candles.Low[i],
			Close:     candles.Close[i],
			Volume:    candles.Volume[i],
		}
	}

	summary := fmt.Sprintf("Retrieved %d candles for %s (%dm timeframe)\n\n", dataLen, symbol, resolution)
	summary += "Timestamp,Open,High,Low,Close,Volume\n"
	for _, candle := range result {
		summary += fmt.Sprintf("%d,%.2f,%.2f,%.2f,%.2f,%.2f\n",
			candle.Timestamp,
			candle.Open,
			candle.High,
			candle.Low,
			candle.Close,
			candle.Volume,
		)
	}

	return utils.ArtifactsResult(summary, map[string]any{"candles": result})
}
