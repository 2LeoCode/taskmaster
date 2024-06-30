package requests

type ShutdownResponse interface {
	Response
	shutdown()
}

type _shutdownResponse struct{ _response }

func (*_shutdownResponse) shutdown() {}

func NewShutdownResponse() ShutdownResponse {
	return &_shutdownResponse{}
}
