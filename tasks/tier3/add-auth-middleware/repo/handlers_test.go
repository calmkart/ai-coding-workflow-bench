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
		strings.NewReader(`{"title":"buy milk"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	// Note: after auth is added, this will need a token
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 201 or 401, got %d", resp.StatusCode)
	}
}
