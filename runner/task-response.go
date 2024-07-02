package runner

import "taskmaster/requests"

type TaskResponse interface {
	taskResponseTag()
}

type taskResponse struct{}

func (*taskResponse) taskResponseTag()

type StatusTaskResponse interface {
	TaskResponse
	statusTag()
	Status() requests.TaskStatus
}

type statusTaskResponse struct {
	taskResponse
	status requests.TaskStatus
}

func (*statusTaskResponse) statusTag()

func (this *statusTaskResponse) Status() requests.TaskStatus {
	return this.status
}

func NewTaskStatusResponse(status requests.TaskStatus) StatusTaskResponse {
	return &statusTaskResponse{status: status}
}
