package runner

type ProcessRequest interface {
	processRequestTag()
}

type processRequest struct{}

func (*processRequest) processRequestTag()

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
