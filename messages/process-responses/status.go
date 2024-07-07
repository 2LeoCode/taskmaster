package process_responses

import "taskmaster/messages/responses"

type StatusProcessResponse interface {
	ProcessResponse
	statusTag()
	Status() responses.ProcessStatus
}

type statusProcessResponse struct {
	processResponse
	status responses.ProcessStatus
}

func (*statusProcessResponse) statusTag() {}

func (this *statusProcessResponse) Status() responses.ProcessStatus {
	return this.status
}

func NewStatusProcessResponse(status responses.ProcessStatus) StatusProcessResponse {
	return &statusProcessResponse{status: status}
}
