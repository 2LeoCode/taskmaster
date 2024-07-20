package responses

type StartProcessResponse interface {
	Response
	startProcessTag()
	TaskId() uint
	ProcessId() uint
}

type StartProcessSuccesResponse interface {
	StartProcessResponse
	successTag()
}

type StartProcessFailureResponse interface {
	StartProcessResponse
	failureTag()
	Reason() string
}

type startProcessResponse struct {
	response
	taskId    uint
	processId uint
}

func (*startProcessResponse) startProcessTag() {}

func (this *startProcessResponse) TaskId() uint {
	return this.taskId
}

func (this *startProcessResponse) ProcessId() uint {
	return this.processId
}

type startProcessSuccessResponse struct {
	startProcessResponse
}

func (*startProcessSuccessResponse) successTag() {}

type startProcessFailureResponse struct {
	startProcessResponse
	reason string
}

func (*startProcessFailureResponse) failureTag() {}

func (this *startProcessFailureResponse) Reason() string {
	return this.reason
}

func NewStartProcessSuccessResponse(taskId, processId uint) StartProcessSuccesResponse {
	return &startProcessSuccessResponse{
		startProcessResponse{
			taskId:    taskId,
			processId: processId,
		},
	}
}

func NewStartProcessFailureResponse(taskId, processId uint, reason string) StartProcessFailureResponse {
	return &startProcessFailureResponse{
		startProcessResponse{
			taskId:    taskId,
			processId: processId,
		},
		reason,
	}
}
