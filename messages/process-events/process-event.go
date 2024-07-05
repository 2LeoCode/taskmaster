package process_events

type ProcessEvent interface {
	processEventTag()
}

type processEvent struct{}

func (*processEvent) processEventTag()
