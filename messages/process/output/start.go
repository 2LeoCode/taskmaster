package output

import "taskmaster/messages/helpers"

type Start interface {
	Message
	helpers.Local
	isStart() bool
}

type StartSuccess interface {
	Start
	helpers.Success
}

type StartFailure interface {
	Start
	helpers.Failure
}

type start struct {
	message
	helpers.BaseLocal
}

type startSuccess struct {
	start
	helpers.BaseSuccess
}

type startFailure struct {
	start
	helpers.BaseFailure
}

func (*start) isStart() bool { return true }

func NewStartSuccess() StartSuccess {
	return &startSuccess{}
}

func NewStartFailure(reason string) StartFailure {
	instance := startFailure{}
	instance.SetReason(reason)
	return &instance
}
