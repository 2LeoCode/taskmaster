package output

import (
	"taskmaster/messages/helpers"
	processOutput "taskmaster/messages/process/output"
)

type StopProcess interface {
	Message
	helpers.Local
	isStopProcess() bool
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
	helpers.BaseLocal
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
func (this *stopProcess) ProcessId() uint { return this.processId }

func (this *stopProcessSuccess) Killed() bool { return this.killed }

func NewStopProcess(processId uint, response processOutput.Stop) StopProcess {
	switch response.(type) {
	case processOutput.StopSuccess:
		response := response.(processOutput.StopSuccess)
		return NewStopProcessSuccess(processId, response.Killed())
	case processOutput.StopFailure:
		response := response.(processOutput.StopFailure)
		return NewStopProcessFailure(processId, response.Reason())
	}
	return nil
}

func NewStopProcessSuccess(processId uint, killed bool) StopProcessSuccess {
	return &stopProcessSuccess{
		stopProcess: stopProcess{processId: processId},
		killed:      killed,
	}
}

func NewStopProcessFailure(processId uint, reason string) StopProcessFailure {
	instance := stopProcessFailure{
		stopProcess: stopProcess{processId: processId},
	}
	instance.SetReason(reason)
	return &instance
}
