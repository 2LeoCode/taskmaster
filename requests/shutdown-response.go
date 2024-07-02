package requests

type ShutdownResponse interface {
	Response
	shutdownTag()
}

type shutdownResponse struct{ response }

func (*shutdownResponse) shutdownTag()

func NewShutdownResponse() ShutdownResponse {
	return &shutdownResponse{}
}
