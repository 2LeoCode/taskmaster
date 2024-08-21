package output

import (
	"taskmaster/messages/helpers"
	taskOutput "taskmaster/messages/task/output"
)

type RestartProcess interface {
	Message
	TaskId() uint
	ProcessId() uint
}

type RestartProcessSuccess interface {
	RestartProcess
	helpers.Success
	Killed() bool
}

type RestartProcessFailure interface {
	RestartProcess
	helpers.Failure
}

type restartProcess struct {
	message
	taskId    uint
	processId uint
}

type restartProcessSuccess struct {
	restartProcess
	helpers.BaseSuccess
	killed bool
}

type restartProcessFailure struct {
	restartProcess
	helpers.BaseFailure
}

func (*restartProcess) isRestartProcess() bool { return true }
func (this *restartProcess) TaskId() uint      { return this.taskId }
func (this *restartProcess) ProcessId() uint   { return this.processId }

func (this *restartProcessSuccess) Killed() bool { return this.killed }

func NewRestartProcess(taskId uint, response taskOutput.RestartProcess) RestartProcess {
	switch response.(type) {
	case taskOutput.RestartProcessSuccess:
		response := response.(RestartProcessSuccess)
		return NewRestartProcessSuccess(taskId, response.ProcessId(), response.Killed())
	case taskOutput.RestartProcessFailure:
		response := response.(taskOutput.RestartProcessFailure)
		return NewRestartProcessFailure(taskId, response.ProcessId(), response.Reason())
	}
	return nil
}

func NewRestartProcessSuccess(taskId, processId uint, killed bool) RestartProcessSuccess {
	return &restartProcessSuccess{
		restartProcess: restartProcess{
			taskId:    taskId,
			processId: processId,
		},
		killed: killed,
	}
}

func NewRestartProcessFailure(taskId, processId uint, reason string) RestartProcessFailure {
	instance := restartProcessFailure{
		restartProcess: restartProcess{
			taskId:    taskId,
			processId: processId,
		},
	}
	instance.SetReason(reason)
	return &instance
}
