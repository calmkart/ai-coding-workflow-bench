package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	handler := setupRouter()
	addr := fmt.Sprintf(":%d", 8080) // hardcoded port
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
