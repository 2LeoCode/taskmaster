package utils

import "fmt"

// If `ptr` is nil, return "(nil)", otherwise
// return fmt.Sprint(*ptr)
func PointerFormat[T any](ptr *T) string {
	if ptr == nil {
		return "(nil)"
	}
	return fmt.Sprint(*ptr)
}
