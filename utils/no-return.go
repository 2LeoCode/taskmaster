package utils

type Void struct{}

func NoReturn(callback func()) func() Void {
	return func() Void {
		callback()
		return Void{}
	}
}
