package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/dvgamerr-app/go-bitkub/market"
	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/rs/zerolog/log"
)

type Candle struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
	Synthetic bool    `json:"synthetic,omitempty"`
}

type CandleHistoryMetadata struct {
	LastActualCandle   int64  `json:"last_actual_closed_candle_timestamp"`
	LastReturnedCandle int64  `json:"last_returned_candle_timestamp"`
	SyntheticCount     int    `json:"synthetic_candles"`
	GapPolicy          string `json:"gap_policy"`
}

func prepareClosedHistory(data *market.HistoryResponse, limit, resolution int, now int64) (*market.HistoryResponse, map[int64]bool, CandleHistoryMetadata, error) {
	real := map[int64]bool{}
	metadata := CandleHistoryMetadata{GapPolicy: "previous close, zero volume; synthetic candles are not trades"}
	if data == nil {
		return nil, real, metadata, fmt.Errorf("empty history response")
	}
	for _, ts := range data.Time {
		real[ts] = true
		if ts+int64(resolution)*60 <= now && ts > metadata.LastActualCandle {
			metadata.LastActualCandle = ts
		}
	}
	filled, err := fillHistoryGaps(data, limit, resolution, now)
	if err != nil {
		return nil, real, metadata, err
	}
	closed, err := closedHistory(filled, limit, resolution, now)
	if err != nil {
		return nil, real, metadata, err
	}
	for _, ts := range closed.Time {
		if !real[ts] {
			metadata.SyntheticCount++
		}
		metadata.LastReturnedCandle = ts
	}
	return closed, real, metadata, nil
}

// minIndicatorCandles covers the largest downstream requirement (calculate_ema
// period=200 for the EMA200 regime check); ATR/RSI/ROC/breakout/pullback all
// need far fewer, so this floor keeps every chained tool call fed.
const minIndicatorCandles = 220

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
		mcp.WithBoolean("closed_only", mcp.Description("Return completed candles on a regular grid, filling no-trade gaps with previous close and zero volume, marked synthetic. Includes source timestamps; unavailable history still errors. Default false.")),
		mcp.WithArray("symbols",
			mcp.Items("string"),
			mcp.Required(),
			mcp.Description("Array of trading pair symbols (e.g., ['btc_thb', 'eth_thb']) or single symbol ['btc_thb']"),
		),
		mcp.WithNumber("resolution",
			mcp.Description("Timeframe resolution in minutes (1, 5, 15, 60, 240, 1440). Default: 60"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of candles to retrieve (1-1000). Default: 220. Bumped up to 220 automatically so ATR/RSI/EMA200/breakout/pullback have enough candles."),
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
	limit := utils.GetIntArg(args, "limit", minIndicatorCandles)
	format := strings.ToLower(utils.GetStringArg(args, "format"))
	if format == "" {
		format = "ohlcv"
	}

	if limit < 1 || limit > 1000 {
		return utils.ErrorResult("limit must be between 1 and 1000")
	}
	if limit < minIndicatorCandles {
		limit = minIndicatorCandles
	}

	if format != "ohlcv" && format != "close" && format != "csv" {
		return utils.ErrorResult("format must be 'ohlcv', 'close', or 'csv'")
	}

	resolutionStr, ok := validResolutions[resolution]
	if !ok {
		return utils.ErrorResult("invalid resolution. Use: 1, 5, 15, 60, 240, or 1440")
	}

	now := time.Now().Unix()
	closedOnly := utils.GetBoolArg(args, "closed_only")
	from := now - int64(limit*resolution*60)

	if len(symbols) == 1 {
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

		var actual map[int64]bool
		var metadata CandleHistoryMetadata
		if closedOnly {
			candles, actual, metadata, err = prepareClosedHistory(candles, limit, resolution, now)
		} else {
			candles, err = latestHistory(candles, limit)
		}
		if err != nil {
			return utils.ErrorResult(err.Error())
		}

		if len(candles.Close) == 0 {
			log.Warn().Str("symbol", symbol).Msg("No candle data found")
			return utils.ErrorResult("no data: " + symbol)
		}

		dataLen := min(limit, len(candles.Close))

		if format == "close" {
			prices := make([]float64, dataLen)
			for i := range dataLen {
				prices[i] = candles.Close[i]
			}
			summary := fmt.Sprintf("Retrieved %d close prices for %s (%s timeframe)", dataLen, symbol, resolutionStr)
			return utils.ArtifactsResult(summary, map[string]any{"symbol": strings.ToLower(symbol), "prices": prices, "data_points": dataLen, "history_metadata": metadata})
		}

		result := make([]*Candle, dataLen)
		for i := range dataLen {
			result[i] = &Candle{
				Timestamp: candles.Time[i],
				Open:      candles.Open[i],
				High:      candles.High[i],
				Low:       candles.Low[i],
				Close:     candles.Close[i],
				Volume:    candles.Volume[i],
				Synthetic: closedOnly && !actual[candles.Time[i]],
			}
		}

		var summary string
		if format == "csv" {
			summary = fmt.Sprintf("Retrieved %d candles for %s (%s timeframe)\n\n", dataLen, symbol, resolutionStr)
			summary += "Timestamp,Open,High,Low,Close,Volume,Synthetic\n"
			for _, candle := range result {
				summary += fmt.Sprintf("%d,%.2f,%.2f,%.2f,%.2f,%.2f,%t\n",
					candle.Timestamp,
					candle.Open,
					candle.High,
					candle.Low,
					candle.Close,
					candle.Volume,
					candle.Synthetic,
				)
			}
		} else {
			summary = fmt.Sprintf("Retrieved %d candles for %s (%s timeframe)", dataLen, symbol, resolutionStr)
		}

		return utils.ArtifactsResult(summary, map[string]any{"candles": result, "history_metadata": metadata})
	}

	pricesMap := make(map[string][]float64)
	candlesMap := make(map[string][]*Candle)
	metadataMap := make(map[string]CandleHistoryMetadata)

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

		var actual map[int64]bool
		var metadata CandleHistoryMetadata
		if closedOnly {
			candles, actual, metadata, err = prepareClosedHistory(candles, limit, resolution, now)
		} else {
			candles, err = latestHistory(candles, limit)
		}
		if err != nil {
			return utils.ErrorResult(err.Error())
		}
		metadataMap[strings.ToLower(symbol)] = metadata

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
					Synthetic: closedOnly && !actual[candles.Time[i]],
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
			"timeframe":        resolutionStr,
			"data_points":      limit,
			"symbols":          pricesMap,
			"history_metadata": metadataMap,
		}
	} else {
		if len(candlesMap) == 0 {
			return utils.ErrorResult("no data retrieved for any symbol")
		}

		if format == "csv" {
			summary = fmt.Sprintf("Retrieved %d symbols (%s timeframe)\n\n", len(candlesMap), resolutionStr)
			for symbol, candles := range candlesMap {
				summary += fmt.Sprintf("%s - Timestamp,Open,High,Low,Close,Volume,Synthetic\n", strings.ToUpper(symbol))
				for _, candle := range candles {
					summary += fmt.Sprintf("%d,%.2f,%.2f,%.2f,%.2f,%.2f,%t\n",
						candle.Timestamp,
						candle.Open,
						candle.High,
						candle.Low,
						candle.Close,
						candle.Volume,
						candle.Synthetic,
					)
				}
				summary += "\n"
			}
		} else {
			summary = fmt.Sprintf("Retrieved candles for %d symbols (%s timeframe)", len(candlesMap), resolutionStr)
		}

		resultData = map[string]any{
			"timeframe":        resolutionStr,
			"symbols":          candlesMap,
			"history_metadata": metadataMap,
		}
	}

	return utils.ArtifactsResult(summary, resultData)
}

func latestHistory(data *market.HistoryResponse, limit int) (*market.HistoryResponse, error) {
	if data == nil || limit < 1 {
		return nil, fmt.Errorf("invalid history response or limit")
	}
	n := len(data.Time)
	if len(data.Open) != n || len(data.High) != n || len(data.Low) != n || len(data.Close) != n || len(data.Volume) != n {
		return nil, fmt.Errorf("history OHLCV arrays have mismatched lengths")
	}
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	sort.SliceStable(indices, func(i, j int) bool { return data.Time[indices[i]] < data.Time[indices[j]] })
	unique := make([]int, 0, len(indices))
	for _, index := range indices {
		if len(unique) > 0 && data.Time[unique[len(unique)-1]] == data.Time[index] {
			unique[len(unique)-1] = index
		} else {
			unique = append(unique, index)
		}
	}
	indices = unique
	if len(indices) > limit {
		indices = indices[len(indices)-limit:]
	}
	result := &market.HistoryResponse{Status: data.Status}
	for _, i := range indices {
		result.Time = append(result.Time, data.Time[i])
		result.Open = append(result.Open, data.Open[i])
		result.High = append(result.High, data.High[i])
		result.Low = append(result.Low, data.Low[i])
		result.Close = append(result.Close, data.Close[i])
		result.Volume = append(result.Volume, data.Volume[i])
	}
	return result, nil
}

func closedHistory(data *market.HistoryResponse, limit, resolution int, now int64) (*market.HistoryResponse, error) {
	if data == nil || resolution <= 0 || limit <= 0 {
		return nil, fmt.Errorf("invalid closed history request")
	}
	ordered, err := latestHistory(data, max(1, len(data.Time)))
	if err != nil {
		return nil, err
	}
	interval := int64(resolution) * 60
	end := 0
	for end < len(ordered.Time) && ordered.Time[end]+interval <= now {
		end++
	}
	if end == 0 || now-(ordered.Time[end-1]+interval) >= interval*2 {
		return nil, fmt.Errorf("no recent completed candles for %d-minute timeframe", resolution)
	}
	ordered.Time = ordered.Time[:end]
	ordered.Open = ordered.Open[:end]
	ordered.High = ordered.High[:end]
	ordered.Low = ordered.Low[:end]
	ordered.Close = ordered.Close[:end]
	ordered.Volume = ordered.Volume[:end]
	return latestHistory(ordered, limit)
}

func fillHistoryGaps(data *market.HistoryResponse, limit, resolution int, now int64) (*market.HistoryResponse, error) {
	if data == nil || limit <= 0 || resolution <= 0 {
		return nil, fmt.Errorf("invalid history gap request")
	}
	ordered, err := latestHistory(data, max(1, len(data.Time)))
	if err != nil {
		return nil, err
	}
	interval := int64(resolution) * 60
	result := &market.HistoryResponse{Status: ordered.Status}
	appendCandle := func(ts int64, open, high, low, close, volume float64) {
		result.Time = append(result.Time, ts)
		result.Open = append(result.Open, open)
		result.High = append(result.High, high)
		result.Low = append(result.Low, low)
		result.Close = append(result.Close, close)
		result.Volume = append(result.Volume, volume)
	}
	for i, ts := range ordered.Time {
		if ts+interval > now {
			continue
		}
		for _, price := range []float64{ordered.Open[i], ordered.High[i], ordered.Low[i], ordered.Close[i]} {
			if price <= 0 || math.IsNaN(price) || math.IsInf(price, 0) {
				return nil, fmt.Errorf("invalid source candle price")
			}
		}
		if ordered.Volume[i] < 0 || math.IsNaN(ordered.Volume[i]) || math.IsInf(ordered.Volume[i], 0) {
			return nil, fmt.Errorf("invalid source candle volume")
		}
		if len(result.Time) > 0 {
			last := result.Time[len(result.Time)-1]
			if ts == last {
				return nil, fmt.Errorf("duplicate source candle timestamp")
			}
			if (ts-last)%interval != 0 {
				return nil, fmt.Errorf("source candles do not align with requested timeframe")
			}
			if (ts-last)/interval > int64(limit) {
				return nil, fmt.Errorf("source candle gap exceeds requested history window")
			}
			for next := last + interval; next < ts; next += interval {
				close := result.Close[len(result.Close)-1]
				appendCandle(next, close, close, close, close, 0)
			}
		}
		appendCandle(ts, ordered.Open[i], ordered.High[i], ordered.Low[i], ordered.Close[i], ordered.Volume[i])
	}
	if len(result.Time) == 0 {
		return nil, fmt.Errorf("no completed source candles")
	}
	last := result.Time[len(result.Time)-1]
	if now-(last+interval) >= int64(limit)*interval {
		return nil, fmt.Errorf("source candles are older than requested history window")
	}
	for next := last + interval; next+interval <= now; next += interval {
		close := result.Close[len(result.Close)-1]
		appendCandle(next, close, close, close, close, 0)
	}
	return latestHistory(result, limit)
}
