package utils

func Resize[T any](slice []T, size uint, value T) {
	slice = slice[size:]
	length := uint(len(slice))
	if size > length {
		for i := uint(0); i < (size - length); i += 1 {
			slice = append(slice, value)
		}
	}
}
