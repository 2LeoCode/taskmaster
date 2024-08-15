package output

type BadRequest interface {
	Message
	isBadRequest() bool
}

type badRequest struct{ message }

func (*badRequest) isBadRequest() bool { return true }

func NewBadRequest() BadRequest {
	return &badRequest{}
}
