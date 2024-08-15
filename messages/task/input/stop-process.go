package input

type StopProcess interface {
	Message
	isStopProcess() bool
	ProcessId() uint
}

type stopProcess struct {
	message
	processId uint
}

func (*stopProcess) isStopProcess() bool  { return true }
func (this *stopProcess) ProcessId() uint { return this.processId }

func NewStopProcess(processId uint) StopProcess {
	return &stopProcess{processId: processId}
}
