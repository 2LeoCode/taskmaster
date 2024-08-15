package utils

// Generic struct that can hold two values of any type.
type Pair[T, U any] struct {
	First  T
	Second U
}

// Create a new pair populated with the provided values.
func NewPair[T, U any](first T, second U) Pair[T, U] {
	return Pair[T, U]{first, second}
}
