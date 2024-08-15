package output

import "taskmaster/messages/helpers"

type Restart interface {
	Message
	helpers.Local
	isRestart() bool
}

type RestartSuccess interface {
	Restart
	helpers.Success
	Killed() bool
}

type RestartFailure interface {
	Restart
	helpers.Failure
}

type restart struct {
	message
	helpers.BaseLocal
}

type restartSuccess struct {
	restart
	helpers.BaseSuccess
	killed bool
}

type restartFailure struct {
	restart
	helpers.BaseFailure
}

func (*restart) isRestart() bool { return true }

func (this *restartSuccess) Killed() bool { return this.killed }

func NewRestartSuccess(killed bool) RestartSuccess {
	return &restartSuccess{killed: killed}
}

func NewRestartFailure(reason string) RestartFailure {
	instance := restartFailure{}
	instance.SetReason(reason)
	return &instance
}
