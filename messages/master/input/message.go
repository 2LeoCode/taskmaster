package input

import (
	"taskmaster/messages/helpers"
	"taskmaster/messages/master"
)

type Message interface {
	helpers.Message
	helpers.Input
	target() master.Target
}

type message struct {
	helpers.BaseMessage
	helpers.BaseInput
}

func (*message) target() master.Target { return master.Target{} }
