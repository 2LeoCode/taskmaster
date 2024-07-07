package process_responses

type StartProcessResponse interface {
	ProcessResponse
	startTag()
	Id() uint
}

type startProcessResponse struct {
	processResponse
	id uint
}

func (*startProcessResponse) startTag() {}

func (this *startProcessResponse) Id() uint {
	return this.id
}

type StartSuccessProcessResponse interface {
	StartProcessResponse
	successTag()
}

type startSuccessProcessResponse struct{ startProcessResponse }

func (*startSuccessProcessResponse) successTag() {}

func NewStartSuccessProcessResponse() StartSuccessProcessResponse {
	return &startSuccessProcessResponse{}
}

type StartFailureProcessResponse interface {
	StartProcessResponse
	failureTag()
	Reason() string
}

type startFailureProcessResponse struct {
	startProcessResponse
	reason string
}

func (*startFailureProcessResponse) failureTag() {}

func (this *startFailureProcessResponse) Reason() string {
	return this.reason
}

func NewStartFailureProcessResponse(reason string) StartFailureProcessResponse {
	return &startFailureProcessResponse{reason: reason}
}
