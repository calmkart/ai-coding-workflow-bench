package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDefaultConfig_HasPricing verifies that DefaultConfig sets default pricing.
func TestDefaultConfig_HasPricing(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Judge.InputPricePerMTok != 3.0 {
		t.Errorf("expected default InputPricePerMTok=3.0, got %v", cfg.Judge.InputPricePerMTok)
	}
	if cfg.Judge.OutputPricePerMTok != 15.0 {
		t.Errorf("expected default OutputPricePerMTok=15.0, got %v", cfg.Judge.OutputPricePerMTok)
	}
}

// TestLoad_PricingFromYAML verifies that custom pricing is loaded from YAML.
func TestLoad_PricingFromYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "bench.yaml")
	yamlContent := `
judge:
  enabled: true
  model: "claude-sonnet-4-20250514"
  input_price_per_mtok: 1.5
  output_price_per_mtok: 7.5
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Judge.InputPricePerMTok != 1.5 {
		t.Errorf("expected InputPricePerMTok=1.5, got %v", cfg.Judge.InputPricePerMTok)
	}
	if cfg.Judge.OutputPricePerMTok != 7.5 {
		t.Errorf("expected OutputPricePerMTok=7.5, got %v", cfg.Judge.OutputPricePerMTok)
	}
}

// TestLoad_PricingDefaultsWhenNotSet verifies that pricing defaults to 3.0/15.0
// when not set in YAML.
func TestLoad_PricingDefaultsWhenNotSet(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "bench.yaml")
	yamlContent := `
judge:
  enabled: true
  model: "claude-sonnet-4-20250514"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Judge.InputPricePerMTok != 3.0 {
		t.Errorf("expected InputPricePerMTok=3.0 (default), got %v", cfg.Judge.InputPricePerMTok)
	}
	if cfg.Judge.OutputPricePerMTok != 15.0 {
		t.Errorf("expected OutputPricePerMTok=15.0 (default), got %v", cfg.Judge.OutputPricePerMTok)
	}
}
