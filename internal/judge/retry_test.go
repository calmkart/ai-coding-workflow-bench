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

// TestScoreRun_Retry429 verifies that scoreRunWithClient retries on 429 status.
func TestScoreRun_Retry429(t *testing.T) {
	var attempts int32

	rubric := mockRubricResponse()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 2 {
			// First two attempts return 429.
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			w.Write([]byte(`{"error": {"type": "rate_limit_error", "message": "rate limited"}}`))
			return
		}
		// Third attempt succeeds.
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
	result, err := scoreRunWithClient(
		context.Background(),
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
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

// TestScoreRun_Retry529 verifies that scoreRunWithClient retries on 529 (overloaded).
func TestScoreRun_Retry529(t *testing.T) {
	var attempts int32

	rubric := mockRubricResponse()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.WriteHeader(529)
			w.Write([]byte(`{"error": {"type": "overloaded_error", "message": "overloaded"}}`))
			return
		}
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
	result, err := scoreRunWithClient(
		context.Background(),
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
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

// TestScoreRun_RetryExhausted verifies that all 3 retries exhausted returns error on 429.
func TestScoreRun_RetryExhausted(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(429)
		w.Write([]byte(`{"error": {"type": "rate_limit_error", "message": "rate limited"}}`))
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
		t.Fatal("expected error after exhausted retries")
	}
	// Should have attempted 3 times.
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
}

// TestScoreRun_RetryContextCancelled verifies that retry respects context cancellation.
func TestScoreRun_RetryContextCancelled(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.Header().Set("Retry-After", "60") // long wait
		w.WriteHeader(429)
		w.Write([]byte(`{"error": {"type": "rate_limit_error", "message": "rate limited"}}`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	client := &redirectClient{server: server}
	_, err := scoreRunWithClient(
		ctx,
		client,
		"test-api-key",
		"claude-sonnet-4-20250514",
		"plan", "code", "diff",
	)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

// TestScoreRun_NoRetryOn500 verifies that non-429/529 errors are NOT retried.
func TestScoreRun_NoRetryOn500(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(500)
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
		t.Fatal("expected error for 500 status")
	}
	// Should NOT have retried -- only 1 attempt.
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("expected 1 attempt (no retry for 500), got %d", atomic.LoadInt32(&attempts))
	}
}
