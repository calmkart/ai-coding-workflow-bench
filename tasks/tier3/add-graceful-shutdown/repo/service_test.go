package servicegroup

import (
	"context"
	"testing"
)

type mockService struct {
	name string
}

func (m *mockService) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (m *mockService) Stop(ctx context.Context) error {
	return nil
}

func (m *mockService) Name() string {
	return m.name
}

func TestNewServiceGroup(t *testing.T) {
	sg := NewServiceGroup()
	sg.Add(&mockService{name: "test"})
	if len(sg.services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(sg.services))
	}
}
