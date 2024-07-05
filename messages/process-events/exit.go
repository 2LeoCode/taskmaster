package process_events

type ExitProcessEvent interface {
	ProcessEvent
	exitTag()
}

type exitProcessEvent struct {
	processEvent
}

func (*exitProcessEvent) exitTag()

func NewExitProcessEvent() ExitProcessEvent {
	return &exitProcessEvent{}
}
