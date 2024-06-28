package utils

import "fmt"

func PointerFormat[T any](ptr *T) string {
	if ptr == nil {
		return "(nil)"
	}
	return fmt.Sprint(*ptr)
}
