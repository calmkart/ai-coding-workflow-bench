// Package judge — pairwise comparison of two code implementations.
//
// PairwiseCompare evaluates two code diffs side-by-side using the Anthropic
// Messages API. It runs the comparison twice (A-B and B-A) to detect position
// bias, and reports a final winner only when both orderings agree.
//
// @implements P17 (Pairwise Comparison)
package judge

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
)

//go:embed prompts/pairwise_compare.md
var pairwisePromptTemplate string

// PairwiseResult holds the outcome of a pairwise comparison between two implementations.
//
// @implements P17 (Pairwise Comparison result structure)
type PairwiseResult struct {
	Winner             string            // "left", "right", "tie"
	Reasoning          string            // LLM reasoning for the overall winner
	Dimensions         map[string]string // per-dimension winner: "left", "right", "tie"
	PositionConsistent bool              // true if AB and BA orderings agree
}

// pairwisePromptData holds the template variables for the pairwise comparison prompt.
type pairwisePromptData struct {
	PlanContent string
	DiffA       string
	DiffB       string
}

// pairwiseJSON is the expected JSON structure from the LLM pairwise response.
type pairwiseJSON struct {
	OverallWinner string            `json:"overall_winner"` // "A", "B", "Tie"
	Reasoning     string            `json:"reasoning"`
	Dimensions    map[string]string `json:"dimensions"` // dimension -> "A"|"B"|"Tie"
}

// PairwiseCompare performs a pairwise comparison of two code diffs against a plan.
// It calls the LLM twice (A=left,B=right then A=right,B=left) and checks for
// position consistency. If both orderings agree, the result is taken as-is;
// otherwise the winner is "tie" with PositionConsistent=false.
//
// @implements P17 (Pairwise Comparison entry point)
func PairwiseCompare(ctx context.Context, apiKey, model, planContent, diffLeft, diffRight string) (*PairwiseResult, error) {
	return pairwiseCompareWithClient(ctx, http.DefaultClient, apiKey, model, planContent, diffLeft, diffRight)
}

// pairwiseCompareWithClient is the internal implementation that accepts an HTTPClient for testing.
func pairwiseCompareWithClient(ctx context.Context, client HTTPClient, apiKey, model, planContent, diffLeft, diffRight string) (*PairwiseResult, error) {
	// Round 1: A=left, B=right.
	r1, err := callPairwise(ctx, client, apiKey, model, planContent, diffLeft, diffRight)
	if err != nil {
		return nil, fmt.Errorf("pairwise round 1: %w", err)
	}

	// Round 2: A=right, B=left (position swap).
	r2, err := callPairwise(ctx, client, apiKey, model, planContent, diffRight, diffLeft)
	if err != nil {
		return nil, fmt.Errorf("pairwise round 2: %w", err)
	}

	// Normalize round 1: A=left, B=right -> winner is as-is.
	r1Winner := normalizeWinner(r1.OverallWinner, false) // "A"->"left", "B"->"right"
	r1Dims := normalizeDimensions(r1.Dimensions, false)

	// Normalize round 2: A=right, B=left -> flip A/B mapping.
	r2Winner := normalizeWinner(r2.OverallWinner, true) // "A"->"right", "B"->"left"
	r2Dims := normalizeDimensions(r2.Dimensions, true)

	// Check position consistency.
	consistent := r1Winner == r2Winner

	result := &PairwiseResult{
		PositionConsistent: consistent,
		Reasoning:          r1.Reasoning,
	}

	if consistent {
		result.Winner = r1Winner
		result.Dimensions = r1Dims
	} else {
		result.Winner = "tie"
		result.Reasoning = fmt.Sprintf("inconsistent: round1=%s, round2=%s. R1: %s; R2: %s",
			r1Winner, r2Winner, r1.Reasoning, r2.Reasoning)
		// Merge dimensions: agree -> keep; disagree -> tie.
		result.Dimensions = mergeDimensions(r1Dims, r2Dims)
	}

	return result, nil
}

// callPairwise makes a single pairwise comparison API call.
func callPairwise(ctx context.Context, client HTTPClient, apiKey, model, planContent, diffA, diffB string) (*pairwiseJSON, error) {
	prompt, err := renderPairwisePrompt(planContent, diffA, diffB)
	if err != nil {
		return nil, fmt.Errorf("render prompt: %w", err)
	}

	reqBody := anthropicRequest{
		Model:     model,
		MaxTokens: 2048,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("API error: %s: %s", apiResp.Error.Type, apiResp.Error.Message)
	}

	var textContent string
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			textContent = block.Text
			break
		}
	}
	if textContent == "" {
		return nil, fmt.Errorf("no text content in response")
	}

	var result pairwiseJSON
	if err := json.Unmarshal([]byte(textContent), &result); err != nil {
		return nil, fmt.Errorf("parse pairwise JSON: %w (response: %.200s)", err, textContent)
	}

	return &result, nil
}

// renderPairwisePrompt renders the pairwise comparison prompt template.
func renderPairwisePrompt(planContent, diffA, diffB string) (string, error) {
	tmpl, err := template.New("pairwise").Parse(pairwisePromptTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	data := pairwisePromptData{
		PlanContent: planContent,
		DiffA:       diffA,
		DiffB:       diffB,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

// normalizeWinner converts "A"/"B"/"Tie" to "left"/"right"/"tie".
// When flipped=true, A maps to "right" and B maps to "left" (position swap).
func normalizeWinner(raw string, flipped bool) string {
	switch raw {
	case "A":
		if flipped {
			return "right"
		}
		return "left"
	case "B":
		if flipped {
			return "left"
		}
		return "right"
	default:
		return "tie"
	}
}

// normalizeDimensions converts per-dimension "A"/"B"/"Tie" values to "left"/"right"/"tie".
func normalizeDimensions(dims map[string]string, flipped bool) map[string]string {
	result := make(map[string]string, len(dims))
	for k, v := range dims {
		result[k] = normalizeWinner(v, flipped)
	}
	return result
}

// mergeDimensions merges two dimension maps: where both agree, keep the value;
// where they disagree, set "tie".
func mergeDimensions(d1, d2 map[string]string) map[string]string {
	result := make(map[string]string)
	// Collect all dimension keys.
	allKeys := make(map[string]bool)
	for k := range d1 {
		allKeys[k] = true
	}
	for k := range d2 {
		allKeys[k] = true
	}
	for k := range allKeys {
		v1, ok1 := d1[k]
		v2, ok2 := d2[k]
		if ok1 && ok2 && v1 == v2 {
			result[k] = v1
		} else {
			result[k] = "tie"
		}
	}
	return result
}
