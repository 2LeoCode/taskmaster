package requests

type StartProcessRequest interface {
	Request
	startProcessTag()
	Id() string
}

type startProcessRequest struct {
	request
	id string
}

func (*startProcessRequest) startProcessTag() {}

func (this *startProcessRequest) Id() string {
	return this.id
}

func NewStartProcessRequest(id string) StartProcessRequest {
	return &startProcessRequest{id: id}
}
