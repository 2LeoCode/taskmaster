package process_requests

type StopProcessRequest interface {
	ProcessRequest
	stopTag()
}

type stopProcessRequest struct {
	processRequest
}

func (*stopProcessRequest) stopTag() {}

func NewStopProcessRequest() StopProcessRequest {
	return &stopProcessRequest{}
}
