package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
)

// Todo only has basic fields. No support for v2 fields (priority, tags).
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

	result := s.todos
	if result == nil {
		result = []Todo{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *TodoStore) createTodo(w http.ResponseWriter, r *http.Request) {
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

func (s *TodoStore) updateTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var update Todo
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.todos {
		if t.ID == id {
			s.todos[i].Title = update.Title
			s.todos[i].Done = update.Done
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(s.todos[i])
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
