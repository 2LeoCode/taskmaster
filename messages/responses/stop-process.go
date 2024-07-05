package responses

type StopProcessResponse interface {
	Response
	stopProcessTag()
}

type StopProcessSuccesResponse interface {
	StopProcessResponse
	successTag()
}

type StopProcessFailureResponse interface {
	StopProcessResponse
	failureTag()
	Reason() string
}

type stopProcessResponse struct {
	response
}

func (*stopProcessResponse) stopProcessTag()

type stopProcessSuccessResponse struct {
	stopProcessResponse
}

func (*stopProcessSuccessResponse) successTag()

type stopProcessFailureResponse struct {
	stopProcessResponse
	reason string
}

func (*stopProcessFailureResponse) failureTag()

func (this *stopProcessFailureResponse) Reason() string {
	return this.reason
}

func NewStopProcessSuccessResponse() StopProcessSuccesResponse {
	return &stopProcessSuccessResponse{}
}

func NewStopProcessFailureResponse(reason string) StopProcessFailureResponse {
	return &stopProcessFailureResponse{reason: reason}
}
