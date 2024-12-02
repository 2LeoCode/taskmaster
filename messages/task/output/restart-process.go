package output

import (
	"taskmaster/messages/helpers"
	processOutput "taskmaster/messages/process/output"
)

type RestartProcess interface {
	Message
	helpers.Local
	isRestartProcess() bool
	ProcessId() uint
}

type RestartProcessSuccess interface {
	RestartProcess
	helpers.Success
	ProcessId() uint
}

type RestartProcessFailure interface {
	RestartProcess
	helpers.Failure
	ProcessId() uint
}

type restartProcess struct {
	message
	helpers.BaseLocal
	processId uint
}

type restartProcessSuccess struct {
	restartProcess
	helpers.BaseSuccess
}

type restartProcessFailure struct {
	restartProcess
	helpers.BaseFailure
}

func (*restartProcess) isRestartProcess() bool { return true }
func (this *restartProcess) ProcessId() uint   { return this.processId }

func NewRestartProcess(processId uint, response processOutput.Restart) RestartProcess {
	switch response.(type) {
	case processOutput.RestartSuccess:
		return NewRestartProcessSuccess(processId)
	case processOutput.RestartFailure:
		response := response.(processOutput.StartFailure)
		return NewRestartProcessFailure(processId, response.Reason())
	}
	return nil
}

func NewRestartProcessSuccess(processId uint) RestartProcessSuccess {
	return &restartProcessSuccess{
		restartProcess: restartProcess{processId: processId},
	}
}

func NewRestartProcessFailure(processId uint, reason string) RestartProcessFailure {
	instance := restartProcessFailure{
		restartProcess: restartProcess{processId: processId},
	}
	instance.SetReason(reason)
	return &instance
}
