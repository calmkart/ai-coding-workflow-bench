// Package strkit provides string and map utility functions.
package strkit

// MergeMaps merges multiple maps into one. Later maps override earlier ones.
// BUG: when only one map is provided, it returns the original map directly
// instead of making a copy. Modifying the result modifies the original.
func MergeMaps(maps ...map[string]string) map[string]string {
	if len(maps) == 0 {
		return map[string]string{}
	}

	// BUG: returns the first map directly, not a copy
	if len(maps) == 1 {
		return maps[0]
	}

	// BUG: uses first map as base, so modifications to result affect maps[0]
	result := maps[0]
	for _, m := range maps[1:] {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// Keys returns all keys from a map in arbitrary order.
func Keys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values from a map in arbitrary order.
func Values(m map[string]string) []string {
	vals := make([]string, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}
