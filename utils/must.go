package utils

import "log"

// To use as a wrapper around functions that return
// a result and error values.
// Check if `err` is nil, if so, just return the value,
// otherwise call log.Fatalln(err).
// Example:
//
//	f := utils.Must(os.Open("some_file"))
//
// This code will return the file handle if os.Open succeeds,
// otherwise it will print the returned error, and exit the program.
func Must[T any](value T, err error) T {
	if err != nil {
		log.Fatalln(err)
	}
	return value
}
