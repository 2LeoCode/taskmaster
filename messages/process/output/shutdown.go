package output

import "taskmaster/messages/helpers"

type Shutdown interface {
	Message
	helpers.Global
	isShutdown() bool
}

type shutdown struct {
	message
	helpers.BaseGlobal
}

func (*shutdown) isShutdown() bool { return true }

func NewShutdown() Shutdown {
	return &shutdown{}
}
