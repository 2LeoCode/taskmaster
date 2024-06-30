package requests

type StartProcessResponse interface {
	Response
	startProcess()
}

type StartProcessSuccesResponse interface {
	StartProcessResponse
	success()
}

type StartProcessFailureResponse interface {
	StartProcessResponse
	failure()
	Reason() string
}

type _startProcessResponse struct {
	_response
}

func (*_startProcessResponse) startProcess() {}

type _startProcessSuccessResponse struct {
	_startProcessResponse
}

func (*_startProcessSuccessResponse) success() {}

type _startProcessFailureResponse struct {
	_startProcessResponse
	reason string
}

func (*_startProcessFailureResponse) failure() {}

func (this *_startProcessFailureResponse) Reason() string {
	return this.reason
}

func NewStartProcessSuccessResponse() StartProcessSuccesResponse {
	return &_startProcessSuccessResponse{}
}

func NewStartProcessFailureResponse(reason string) StartProcessFailureResponse {
	return &_startProcessFailureResponse{reason: reason}
}
