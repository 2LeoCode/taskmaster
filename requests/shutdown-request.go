package requests

type ShutdownRequest interface {
	Request
	shutdown()
}

type _shutdownRequest struct{ _request }

func (*_shutdownRequest) shutdown() {}

func NewShutdownRequest() ShutdownRequest {
	return &_shutdownRequest{}
}
