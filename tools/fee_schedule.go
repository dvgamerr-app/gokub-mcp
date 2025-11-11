package tools

import (
	"context"
	"fmt"
	"gokub/utils"

	"github.com/dvgamerr-app/go-bitkub/market"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rs/zerolog/log"
)

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

	feeSchedule := determineFeeSchedule(credits)

	log.Info().
		Float64("credits", credits).
		Str("level", feeSchedule.Level).
		Float64("maker", feeSchedule.MakerFee).
		Float64("taker", feeSchedule.TakerFee).
		Msg("Retrieved fee schedule")

	result := fmt.Sprintf(`💰 Fee Schedule: Trading Credits %.2f | Level: %s | Maker Fee: %.2f%% | Taker Fee: %.2f%% | %s`,
		feeSchedule.TradingCredits,
		feeSchedule.Level,
		feeSchedule.MakerFee*100,
		feeSchedule.TakerFee*100,
		feeSchedule.Description,
	)

	return utils.TextResult(result)
}

func determineFeeSchedule(credits float64) FeeSchedule {
	switch {
	case credits >= 50000000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "VIP 4",
			MakerFee:       0.0000,
			TakerFee:       0.0010,
			Description:    "Trading Credits ≥ 50M - Highest tier",
		}
	case credits >= 10000000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "VIP 3",
			MakerFee:       0.0000,
			TakerFee:       0.0015,
			Description:    "Trading Credits ≥ 10M",
		}
	case credits >= 5000000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "VIP 2",
			MakerFee:       0.0005,
			TakerFee:       0.0020,
			Description:    "Trading Credits ≥ 5M",
		}
	case credits >= 1000000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "VIP 1",
			MakerFee:       0.0010,
			TakerFee:       0.0023,
			Description:    "Trading Credits ≥ 1M",
		}
	case credits >= 500000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Level 5",
			MakerFee:       0.0015,
			TakerFee:       0.0023,
			Description:    "Trading Credits ≥ 500K",
		}
	case credits >= 100000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Level 4",
			MakerFee:       0.0020,
			TakerFee:       0.0023,
			Description:    "Trading Credits ≥ 100K",
		}
	case credits >= 50000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Level 3",
			MakerFee:       0.0023,
			TakerFee:       0.0023,
			Description:    "Trading Credits ≥ 50K",
		}
	case credits >= 10000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Level 2",
			MakerFee:       0.0024,
			TakerFee:       0.0024,
			Description:    "Trading Credits ≥ 10K",
		}
	case credits >= 1000:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Level 1",
			MakerFee:       0.0025,
			TakerFee:       0.0025,
			Description:    "Trading Credits ≥ 1K",
		}
	default:
		return FeeSchedule{
			TradingCredits: credits,
			Level:          "Standard",
			MakerFee:       0.0025,
			TakerFee:       0.0025,
			Description:    "Standard tier - No trading credits",
		}
	}
}

func GetFeeSchedule(credits float64) FeeSchedule {
	return determineFeeSchedule(credits)
}
