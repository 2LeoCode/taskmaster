package utils

func New[T any](value T) *T {
	result := new(T)
	*result = value
	return result
}
