package tools

import (
	"context"
	"fmt"
	"gokub/utils"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

func parseFloat(args map[string]any, key string) (float64, error) {
	val, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("%s is required", key)
	}

	switch v := val.(type) {
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	case int:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("%s must be a number", key)
	}
}

type PositionSizeInput struct {
	Balance     float64 `json:"balance" jsonschema:"description=Total available balance in THB"`
	RiskPercent float64 `json:"risk_percent" jsonschema:"description=Risk percentage per trade (e.g. 2 for 2%)"`
	Entry       float64 `json:"entry" jsonschema:"description=Entry price"`
	Stop        float64 `json:"stop" jsonschema:"description=Stop loss price"`
	MakerFee    float64 `json:"maker_fee,omitempty" jsonschema:"description=Maker fee percentage (optional, default 0.25%)"`
	TakerFee    float64 `json:"taker_fee,omitempty" jsonschema:"description=Taker fee percentage (optional, default 0.25%)"`
}

type PositionSizeOutput struct {
	RiskTHB          float64 `json:"risk_thb"`
	StopFrac         float64 `json:"stop_frac"`
	PositionValueTHB float64 `json:"position_value_thb"`
	Qty              float64 `json:"qty"`
	TakeProfit2R     float64 `json:"take_profit_2R"`
}

func NewCalculatePositionSizeTool() mcp.Tool {
	return mcp.NewTool("calculate_position_size",
		mcp.WithDescription("Calculate position size based on risk management. Formula: stop_frac = (entry - stop)/entry, position = risk_thb/stop_frac"),
		mcp.WithString("balance",
			mcp.Required(),
			mcp.Description("Total available balance in THB"),
		),
		mcp.WithString("risk_percent",
			mcp.Required(),
			mcp.Description("Risk percentage per trade (e.g. 2 for 2%)"),
		),
		mcp.WithString("entry",
			mcp.Required(),
			mcp.Description("Entry price"),
		),
		mcp.WithString("stop",
			mcp.Required(),
			mcp.Description("Stop loss price"),
		),
		mcp.WithString("maker_fee",
			mcp.Description("Maker fee percentage (optional, default 0.25%)"),
		),
		mcp.WithString("taker_fee",
			mcp.Description("Taker fee percentage (optional, default 0.25%)"),
		),
	)
}

func CalculatePositionSizeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	balance, err := parseFloat(args, "balance")
	if err != nil || balance <= 0 {
		return utils.ErrorResult("balance must be a positive number")
	}

	riskPercent, err := parseFloat(args, "risk_percent")
	if err != nil || riskPercent <= 0 || riskPercent > 100 {
		return utils.ErrorResult("risk_percent must be between 0 and 100")
	}

	entry, err := parseFloat(args, "entry")
	if err != nil || entry <= 0 {
		return utils.ErrorResult("entry price must be positive")
	}

	stop, err := parseFloat(args, "stop")
	if err != nil || stop <= 0 {
		return utils.ErrorResult("stop price must be positive")
	}

	if stop >= entry {
		return utils.ErrorResult("stop price must be lower than entry price (long position)")
	}

	makerFee := 0.25
	if feeVal, err := parseFloat(args, "maker_fee"); err == nil && feeVal > 0 {
		makerFee = feeVal
	}

	takerFee := 0.25
	if feeVal, err := parseFloat(args, "taker_fee"); err == nil && feeVal > 0 {
		takerFee = feeVal
	}

	riskTHB := balance * (riskPercent / 100)
	stopFrac := (entry - stop) / entry
	positionValueTHB := riskTHB / stopFrac
	qty := positionValueTHB / entry

	entryFeeFrac := makerFee / 100
	exitFeeFrac := takerFee / 100
	totalFeeFrac := entryFeeFrac + exitFeeFrac

	riskPerCoin := entry - stop
	targetGainPerCoin := 2 * riskPerCoin
	takeProfitPrice := entry + targetGainPerCoin

	feeAdjustedTP := takeProfitPrice * (1 + totalFeeFrac)

	output := PositionSizeOutput{
		RiskTHB:          riskTHB,
		StopFrac:         stopFrac,
		PositionValueTHB: positionValueTHB,
		Qty:              qty,
		TakeProfit2R:     feeAdjustedTP,
	}

	log.Info().
		Float64("balance", balance).
		Float64("risk_percent", riskPercent).
		Float64("entry", entry).
		Float64("stop", stop).
		Float64("risk_thb", riskTHB).
		Float64("stop_frac", stopFrac).
		Float64("position_value_thb", positionValueTHB).
		Float64("qty", qty).
		Float64("take_profit_2R", feeAdjustedTP).
		Msg("Calculated position size")

	contents := fmt.Sprintf(`📊 Position Size Calculation:
• Balance: %.2f THB
• Risk: %.2f%% = %.2f THB
• Entry: %.2f | Stop: %.2f
• Stop Distance: %.2f%% (%.4f fraction)
• Position Value: %.2f THB
• Quantity: %.6f coins
• Take Profit (2R): %.2f
• Fees: Maker %.2f%% + Taker %.2f%% = %.2f%%`,
		balance,
		riskPercent,
		riskTHB,
		entry,
		stop,
		stopFrac*100,
		stopFrac,
		positionValueTHB,
		qty,
		feeAdjustedTP,
		makerFee,
		takerFee,
		totalFeeFrac*100,
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: contents,
			},
		},
		StructuredContent: output,
	}, nil
}
