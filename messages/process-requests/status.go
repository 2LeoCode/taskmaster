package process_requests

type StatusProcessRequest interface {
	ProcessRequest
	statusTag()
}

type statusProcessRequest struct {
	processRequest
}

func (*statusProcessRequest) statusTag()

func NewStatusProcessRequest() StatusProcessRequest {
	return &statusProcessRequest{}
}
