package responses

type StartProcessResponse interface {
	Response
	startProcessTag()
}

type StartProcessSuccesResponse interface {
	StartProcessResponse
	successTag()
}

type StartProcessFailureResponse interface {
	StartProcessResponse
	failureTag()
	Reason() string
}

type startProcessResponse struct {
	response
}

func (*startProcessResponse) startProcessTag()

type startProcessSuccessResponse struct {
	startProcessResponse
}

func (*startProcessSuccessResponse) successTag()

type startProcessFailureResponse struct {
	startProcessResponse
	reason string
}

func (*startProcessFailureResponse) failureTag()

func (this *startProcessFailureResponse) Reason() string {
	return this.reason
}

func NewStartProcessSuccessResponse() StartProcessSuccesResponse {
	return &startProcessSuccessResponse{}
}

func NewStartProcessFailureResponse(reason string) StartProcessFailureResponse {
	return &startProcessFailureResponse{reason: reason}
}
