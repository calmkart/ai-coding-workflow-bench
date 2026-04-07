package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	ProjectID int    `json:"project_id,omitempty"`
}

type Project struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"` // "pending", "active", "failed"
	TodoIDs []int  `json:"todo_ids"`
}

type TodoStore struct {
	mu        sync.RWMutex
	todos     map[int]*Todo
	projects  map[int]*Project
	nextTodoID    atomic.Int64
	nextProjectID atomic.Int64
}

func NewTodoStore() *TodoStore {
	s := &TodoStore{
		todos:    make(map[int]*Todo),
		projects: make(map[int]*Project),
	}
	s.nextTodoID.Store(1)
	s.nextProjectID.Store(1)
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
	todo.ID = int(s.nextTodoID.Add(1) - 1)
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

func (s *TodoStore) deleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.todos[id]; !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	delete(s.todos, id)
	w.WriteHeader(http.StatusNoContent)
}

// createProject does multiple steps with NO compensation/rollback.
// Step 1: Create the project record
// Step 2: Create associated todos
// Step 3: Update project status to "active"
// BUG: If step 2 or 3 fails, step 1 is NOT rolled back.
func (s *TodoStore) createProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string   `json:"name"`
		Todos []string `json:"todos"` // todo titles to create
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	// Step 1: Create project
	project := &Project{
		ID:     int(s.nextProjectID.Add(1) - 1),
		Name:   req.Name,
		Status: "pending",
	}
	s.mu.Lock()
	s.projects[project.ID] = project
	s.mu.Unlock()

	// Step 2: Create todos — NO rollback if this fails
	var todoIDs []int
	for _, title := range req.Todos {
		todo := &Todo{
			ID:        int(s.nextTodoID.Add(1) - 1),
			Title:     title,
			ProjectID: project.ID,
		}

		// Simulate potential failure: empty title
		if title == "" {
			// BUG: project already created but we just return error
			// Previous todos also remain. No rollback!
			http.Error(w, fmt.Sprintf("todo creation failed: empty title"), http.StatusBadRequest)
			return
		}

		s.mu.Lock()
		s.todos[todo.ID] = todo
		s.mu.Unlock()
		todoIDs = append(todoIDs, todo.ID)
	}

	// Step 3: Update status — if this somehow failed, partial state
	s.mu.Lock()
	project.TodoIDs = todoIDs
	project.Status = "active"
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (s *TodoStore) listProjects(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Project, 0, len(s.projects))
	for _, p := range s.projects {
		result = append(result, *p)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *TodoStore) getProject(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projects[id]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}
