package input

type RestartProcess interface {
	Message
	isRestartProcess() bool
	ProcessId() uint
}

type restartProcess struct {
	message
	processId uint
}

func (*restartProcess) isRestartProcess() bool { return true }
func (this *restartProcess) ProcessId() uint   { return this.processId }

func NewRestartProcess(processId uint) RestartProcess {
	return &restartProcess{processId: processId}
}
