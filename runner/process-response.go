package runner

import "taskmaster/requests"

type ProcessResponse interface {
	processResponseTag()
}

type processResponse struct{}

func (*processResponse) processResponseTag()

type StatusProcessResponse interface {
	ProcessResponse
	statusTag()
	Status() requests.ProcessStatus
}

type statusProcessResponse struct {
	processResponse
	status requests.ProcessStatus
}

func (*statusProcessResponse) statusTag()

func (this *statusProcessResponse) Status() requests.ProcessStatus {
	return this.status
}

func NewStatusProcessResponse(status requests.ProcessStatus) StatusProcessResponse {
	return &statusProcessResponse{status: status}
}
