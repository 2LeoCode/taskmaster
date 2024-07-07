package process_requests

type ProcessRequest interface {
	processRequestTag()
}

type processRequest struct{}

func (*processRequest) processRequestTag() {}
