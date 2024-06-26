package utils

import "log"

func Must[T any](value T, error error) T {
	if error != nil {
		log.Fatalln(error)
	}
	return value
}
