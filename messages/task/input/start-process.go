package input

import "taskmaster/messages/helpers"

type StartProcess interface {
	Message
	helpers.Local
	isStartProcess() bool
	ProcessId() uint
}

type startProcess struct {
	message
	helpers.BaseLocal
	processId uint
}

func (*startProcess) isStartProcess() bool { return true }
func (this *startProcess) ProcessId() uint { return this.processId }

func NewStartProcess(processId uint) StartProcess {
	return &startProcess{processId: processId}
}
