package input

type Reload interface {
	Message
	isReload() bool
}

type reload struct{ message }

func (*reload) isReload() bool { return true }

func NewReload() Reload { return &reload{} }
