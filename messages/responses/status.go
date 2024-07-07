package responses

type ProcessStatus struct {
	Id     uint
	Status string
}

type TaskStatus struct {
	Id        uint
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

func (*statusResponse) statusTag() {}

func (this *statusResponse) Tasks() []TaskStatus {
	return this.tasks
}

func NewStatusResponse(tasks []TaskStatus) StatusResponse {
	return &statusResponse{tasks: tasks}
}
