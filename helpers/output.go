package helpers

type Output interface {
	isOutput() bool
}

type BaseOutput struct{}

func (*BaseOutput) isOutput() bool { return true }
