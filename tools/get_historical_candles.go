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
		mcp.WithDescription(`Get historical candlestick/OHLCV data for symbols with specified timeframe and limit. Supports single or multiple symbols.`),
		mcp.WithArray("symbols",
			mcp.Required(),
			mcp.Description("Array of trading pair symbols (e.g., ['btc_thb', 'eth_thb']) or single symbol ['btc_thb']"),
		),
		mcp.WithNumber("resolution",
			mcp.Description("Timeframe resolution in minutes (1, 5, 15, 60, 240, 1440). Default: 60"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of candles to retrieve (1-1000). Default: 100"),
		),
		mcp.WithString("format",
			mcp.Description("Output format: 'ohlcv' (default), 'close' (close prices only), 'csv' (CSV format)"),
		),
	)
}

func HistoricalCandlesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for historical candles")
		return utils.ErrorResult("invalid arguments")
	}

	symbolsRaw, ok := args["symbols"].([]any)
	if !ok {
		return utils.ErrorResult("symbols must be an array")
	}

	symbols := make([]string, len(symbolsRaw))
	for i, s := range symbolsRaw {
		if str, ok := s.(string); ok {
			symbols[i] = strings.ToUpper(str)
		} else {
			return utils.ErrorResult("symbols array must contain strings only")
		}
	}

	resolution := utils.GetIntArg(args, "resolution", 60)
	limit := utils.GetIntArg(args, "limit", 100)
	format := strings.ToLower(utils.GetStringArg(args, "format"))
	if format == "" {
		format = "ohlcv"
	}

	if limit < 1 || limit > 1000 {
		return utils.ErrorResult("limit must be between 1 and 1000")
	}

	if format != "ohlcv" && format != "close" && format != "csv" {
		return utils.ErrorResult("format must be 'ohlcv', 'close', or 'csv'")
	}

	resolutionStr, ok := validResolutions[resolution]
	if !ok {
		return utils.ErrorResult("invalid resolution. Use: 1, 5, 15, 60, 240, or 1440")
	}

	now := time.Now().Unix()
	from := now - int64(limit*resolution*60)

	if len(symbols) == 1 && format != "close" {
		symbol := symbols[0]
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

		var summary string
		if format == "csv" {
			summary = fmt.Sprintf("Retrieved %d candles for %s (%s timeframe)\n\n", dataLen, symbol, resolutionStr)
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
		} else {
			summary = fmt.Sprintf("Retrieved %d candles for %s (%s timeframe)", dataLen, symbol, resolutionStr)
		}

		return utils.ArtifactsResult(summary, map[string]any{"candles": result})
	}

	pricesMap := make(map[string][]float64)
	candlesMap := make(map[string][]*Candle)

	for _, symbol := range symbols {
		candles, err := market.GetHistory(market.HistoryRequest{
			Symbol:     symbol,
			Resolution: resolutionStr,
			From:       from,
			To:         now,
		})

		if err != nil {
			log.Warn().Err(err).Str("symbol", symbol).Msg("Failed to get historical candles")
			continue
		}

		if len(candles.Close) == 0 {
			log.Warn().Str("symbol", symbol).Msg("No candle data found")
			continue
		}

		dataLen := min(limit, len(candles.Close))

		if format == "close" {
			prices := make([]float64, dataLen)
			for i := range dataLen {
				prices[i] = candles.Close[i]
			}
			pricesMap[strings.ToLower(symbol)] = prices
		} else {
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
			candlesMap[strings.ToLower(symbol)] = result
		}
	}

	var summary string
	var resultData any

	if format == "close" {
		if len(pricesMap) == 0 {
			return utils.ErrorResult("no data retrieved for any symbol")
		}
		summary = fmt.Sprintf("Retrieved close prices for %d symbols (Timeframe: %s, Limit: %d)", len(pricesMap), resolutionStr, limit)
		resultData = map[string]any{
			"timeframe":   resolutionStr,
			"data_points": limit,
			"symbols":     pricesMap,
		}
	} else {
		if len(candlesMap) == 0 {
			return utils.ErrorResult("no data retrieved for any symbol")
		}

		if format == "csv" {
			summary = fmt.Sprintf("Retrieved %d symbols (%s timeframe)\n\n", len(candlesMap), resolutionStr)
			for symbol, candles := range candlesMap {
				summary += fmt.Sprintf("%s - Timestamp,Open,High,Low,Close,Volume\n", strings.ToUpper(symbol))
				for _, candle := range candles {
					summary += fmt.Sprintf("%d,%.2f,%.2f,%.2f,%.2f,%.2f\n",
						candle.Timestamp,
						candle.Open,
						candle.High,
						candle.Low,
						candle.Close,
						candle.Volume,
					)
				}
				summary += "\n"
			}
		} else {
			summary = fmt.Sprintf("Retrieved candles for %d symbols (%s timeframe)", len(candlesMap), resolutionStr)
		}

		resultData = map[string]any{
			"timeframe": resolutionStr,
			"symbols":   candlesMap,
		}
	}

	return utils.ArtifactsResult(summary, resultData)
}
