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

// listTodos handles GET /todos.
// PROBLEM: No caching - every request scans the full list.
func (s *TodoStore) listTodos(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 10

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			http.Error(w, `{"error":"invalid page"}`, http.StatusBadRequest)
			return
		}
		page = p
	}
	if pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil || ps < 1 {
			http.Error(w, `{"error":"invalid page_size"}`, http.StatusBadRequest)
			return
		}
		pageSize = ps
	}

	s.mu.RLock()
	total := len(s.todos)
	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset > total {
		offset = total
	}
	if end > total {
		end = total
	}
	result := make([]Todo, len(s.todos[offset:end]))
	copy(result, s.todos[offset:end])
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"items": result,
		"total": total,
		"page":  page,
	})
}

// createTodo handles POST /todos.
func (s *TodoStore) createTodo(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if input.Title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
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

	// PROBLEM: No cache invalidation after write

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

// getTodo handles GET /todos/{id}.
// PROBLEM: No caching - every request does linear scan.
func (s *TodoStore) getTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
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

	http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
}

// updateTodo handles PUT /todos/{id}.
func (s *TodoStore) updateTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	var input struct {
		Title string `json:"title"`
		Done  *bool  `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
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
			// PROBLEM: No cache invalidation
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(s.todos[i])
			return
		}
	}

	http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
}

// deleteTodo handles DELETE /todos/{id}.
func (s *TodoStore) deleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.todos {
		if t.ID == id {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			// PROBLEM: No cache invalidation
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
}
