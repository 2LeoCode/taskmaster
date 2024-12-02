package output

import (
	"taskmaster/messages/helpers"
	taskOutput "taskmaster/messages/task/output"
)

type StopProcess interface {
	Message
	TaskId() uint
	ProcessId() uint
}

type StopProcessSuccess interface {
	StopProcess
	helpers.Success
}

type StopProcessFailure interface {
	StopProcess
	helpers.Failure
}

type stopProcess struct {
	message
	taskId    uint
	processId uint
}

type stopProcessSuccess struct {
	stopProcess
	helpers.BaseSuccess
}

type stopProcessFailure struct {
	stopProcess
	helpers.BaseFailure
}

func (*stopProcess) isStopProcess() bool  { return true }
func (this *stopProcess) TaskId() uint    { return this.taskId }
func (this *stopProcess) ProcessId() uint { return this.processId }

func NewStopProcess(taskId uint, response taskOutput.StopProcess) StopProcess {
	switch response.(type) {
	case taskOutput.StopProcessSuccess:
		response := response.(taskOutput.StopProcessSuccess)
		return NewStopProcessSuccess(taskId, response.ProcessId())
	case taskOutput.StopProcessFailure:
		response := response.(taskOutput.StopProcessFailure)
		return NewStopProcessFailure(taskId, response.ProcessId(), response.Reason())
	}
	return nil
}

func NewStopProcessSuccess(taskId, processId uint) StopProcessSuccess {
	return &stopProcessSuccess{
		stopProcess: stopProcess{
			taskId:    taskId,
			processId: processId,
		},
	}
}

func NewStopProcessFailure(taskId, processId uint, reason string) StopProcessFailure {
	instance := stopProcessFailure{
		stopProcess: stopProcess{
			taskId:    taskId,
			processId: processId,
		},
	}
	instance.SetReason(reason)
	return &instance
}
