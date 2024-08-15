package output

import (
	"taskmaster/messages/helpers"
	processOutput "taskmaster/messages/process/output"
)

type StartProcess interface {
	Message
	helpers.Local
	isStartProcess() bool
	ProcessId() uint
}

type StartProcessSuccess interface {
	StartProcess
	helpers.Success
	ProcessId() uint
}

type StartProcessFailure interface {
	StartProcess
	helpers.Failure
	ProcessId() uint
}

type startProcess struct {
	message
	helpers.BaseLocal
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
func (this *startProcess) ProcessId() uint { return this.processId }

func NewStartProcess(processId uint, response processOutput.Start) StartProcess {
	switch response.(type) {
	case processOutput.StartSuccess:
		return NewStartProcessSuccess(processId)
	case processOutput.StartFailure:
		response := response.(processOutput.StartFailure)
		return NewStartProcessFailure(processId, response.Reason())
	}
	return nil
}

func NewStartProcessSuccess(processId uint) StartProcessSuccess {
	return &startProcessSuccess{
		startProcess: startProcess{processId: processId},
	}
}

func NewStartProcessFailure(processId uint, reason string) StartProcessFailure {
	instance := startProcessFailure{
		startProcess: startProcess{processId: processId},
	}
	instance.SetReason(reason)
	return &instance
}
