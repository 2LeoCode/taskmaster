package responses

type RestartProcessResponse interface {
	Response
	restartProcessTag()
}

type RestartProcessSuccesResponse interface {
	RestartProcessResponse
	successTag()
}

type RestartProcessFailureResponse interface {
	RestartProcessResponse
	failureTag()
	Reason() string
}

type restartProcessResponse struct {
	response
}

func (*restartProcessResponse) restartProcessTag() {}

type restartProcessSuccessResponse struct {
	restartProcessResponse
}

func (*restartProcessSuccessResponse) successTag() {}

type restartProcessFailureResponse struct {
	restartProcessResponse
	reason string
}

func (*restartProcessFailureResponse) failureTag() {}

func (this *restartProcessFailureResponse) Reason() string {
	return this.reason
}

func NewRestartProcessSuccessResponse() RestartProcessSuccesResponse {
	return &restartProcessSuccessResponse{}
}

func NewRestartProcessFailureResponse(reason string) RestartProcessFailureResponse {
	return &restartProcessFailureResponse{reason: reason}
}
