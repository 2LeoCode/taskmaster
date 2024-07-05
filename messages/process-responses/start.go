package process_responses

type StartProcessResponse interface {
	ProcessResponse
	start()
	Id() uint
}

type startProcessResponse struct {
	processResponse
	id uint
}

func (*startProcessResponse) start()

func (this *startProcessResponse) Id() uint {
	return this.id
}

type StartSuccessProcessResponse interface {
	StartProcessResponse
	success()
}

type startSuccessProcessResponse struct{ startProcessResponse }

func (*startSuccessProcessResponse) success()

func NewStartSuccessProcessResponse() StartSuccessProcessResponse {
	return &startSuccessProcessResponse{}
}

type StartFailureProcessResponse interface {
	StartProcessResponse
	failure()
	Reason() string
}

type startFailureProcessResponse struct {
	startProcessResponse
	reason string
}

func (*startFailureProcessResponse) failure()

func (this *startFailureProcessResponse) Reason() string {
	return this.reason
}

func NewStartFailureProcessResponse(reason string) StartFailureProcessResponse {
	return &startFailureProcessResponse{reason: reason}
}
