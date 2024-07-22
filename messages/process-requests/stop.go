package process_requests

type StopProcessProcessRequest interface {
	ProcessRequest
	stopTag()
}

type stopProcessProcessRequest struct {
	processRequest
}

func (*stopProcessProcessRequest) stopTag() {}

func NewStopProcessProcessRequest() StopProcessProcessRequest {
	return &stopProcessProcessRequest{}
}
