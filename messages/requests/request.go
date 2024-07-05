package requests

type Request interface {
	requestTag()
}

type request struct{}

func (*request) requestTag()
