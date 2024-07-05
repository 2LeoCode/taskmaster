package process_events

type StartedProcessEvent interface {
	ProcessEvent
	startedTag()
}

type startedProcessEvent struct {
	processEvent
}

func (*startedProcessEvent) startedTag()

func NewStartedProcessEvent() StartedProcessEvent {
	return &startedProcessEvent{}
}
