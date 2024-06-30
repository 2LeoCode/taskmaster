package requests

type Request interface {
	request()
}

type _request struct{}

func (*_request) request() {}
