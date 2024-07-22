package task_requests

type StopProcessTaskRequest interface {
	TaskRequest
	Process_id() int
	stopProcessTag()
}

type stopProcessTaskRequest struct {
	taskRequest
	process_id int
}

func (this *stopProcessTaskRequest) Process_id() int {
	return this.process_id
}

func (*stopProcessTaskRequest) stopProcessTag() {}

func NewStopProcessTaskRequest(process_id int) StopProcessTaskRequest {
	return &stopProcessTaskRequest{process_id: process_id}
}
