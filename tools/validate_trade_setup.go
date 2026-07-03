package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// ponytail: pure validator — takes metrics already computed by the other tools
// (check_market_regime, calculate_relative_strength_rank, calculate_atr,
// detect_*_signal, calculate_position_size) instead of re-running the whole
// pipeline. Keeps it network-free and testable; the LLM chains the inputs.

type ValidateSetupInput struct {
	Regime        string  `json:"regime"`
	RSTop         bool    `json:"rs_top"`
	ATRPercent    float64 `json:"atr_percent"`
	ATRMin        float64 `json:"atr_min"`
	ATRMax        float64 `json:"atr_max"`
	HasSignal     bool    `json:"has_signal"`
	VolumeOK      bool    `json:"volume_ok"`
	PositionValue float64 `json:"position_value_thb"`
	Balance       float64 `json:"balance"`
}

type ValidateSetupOutput struct {
	Checklist map[string]bool `json:"checklist"`
	Score     string          `json:"score"`
	CanTrade  bool            `json:"can_trade"`
	Warnings  []string        `json:"warnings"`
}

func NewValidateTradeSetupTool() mcp.Tool {
	return mcp.NewTool("validate_trade_setup",
		mcp.WithDescription("Validate a pre-trade checklist from metrics produced by the other tools. Returns checklist, score, can_trade, and warnings. can_trade is true only when ALL checks pass."),
		mcp.WithString("regime",
			mcp.Required(),
			mcp.Description("Market regime from check_market_regime (UPTREND|DOWNTREND|SIDEWAYS)"),
		),
		mcp.WithBoolean("rs_top",
			mcp.Required(),
			mcp.Description("Is the symbol in the top relative-strength rank?"),
		),
		mcp.WithNumber("atr_percent",
			mcp.Required(),
			mcp.Description("ATR% from calculate_atr"),
		),
		mcp.WithNumber("atr_min",
			mcp.Description("Lower bound of acceptable ATR%% zone (default 2)"),
			mcp.DefaultNumber(2),
		),
		mcp.WithNumber("atr_max",
			mcp.Description("Upper bound of acceptable ATR%% zone (default 6)"),
			mcp.DefaultNumber(6),
		),
		mcp.WithBoolean("has_signal",
			mcp.Required(),
			mcp.Description("Did a breakout/pullback entry signal fire?"),
		),
		mcp.WithBoolean("volume_ok",
			mcp.Required(),
			mcp.Description("Is volume confirming the signal (e.g. >=1.5x average)?"),
		),
		mcp.WithNumber("position_value_thb",
			mcp.Required(),
			mcp.Description("position_value_thb from calculate_position_size"),
		),
		mcp.WithNumber("balance",
			mcp.Required(),
			mcp.Description("Available balance in THB"),
		),
	)
}

func evaluateTradeSetup(in ValidateSetupInput) ValidateSetupOutput {
	checklist := map[string]bool{
		"regime_uptrend":     strings.EqualFold(in.Regime, "UPTREND"),
		"relative_strength":  in.RSTop,
		"atr_in_zone":        in.ATRPercent >= in.ATRMin && in.ATRPercent <= in.ATRMax,
		"entry_signal":       in.HasSignal,
		"volume_confirmed":   in.VolumeOK,
		"position_in_budget": in.PositionValue > 0 && in.PositionValue <= in.Balance,
	}

	warnings := []string{}
	passed := 0
	for name, ok := range checklist {
		if ok {
			passed++
		} else {
			warnings = append(warnings, "failed: "+name)
		}
	}

	return ValidateSetupOutput{
		Checklist: checklist,
		Score:     fmt.Sprintf("%d/%d", passed, len(checklist)),
		CanTrade:  passed == len(checklist),
		Warnings:  warnings,
	}
}

func ValidateTradeSetupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		return utils.ErrorResult("invalid arguments")
	}

	balance := utils.GetFloat64Arg(args, "balance")
	if balance <= 0 {
		return utils.ErrorResult("balance must be positive")
	}

	in := ValidateSetupInput{
		Regime:        utils.GetStringArg(args, "regime"),
		RSTop:         utils.GetBoolArg(args, "rs_top"),
		ATRPercent:    utils.GetFloat64Arg(args, "atr_percent"),
		ATRMin:        utils.GetFloat64Arg(args, "atr_min", 2),
		ATRMax:        utils.GetFloat64Arg(args, "atr_max", 6),
		HasSignal:     utils.GetBoolArg(args, "has_signal"),
		VolumeOK:      utils.GetBoolArg(args, "volume_ok"),
		PositionValue: utils.GetFloat64Arg(args, "position_value_thb"),
		Balance:       balance,
	}

	out := evaluateTradeSetup(in)

	verdict := "❌ DO NOT TRADE"
	if out.CanTrade {
		verdict = "✅ OK TO TRADE"
	}

	return utils.ArtifactsResult(fmt.Sprintf(`🧪 Trade Setup: %s | score %s | warnings: %v`,
		verdict,
		out.Score,
		out.Warnings,
	), out)
}
