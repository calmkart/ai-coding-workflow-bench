package coordinator

import (
	"context"
	"testing"
)

func TestBasic(t *testing.T) {
	c := NewCoordinator(2, func(task Task) Result {
		return Result{TaskID: task.ID, Output: "done"}
	})
	c.Start(context.Background())
	c.Submit(Task{ID: 1, Payload: "test"})
	c.Wait()
}
