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
	status()
	Tasks() []TaskStatus
}

type _statusResponse struct {
	_response
	tasks []TaskStatus
}

func (*_statusResponse) status() {}

func (this *_statusResponse) Tasks() []TaskStatus {
	return this.tasks
}

func NewStatusResponse(tasks []TaskStatus) StatusResponse {
	return &_statusResponse{tasks: tasks}
}
