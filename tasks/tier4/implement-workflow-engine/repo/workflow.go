package workflow

import (
	"context"
	"errors"
	"time"
)

// Step defines a single step in a workflow.
type Step struct {
	Name      string
	Command   string
	DependsOn string // name of step this depends on
	Parallel  bool
	Timeout   int // seconds, 0 = no timeout
}

// Workflow is a named collection of steps.
type Workflow struct {
	Name  string
	Steps []Step
}

// StepResult holds the result of a single step execution.
type StepResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"` // "success", "failed", "skipped", "timeout"
	Output   string        `json:"output"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// WorkflowResult holds the result of an entire workflow execution.
type WorkflowResult struct {
	Name    string       `json:"name"`
	Success bool         `json:"success"`
	Steps   []StepResult `json:"steps"`
}

// StepRunner executes a step command and returns output.
// This interface allows injecting mock runners for testing.
type StepRunner interface {
	Run(ctx context.Context, command string) (output string, err error)
}

// Errors
var (
	ErrParse       = errors.New("parse error")
	ErrStepTimeout = errors.New("step timeout")
	ErrStepFailed  = errors.New("step failed")
)

// ParseWorkflow parses a simplified YAML-like workflow definition.
// Format:
//   name: workflow-name
//   steps:
//     - name: step-name
//       command: some command
//       depends_on: other-step
//       parallel: true
//       timeout: 60
//
// TODO: Implement parser. This is a simplified YAML parser, not full YAML.
func ParseWorkflow(data []byte) (*Workflow, error) {
	return nil, ErrParse
}

// Execute runs all workflow steps in order, respecting dependencies and parallelism.
// TODO: Implement execution engine.
func (w *Workflow) Execute(ctx context.Context, runner StepRunner) (*WorkflowResult, error) {
	return nil, errors.New("not implemented")
}
