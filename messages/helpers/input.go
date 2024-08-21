package helpers

type Input interface {
	isInput() bool
}

type BaseInput struct{}

func (*BaseInput) isInput() bool { return true }
