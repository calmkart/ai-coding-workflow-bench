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
}

func NewTodoStore() *TodoStore {
	s := &TodoStore{}
	s.nextID.Store(1)
	return s
}

// SMELL: Inconsistent error response formats throughout all handlers.
// Some use plain text, some use JSON strings, no error codes.

func (s *TodoStore) listTodos(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	result := make([]Todo, len(s.todos))
	copy(result, s.todos)
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *TodoStore) createTodo(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// SMELL: Plain text error, no code
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if input.Title == "" {
		// SMELL: Different format from above
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if len(input.Title) > 200 {
		http.Error(w, "title too long (max 200)", http.StatusBadRequest)
		return
	}

	todo := Todo{
		ID:    int(s.nextID.Add(1) - 1),
		Title: input.Title,
		Done:  false,
	}

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
		// SMELL: Yet another error format
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

	// SMELL: Plain text "not found"
	http.Error(w, "not found", http.StatusNotFound)
}

func (s *TodoStore) updateTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var input struct {
		Title string `json:"title"`
		Done  *bool  `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.todos {
		if t.ID == id {
			if input.Title != "" {
				s.todos[i].Title = input.Title
			}
			if input.Done != nil {
				s.todos[i].Done = *input.Done
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(s.todos[i])
			return
		}
	}

	http.Error(w, "todo not found", http.StatusNotFound)
}

func (s *TodoStore) deleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
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

	http.Error(w, "resource not found", http.StatusNotFound)
}
