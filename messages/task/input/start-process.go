package input

type StartProcess interface {
	Message
	isStartProcess() bool
	ProcessId() uint
}

type startProcess struct {
	message
	processId uint
}

func (*startProcess) isStartProcess() bool { return true }
func (this *startProcess) ProcessId() uint { return this.processId }

func NewStartProcess(processId uint) StartProcess {
	return &startProcess{processId: processId}
}
