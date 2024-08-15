package input

type StartProcess interface {
	Message
	isStartProcess() bool
	TaskId() uint
	ProcessId() uint
}

type startProcess struct {
	message
	taskId    uint
	processId uint
}

func (*startProcess) isStartProcess() bool { return true }
func (this *startProcess) TaskId() uint    { return this.taskId }
func (this *startProcess) ProcessId() uint { return this.processId }

func NewStartProcess(taskId, processId uint) StartProcess {
	return &startProcess{taskId: taskId, processId: processId}
}
