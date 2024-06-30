package requests

type StopProcessResponse interface {
	Response
	stopProcess()
}

type StopProcessSuccesResponse interface {
	StopProcessResponse
	success()
}

type StopProcessFailureResponse interface {
	StopProcessResponse
	failure()
	Reason() string
}

type _stopProcessResponse struct {
	_response
}

func (*_stopProcessResponse) stopProcess() {}

type _stopProcessSuccessResponse struct {
	_stopProcessResponse
}

func (*_stopProcessSuccessResponse) success() {}

type _stopProcessFailureResponse struct {
	_stopProcessResponse
	reason string
}

func (*_stopProcessFailureResponse) failure() {}

func (this *_stopProcessFailureResponse) Reason() string {
	return this.reason
}

func NewStopProcessSuccessResponse() StopProcessSuccesResponse {
	return &_stopProcessSuccessResponse{}
}

func NewStopProcessFailureResponse(reason string) StopProcessFailureResponse {
	return &_stopProcessFailureResponse{reason: reason}
}
