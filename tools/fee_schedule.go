package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"

	"github.com/dvgamerr-app/go-bitkub/market"
	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/rs/zerolog/log"
)

// MakerFee/TakerFee are percentage points (e.g. 0.25 means 0.25%), matching
// what calculate_position_size/check_trade_pnl/pnl_with_fees/simulate_trade/
// log_trade_exit expect for their maker_fee/taker_fee inputs (they divide by
// 100 internally) — chain this straight through without re-scaling.
type FeeSchedule struct {
	TradingCredits float64 `json:"trading_credits"`
	Level          string  `json:"level"`
	MakerFee       float64 `json:"maker_fee"`
	TakerFee       float64 `json:"taker_fee"`
	Description    string  `json:"description"`
}

func NewFeeScheduleTool() mcp.Tool {
	return mcp.NewTool("get_fee_schedule",
		mcp.WithDescription("Get trading fee schedule (maker/taker rates) based on user's trading level and credits"),
	)
}

func FeeScheduleHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Debug().Msg("Getting fee schedule")

	credits, err := market.GetTradingCredits()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get trading credits")
		return utils.ErrorResult(fmt.Sprintf("error: %v", err))
	}

	fee := determineFeeSchedule(credits)

	return utils.ArtifactsResult(fmt.Sprintf(`💰 Fee Schedule: Trading Credits %.2f | Level: %s | Maker Fee: %.2f%% | Taker Fee: %.2f%% | %s`,
		fee.TradingCredits,
		fee.Level,
		fee.MakerFee,
		fee.TakerFee,
		fee.Description,
	), fee)
}

func determineFeeSchedule(credits float64) FeeSchedule {
	switch {
	case credits >= 50000000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "VIP 4",
			MakerFee:       0.00,
			TakerFee:       0.10,
			Description:    "Trading Credits ≥ 50M - Highest tier",
		}
	case credits >= 10000000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "VIP 3",
			MakerFee:       0.00,
			TakerFee:       0.15,
			Description:    "Trading Credits ≥ 10M",
		}
	case credits >= 5000000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "VIP 2",
			MakerFee:       0.05,
			TakerFee:       0.20,
			Description:    "Trading Credits ≥ 5M",
		}
	case credits >= 1000000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "VIP 1",
			MakerFee:       0.10,
			TakerFee:       0.23,
			Description:    "Trading Credits ≥ 1M",
		}
	case credits >= 500000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Level 5",
			MakerFee:       0.15,
			TakerFee:       0.23,
			Description:    "Trading Credits ≥ 500K",
		}
	case credits >= 100000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Level 4",
			MakerFee:       0.20,
			TakerFee:       0.23,
			Description:    "Trading Credits ≥ 100K",
		}
	case credits >= 50000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Level 3",
			MakerFee:       0.23,
			TakerFee:       0.23,
			Description:    "Trading Credits ≥ 50K",
		}
	case credits >= 10000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Level 2",
			MakerFee:       0.24,
			TakerFee:       0.24,
			Description:    "Trading Credits ≥ 10K",
		}
	case credits >= 1000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Level 1",
			MakerFee:       0.25,
			TakerFee:       0.25,
			Description:    "Trading Credits ≥ 1K",
		}
	default:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Standard",
			MakerFee:       0.25,
			TakerFee:       0.25,
			Description:    "Standard tier - No trading credits",
		}
	}
}
