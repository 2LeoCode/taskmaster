package requests

type RestartProcessResponse interface {
	Response
	restartProcess()
}

type RestartProcessSuccesResponse interface {
	RestartProcessResponse
	success()
}

type RestartProcessFailureResponse interface {
	RestartProcessResponse
	failure()
	Reason() string
}

type _restartProcessResponse struct {
	_response
}

func (*_restartProcessResponse) restartProcess() {}

type _restartProcessSuccessResponse struct {
	_restartProcessResponse
}

func (*_restartProcessSuccessResponse) success() {}

type _restartProcessFailureResponse struct {
	_restartProcessResponse
	reason string
}

func (*_restartProcessFailureResponse) failure() {}

func (this *_restartProcessFailureResponse) Reason() string {
	return this.reason
}

func NewRestartProcessSuccessResponse() RestartProcessSuccesResponse {
	return &_restartProcessSuccessResponse{}
}

func NewRestartProcessFailureResponse(reason string) RestartProcessFailureResponse {
	return &_restartProcessFailureResponse{reason: reason}
}
