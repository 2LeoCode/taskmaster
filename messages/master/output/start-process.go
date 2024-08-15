package output

import (
	"taskmaster/messages/helpers"
	taskOutput "taskmaster/messages/task/output"
)

type StartProcess interface {
	Message
	TaskId() uint
	ProcessId() uint
}

type StartProcessSuccess interface {
	StartProcess
	helpers.Success
}

type StartProcessFailure interface {
	StartProcess
	helpers.Failure
}

type startProcess struct {
	message
	taskId    uint
	processId uint
}

type startProcessSuccess struct {
	startProcess
	helpers.BaseSuccess
}

type startProcessFailure struct {
	startProcess
	helpers.BaseFailure
}

func (*startProcess) isStartProcess() bool { return true }
func (this *startProcess) TaskId() uint    { return this.taskId }
func (this *startProcess) ProcessId() uint { return this.processId }

func NewStartProcess(taskId uint, response taskOutput.StartProcess) StartProcess {
	switch response.(type) {
	case taskOutput.StartProcessSuccess:
		response := response.(taskOutput.StartProcessSuccess)
		return NewStartProcessSuccess(taskId, response.ProcessId())
	case taskOutput.StartProcessFailure:
		response := response.(taskOutput.StartProcessFailure)
		return NewStartProcessFailure(taskId, response.ProcessId(), response.Reason())
	}
	return nil
}

func NewStartProcessSuccess(taskId, processId uint) StartProcessSuccess {
	return &startProcessSuccess{
		startProcess: startProcess{
			taskId:    taskId,
			processId: processId,
		},
	}
}

func NewStartProcessFailure(taskId, processId uint, reason string) StartProcessFailure {
	instance := startProcessFailure{
		startProcess: startProcess{
			taskId:    taskId,
			processId: processId,
		},
	}
	instance.SetReason(reason)
	return &instance
}
