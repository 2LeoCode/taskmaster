package task_requests

type StartProcessTaskRequest interface {
	TaskRequest
	startTag()
	Id() uint
}

type startProcessTaskRequest struct {
	taskRequest
	id uint
}

func (*startProcessTaskRequest) startTag() {}

func (this *startProcessTaskRequest) Id() uint {
	return this.id
}

func NewStartProcessTaskRequest(id uint) StartProcessTaskRequest {
	return &startProcessTaskRequest{id: id}
}
