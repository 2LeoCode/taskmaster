package output

import (
	"taskmaster/messages/helpers"
	"taskmaster/messages/task"
)

type Message interface {
	helpers.Message
	helpers.Output
	target() task.Target
}

type message struct {
	helpers.BaseMessage
	helpers.BaseOutput
}

func (*message) target() task.Target { return task.Target{} }
