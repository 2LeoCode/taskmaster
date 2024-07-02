package requests

type RestartProcessRequest interface {
	Request
	restartProcessTag()
	Id() string
}

type restartProcessRequest struct {
	request
	id string
}

func (*restartProcessRequest) restartProcessTag()

func (this *restartProcessRequest) Id() string {
	return this.id
}

func NewRestartProcessRequest(id string) RestartProcessRequest {
	return &restartProcessRequest{
		id: id,
	}
}
