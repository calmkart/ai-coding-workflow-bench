package pool

import "testing"

func TestNewPool(t *testing.T) {
	p := NewPool(func() []byte {
		return make([]byte, 1024)
	})

	buf := p.Get()
	if len(buf) != 1024 {
		t.Fatalf("expected 1024, got %d", len(buf))
	}
}
