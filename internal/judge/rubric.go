// Package judge implements LLM-as-Judge code quality scoring.
//
// It calls the Anthropic Messages API to evaluate code changes against
// a rubric of quality dimensions (correctness, readability, simplicity,
// robustness, minimality, maintainability) plus Go idioms.
package judge

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"text/template"
	"time"
)

//go:embed prompts/rubric_universal.md
var rubricPromptTemplate string

// Dimension weights for composite score calculation.
var dimensionWeights = map[string]float64{
	"correctness":     0.25,
	"readability":     0.15,
	"simplicity":      0.15,
	"robustness":      0.15,
	"minimality":      0.15,
	"maintainability": 0.15,
}

// RubricScore holds the complete LLM Judge evaluation result.
//
// @implements REQ-JUDGE-RUBRIC (LLM-as-Judge rubric scoring for code quality)
type RubricScore struct {
	Dimensions          map[string]DimensionScore // correctness, readability, simplicity, robustness, minimality, maintainability
	GoIdioms            *DimensionScore           // reported separately, not in composite
	Composite           float64                   // weighted average of dimension scores (0-5)
	Summary             string
	ConsistencyWarnings []string // P13: boolean-to-score consistency warnings per dimension
}

// DimensionScore holds the evaluation for a single quality dimension.
type DimensionScore struct {
	Score         int             // 0-5
	Booleans      map[string]bool // 6 sub-questions (q1-q6)
	Justification string
}

// CheckConsistency validates that the boolean sub-question results are consistent
// with the assigned numeric score for this dimension. It compares the count of
// true booleans against the score using an expected mapping:
//
//	0 true -> score 0, 1 true -> score 1, ..., 4 true -> score 4, 5-6 true -> score 5
//
// Returns a non-empty warning string if the score deviates by more than 1 from
// the expected value, indicating the LLM judge may have been inconsistent.
//
// @implements P13 (Boolean-to-Score consistency check)
func (ds *DimensionScore) CheckConsistency() string {
	trueCount := 0
	for _, v := range ds.Booleans {
		if v {
			trueCount++
		}
	}
	// Expected mapping: 0->0, 1->1, 2->2, 3->3, 4->4, 5-6->5
	expectedMin := trueCount
	if trueCount >= 5 {
		expectedMin = 5
	}

	if ds.Score < expectedMin-1 || ds.Score > expectedMin+1 {
		return fmt.Sprintf("inconsistent: %d/%d booleans true but score=%d",
			trueCount, len(ds.Booleans), ds.Score)
	}
	return ""
}

// promptData holds the template variables for the rubric prompt.
type promptData struct {
	PlanContent  string
	OriginalCode string
	Diff         string
}

// anthropicRequest is the Messages API request body.
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

// anthropicMessage is a single message in the Messages API.
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse is the Messages API response body.
type anthropicResponse struct {
	Content []anthropicContentBlock `json:"content"`
	Error   *anthropicError         `json:"error,omitempty"`
}

// anthropicContentBlock is a content block in the response.
type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// anthropicError represents an API error response.
type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// rubricJSON is the expected JSON structure from the LLM response.
type rubricJSON struct {
	Dimensions map[string]dimensionJSON `json:"dimensions"`
	GoIdioms   dimensionJSON            `json:"go_idioms"`
	Summary    string                   `json:"summary"`
}

// dimensionJSON is the JSON structure for a single dimension.
type dimensionJSON struct {
	Score         int             `json:"score"`
	Booleans      map[string]bool `json:"booleans"`
	Justification string          `json:"justification"`
}

// HTTPClient is the interface for making HTTP requests, allowing test injection.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ScoreRun evaluates code changes using the Anthropic Messages API.
// It renders the rubric prompt template with the provided inputs, calls the API,
// and parses the JSON response into a RubricScore.
//
// Parameters:
//   - ctx: context for cancellation/timeout
//   - apiKey: Anthropic API key (x-api-key header)
//   - model: model identifier (e.g. "claude-sonnet-4-20250514")
//   - planContent: the task plan/requirements
//   - originalCode: the code before changes
//   - diff: the git diff of changes
//
// @implements REQ-JUDGE-SCORE (score a run's code changes via LLM-as-Judge)
func ScoreRun(ctx context.Context, apiKey string, model string, planContent string, originalCode string, diff string) (*RubricScore, error) {
	return scoreRunWithClient(ctx, http.DefaultClient, apiKey, model, planContent, originalCode, diff)
}

// scoreRunWithClient is the internal implementation that accepts an HTTPClient for testing.
func scoreRunWithClient(ctx context.Context, client HTTPClient, apiKey string, model string, planContent string, originalCode string, diff string) (*RubricScore, error) {
	// Render prompt template.
	prompt, err := renderPrompt(planContent, originalCode, diff)
	if err != nil {
		return nil, fmt.Errorf("render prompt: %w", err)
	}

	// Build API request.
	reqBody := anthropicRequest{
		Model:     model,
		MaxTokens: 4096,
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

	// Call API with retry on 429/529 (rate limit / overloaded).
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("call anthropic API: %w", err)
		}
		if resp.StatusCode != 429 && resp.StatusCode != 529 {
			break
		}
		resp.Body.Close()
		retryAfter := resp.Header.Get("Retry-After")
		wait := time.Duration(attempt+1) * 2 * time.Second
		if retryAfter != "" {
			if secs, e := strconv.Atoi(retryAfter); e == nil {
				wait = time.Duration(secs) * time.Second
			}
		}
		// P10: Cap Retry-After at 30 seconds to prevent excessive waits.
		const maxRetryWait = 30 * time.Second
		if wait > maxRetryWait {
			wait = maxRetryWait
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
		// Rebuild request for retry (body was consumed).
		httpReq, err = http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("create retry request: %w", err)
		}
		httpReq.Header.Set("x-api-key", apiKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
		httpReq.Header.Set("content-type", "application/json")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse API response.
	var apiResp anthropicResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("parse API response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("anthropic API error: %s: %s", apiResp.Error.Type, apiResp.Error.Message)
	}

	// Extract text content from response.
	var textContent string
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			textContent = block.Text
			break
		}
	}
	if textContent == "" {
		return nil, fmt.Errorf("no text content in API response")
	}

	// Parse the JSON from the LLM's response.
	return parseRubricResponse(textContent)
}

// renderPrompt renders the rubric prompt template with the given inputs.
func renderPrompt(planContent, originalCode, diff string) (string, error) {
	tmpl, err := template.New("rubric").Parse(rubricPromptTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	data := promptData{
		PlanContent:  planContent,
		OriginalCode: originalCode,
		Diff:         diff,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

// parseRubricResponse parses the LLM's JSON response into a RubricScore.
func parseRubricResponse(text string) (*RubricScore, error) {
	var raw rubricJSON
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return nil, fmt.Errorf("parse rubric JSON: %w (response: %.200s)", err, text)
	}

	result := &RubricScore{
		Dimensions: make(map[string]DimensionScore),
		Summary:    raw.Summary,
	}

	// Convert dimensions.
	for name, dim := range raw.Dimensions {
		score := clampScore(dim.Score)
		result.Dimensions[name] = DimensionScore{
			Score:         score,
			Booleans:      dim.Booleans,
			Justification: dim.Justification,
		}
	}

	// Convert Go idioms.
	goScore := clampScore(raw.GoIdioms.Score)
	result.GoIdioms = &DimensionScore{
		Score:         goScore,
		Booleans:      raw.GoIdioms.Booleans,
		Justification: raw.GoIdioms.Justification,
	}

	// Calculate weighted composite score.
	result.Composite = calculateComposite(result.Dimensions)

	// P13: Check boolean-to-score consistency for each dimension.
	// Iterate in sorted order for deterministic warning output.
	dimNames := make([]string, 0, len(result.Dimensions))
	for name := range result.Dimensions {
		dimNames = append(dimNames, name)
	}
	sort.Strings(dimNames)
	for _, name := range dimNames {
		dim := result.Dimensions[name]
		if warning := dim.CheckConsistency(); warning != "" {
			result.ConsistencyWarnings = append(result.ConsistencyWarnings,
				fmt.Sprintf("%s: %s", name, warning))
		}
	}
	if result.GoIdioms != nil {
		if warning := result.GoIdioms.CheckConsistency(); warning != "" {
			result.ConsistencyWarnings = append(result.ConsistencyWarnings,
				fmt.Sprintf("go_idioms: %s", warning))
		}
	}

	return result, nil
}

// calculateComposite computes the weighted average of dimension scores.
func calculateComposite(dimensions map[string]DimensionScore) float64 {
	var weightedSum float64
	var totalWeight float64

	for name, weight := range dimensionWeights {
		if dim, ok := dimensions[name]; ok {
			weightedSum += float64(dim.Score) * weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}
	return weightedSum / totalWeight
}

// clampScore ensures a score is within the valid range [0, 5].
func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 5 {
		return 5
	}
	return score
}
