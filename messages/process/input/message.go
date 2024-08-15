package input

import (
	"taskmaster/helpers"
	"taskmaster/messages/process"
)

type Message interface {
	helpers.Message
	helpers.Input
	target() process.Target
}

type message struct {
	helpers.BaseMessage
	helpers.BaseInput
}

func (*message) target() process.Target { return process.Target{} }
