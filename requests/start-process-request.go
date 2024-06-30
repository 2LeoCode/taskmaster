package requests

type StartProcessRequest interface {
	Request
	startProcess()
	Id() string
}

type _startProcessRequest struct {
	_request
	id string
}

func (*_startProcessRequest) startProcess() {}

func (this *_startProcessRequest) Id() string {
	return this.id
}

func NewStartProcessRequest(id string) StartProcessRequest {
	return &_startProcessRequest{
		id: id,
	}
}
