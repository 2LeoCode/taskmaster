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
	Killed() bool
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
	killed bool
}

type stopProcessFailure struct {
	stopProcess
	helpers.BaseFailure
}

func (*stopProcess) isStopProcess() bool  { return true }
func (this *stopProcess) TaskId() uint    { return this.taskId }
func (this *stopProcess) ProcessId() uint { return this.processId }

func (this *stopProcessSuccess) Killed() bool { return this.killed }

func NewStopProcess(taskId uint, response taskOutput.StopProcess) StopProcess {
	switch response.(type) {
	case taskOutput.StopProcessSuccess:
		response := response.(taskOutput.StopProcessSuccess)
		return NewStopProcessSuccess(taskId, response.ProcessId(), response.Killed())
	case taskOutput.StopProcessFailure:
		response := response.(taskOutput.StopProcessFailure)
		return NewStopProcessFailure(taskId, response.ProcessId(), response.Reason())
	}
	return nil
}

func NewStopProcessSuccess(taskId, processId uint, killed bool) StopProcessSuccess {
	return &stopProcessSuccess{
		stopProcess: stopProcess{
			taskId:    taskId,
			processId: processId,
		},
		killed: killed,
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
