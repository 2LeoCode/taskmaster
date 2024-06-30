package requests

type StatusRequest interface {
	Request
	status()
}

type _statusRequest struct{ _request }

func (*_statusRequest) status() {}

func NewStatusRequest() StatusRequest {
	return &_statusRequest{}
}
