package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	handler := setupRouter()
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

// setupRouter returns the configured HTTP handler.
// This function signature is part of the API contract and must not be changed.
func setupRouter() http.Handler {
	mux := http.NewServeMux()
	store := NewTodoStore()

	mux.HandleFunc("GET /todos", store.listTodos)
	mux.HandleFunc("POST /todos", store.createTodo)
	mux.HandleFunc("GET /todos/{id}", store.getTodo)
	mux.HandleFunc("DELETE /todos/{id}", store.deleteTodo)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return mux
}
