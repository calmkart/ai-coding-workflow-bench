package result

// Package result provides a generic Result type.
// TODO: Implement Result[T] with Map, FlatMap, Unwrap, UnwrapOr.

import "errors"

// Result represents a value or an error.
// BUG: Stub - needs Map, FlatMap, Unwrap, UnwrapOr methods.
type Result[T any] struct {
	Value T
	Err   error
}

// Ok creates a successful Result.
func Ok[T any](v T) Result[T] {
	return Result[T]{Value: v}
}

// Fail creates a failed Result.
func Fail[T any](err error) Result[T] {
	if err == nil {
		err = errors.New("unknown error")
	}
	return Result[T]{Err: err}
}

// IsOk returns true if the Result is successful.
func (r Result[T]) IsOk() bool {
	return r.Err == nil
}

// IsErr returns true if the Result is an error.
func (r Result[T]) IsErr() bool {
	return r.Err != nil
}
