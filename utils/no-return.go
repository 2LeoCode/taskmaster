package utils

type void struct{}

func NoReturn(callback func()) func() void {
	return func() void {
		callback()
		return void{}
	}
}
