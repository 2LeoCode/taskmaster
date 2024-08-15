package process_events

import "taskmaster/helpers"

type Init interface {
	ProcessEvent
	isInit()
}

type InitSuccess interface {
	Init
	helpers.Success
}

type InitFailure interface {
	Init
	failureTag()
	helpers.Failure
}
