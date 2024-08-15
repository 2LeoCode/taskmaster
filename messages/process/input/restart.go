package input

type Restart interface {
	Message
	isRestart() bool
}

type restart struct{ message }

func (*restart) isRestart() bool { return true }

func NewRestart() Restart { return &restart{} }
