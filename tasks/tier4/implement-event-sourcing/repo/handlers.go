package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
)

// Todo is directly mutated — no event history.
type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// TodoStore directly modifies state. No events, no history, no audit.
type TodoStore struct {
	mu     sync.RWMutex
	todos  map[int]*Todo
	nextID atomic.Int64
}

func NewTodoStore() *TodoStore {
	s := &TodoStore{
		todos: make(map[int]*Todo),
	}
	s.nextID.Store(1)
	return s
}

func (s *TodoStore) listTodos(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Todo, 0, len(s.todos))
	for _, t := range s.todos {
		result = append(result, *t)
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

	// Direct mutation — no event recorded
	s.mu.Lock()
	s.todos[todo.ID] = &todo
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

	t, ok := s.todos[id]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
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

	// Direct mutation — no event, no version check
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.todos[id]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	t.Title = update.Title
	t.Done = update.Done

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (s *TodoStore) deleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Direct deletion — no tombstone event
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.todos[id]; !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	delete(s.todos, id)
	w.WriteHeader(http.StatusNoContent)
}
