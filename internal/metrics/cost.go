// Package metrics provides benchmark score calculations.
//
// Spec: docs/metrics/spec.md
package metrics

// EstimateCost calculates estimated USD cost from input/output token counts and pricing.
//
// Parameters:
//   - inputTokens: number of input tokens consumed
//   - outputTokens: number of output tokens generated
//   - inputPricePerMTok: price per million input tokens (USD)
//   - outputPricePerMTok: price per million output tokens (USD)
//
// Returns the estimated cost in USD.
//
// @implements P8 (unified cost estimation in metrics package)
func EstimateCost(inputTokens, outputTokens int, inputPricePerMTok, outputPricePerMTok float64) float64 {
	return float64(inputTokens)/1_000_000*inputPricePerMTok +
		float64(outputTokens)/1_000_000*outputPricePerMTok
}
