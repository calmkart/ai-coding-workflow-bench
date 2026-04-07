package adapter

import (
	"testing"
)

func TestParseClaudeJSON_ValidUsage(t *testing.T) {
	stdout := []byte(`{"usage":{"input_tokens":100,"output_tokens":50},"tool_uses":5}`)
	result := &RunOutput{}
	parseClaudeJSON(stdout, result)

	if result.TokenUsage == nil {
		t.Fatal("expected token usage to be parsed")
	}
	if result.TokenUsage.InputTokens != 100 {
		t.Errorf("expected input_tokens=100, got %d", result.TokenUsage.InputTokens)
	}
	if result.TokenUsage.OutputTokens != 50 {
		t.Errorf("expected output_tokens=50, got %d", result.TokenUsage.OutputTokens)
	}
	if result.ToolUses != 5 {
		t.Errorf("expected tool_uses=5, got %d", result.ToolUses)
	}
}

func TestParseClaudeJSON_NoUsage(t *testing.T) {
	stdout := []byte(`{"result":"ok"}`)
	result := &RunOutput{}
	parseClaudeJSON(stdout, result)

	if result.TokenUsage != nil {
		t.Errorf("expected nil token usage for JSON without usage, got %+v", result.TokenUsage)
	}
	if result.ToolUses != 0 {
		t.Errorf("expected tool_uses=0, got %d", result.ToolUses)
	}
}

func TestParseClaudeJSON_InvalidJSON(t *testing.T) {
	stdout := []byte(`not json at all`)
	result := &RunOutput{}
	parseClaudeJSON(stdout, result)

	if result.TokenUsage != nil {
		t.Errorf("expected nil token usage for invalid JSON, got %+v", result.TokenUsage)
	}
}

func TestParseClaudeJSON_EmptyInput(t *testing.T) {
	result := &RunOutput{}
	parseClaudeJSON([]byte{}, result)

	if result.TokenUsage != nil {
		t.Errorf("expected nil token usage for empty input, got %+v", result.TokenUsage)
	}
}

func TestJsonInt_Float64(t *testing.T) {
	if got := jsonInt(float64(42)); got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
}

func TestJsonInt_NonNumeric(t *testing.T) {
	if got := jsonInt("not a number"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestJsonInt_Nil(t *testing.T) {
	if got := jsonInt(nil); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

// TestGetWithCustomWorkflowName verifies that a custom workflow name (like "my-workflow")
// must be resolved to the adapter type ("custom") before calling Get. This test documents
// the correct usage pattern that runner.go implements.
func TestGetWithCustomWorkflowName_Fails(t *testing.T) {
	// Calling Get with a custom workflow name (not a registry key) should fail.
	_, err := Get("my-workflow", map[string]any{
		"adapter":       "custom",
		"entry_command": "echo test",
	})
	if err == nil {
		t.Fatal("expected error for unknown adapter name 'my-workflow'")
	}
}

// TestResolveAdapterName_CustomWorkflow verifies the adapter name resolution logic
// that runner.go uses: extract the "adapter" field from WorkflowCfg to resolve
// custom workflow names to their adapter type.
func TestResolveAdapterName_CustomWorkflow(t *testing.T) {
	workflowCfg := map[string]any{
		"adapter":       "custom",
		"entry_command": "echo test",
		"name":          "my-workflow",
	}

	// Simulate the resolution logic from runner.go.
	adapterName := "my-workflow" // This is what cfg.Workflow would be.
	if a, ok := workflowCfg["adapter"].(string); ok && a != "" {
		adapterName = a
	}

	adpt, err := Get(adapterName, workflowCfg)
	if err != nil {
		t.Fatalf("Get(%q) failed: %v", adapterName, err)
	}
	if adpt.Name() != "my-workflow" {
		t.Errorf("expected adapter name 'my-workflow', got %q", adpt.Name())
	}
}

// TestResolveAdapterName_BuiltinWorkflow verifies that built-in workflow names
// (like "vanilla") work without needing adapter field resolution.
func TestResolveAdapterName_BuiltinWorkflow(t *testing.T) {
	// For built-in workflows, WorkflowCfg may be nil or may have adapter matching the name.
	adapterName := "vanilla"
	var workflowCfg map[string]any // nil

	if workflowCfg != nil {
		if a, ok := workflowCfg["adapter"].(string); ok && a != "" {
			adapterName = a
		}
	}

	adpt, err := Get(adapterName, workflowCfg)
	if err != nil {
		t.Fatalf("Get(%q) failed: %v", adapterName, err)
	}
	if adpt.Name() != "vanilla" {
		t.Errorf("expected adapter name 'vanilla', got %q", adpt.Name())
	}
}
