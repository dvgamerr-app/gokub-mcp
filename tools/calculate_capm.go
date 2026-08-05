package tools

import (
	"context"
	"fmt"
	"github.com/dvgamerr-app/gokub-mcp/utils"

	mcp "github.com/dvgamerr-app/gokub-mcp/internal/mcpcompat"
	"github.com/rs/zerolog/log"
)

type CAPMResult struct {
	ExpectedReturn float64 `json:"expected_return"`
	RiskFreeRate   float64 `json:"risk_free_rate"`
	Beta           float64 `json:"beta"`
	MarketReturn   float64 `json:"market_return"`
	MarketPremium  float64 `json:"market_premium"`
	RiskPremium    float64 `json:"risk_premium"`
	Interpretation string  `json:"interpretation"`
}

func NewCalculateCAPMTool() mcp.Tool {
	return mcp.NewTool("calculate_capm",
		mcp.WithDescription(`Calculate Capital Asset Pricing Model (CAPM) expected return.
CAPM Formula: E(Ri) = Rf + β * (E(Rm) - Rf)
Where:
- E(Ri) = Expected return of the asset
- Rf = Risk-free rate
- β = Beta (sensitivity to market movements)
- E(Rm) = Expected market return
- E(Rm) - Rf = Market risk premium`),
		mcp.WithNumber("risk_free_rate",
			mcp.Required(),
			mcp.Description("Risk-free rate of return (e.g., government bond yield, as decimal: 0.02 for 2%)"),
		),
		mcp.WithNumber("beta",
			mcp.Required(),
			mcp.Description("Beta coefficient (asset's sensitivity to market risk, 1.0 = market average)"),
		),
		mcp.WithNumber("market_return",
			mcp.Required(),
			mcp.Description("Expected market return (as decimal: 0.08 for 8%)"),
		),
	)
}

func CalculateCAPMHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := utils.ValidateArgs(request.Params.Arguments)
	if err != nil {
		log.Warn().Msg("Invalid arguments format for calculate CAPM")
		return utils.ErrorResult("invalid arguments")
	}

	riskFreeRate := utils.GetFloat64Arg(args, "risk_free_rate", 0)
	beta := utils.GetFloat64Arg(args, "beta", 0)
	marketReturn := utils.GetFloat64Arg(args, "market_return", 0)

	if riskFreeRate < 0 {
		return utils.ErrorResult("risk_free_rate must be non-negative")
	}

	if marketReturn < riskFreeRate {
		return utils.ErrorResult("market_return should typically be greater than risk_free_rate")
	}

	marketPremium := marketReturn - riskFreeRate
	riskPremium := beta * marketPremium
	expectedReturn := riskFreeRate + riskPremium

	interpretation := getInterpretation(beta, expectedReturn, marketReturn, riskFreeRate)

	result := &CAPMResult{
		ExpectedReturn: utils.Round(expectedReturn, 6),
		RiskFreeRate:   utils.Round(riskFreeRate, 6),
		Beta:           utils.Round(beta, 4),
		MarketReturn:   utils.Round(marketReturn, 6),
		MarketPremium:  utils.Round(marketPremium, 6),
		RiskPremium:    utils.Round(riskPremium, 6),
		Interpretation: interpretation,
	}

	summary := "CAPM Expected Return Calculation\n"
	summary += fmt.Sprintf("Risk-free Rate (Rf): %.2f%%\n", riskFreeRate*100)
	summary += fmt.Sprintf("Beta (β): %.4f\n", beta)
	summary += fmt.Sprintf("Market Return (Rm): %.2f%%\n", marketReturn*100)
	summary += fmt.Sprintf("Market Premium (Rm - Rf): %.2f%%\n", marketPremium*100)
	summary += fmt.Sprintf("Risk Premium (β × Market Premium): %.2f%%\n", riskPremium*100)
	summary += fmt.Sprintf("Expected Return: %.2f%%\n", expectedReturn*100)
	summary += fmt.Sprintf("\n%s", interpretation)

	return utils.ArtifactsResult(summary, result)
}

func getInterpretation(beta, expectedReturn, marketReturn, riskFreeRate float64) string {
	var betaInterpretation string
	var returnInterpretation string

	if beta < 0 {
		betaInterpretation = "Negative beta: Asset moves inversely to the market (rare, defensive)"
	} else if beta < 0.5 {
		betaInterpretation = "Low beta: Asset is less volatile than market (defensive)"
	} else if beta < 1 {
		betaInterpretation = "Moderate beta: Asset is less volatile than market"
	} else if beta == 1 {
		betaInterpretation = "Beta = 1: Asset moves in line with the market"
	} else if beta < 1.5 {
		betaInterpretation = "Moderate-high beta: Asset is more volatile than market"
	} else {
		betaInterpretation = "High beta: Asset is significantly more volatile than market (aggressive)"
	}

	if expectedReturn > marketReturn {
		returnInterpretation = "Expected return exceeds market return - higher risk, higher potential reward"
	} else if expectedReturn < marketReturn {
		returnInterpretation = "Expected return below market return - lower risk, lower potential reward"
	} else {
		returnInterpretation = "Expected return equals market return - average market risk/reward"
	}

	return fmt.Sprintf("%s\n%s", betaInterpretation, returnInterpretation)
}
