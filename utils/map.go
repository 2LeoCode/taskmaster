package utils

func Map[T, U any](slice []T, mapFunction func(int, *T) U) []U {
	result := make([]U, len(slice))

	for i, value := range slice {
		result[i] = mapFunction(i, &value)
	}
	return result
}

func GoMap[T, U any](slice []T, mapFunction func(int, *T) U) []U {
	pipe := make(chan Pair[int, U], len(slice))
	for i, value := range slice {
		i := i
		value := value
		go func() {
			pipe <- NewPair(i, mapFunction(i, &value))
		}()
	}

	result := make([]U, len(slice))
	for range result {
		value := <-pipe
		result[value.First] = value.Second
	}
	return result
}
