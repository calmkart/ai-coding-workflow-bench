package mapreduce

import "testing"

func TestNewJob(t *testing.T) {
	mapper := MapperFunc[string, string, int](func(s string) []KeyValue[string, int] {
		return nil
	})
	reducer := ReducerFunc[string, int, int](func(k string, vs []int) int {
		return 0
	})
	job := NewJob[string, string, int, int](mapper, reducer, 4)
	if job == nil {
		t.Fatal("job should not be nil")
	}
}
