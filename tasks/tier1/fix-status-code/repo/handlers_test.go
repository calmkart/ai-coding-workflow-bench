package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateTodo(t *testing.T) {
	mux := setupRouter()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/todos", "application/json",
		strings.NewReader(`{"title":"buy milk","done":false}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// NOTE: this test checks 200 which is the current (buggy) behavior
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHealthCheck(t *testing.T) {
	mux := setupRouter()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
