package utils

func Map[T, U any](slice []T, mapFunction func(int, *T) U) []U {
	result := make([]U, len(slice))

	for i, value := range slice {
		result[i] = mapFunction(i, &value)
	}
	return result
}

func GoMap[T, U any](slice []T, mapFunction func(int, *T) U) []U {
	pipe := make(chan goMapResult[U])
	for i, value := range slice {
		i := i
		value := value
		go func() {
			pipe <- goMapResult[U]{i, mapFunction(i, &value)}
		}()
	}
	processed := 0

	result := make([]U, len(slice))
	for processed != len(slice) {
		select {
		case value := <-pipe:
			result[value.Idx] = value.Result
		}
	}
	return result
}

type goMapResult[T any] struct {
	Idx    int
	Result T
}
