package input

import "taskmaster/messages/helpers"

type Status interface {
	Message
	helpers.Global
	isStatus() bool
}

type status struct {
	message
	helpers.BaseGlobal
}

func (*status) isStatus() bool { return true }

func NewStatus() Status { return &status{} }
