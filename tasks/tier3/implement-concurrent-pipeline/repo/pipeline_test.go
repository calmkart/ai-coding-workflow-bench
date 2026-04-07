package concpipeline

import (
	"context"
	"testing"
)

func TestBasic(t *testing.T) {
	_, err := RunPipeline(
		context.Background(),
		func(ctx context.Context, out chan<- int) {
			out <- 1
		},
		func(x int) (int, error) { return x * 2, nil },
		func(x int) (int, error) { return x, nil },
	)
	if err != nil {
		t.Fatal(err)
	}
}
