package utils

// Create a new slice populated with the results of calling a provided
// function on every element in the provided slice.
func Transform[T, U any](slice []T, mapFunction func(int, *T) U) []U {
	result := make([]U, len(slice))

	for i, value := range slice {
		result[i] = mapFunction(i, &value)
	}
	return result
}

// Do the same thing as Transform, but wrap each function call in a
// goroutine, and wait asynchronously for the results. This can improve
// performance in some cases (for example if a lot of work is done inside
// the provided function).
func GoTransform[T, U any](slice []T, mapFunction func(int, *T) U) []U {
	// Create a channel of Pairs, to store index-value pairs.
	pipe := make(chan Pair[int, U], len(slice))

	for i, value := range slice {
		// Store index and value in local variable, so we can safely access
		// them inside the goroutine
		i := i
		value := value

		// Call the function inside the goroutine and write the index and result
		// to our channel using a Pair.
		go func() {
			pipe <- NewPair(i, mapFunction(i, &value))
		}()
	}

	// Create our result slice
	result := make([]U, len(slice))
	for range result {
		// Read from the channel
		value := <-pipe
		// Insert the value into result at the right index.
		result[value.First] = value.Second
	}
	return result
}
