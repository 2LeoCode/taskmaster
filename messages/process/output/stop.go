package output

import "taskmaster/messages/helpers"

type Stop interface {
	Message
	helpers.Local
	isStop() bool
}

type StopSuccess interface {
	Stop
	helpers.Success
}

type StopFailure interface {
	Stop
	helpers.Failure
}

type stop struct {
	message
	helpers.BaseLocal
}

type stopSuccess struct {
	stop
	helpers.BaseSuccess
}

type stopFailure struct {
	stop
	helpers.BaseFailure
}

func (*stop) isStop() bool { return true }

func NewStopSuccess() StopSuccess {
	return &stopSuccess{}
}

func NewStopFailure(reason string) StopFailure {
	instance := stopFailure{}
	instance.SetReason(reason)
	return &instance
}
