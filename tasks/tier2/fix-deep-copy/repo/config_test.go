package config

import "testing"

func TestNewConfig(t *testing.T) {
	c := NewConfig()
	if c.Name != "default" {
		t.Fatalf("expected 'default', got '%s'", c.Name)
	}
}

func TestCloneBasic(t *testing.T) {
	c := NewConfig()
	clone := c.Clone()
	if clone.Name != c.Name {
		t.Fatalf("expected '%s', got '%s'", c.Name, clone.Name)
	}
}
