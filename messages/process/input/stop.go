package input

import "taskmaster/messages/helpers"

type Stop interface {
	Message
	helpers.Local
	isStop() bool
}

type stop struct {
	message
	helpers.BaseLocal
}

func (*stop) isStop() bool { return true }

func NewStop() Stop { return &stop{} }
