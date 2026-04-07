package mapreduce

import "context"

// KeyValue represents a key-value pair emitted by Mapper.
type KeyValue[K, V any] struct {
	Key   K
	Value V
}

// Mapper transforms an input into zero or more key-value pairs.
type Mapper[In, K, V any] interface {
	Map(input In) []KeyValue[K, V]
}

// Reducer aggregates values for a key into a result.
type Reducer[K, V, Out any] interface {
	Reduce(key K, values []V) Out
}

// MapperFunc wraps a function as a Mapper.
type MapperFunc[In, K, V any] func(In) []KeyValue[K, V]

func (f MapperFunc[In, K, V]) Map(input In) []KeyValue[K, V] {
	return f(input)
}

// ReducerFunc wraps a function as a Reducer.
type ReducerFunc[K, V, Out any] func(K, []V) Out

func (f ReducerFunc[K, V, Out]) Reduce(key K, values []V) Out {
	return f(key, values)
}

// MapReduceJob coordinates the MapReduce execution.
// TODO: Implement Run with parallel Map and Reduce phases.
type MapReduceJob[In, K comparable, V, Out any] struct {
	mapper  Mapper[In, K, V]
	reducer Reducer[K, V, Out]
	workers int
}

// NewJob creates a new MapReduce job.
func NewJob[In, K comparable, V, Out any](
	mapper Mapper[In, K, V],
	reducer Reducer[K, V, Out],
	workers int,
) *MapReduceJob[In, K, V, Out] {
	if workers < 1 {
		workers = 1
	}
	return &MapReduceJob[In, K, V, Out]{
		mapper:  mapper,
		reducer: reducer,
		workers: workers,
	}
}

// Run executes the MapReduce job:
// 1. Map phase: parallel map over inputs
// 2. Shuffle phase: group by key
// 3. Reduce phase: parallel reduce per key
// TODO: Implement.
func (j *MapReduceJob[In, K, V, Out]) Run(ctx context.Context, input []In) (map[K]Out, error) {
	return nil, nil
}
