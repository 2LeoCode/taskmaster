package task_requests

type TaskRequest interface {
	taskRequestTag()
}

type taskRequest struct{}

func (*taskRequest) taskRequestTag()
