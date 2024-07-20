package requests

type StartProcessRequest interface {
	Request
	startProcessTag()
	TaskId() uint
	ProcessId() uint
}

type startProcessRequest struct {
	request
	taskId    uint
	processId uint
}

func (*startProcessRequest) startProcessTag() {}

func (this *startProcessRequest) TaskId() uint {
	return this.taskId
}

func (this *startProcessRequest) ProcessId() uint {
	return this.processId
}

func NewStartProcessRequest(taskId, processId uint) StartProcessRequest {
	return &startProcessRequest{
		taskId:    taskId,
		processId: processId,
	}
}
