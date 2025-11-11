package tools

import (
	"context"
	"fmt"
	"gokub/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

type PositionSizeInput struct {
	Balance     float64 `json:"balance" jsonschema:"description=Total available balance in THB"`
	RiskPercent float64 `json:"risk_percent" jsonschema:"description=Risk percentage per trade (e.g. 2 for 2%)"`
	Entry       float64 `json:"entry" jsonschema:"description=Entry price"`
	Stop        float64 `json:"stop" jsonschema:"description=Stop loss price"`
	MakerFee    float64 `json:"maker_fee,omitempty" jsonschema:"description=Maker fee percentage (optional, default 0.25%)"`
	TakerFee    float64 `json:"taker_fee,omitempty" jsonschema:"description=Taker fee percentage (optional, default 0.25%)"`
}

type PositionSizeOutput struct {
	Balance          float64 `json:"balance"`
	RiskPercent      float64 `json:"risk_percent"`
	RiskTHB          float64 `json:"risk_thb"`
	Entry            float64 `json:"entry"`
	Stop             float64 `json:"stop"`
	StopFrac         float64 `json:"stop_frac"`
	PositionValueTHB float64 `json:"position_value_thb"`
	Qty              float64 `json:"qty"`
	TakeProfit2R     float64 `json:"take_profit_2R"`
	MakerFee         float64 `json:"maker_fee"`
	TakerFee         float64 `json:"taker_fee"`
	TotalFee         float64 `json:"total_fee"`
}

func NewCalculatePositionSizeTool() mcp.Tool {
	return mcp.NewTool("calculate_position_size",
		mcp.WithDescription("Calculate position size based on risk management. Formula: stop_frac = (entry - stop)/entry, position = risk_thb/stop_frac"),
		mcp.WithNumber("balance",
			mcp.Required(),
			mcp.Description("Total available balance in THB"),
		),
		mcp.WithNumber("risk_percent",
			mcp.Required(),
			mcp.Description("Risk percentage per trade (e.g. 2 for 2%)"),
		),
		mcp.WithNumber("entry",
			mcp.Required(),
			mcp.Description("Entry price"),
		),
		mcp.WithNumber("stop",
			mcp.Required(),
			mcp.Description("Stop loss price"),
		),
		mcp.WithNumber("maker_fee",
			mcp.Description("Maker fee percentage (optional, default 0.25%)"),
			mcp.DefaultNumber(0.25),
		),
		mcp.WithNumber("taker_fee",
			mcp.Description("Taker fee percentage (optional, default 0.25%)"),
			mcp.DefaultNumber(0.25),
		),
	)
}

func CalculatePositionSizeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	balance := utils.GetFloat64Arg(args, "balance")
	if balance <= 0 {
		return utils.ErrorResult("balance must be a positive number")
	}

	riskPercent := utils.GetFloat64Arg(args, "risk_percent")
	if riskPercent <= 0 || riskPercent > 100 {
		return utils.ErrorResult("risk_percent must be between 0 and 100")
	}

	entry := utils.GetFloat64Arg(args, "entry")
	if entry <= 0 {
		return utils.ErrorResult("entry price must be positive")
	}

	stop := utils.GetFloat64Arg(args, "stop")
	if stop <= 0 {
		return utils.ErrorResult("stop price must be positive")
	}

	if stop >= entry {
		return utils.ErrorResult("stop price must be lower than entry price (long position)")
	}

	makerFee := utils.GetFloat64Arg(args, "maker_fee", 0.25)
	takerFee := utils.GetFloat64Arg(args, "taker_fee", 0.25)

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
		Balance:          utils.Round(balance),
		RiskPercent:      utils.Round(riskPercent, 2),
		RiskTHB:          utils.Round(riskTHB),
		Entry:            utils.Round(entry),
		Stop:             utils.Round(stop),
		StopFrac:         utils.Round(stopFrac),
		PositionValueTHB: utils.Round(positionValueTHB),
		Qty:              utils.Round(qty),
		TakeProfit2R:     utils.Round(feeAdjustedTP),
		MakerFee:         utils.Round(makerFee, 2),
		TakerFee:         utils.Round(takerFee, 2),
		TotalFee:         utils.Round(totalFeeFrac*100, 2),
	}

	return utils.ArtifactsResult(fmt.Sprintf(`📊 Position Size Calculation:
• Balance: %.2f THB
• Risk: %.2f%% = %.2f THB
• Entry: %.2f | Stop: %.2f
• Stop Distance: %.2f%% (%.4f fraction)
• Position Value: %.2f THB
• Quantity: %.6f coins
• Take Profit (2R): %.2f
• Fees: Maker %.2f%% + Taker %.2f%% = %.2f%%`,
		output.Balance,
		output.RiskPercent,
		output.RiskTHB,
		output.Entry,
		output.Stop,
		utils.Round(output.StopFrac*100, 2),
		output.StopFrac,
		output.PositionValueTHB,
		output.Qty,
		output.TakeProfit2R,
		output.MakerFee,
		output.TakerFee,
		output.TotalFee,
	), output)
}
