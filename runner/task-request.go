package runner

type TaskRequest interface {
	taskRequestTag()
}

type taskRequest struct{}

func (*taskRequest) taskRequestTag()

type StatusTaskRequest interface {
	TaskRequest
	statusTag()
}

type statusTaskRequest struct {
	taskRequest
}

func (*statusTaskRequest) statusTag()

func NewStatusTaskRequest() StatusTaskRequest {
	return &statusTaskRequest{}
}
