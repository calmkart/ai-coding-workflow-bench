package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBasicCRUD(t *testing.T) {
	mux := setupRouter()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Create
	body := `{"title":"test","done":false}`
	resp, err := http.Post(ts.URL+"/todos", "application/json",
		strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var created map[string]any
	json.NewDecoder(resp.Body).Decode(&created)

	// List
	resp2, err := http.Get(ts.URL + "/todos")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", resp2.StatusCode)
	}
}
