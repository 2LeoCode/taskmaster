package input

import "taskmaster/messages/helpers"

type RestartProcess interface {
	Message
	helpers.Local
	isRestartProcess() bool
	ProcessId() uint
}

type restartProcess struct {
	message
	helpers.BaseLocal
	processId uint
}

func (*restartProcess) isRestartProcess() bool { return true }
func (this *restartProcess) ProcessId() uint   { return this.processId }

func NewRestartProcess(processId uint) RestartProcess {
	return &restartProcess{processId: processId}
}
