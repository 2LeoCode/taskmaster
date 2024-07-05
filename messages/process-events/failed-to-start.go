package process_events

type FailedToStartProcessEvent interface {
	ProcessEvent
	failedToStartTag()
	Reason() string
}

type failedToStartProcessEvent struct {
	processEvent
	reason string
}

func (*failedToStartProcessEvent) failedToStartTag()

func (this *failedToStartProcessEvent) Reason() string {
	return this.reason
}

func NewFailedToStartProcessEvent(reason string) FailedToStartProcessEvent {
	return &failedToStartProcessEvent{reason: reason}
}
