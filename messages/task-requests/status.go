package task_requests

type StatusTaskRequest interface {
	TaskRequest
	statusTag()
}

type statusTaskRequest struct {
	taskRequest
}

func (*statusTaskRequest) statusTag() {}

func NewStatusTaskRequest() StatusTaskRequest {
	return &statusTaskRequest{}
}
