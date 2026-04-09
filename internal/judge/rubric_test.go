package judge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockRubricResponse returns a valid JSON rubric response for testing.
func mockRubricResponse() rubricJSON {
	return rubricJSON{
		Dimensions: map[string]dimensionJSON{
			"correctness": {
				Score:         4,
				Booleans:      map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": true, "q6": false},
				Justification: "Handles the primary path and edge cases well.",
			},
			"readability": {
				Score:         3,
				Booleans:      map[string]bool{"q1": true, "q2": true, "q3": false, "q4": true, "q5": true, "q6": false},
				Justification: "Good naming, could use more comments.",
			},
			"simplicity": {
				Score:         4,
				Booleans:      map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": true, "q6": false},
				Justification: "Straightforward approach.",
			},
			"robustness": {
				Score:         3,
				Booleans:      map[string]bool{"q1": true, "q2": true, "q3": true, "q4": false, "q5": true, "q6": false},
				Justification: "Good error handling, missing concurrent safety.",
			},
			"minimality": {
				Score:         5,
				Booleans:      map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": true, "q6": true},
				Justification: "Only necessary changes.",
			},
			"maintainability": {
				Score:         4,
				Booleans:      map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": false, "q6": true},
				Justification: "Modular and testable.",
			},
		},
		GoIdioms: dimensionJSON{
			Score:         4,
			Booleans:      map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": true, "q6": false},
			Justification: "Follows Go conventions well.",
		},
		Summary: "Solid implementation with room for improvement in readability and robustness.",
	}
}

// newMockAnthropicServer creates a test server that returns the given rubric response.
func newMockAnthropicServer(t *testing.T, rubric rubricJSON) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request structure.
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("x-api-key") == "" {
			t.Error("missing x-api-key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected anthropic-version 2023-06-01, got %s", r.Header.Get("anthropic-version"))
		}
		if r.Header.Get("content-type") != "application/json" {
			t.Errorf("expected content-type application/json, got %s", r.Header.Get("content-type"))
		}

		// Parse request body to verify structure.
		var req anthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if len(req.Messages) == 0 {
			t.Error("expected at least one message")
		}
		if req.Model == "" {
			t.Error("expected non-empty model")
		}

		// Build response.
		rubricBytes, _ := json.Marshal(rubric)
		resp := anthropicResponse{
			Content: []anthropicContentBlock{
				{Type: "text", Text: string(rubricBytes)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestScoreRun_Success(t *testing.T) {
	rubric := mockRubricResponse()
	server := newMockAnthropicServer(t, rubric)
	defer server.Close()

	// Override the API URL by using a custom HTTP client that redirects to our server.
	client := &redirectClient{server: server}

	result, err := scoreRunWithClient(
		context.Background(),
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"Implement a REST endpoint",
		"package main\n\nfunc main() {}",
		"+ func handler(w http.ResponseWriter, r *http.Request) {}",
	)
	if err != nil {
		t.Fatalf("ScoreRun: %v", err)
	}

	// Verify dimensions.
	if len(result.Dimensions) != 6 {
		t.Errorf("expected 6 dimensions, got %d", len(result.Dimensions))
	}

	corr, ok := result.Dimensions["correctness"]
	if !ok {
		t.Fatal("missing correctness dimension")
	}
	if corr.Score != 4 {
		t.Errorf("expected correctness score 4, got %d", corr.Score)
	}
	if !corr.Booleans["q1"] {
		t.Error("expected q1=true for correctness")
	}

	// Verify Go idioms.
	if result.GoIdioms == nil {
		t.Fatal("missing go_idioms")
	}
	if result.GoIdioms.Score != 4 {
		t.Errorf("expected go_idioms score 4, got %d", result.GoIdioms.Score)
	}

	// Verify composite.
	// Expected: (4*0.25 + 3*0.15 + 4*0.15 + 3*0.15 + 5*0.15 + 4*0.15) / 1.0
	//         = (1.0 + 0.45 + 0.6 + 0.45 + 0.75 + 0.6) / 1.0
	//         = 3.85
	expectedComposite := 3.85
	if result.Composite < expectedComposite-0.01 || result.Composite > expectedComposite+0.01 {
		t.Errorf("expected composite ~%.2f, got %.2f", expectedComposite, result.Composite)
	}

	// Verify summary.
	if result.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestScoreRun_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"type": "server_error", "message": "internal error"}}`))
	}))
	defer server.Close()

	client := &redirectClient{server: server}
	_, err := scoreRunWithClient(
		context.Background(),
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"plan", "code", "diff",
	)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected status 500 in error, got: %v", err)
	}
}

func TestScoreRun_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Content: []anthropicContentBlock{
				{Type: "text", Text: "This is not JSON"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &redirectClient{server: server}
	_, err := scoreRunWithClient(
		context.Background(),
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"plan", "code", "diff",
	)
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
	if !strings.Contains(err.Error(), "parse rubric JSON") {
		t.Errorf("expected 'parse rubric JSON' in error, got: %v", err)
	}
}

func TestScoreRun_EmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := anthropicResponse{
			Content: []anthropicContentBlock{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &redirectClient{server: server}
	_, err := scoreRunWithClient(
		context.Background(),
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"plan", "code", "diff",
	)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
	if !strings.Contains(err.Error(), "no text content") {
		t.Errorf("expected 'no text content' in error, got: %v", err)
	}
}

func TestScoreRun_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This handler should not be reached if context is already cancelled.
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	client := &redirectClient{server: server}
	_, err := scoreRunWithClient(
		ctx,
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"plan", "code", "diff",
	)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestRenderPrompt(t *testing.T) {
	prompt, err := renderPrompt("Test plan", "original code", "diff content")
	if err != nil {
		t.Fatalf("renderPrompt: %v", err)
	}

	if !strings.Contains(prompt, "Test plan") {
		t.Error("expected plan content in prompt")
	}
	if !strings.Contains(prompt, "original code") {
		t.Error("expected original code in prompt")
	}
	if !strings.Contains(prompt, "diff content") {
		t.Error("expected diff in prompt")
	}
	if !strings.Contains(prompt, "Correctness") {
		t.Error("expected dimension names in prompt")
	}
}

func TestParseRubricResponse(t *testing.T) {
	rubric := mockRubricResponse()
	rubricBytes, _ := json.Marshal(rubric)

	result, err := parseRubricResponse(string(rubricBytes))
	if err != nil {
		t.Fatalf("parseRubricResponse: %v", err)
	}

	if len(result.Dimensions) != 6 {
		t.Errorf("expected 6 dimensions, got %d", len(result.Dimensions))
	}
	if result.GoIdioms == nil {
		t.Error("expected go_idioms")
	}
	if result.Summary != rubric.Summary {
		t.Errorf("expected summary %q, got %q", rubric.Summary, result.Summary)
	}
}

func TestParseRubricResponse_InvalidJSON(t *testing.T) {
	_, err := parseRubricResponse("not json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// TestCheckConsistency_Consistent verifies that no warning is produced when
// booleans and score are consistent (within +/- 1 tolerance).
func TestCheckConsistency_Consistent(t *testing.T) {
	tests := []struct {
		name     string
		booleans map[string]bool
		score    int
	}{
		{"all true score 5", map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": true, "q6": true}, 5},
		{"5 true score 5", map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": true, "q6": false}, 5},
		{"5 true score 4", map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": true, "q6": false}, 4},
		{"4 true score 4", map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": false, "q6": false}, 4},
		{"4 true score 3", map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": false, "q6": false}, 3},
		{"3 true score 3", map[string]bool{"q1": true, "q2": true, "q3": true, "q4": false, "q5": false, "q6": false}, 3},
		{"0 true score 0", map[string]bool{"q1": false, "q2": false, "q3": false, "q4": false, "q5": false, "q6": false}, 0},
		{"0 true score 1", map[string]bool{"q1": false, "q2": false, "q3": false, "q4": false, "q5": false, "q6": false}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := DimensionScore{Score: tt.score, Booleans: tt.booleans}
			if w := ds.CheckConsistency(); w != "" {
				t.Errorf("expected no warning, got: %s", w)
			}
		})
	}
}

// TestCheckConsistency_Inconsistent verifies that a warning is produced when
// booleans and score disagree beyond the +/- 1 tolerance.
func TestCheckConsistency_Inconsistent(t *testing.T) {
	tests := []struct {
		name     string
		booleans map[string]bool
		score    int
	}{
		{"5 true score 2", map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": true, "q6": false}, 2},
		{"0 true score 3", map[string]bool{"q1": false, "q2": false, "q3": false, "q4": false, "q5": false, "q6": false}, 3},
		{"1 true score 4", map[string]bool{"q1": true, "q2": false, "q3": false, "q4": false, "q5": false, "q6": false}, 4},
		{"6 true score 2", map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": true, "q6": true}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := DimensionScore{Score: tt.score, Booleans: tt.booleans}
			w := ds.CheckConsistency()
			if w == "" {
				t.Error("expected warning for inconsistent score")
			}
			if !strings.Contains(w, "inconsistent") {
				t.Errorf("expected 'inconsistent' in warning, got: %s", w)
			}
		})
	}
}

// TestParseRubricResponse_ConsistencyWarnings verifies that parseRubricResponse
// populates ConsistencyWarnings when a dimension has inconsistent booleans vs score.
func TestParseRubricResponse_ConsistencyWarnings(t *testing.T) {
	// Create a rubric where readability has 5/6 true booleans but score=2 (inconsistent).
	rubric := mockRubricResponse()
	rubric.Dimensions["readability"] = dimensionJSON{
		Score:         2,
		Booleans:      map[string]bool{"q1": true, "q2": true, "q3": true, "q4": true, "q5": true, "q6": false},
		Justification: "Inconsistent score.",
	}
	rubricBytes, _ := json.Marshal(rubric)

	result, err := parseRubricResponse(string(rubricBytes))
	if err != nil {
		t.Fatalf("parseRubricResponse: %v", err)
	}

	if len(result.ConsistencyWarnings) == 0 {
		t.Fatal("expected at least one consistency warning")
	}

	found := false
	for _, w := range result.ConsistencyWarnings {
		if strings.Contains(w, "readability") && strings.Contains(w, "inconsistent") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected readability inconsistency warning, got: %v", result.ConsistencyWarnings)
	}
}

// TestParseRubricResponse_NoConsistencyWarnings verifies no warnings are produced
// for a normal consistent rubric.
func TestParseRubricResponse_NoConsistencyWarnings(t *testing.T) {
	rubric := mockRubricResponse()
	rubricBytes, _ := json.Marshal(rubric)

	result, err := parseRubricResponse(string(rubricBytes))
	if err != nil {
		t.Fatalf("parseRubricResponse: %v", err)
	}

	if len(result.ConsistencyWarnings) != 0 {
		t.Errorf("expected no consistency warnings for consistent rubric, got: %v", result.ConsistencyWarnings)
	}
}

func TestClampScore(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{-1, 0},
		{0, 0},
		{3, 3},
		{5, 5},
		{6, 5},
		{100, 5},
	}
	for _, tt := range tests {
		got := clampScore(tt.input)
		if got != tt.expected {
			t.Errorf("clampScore(%d): expected %d, got %d", tt.input, tt.expected, got)
		}
	}
}

func TestCalculateComposite(t *testing.T) {
	dimensions := map[string]DimensionScore{
		"correctness":     {Score: 5},
		"readability":     {Score: 5},
		"simplicity":      {Score: 5},
		"robustness":      {Score: 5},
		"minimality":      {Score: 5},
		"maintainability": {Score: 5},
	}

	composite := calculateComposite(dimensions)
	if composite < 4.99 || composite > 5.01 {
		t.Errorf("expected composite ~5.0, got %.2f", composite)
	}
}

func TestCalculateComposite_Empty(t *testing.T) {
	composite := calculateComposite(map[string]DimensionScore{})
	if composite != 0 {
		t.Errorf("expected composite 0 for empty dimensions, got %.2f", composite)
	}
}

func TestCalculateComposite_Partial(t *testing.T) {
	// Only correctness present: composite = 4 * 0.25 / 0.25 = 4.0
	dimensions := map[string]DimensionScore{
		"correctness": {Score: 4},
	}
	composite := calculateComposite(dimensions)
	if composite < 3.99 || composite > 4.01 {
		t.Errorf("expected composite ~4.0, got %.2f", composite)
	}
}

func TestScoreRun_VerifiesRequestHeaders(t *testing.T) {
	var capturedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()

		rubric := mockRubricResponse()
		rubricBytes, _ := json.Marshal(rubric)
		resp := anthropicResponse{
			Content: []anthropicContentBlock{
				{Type: "text", Text: string(rubricBytes)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &redirectClient{server: server}
	_, err := scoreRunWithClient(
		context.Background(),
		client,
		"sk-test-key-123",
		"claude-sonnet-4-20250514",
		"plan", "code", "diff",
	)
	if err != nil {
		t.Fatalf("ScoreRun: %v", err)
	}

	if capturedHeaders.Get("x-api-key") != "sk-test-key-123" {
		t.Errorf("expected x-api-key=sk-test-key-123, got %s", capturedHeaders.Get("x-api-key"))
	}
	if capturedHeaders.Get("anthropic-version") != "2023-06-01" {
		t.Errorf("expected anthropic-version=2023-06-01, got %s", capturedHeaders.Get("anthropic-version"))
	}
}

// redirectClient is an HTTPClient that redirects all requests to a test server
// while preserving original headers.
type redirectClient struct {
	server *httptest.Server
}

func (c *redirectClient) Do(req *http.Request) (*http.Response, error) {
	// Redirect request to test server.
	redirectURL := c.server.URL + req.URL.Path
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, redirectURL, req.Body)
	if err != nil {
		return nil, err
	}
	newReq.Header = req.Header
	return c.server.Client().Do(newReq)
}
