package input

import "taskmaster/messages/helpers"

type Shutdown interface {
	Message
	helpers.Local
	isShutdown() bool
}

type shutdown struct {
	message
	helpers.BaseLocal
}

func (*shutdown) isShutdown() bool { return true }

func NewShutdown() Shutdown { return &shutdown{} }
