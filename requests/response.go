package requests

type Response interface {
	response()
}

type _response struct{}

func (*_response) response() {}
