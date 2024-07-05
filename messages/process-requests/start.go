package process_requests

type StartProcessRequest interface {
	ProcessRequest
	startTag()
}

type startProcessRequest struct {
	processRequest
	id uint
}

func (*startProcessRequest) startTag()

func NewStartProcessRequest() StartProcessRequest {
	return &startProcessRequest{}
}
