package input

import (
	"taskmaster/messages/helpers"
	"taskmaster/messages/task"
)

type Message interface {
	helpers.Message
	helpers.Input
	target() task.Target
}

type message struct {
	helpers.BaseMessage
	helpers.BaseInput
}

func (*message) target() task.Target { return task.Target{} }
