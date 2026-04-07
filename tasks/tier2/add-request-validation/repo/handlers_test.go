package main

import (
	"encoding/json"
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

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestListTodos(t *testing.T) {
	mux := setupRouter()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Create one first
	http.Post(ts.URL+"/todos", "application/json",
		strings.NewReader(`{"title":"test","done":false}`))

	resp, err := http.Get(ts.URL + "/todos")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var todos []Todo
	json.NewDecoder(resp.Body).Decode(&todos)
	if len(todos) != 1 {
		t.Fatalf("expected 1, got %d", len(todos))
	}
}
