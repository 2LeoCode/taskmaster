package responses

type Response interface {
	responseTag()
}

type response struct{}

func (*response) responseTag()
