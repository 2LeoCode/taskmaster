package input

type Stop interface {
	Message
	isStop() bool
}

type stop struct{ message }

func (*stop) isStop() bool { return true }

func NewStop() Stop { return &stop{} }
