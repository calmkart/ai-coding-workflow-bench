package pipeline

import "testing"

func TestUnsafePipelineBasic(t *testing.T) {
	p := NewUnsafePipeline()
	p.AddStage("double", func(v interface{}) (interface{}, error) {
		return v.(int) * 2, nil
	})

	result, err := p.Execute(5)
	if err != nil {
		t.Fatal(err)
	}
	if result.(int) != 10 {
		t.Fatalf("expected 10, got %v", result)
	}
}
