package task_requests

type StopProcessRequest interface {
	TaskRequest
	ProcessId() int
	statusTag()
}

type stopProcessRequest struct {
	taskRequest
	process_id int
}

func (this *stopProcessRequest) ProcessId() int {
	return this.process_id
}

func (*stopProcessRequest) statusTag() {}

func NewStopProcessRequest(process_id int) StopProcessRequest {
	return &stopProcessRequest{process_id: process_id}
}
