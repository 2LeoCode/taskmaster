package process_events

type Exit interface {
	ProcessEvent
	isExit() bool
}

type exit struct{ processEvent }

func (*exit) isExit() bool { return true }

func NewExit() Exit { return &exit{} }
