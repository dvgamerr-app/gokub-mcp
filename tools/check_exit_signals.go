package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"math"
	"slices"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

type ExitSignalsOutput struct {
	ShouldExit bool     `json:"should_exit"`
	Reasons    []string `json:"reasons"`
	Urgency    string   `json:"urgency"`
	Detail     string   `json:"detail"`
}

// evaluateExitSignals inspects the most recent candle against the prior `lookback`
// candles for: STRUCTURE_BREAK (close below recent swing low), VOLUME_DRY
// (volume < 0.5x average), REJECTION_AT_RESISTANCE (upper-wick rejection near high).
func evaluateExitSignals(candles []Candle, lookback int) ExitSignalsOutput {
	n := len(candles)
	reasons := []string{}
	if n < lookback+1 {
		return ExitSignalsOutput{Detail: "not enough candles to evaluate"}
	}

	last := candles[n-1]
	prior := candles[n-1-lookback : n-1]

	swingLow := math.MaxFloat64
	recentHigh := 0.0
	volSum := 0.0
	for _, c := range prior {
		swingLow = math.Min(swingLow, c.Low)
		recentHigh = math.Max(recentHigh, c.High)
		volSum += c.Volume
	}
	avgVol := volSum / float64(len(prior))

	if last.Close < swingLow {
		reasons = append(reasons, "STRUCTURE_BREAK")
	}
	if avgVol > 0 && last.Volume < 0.5*avgVol {
		reasons = append(reasons, "VOLUME_DRY")
	}
	upperWick := last.High - math.Max(last.Open, last.Close)
	body := math.Abs(last.Close - last.Open)
	if recentHigh > 0 && last.High >= recentHigh*0.99 && last.Close < last.Open && upperWick > body {
		reasons = append(reasons, "REJECTION_AT_RESISTANCE")
	}

	out := ExitSignalsOutput{
		ShouldExit: len(reasons) > 0,
		Reasons:    reasons,
	}
	switch {
	case len(reasons) == 0:
		out.Urgency = "none"
		out.Detail = "no exit signal"
	case slices.Contains(reasons, "STRUCTURE_BREAK") || len(reasons) >= 2:
		out.Urgency = "high"
		out.Detail = "consider exiting now"
	default:
		out.Urgency = "medium"
		out.Detail = "tighten stop / watch closely"
	}
	return out
}

func NewCheckExitSignalsTool() mcp.Tool {
	return mcp.NewTool("check_exit_signals",
		mcp.WithDescription("Check emergency exit signals from OHLCV candles: STRUCTURE_BREAK (close below recent swing low), VOLUME_DRY (<0.5x avg volume), REJECTION_AT_RESISTANCE (upper-wick rejection near recent high)."),
		mcp.WithArray("candles",
			mcp.Required(),
			mcp.Description("Array of OHLCV objects (high, low, close, open, volume)"),
		),
		mcp.WithNumber("lookback",
			mcp.Description("Number of prior candles for swing/volume reference (default 20)"),
			mcp.DefaultNumber(20),
		),
	)
}

func CheckExitSignalsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	raw, ok := args["candles"].([]any)
	if !ok {
		return utils.ErrorResult("candles must be an array")
	}

	lookback := utils.GetIntArg(args, "lookback", 20)
	if lookback < 1 {
		return utils.ErrorResult("lookback must be greater than 0")
	}

	candles := parseOHLCV(raw)
	out := evaluateExitSignals(candles, lookback)

	icon := "✅"
	if out.ShouldExit {
		icon = "🚪"
	}
	return utils.ArtifactsResult(fmt.Sprintf(`%s Exit check: should_exit=%v | urgency=%s | %s [%s]`,
		icon, out.ShouldExit, out.Urgency, out.Detail, strings.Join(out.Reasons, ","),
	), out)
}
