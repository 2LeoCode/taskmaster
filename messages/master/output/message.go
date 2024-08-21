package output

import (
	"taskmaster/messages/helpers"
	"taskmaster/messages/master"
)

type Message interface {
	helpers.Message
	helpers.Output
	target() master.Target
}

type message struct {
	helpers.BaseMessage
	helpers.BaseOutput
}

func (*message) target() master.Target { return master.Target{} }
