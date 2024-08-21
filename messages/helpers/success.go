package helpers

type Success interface {
	isSuccess() bool
}

type BaseSuccess struct{}

func (*BaseSuccess) isSuccess() bool { return true }
