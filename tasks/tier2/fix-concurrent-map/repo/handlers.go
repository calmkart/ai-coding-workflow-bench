package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type TodoStore struct {
	mu     sync.RWMutex
	todos  []Todo
	nextID atomic.Int64
	// BUG: stats map is accessed concurrently without synchronization
	stats map[string]int
}

func NewTodoStore() *TodoStore {
	s := &TodoStore{
		stats: make(map[string]int),
	}
	s.nextID.Store(1)
	return s
}

func (s *TodoStore) listTodos(w http.ResponseWriter, r *http.Request) {
	// BUG: concurrent map write without lock
	s.stats["list"]++

	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.todos)
}

func (s *TodoStore) createTodo(w http.ResponseWriter, r *http.Request) {
	// BUG: concurrent map write without lock
	s.stats["create"]++

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	todo.ID = int(s.nextID.Add(1) - 1)

	s.mu.Lock()
	s.todos = append(s.todos, todo)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func (s *TodoStore) getTodo(w http.ResponseWriter, r *http.Request) {
	// BUG: concurrent map write without lock
	s.stats["get"]++

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.todos {
		if t.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(t)
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (s *TodoStore) deleteTodo(w http.ResponseWriter, r *http.Request) {
	// BUG: concurrent map write without lock
	s.stats["delete"]++

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.todos {
		if t.ID == id {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (s *TodoStore) getStats(w http.ResponseWriter, r *http.Request) {
	// BUG: concurrent map read without lock
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.stats)
}
