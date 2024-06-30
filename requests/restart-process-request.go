package requests

type RestartProcessRequest interface {
	Request
	restartProcess()
	Id() string
}

type _restartProcessRequest struct {
	_request
	id string
}

func (*_restartProcessRequest) restartProcess() {}

func (this *_restartProcessRequest) Id() string {
	return this.id
}

func NewRestartProcessRequest(id string) RestartProcessRequest {
	return &_restartProcessRequest{
		id: id,
	}
}
