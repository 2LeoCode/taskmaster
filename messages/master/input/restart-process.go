package input

type RestartProcess interface {
	Message
	isRestartProcess() bool
	TaskId() uint
	ProcessId() uint
}

type restartProcess struct {
	message
	taskId    uint
	processId uint
}

func (*restartProcess) isRestartProcess() bool { return true }
func (this *restartProcess) TaskId() uint      { return this.taskId }
func (this *restartProcess) ProcessId() uint   { return this.processId }

func NewRestartProcess(taskId, processId uint) RestartProcess {
	return &restartProcess{taskId: taskId, processId: processId}
}
