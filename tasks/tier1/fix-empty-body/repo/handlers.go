package main

import (
	"encoding/json"
	"io"
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
}

func NewTodoStore() *TodoStore {
	s := &TodoStore{}
	s.nextID.Store(1)
	return s
}

func (s *TodoStore) listTodos(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.todos)
}

func (s *TodoStore) createTodo(w http.ResponseWriter, r *http.Request) {
	// BUG: reads body into bytes, then tries to use it without checking for empty
	body, _ := io.ReadAll(r.Body)

	var todo Todo
	// BUG: when body is empty, json.Unmarshal returns "unexpected end of JSON input"
	// but the error is silently ignored, leaving todo as zero-value
	// Then we access todo.Title which could cause issues downstream
	json.Unmarshal(body, &todo)

	// BUG: no validation - empty title is silently accepted
	// and if body was completely empty, we still create a todo
	if len(body) == 0 {
		// Attempt to read from nil-like source triggers issue
		// This simulates a panic scenario - accessing uninitialized data
		panic("unexpected nil body")
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
