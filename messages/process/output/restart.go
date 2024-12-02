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
}

type restartFailure struct {
	restart
	helpers.BaseFailure
}

func (*restart) isRestart() bool { return true }

func NewRestartSuccess() RestartSuccess {
	return &restartSuccess{}
}

func NewRestartFailure(reason string) RestartFailure {
	instance := restartFailure{}
	instance.SetReason(reason)
	return &instance
}
