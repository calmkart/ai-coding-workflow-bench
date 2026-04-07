package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Everything is in one file. This is the God Handler anti-pattern.
// All concerns mixed together: routing, business logic, data storage,
// error handling, validation, configuration, and formatting.

// Hardcoded configuration values scattered throughout
const maxPageSize = 50
const defaultPageSize = 10
const maxTitleLength = 200

type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tag struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	TodoID int    `json:"todo_id"`
}

// Global state — no dependency injection
var (
	mu       sync.RWMutex
	todos    = make(map[int]*Todo)
	tags     = make(map[int]*Tag)
	nextTodoID atomic.Int64
	nextTagID  atomic.Int64
)

func init() {
	nextTodoID.Store(1)
	nextTagID.Store(1)
}

// setupRouter configures all routes.
// This function signature is part of the API contract and must not be changed.
func setupRouter() http.Handler {
	mux := http.NewServeMux()

	// No middleware at all — no logging, no recovery, no request ID

	mux.HandleFunc("GET /todos", handleListTodos)
	mux.HandleFunc("POST /todos", handleCreateTodo)
	mux.HandleFunc("GET /todos/{id}", handleGetTodo)
	mux.HandleFunc("PUT /todos/{id}", handleUpdateTodo)
	mux.HandleFunc("DELETE /todos/{id}", handleDeleteTodo)
	mux.HandleFunc("POST /todos/{id}/tags", handleAddTag)
	mux.HandleFunc("GET /todos/{id}/tags", handleGetTags)
	mux.HandleFunc("DELETE /tags/{id}", handleDeleteTag)
	mux.HandleFunc("GET /search", handleSearch)
	mux.HandleFunc("GET /stats", handleStats)
	mux.HandleFunc("GET /health", handleHealth)

	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// Business logic mixed with HTTP handling. No separation of concerns.
func handleListTodos(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")
	sortBy := r.URL.Query().Get("sort")
	filterDone := r.URL.Query().Get("done")

	page := 1
	pageSize := defaultPageSize

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			// Inconsistent error format — plain text
			http.Error(w, "invalid page parameter", http.StatusBadRequest)
			return
		}
		if p < 1 {
			http.Error(w, "page must be >= 1", http.StatusBadRequest)
			return
		}
		page = p
	}

	if pageSizeStr != "" {
		ps, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			http.Error(w, "invalid page_size", http.StatusBadRequest)
			return
		}
		if ps < 1 || ps > maxPageSize {
			http.Error(w, fmt.Sprintf("page_size must be 1-%d", maxPageSize), http.StatusBadRequest)
			return
		}
		pageSize = ps
	}

	mu.RLock()
	result := make([]Todo, 0, len(todos))
	for _, t := range todos {
		// Filter logic mixed with handler
		if filterDone != "" {
			isDone := filterDone == "true"
			if t.Done != isDone {
				continue
			}
		}
		result = append(result, *t)
	}
	mu.RUnlock()

	// Sort logic mixed with handler
	switch sortBy {
	case "title":
		sort.Slice(result, func(i, j int) bool {
			return result[i].Title < result[j].Title
		})
	case "priority":
		sort.Slice(result, func(i, j int) bool {
			return result[i].Priority > result[j].Priority
		})
	case "created_at":
		sort.Slice(result, func(i, j int) bool {
			return result[i].CreatedAt.Before(result[j].CreatedAt)
		})
	default:
		sort.Slice(result, func(i, j int) bool {
			return result[i].ID < result[j].ID
		})
	}

	// Pagination logic
	total := len(result)
	offset := (page - 1) * pageSize
	if offset > total {
		offset = total
	}
	end := offset + pageSize
	if end > total {
		end = total
	}
	paged := result[offset:end]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Total-Count", strconv.Itoa(total))
	json.NewEncoder(w).Encode(paged)
}

func handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		// Another inconsistent error format
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "bad json: %v", err)
		return
	}

	// Validation mixed with handler
	todo.Title = strings.TrimSpace(todo.Title)
	if todo.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if len(todo.Title) > maxTitleLength {
		http.Error(w, fmt.Sprintf("title too long (max %d)", maxTitleLength), http.StatusBadRequest)
		return
	}
	if todo.Priority < 0 || todo.Priority > 5 {
		http.Error(w, "priority must be 0-5", http.StatusBadRequest)
		return
	}

	todo.ID = int(nextTodoID.Add(1) - 1)
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = todo.CreatedAt

	mu.Lock()
	todos[todo.ID] = &todo
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func handleGetTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	mu.RLock()
	t, ok := todos[id]
	mu.RUnlock()

	if !ok {
		// Yet another error format — JSON this time
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "todo not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func handleUpdateTodo(w http.ResponseWriter, r *http.Request) {
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

	// Duplicate validation — same as create
	update.Title = strings.TrimSpace(update.Title)
	if update.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if len(update.Title) > maxTitleLength {
		http.Error(w, fmt.Sprintf("title too long (max %d)", maxTitleLength), http.StatusBadRequest)
		return
	}
	if update.Priority < 0 || update.Priority > 5 {
		http.Error(w, "priority must be 0-5", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	t, ok := todos[id]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	t.Title = update.Title
	t.Done = update.Done
	t.Priority = update.Priority
	t.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ok := todos[id]; !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	delete(todos, id)

	// Also delete associated tags — scattered logic
	for tagID, tag := range tags {
		if tag.TodoID == id {
			delete(tags, tagID)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleAddTag(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	todoID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	mu.RLock()
	_, ok := todos[todoID]
	mu.RUnlock()
	if !ok {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	var tag Tag
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	tag.Name = strings.TrimSpace(tag.Name)
	if tag.Name == "" {
		http.Error(w, "tag name required", http.StatusBadRequest)
		return
	}

	tag.ID = int(nextTagID.Add(1) - 1)
	tag.TodoID = todoID

	mu.Lock()
	tags[tag.ID] = &tag
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tag)
}

func handleGetTags(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	todoID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	mu.RLock()
	_, ok := todos[todoID]
	mu.RUnlock()
	if !ok {
		http.Error(w, "todo not found", http.StatusNotFound)
		return
	}

	mu.RLock()
	result := make([]Tag, 0)
	for _, tag := range tags {
		if tag.TodoID == todoID {
			result = append(result, *tag)
		}
	}
	mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleDeleteTag(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ok := tags[id]; !ok {
		http.Error(w, "tag not found", http.StatusNotFound)
		return
	}

	delete(tags, id)
	w.WriteHeader(http.StatusNoContent)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "q parameter required", http.StatusBadRequest)
		return
	}

	q = strings.ToLower(q)

	mu.RLock()
	defer mu.RUnlock()

	result := make([]Todo, 0)
	for _, t := range todos {
		if strings.Contains(strings.ToLower(t.Title), q) {
			result = append(result, *t)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	total := len(todos)
	done := 0
	totalPriority := 0

	for _, t := range todos {
		if t.Done {
			done++
		}
		totalPriority += t.Priority
	}

	avgPriority := 0.0
	if total > 0 {
		avgPriority = float64(totalPriority) / float64(total)
	}

	stats := map[string]any{
		"total":            total,
		"done":             done,
		"pending":          total - done,
		"average_priority": avgPriority,
		"total_tags":       len(tags),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
