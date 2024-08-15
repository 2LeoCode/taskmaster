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
	Killed() bool
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
	killed bool
}

type stopFailure struct {
	stop
	helpers.BaseFailure
}

func (*stop) isStop() bool { return true }

func (this *stopSuccess) Killed() bool { return this.killed }

func NewStopSuccess(killed bool) StopSuccess {
	return &stopSuccess{killed: killed}
}

func NewStopFailure(reason string) StopFailure {
	instance := stopFailure{}
	instance.SetReason(reason)
	return &instance
}
