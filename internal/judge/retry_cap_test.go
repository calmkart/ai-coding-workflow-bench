package judge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// TestScoreRun_RetryAfterCapped verifies that Retry-After values > 30 seconds
// are capped at 30 seconds (P10).
func TestScoreRun_RetryAfterCapped(t *testing.T) {
	var attempts int32
	var firstRetryAt time.Time

	rubric := mockRubricResponse()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			// First attempt returns 429 with a very long Retry-After.
			w.Header().Set("Retry-After", "300") // 5 minutes
			w.WriteHeader(429)
			w.Write([]byte(`{"error": {"type": "rate_limit_error", "message": "rate limited"}}`))
			firstRetryAt = time.Now()
			return
		}
		// Second attempt succeeds.
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

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	result, err := scoreRunWithClient(
		ctx,
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"plan", "code", "diff",
	)
	if err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Verify the retry happened within 35 seconds (capped at 30s + some slack).
	retryDuration := time.Since(firstRetryAt)
	if retryDuration > 35*time.Second {
		t.Errorf("retry took %v, expected <= 35s (capped at 30s)", retryDuration)
	}

	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}
