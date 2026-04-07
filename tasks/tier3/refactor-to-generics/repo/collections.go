package collections

// SMELL: All functions use interface{} with type assertions.
// This loses compile-time type safety and can panic at runtime.
// Need to refactor to Go generics.

// Map applies fn to each element of the slice.
// PROBLEM: Uses interface{}, requires type assertion in caller.
func Map(slice []interface{}, fn func(interface{}) interface{}) []interface{} {
	if slice == nil {
		return []interface{}{}
	}
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter returns elements where fn returns true.
// PROBLEM: Uses interface{}, no type safety.
func Filter(slice []interface{}, fn func(interface{}) bool) []interface{} {
	if slice == nil {
		return []interface{}{}
	}
	var result []interface{}
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	if result == nil {
		return []interface{}{}
	}
	return result
}

// Reduce accumulates a value over the slice.
// PROBLEM: Uses interface{}, requires multiple type assertions.
func Reduce(slice []interface{}, init interface{}, fn func(interface{}, interface{}) interface{}) interface{} {
	acc := init
	for _, v := range slice {
		acc = fn(acc, v)
	}
	return acc
}

// Contains checks if elem is in slice.
// PROBLEM: Uses interface{}, comparison may panic.
func Contains(slice []interface{}, elem interface{}) bool {
	for _, v := range slice {
		if v == elem {
			return true
		}
	}
	return false
}

// Unique returns deduplicated slice.
// PROBLEM: Uses interface{}, map key must be comparable but no compile-time check.
func Unique(slice []interface{}) []interface{} {
	if slice == nil {
		return []interface{}{}
	}
	seen := make(map[interface{}]bool)
	var result []interface{}
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	if result == nil {
		return []interface{}{}
	}
	return result
}

// GroupBy groups elements by a key function.
// PROBLEM: Uses interface{}, no type safety for keys or values.
func GroupBy(slice []interface{}, fn func(interface{}) interface{}) map[interface{}][]interface{} {
	result := make(map[interface{}][]interface{})
	for _, v := range slice {
		key := fn(v)
		result[key] = append(result[key], v)
	}
	return result
}
