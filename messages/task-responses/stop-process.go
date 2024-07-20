package task_responses

type StopProcessTaskResponse interface {
	TaskResponse
	stopProcessTag()
}

type StopProcessSuccessTaskResponse interface {
	StopProcessTaskResponse
	successTag()
}

type StopProcessFailureTaskResponse interface {
	StopProcessTaskResponse
	failureTag()
	Reason() string
}

type stopProcessTaskResponse struct {
	taskResponse
}

func (*stopProcessTaskResponse) stopProcessTag() {}

type stopProcessSuccessTaskResponse struct {
	stopProcessTaskResponse
}

func (*stopProcessSuccessTaskResponse) successTag() {}

type stopProcessFailureTaskResponse struct {
	stopProcessTaskResponse
	reason string
}

func (*stopProcessFailureTaskResponse) failureTag() {}

func (this *stopProcessFailureTaskResponse) Reason() string {
	return this.reason
}

func NewStopProcessSuccessTaskResponse() StopProcessSuccessTaskResponse {
	return &stopProcessSuccessTaskResponse{}
}

func NewStopProcessFailureTaskResponse(reason string) StopProcessFailureTaskResponse {
	return &stopProcessFailureTaskResponse{reason: reason}
}
