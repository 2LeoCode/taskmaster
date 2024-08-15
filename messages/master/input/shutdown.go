package input

type Shutdown interface {
	Message
	isShutdown() bool
}

type shutdown struct{ message }

func (*shutdown) isShutdown() bool { return true }

func NewShutdown() Shutdown { return &shutdown{} }
