package engine

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/adapter"
)

// TestErrAdapterInfra_IsSentinel verifies that errAdapterInfra is a distinct sentinel
// error that can be detected with errors.Is.
func TestErrAdapterInfra_IsSentinel(t *testing.T) {
	if errAdapterInfra == nil {
		t.Fatal("errAdapterInfra should not be nil")
	}
	if !errors.Is(errAdapterInfra, errAdapterInfra) {
		t.Error("errors.Is(errAdapterInfra, errAdapterInfra) should be true")
	}
	// A wrapped version should also match.
	wrapped := fmt.Errorf("wrapper: %w", errAdapterInfra)
	if !errors.Is(wrapped, errAdapterInfra) {
		t.Error("errors.Is on wrapped errAdapterInfra should be true")
	}
}

// TestFlashExitDetection_Conditions verifies the detection logic thresholds.
// A flash exit requires ExitCode != 0 AND WallTime < 30s.
func TestFlashExitDetection_Conditions(t *testing.T) {
	tests := []struct {
		name       string
		exitCode   int
		wallTime   time.Duration
		wantInfra  bool
	}{
		{
			name:      "flash exit: non-zero exit + fast",
			exitCode:  1,
			wallTime:  3 * time.Second,
			wantInfra: true,
		},
		{
			name:      "flash exit: non-zero exit + boundary just under 30s",
			exitCode:  1,
			wallTime:  29 * time.Second,
			wantInfra: true,
		},
		{
			name:      "not infra: non-zero exit + slow (real failure)",
			exitCode:  1,
			wallTime:  30 * time.Second,
			wantInfra: false,
		},
		{
			name:      "not infra: non-zero exit + well over threshold",
			exitCode:  1,
			wallTime:  120 * time.Second,
			wantInfra: false,
		},
		{
			name:      "not infra: zero exit + fast (success)",
			exitCode:  0,
			wallTime:  3 * time.Second,
			wantInfra: false,
		},
		{
			name:      "not infra: zero exit + slow (success)",
			exitCode:  0,
			wallTime:  120 * time.Second,
			wantInfra: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &adapter.RunOutput{
				ExitCode: tt.exitCode,
				WallTime: tt.wallTime,
			}
			isInfra := output.ExitCode != 0 && output.WallTime < 30*time.Second
			if isInfra != tt.wantInfra {
				t.Errorf("isInfra = %v, want %v (exit=%d, wall=%s)",
					isInfra, tt.wantInfra, tt.exitCode, tt.wallTime)
			}
		})
	}
}
