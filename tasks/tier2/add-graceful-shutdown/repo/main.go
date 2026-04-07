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
	// BUG: No graceful shutdown - just Fatal on ListenAndServe
	log.Fatal(http.ListenAndServe(addr, handler))
}

func setupRouter() http.Handler {
	mux := http.NewServeMux()
	store := NewTodoStore()

	mux.HandleFunc("GET /todos", store.listTodos)
	mux.HandleFunc("POST /todos", store.createTodo)
	mux.HandleFunc("GET /todos/{id}", store.getTodo)
	mux.HandleFunc("DELETE /todos/{id}", store.deleteTodo)

	return mux
}
