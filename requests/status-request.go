package requests

type StatusRequest interface {
	Request
	statusTag()
}

type statusRequest struct{ request }

func (*statusRequest) statusTag()

func NewStatusRequest() StatusRequest {
	return &statusRequest{}
}
