package helpers

type Global interface {
	isGlobal() bool
}

type BaseGlobal struct{}

func (*BaseGlobal) isGlobal() bool { return true }
