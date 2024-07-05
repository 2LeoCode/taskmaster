package requests

type ShutdownRequest interface {
	Request
	shutdownTag()
}

type shutdownRequest struct{ request }

func (*shutdownRequest) shutdownTag()

func NewShutdownRequest() ShutdownRequest {
	return &shutdownRequest{}
}
