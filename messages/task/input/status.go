package input

type Status interface {
	Message
	isStatus() bool
}

type status struct {
	message
}

func (*status) isStatus() bool { return true }

func NewStatus() Status { return &status{} }
