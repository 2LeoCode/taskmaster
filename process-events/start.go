package process_events

type StartProcessEvent interface {
	ProcessEvent
	startTag()
}

type StartSuccessProcessEvent interface {
	StartProcessEvent
	successTag()
}

type StartFailureProcessEvent interface {
	StartProcessEvent
	failureTag()
	Reason() string
}

type startProcessEvent struct {
	processEvent
}

func (*startProcessEvent) startTag() {}

type startSuccessProcessEvent struct {
	startProcessEvent
}

func (*startSuccessProcessEvent) successTag() {}

func NewStartSuccessProcessEvent() StartSuccessProcessEvent {
	return &startSuccessProcessEvent{}
}

type startFailureProcessEvent struct {
	startProcessEvent
	reason string
}

func (*startFailureProcessEvent) failureTag() {}

func (this *startFailureProcessEvent) Reason() string {
	return this.reason
}

func NewStartFailureProcessEvent(reason string) StartFailureProcessEvent {
	return &startFailureProcessEvent{reason: reason}
}
