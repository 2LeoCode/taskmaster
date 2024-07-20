package process_events

type InitProcessEvent interface {
	ProcessEvent
	initTag()
}

type InitSuccessProcessEvent interface {
	InitProcessEvent
	sucessTag()
}

type InitFailureProcessEvent interface {
	InitProcessEvent
	failureTag()
	Reason() string
}
