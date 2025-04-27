package utils

// Ternary is a helper function that mimics the ternary operator
// Returns a if condition is true, otherwise returns b
func Ternary[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}
