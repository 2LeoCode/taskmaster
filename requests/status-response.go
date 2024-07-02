package requests

type ProcessStatus struct {
	Id     string
	Status string
}

type TaskStatus struct {
	Id        string
	Processes []ProcessStatus
}

type StatusResponse interface {
	Response
	statusTag()
	Tasks() []TaskStatus
}

type statusResponse struct {
	response
	tasks []TaskStatus
}

func (*statusResponse) statusTag()

func (this *statusResponse) Tasks() []TaskStatus {
	return this.tasks
}

func NewStatusResponse(tasks []TaskStatus) StatusResponse {
	return &statusResponse{tasks: tasks}
}
