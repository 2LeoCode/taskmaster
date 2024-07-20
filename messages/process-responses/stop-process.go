package process_responses

type StopProcessProcessResponse interface {
	ProcessResponse
	stopProcessTag()
}

type StopProcessSuccessProcessResponse interface {
	StopProcessProcessResponse
	successTag()
}

type StopProcessFailureProcessResponse interface {
	StopProcessProcessResponse
	failureTag()
	Reason() string
}

type stopProcessProcessResponse struct {
	processResponse
}

func (*stopProcessProcessResponse) stopProcessTag() {}

type stopProcessSuccessProcessResponse struct {
	stopProcessProcessResponse
}

func (*stopProcessSuccessProcessResponse) successTag() {}

type stopProcessFailureProcessResponse struct {
	stopProcessProcessResponse
	reason string
}

func (*stopProcessFailureProcessResponse) failureTag() {}

func (this *stopProcessFailureProcessResponse) Reason() string {
	return this.reason
}

func NewStopProcessSuccessProcessResponse() StopProcessSuccessProcessResponse {
	return &stopProcessSuccessProcessResponse{}
}

func NewStopProcessFailureProcessResponse(reason string) StopProcessFailureProcessResponse {
	return &stopProcessFailureProcessResponse{reason: reason}
}
