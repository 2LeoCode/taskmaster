package requests

type StopProcessRequest interface {
	Request
	stopProcess()
	Id() string
}

type _stopProcessRequest struct {
	_request
	id string
}

func (*_stopProcessRequest) stopProcess() {}

func (this *_stopProcessRequest) Id() string {
	return this.id
}

func NewStopProcessRequest(id string) StopProcessRequest {
	return &_stopProcessRequest{
		id: id,
	}
}
