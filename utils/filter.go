package utils

type FilterFunc[T any] func(i int, elem *T) bool

func Filter[T any](slice *[]T, filterFunc FilterFunc[T]) {
	for i := 0; i < len(*slice); i++ {
		if !filterFunc(i, &(*slice)[i]) {
			(*slice)[i] = (*slice)[len(*slice)-1]
			*slice = (*slice)[:len(*slice)-1]
			i--
		}
	}
}
