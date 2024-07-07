package task_responses

import "taskmaster/messages/responses"

type StatusTaskResponse interface {
	TaskResponse
	statusTag()
	Status() responses.TaskStatus
}

type statusTaskResponse struct {
	taskResponse
	status responses.TaskStatus
}

func (*statusTaskResponse) statusTag() {}

func (this *statusTaskResponse) Status() responses.TaskStatus {
	return this.status
}

func NewStatusTaskResponse(status responses.TaskStatus) StatusTaskResponse {
	return &statusTaskResponse{status: status}
}
