package judge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// mockPairwiseResponse creates a pairwise JSON response with the given winner.
func mockPairwiseResponse(winner string) pairwiseJSON {
	return pairwiseJSON{
		OverallWinner: winner,
		Reasoning:     "Implementation " + winner + " is better overall.",
		Dimensions: map[string]string{
			"correctness":     winner,
			"readability":     winner,
			"simplicity":      "Tie",
			"robustness":      winner,
			"minimality":      "Tie",
			"maintainability": winner,
		},
	}
}

// newMockPairwiseServer creates a test server that returns pairwise responses.
// On even calls (1st, 3rd, ...) it returns round1Winner, on odd calls round2Winner.
func newMockPairwiseServer(t *testing.T, round1Winner, round2Winner string) *httptest.Server {
	t.Helper()
	var callCount int32

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)

		winner := round1Winner
		if n%2 == 0 {
			winner = round2Winner
		}

		pw := mockPairwiseResponse(winner)
		pwBytes, _ := json.Marshal(pw)
		resp := anthropicResponse{
			Content: []anthropicContentBlock{
				{Type: "text", Text: string(pwBytes)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestPairwiseCompare_ConsistentWinnerA(t *testing.T) {
	// Round 1: A=left wins -> "A"
	// Round 2: A=right, B=left -> left wins -> response says "B" (because B=left in round 2)
	server := newMockPairwiseServer(t, "A", "B")
	defer server.Close()

	client := &redirectClient{server: server}
	result, err := pairwiseCompareWithClient(
		context.Background(),
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"Implement a REST endpoint",
		"+ func handlerA() {}",
		"+ func handlerB() {}",
	)
	if err != nil {
		t.Fatalf("PairwiseCompare: %v", err)
	}

	if result.Winner != "left" {
		t.Errorf("expected winner 'left', got %q", result.Winner)
	}
	if !result.PositionConsistent {
		t.Error("expected position consistent")
	}
	if result.Reasoning == "" {
		t.Error("expected non-empty reasoning")
	}
	if len(result.Dimensions) == 0 {
		t.Error("expected non-empty dimensions")
	}
}

func TestPairwiseCompare_ConsistentWinnerB(t *testing.T) {
	// Round 1: B=right wins -> "B"
	// Round 2: A=right, B=left -> right wins -> response says "A" (because A=right in round 2)
	server := newMockPairwiseServer(t, "B", "A")
	defer server.Close()

	client := &redirectClient{server: server}
	result, err := pairwiseCompareWithClient(
		context.Background(),
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"Implement a REST endpoint",
		"+ func handlerA() {}",
		"+ func handlerB() {}",
	)
	if err != nil {
		t.Fatalf("PairwiseCompare: %v", err)
	}

	if result.Winner != "right" {
		t.Errorf("expected winner 'right', got %q", result.Winner)
	}
	if !result.PositionConsistent {
		t.Error("expected position consistent")
	}
}

func TestPairwiseCompare_ConsistentTie(t *testing.T) {
	server := newMockPairwiseServer(t, "Tie", "Tie")
	defer server.Close()

	client := &redirectClient{server: server}
	result, err := pairwiseCompareWithClient(
		context.Background(),
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"Implement a REST endpoint",
		"+ func handlerA() {}",
		"+ func handlerB() {}",
	)
	if err != nil {
		t.Fatalf("PairwiseCompare: %v", err)
	}

	if result.Winner != "tie" {
		t.Errorf("expected winner 'tie', got %q", result.Winner)
	}
	if !result.PositionConsistent {
		t.Error("expected position consistent for tie")
	}
}

func TestPairwiseCompare_Inconsistent(t *testing.T) {
	// Round 1: A wins -> left
	// Round 2: A wins -> right (because A=right in round 2)
	// These disagree: left vs right -> inconsistent -> tie
	server := newMockPairwiseServer(t, "A", "A")
	defer server.Close()

	client := &redirectClient{server: server}
	result, err := pairwiseCompareWithClient(
		context.Background(),
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"Implement a REST endpoint",
		"+ func handlerA() {}",
		"+ func handlerB() {}",
	)
	if err != nil {
		t.Fatalf("PairwiseCompare: %v", err)
	}

	if result.Winner != "tie" {
		t.Errorf("expected winner 'tie' for inconsistent, got %q", result.Winner)
	}
	if result.PositionConsistent {
		t.Error("expected position inconsistent")
	}
	if !strings.Contains(result.Reasoning, "inconsistent") {
		t.Errorf("expected 'inconsistent' in reasoning, got %q", result.Reasoning)
	}
}

func TestPairwiseCompare_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"type": "server_error", "message": "internal error"}}`))
	}))
	defer server.Close()

	client := &redirectClient{server: server}
	_, err := pairwiseCompareWithClient(
		context.Background(),
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"plan", "diffA", "diffB",
	)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("expected 'status 500' in error, got: %v", err)
	}
}

func TestPairwiseCompare_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	client := &redirectClient{server: server}
	_, err := pairwiseCompareWithClient(
		ctx,
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"plan", "diffA", "diffB",
	)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestRenderPairwisePrompt(t *testing.T) {
	prompt, err := renderPairwisePrompt("Test plan", "diff A content", "diff B content")
	if err != nil {
		t.Fatalf("renderPairwisePrompt: %v", err)
	}

	if !strings.Contains(prompt, "Test plan") {
		t.Error("expected plan content in prompt")
	}
	if !strings.Contains(prompt, "diff A content") {
		t.Error("expected diff A in prompt")
	}
	if !strings.Contains(prompt, "diff B content") {
		t.Error("expected diff B in prompt")
	}
	if !strings.Contains(prompt, "Implementation A") {
		t.Error("expected 'Implementation A' in prompt")
	}
	if !strings.Contains(prompt, "Implementation B") {
		t.Error("expected 'Implementation B' in prompt")
	}
}

func TestNormalizeWinner(t *testing.T) {
	tests := []struct {
		raw     string
		flipped bool
		want    string
	}{
		{"A", false, "left"},
		{"B", false, "right"},
		{"Tie", false, "tie"},
		{"A", true, "right"},
		{"B", true, "left"},
		{"Tie", true, "tie"},
		{"unknown", false, "tie"},
	}
	for _, tt := range tests {
		got := normalizeWinner(tt.raw, tt.flipped)
		if got != tt.want {
			t.Errorf("normalizeWinner(%q, %v) = %q, want %q", tt.raw, tt.flipped, got, tt.want)
		}
	}
}

func TestMergeDimensions(t *testing.T) {
	d1 := map[string]string{
		"correctness": "left",
		"readability": "right",
		"simplicity":  "tie",
	}
	d2 := map[string]string{
		"correctness": "left",  // agree
		"readability": "left",  // disagree
		"simplicity":  "tie",   // agree
		"robustness":  "right", // only in d2
	}

	merged := mergeDimensions(d1, d2)

	if merged["correctness"] != "left" {
		t.Errorf("correctness: expected 'left', got %q", merged["correctness"])
	}
	if merged["readability"] != "tie" {
		t.Errorf("readability: expected 'tie' (disagreement), got %q", merged["readability"])
	}
	if merged["simplicity"] != "tie" {
		t.Errorf("simplicity: expected 'tie', got %q", merged["simplicity"])
	}
	if merged["robustness"] != "tie" {
		t.Errorf("robustness: expected 'tie' (only in d2), got %q", merged["robustness"])
	}
}
