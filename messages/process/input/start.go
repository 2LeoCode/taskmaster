package input

type Start interface {
	Message
	isStart() bool
}

type start struct{ message }

func (*start) isStart() bool { return true }

func NewStart() Start { return &start{} }
