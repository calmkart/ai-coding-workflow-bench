package workflow

import "testing"

func TestParseStub(t *testing.T) {
	_, err := ParseWorkflow([]byte("name: test"))
	if err == nil {
		// Will pass once implemented
	}
	_ = err
}
