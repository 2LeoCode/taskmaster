package requests

type StopProcessRequest interface {
	Request
	stopProcessTag()
	Task_id() int 
	Process_id() int 
}

type stopProcessRequest struct {
	request
	task_id int 
	process_id int
}

func (*stopProcessRequest) stopProcessTag() {}

func (this *stopProcessRequest) Process_id() int {
	return this.process_id
}

func (this *stopProcessRequest) Task_id() int {
	return this.task_id
}

func NewStopProcessRequest(task_id int, process_id int) StopProcessRequest {
	return &stopProcessRequest{task_id: task_id, process_id: process_id}
}
