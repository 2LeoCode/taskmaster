package requests

type Response interface {
	responseTag()
}

type response struct{}

func (*response) responseTag()
