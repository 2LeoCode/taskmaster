package task_responses

type StartProcessTaskResponse interface {
	TaskResponse
	startProcessTag()
	TaskId() uint
	ProcessId() uint
}

type StartProcessSuccessTaskResponse interface {
	StartProcessTaskResponse
	successTag()
}

type StartProcessFailureTaskResponse interface {
	StartProcessTaskResponse
	failureTag()
	Reason() string
}

type startProcessTaskResponse struct {
	taskResponse
	taskId    uint
	processId uint
}

func (*startProcessTaskResponse) startProcessTag() {}

func (this *startProcessTaskResponse) TaskId() uint {
	return this.taskId
}

func (this *startProcessTaskResponse) ProcessId() uint {
	return this.processId
}

type startProcessSuccessTaskResponse struct {
	startProcessTaskResponse
}

func (*startProcessSuccessTaskResponse) successTag() {}

func NewStartProcessSuccessTaskResponse(taskId, processId uint) StartProcessSuccessTaskResponse {
	return &startProcessSuccessTaskResponse{
		startProcessTaskResponse{
			taskId:    taskId,
			processId: processId,
		},
	}
}

type startProcessFailureTaskResponse struct {
	startProcessTaskResponse
	reason string
}

func (*startProcessFailureTaskResponse) failureTag() {}

func (this *startProcessFailureTaskResponse) Reason() string {
	return this.reason
}

func NewStartProcessFailureTaskResponse(taskId, processId uint, reason string) StartProcessFailureTaskResponse {
	return &startProcessFailureTaskResponse{
		startProcessTaskResponse{
			taskId:    taskId,
			processId: processId,
		},
		reason,
	}
}
