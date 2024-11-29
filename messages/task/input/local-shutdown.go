package input

import "taskmaster/messages/helpers"

type LocalShutdown interface {
	Message
	helpers.Local
}

type localShutdown struct {
	message
	helpers.BaseLocal
}

func NewLocalShutdown() LocalShutdown {
	return &localShutdown{}
}
