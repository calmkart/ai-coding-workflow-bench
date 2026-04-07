package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"
)

// CronExpr represents a parsed cron expression.
// Simplified format: "second minute hour"
// TODO: Implement Next and Matches.
type CronExpr struct {
	Seconds []int // matched seconds (0-59)
	Minutes []int // matched minutes (0-59)
	Hours   []int // matched hours (0-23)
	Raw     string
}

// Job represents a scheduled task.
type Job struct {
	ID       string
	Name     string
	CronExpr *CronExpr
	Fn       func()
	Active   bool
}

// Scheduler manages scheduled jobs.
// TODO: Implement Schedule, Cancel, Start, ListJobs.
type Scheduler struct {
	mu   sync.RWMutex
	jobs map[string]*Job
	nextID int
}

// Errors
var (
	ErrInvalidCron = errors.New("invalid cron expression")
	ErrJobNotFound = errors.New("job not found")
)

// ParseCron parses a simplified cron expression: "second minute hour"
// Each field supports: *, N, N1,N2, */N
// TODO: Implement.
func ParseCron(expr string) (*CronExpr, error) {
	return nil, ErrInvalidCron
}

// Next returns the next time after 'from' that matches this cron expression.
// TODO: Implement.
func (c *CronExpr) Next(from time.Time) time.Time {
	return time.Time{}
}

// Matches returns true if the given time matches this cron expression.
// TODO: Implement.
func (c *CronExpr) Matches(t time.Time) bool {
	return false
}

// NewScheduler creates a new scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs: make(map[string]*Job),
	}
}

// Schedule adds a new job. Returns the job ID.
// TODO: Parse cron, create job, add to scheduler.
func (s *Scheduler) Schedule(name string, cron string, fn func()) (string, error) {
	return "", ErrInvalidCron
}

// Cancel removes a job by ID.
// TODO: Implement.
func (s *Scheduler) Cancel(id string) error {
	return ErrJobNotFound
}

// Start runs the scheduler loop. Blocks until ctx is done.
// Checks every second for matching jobs.
// TODO: Implement.
func (s *Scheduler) Start(ctx context.Context) {
	// stub
}

// ListJobs returns all registered jobs.
// TODO: Implement.
func (s *Scheduler) ListJobs() []Job {
	return nil
}
