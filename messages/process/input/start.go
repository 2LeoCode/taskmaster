package input

import "taskmaster/messages/helpers"

type Start interface {
	Message
	helpers.Local
	isStart() bool
}

type start struct {
	message
	helpers.BaseLocal
}

func (*start) isStart() bool { return true }

func NewStart() Start { return &start{} }
