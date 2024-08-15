package helpers

type Local interface {
	isLocal() bool
}

type BaseLocal struct{}

func (*BaseLocal) isLocal() bool { return true }
