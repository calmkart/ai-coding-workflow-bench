package servicegroup

import (
	"context"
	"fmt"
	"time"
)

// Service represents a long-running service.
type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

// ServiceGroup manages multiple services.
// TODO: Implement Add, Run, and graceful shutdown.
// Currently services are just stored but not managed properly.
type ServiceGroup struct {
	services        []Service
	shutdownTimeout time.Duration
}

// NewServiceGroup creates a new ServiceGroup.
func NewServiceGroup() *ServiceGroup {
	return &ServiceGroup{
		shutdownTimeout: 30 * time.Second,
	}
}

// Add registers a service.
func (sg *ServiceGroup) Add(svc Service) {
	sg.services = append(sg.services, svc)
}

// WithShutdownTimeout sets the shutdown timeout.
func (sg *ServiceGroup) WithShutdownTimeout(d time.Duration) *ServiceGroup {
	sg.shutdownTimeout = d
	return sg
}

// Run starts all services and blocks until shutdown.
// PROBLEM: No graceful shutdown logic.
// - Does not propagate context cancellation
// - Does not call Stop on services
// - Does not handle service failures
func (sg *ServiceGroup) Run(ctx context.Context) error {
	for _, svc := range sg.services {
		go func(s Service) {
			// PROBLEM: No error handling, no shutdown
			s.Start(ctx)
		}(svc)
	}

	// PROBLEM: Just waits on context, never stops services
	<-ctx.Done()
	fmt.Println("context cancelled, but no cleanup performed")
	return ctx.Err()
}
