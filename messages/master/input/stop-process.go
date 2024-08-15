package input

type StopProcess interface {
	Message
	isStopProcess() bool
	TaskId() uint
	ProcessId() uint
}

type stopProcess struct {
	message
	taskId    uint
	processId uint
}

func (*stopProcess) isStopProcess() bool  { return true }
func (this *stopProcess) TaskId() uint    { return this.taskId }
func (this *stopProcess) ProcessId() uint { return this.processId }

func NewStopProcess(taskId, processId uint) StopProcess {
	return &stopProcess{taskId: taskId, processId: processId}
}
