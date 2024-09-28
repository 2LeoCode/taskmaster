package input

import "taskmaster/messages/helpers"

type Restart interface {
	Message
	helpers.Local
	isRestart() bool
}

type restart struct {
	message
	helpers.BaseLocal
}

func (*restart) isRestart() bool { return true }

func NewRestart() Restart { return &restart{} }
