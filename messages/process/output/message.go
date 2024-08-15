package output

import (
	"taskmaster/messages/helpers"
	"taskmaster/messages/process"
)

type Message interface {
	helpers.Message
	helpers.Output
	target() process.Target
}

type message struct {
	helpers.BaseMessage
	helpers.BaseOutput
}

func (*message) target() process.Target { return process.Target{} }
