package requests

type StopProcessRequest interface {
	Request
	stopProcessTag()
	Id() string
}

type stopProcessRequest struct {
	request
	id string
}

func (*stopProcessRequest) stopProcessTag() {}

func (this *stopProcessRequest) Id() string {
	return this.id
}

func NewStopProcessRequest(id string) StopProcessRequest {
	return &stopProcessRequest{id: id}
}
