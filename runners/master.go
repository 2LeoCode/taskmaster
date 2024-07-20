package runners

import (
	"sync"
	"taskmaster/config"
	"taskmaster/messages/requests"
	"taskmaster/messages/responses"
	"taskmaster/messages/task-requests"
	"taskmaster/messages/task-responses"
)

type MasterRunner struct {
	ConfigPath string
	Tasks      []TaskRunner
}

func NewMasterRunner(configPath string, taskCount uint) MasterRunner {
	return MasterRunner{
		ConfigPath: configPath,
		Tasks:      make([]TaskRunner, taskCount),
	}
}

func (this *MasterRunner) Run(
	config *config.Config,
	input <-chan requests.Request,
	output chan<- responses.Response,
) {
	var waitGroup sync.WaitGroup
	defer waitGroup.Wait()

	for i, task := range config.Tasks {
		this.Tasks[i] = NewTaskRunner(uint(i), task.Instances)
	}
	taskInputs := make([]chan task_requests.TaskRequest, len(this.Tasks))
	taskOutputs := make([]chan task_responses.TaskResponse, len(this.Tasks))
	agg := make(chan task_responses.TaskResponse)
	for i, task := range this.Tasks {
		waitGroup.Add(1)
		i := i
		taskInputs[i] = make(chan task_requests.TaskRequest)
		taskOutputs[i] = make(chan task_responses.TaskResponse)

		go func() {
			defer waitGroup.Done()

			go task.Run(config, uint(i), taskInputs[i], taskOutputs[i])
			for msg := range taskOutputs[i] {
				agg <- msg // forward responses to aggregator channel
			}
		}()
	}

	for {
		req := <-input
		if _, ok := req.(requests.StatusRequest); ok {
			for _, ch := range taskInputs {
				ch <- task_requests.NewStatusTaskRequest() // forward to each task
			}

			res := make([]responses.TaskStatus, len(taskOutputs))
			for i := range taskOutputs {
				value, _ := (<-agg).(task_responses.StatusTaskResponse)
				res[i] = value.Status()
			}
			output <- responses.NewStatusResponse(res)
		} else if start, ok := req.(requests.StartProcessRequest); ok {
			if start.TaskId() > uint(len(taskInputs)) {
				output <- responses.NewStartProcessFailureResponse(start.TaskId(), start.ProcessId(), "Invalid task id")
			} else {
				taskInputs[start.TaskId()] <- task_requests.NewStartProcessTaskRequest(start.ProcessId())
				res := <-agg
				if _, ok := res.(task_responses.StartProcessSuccessTaskResponse); ok {
					output <- responses.NewStartProcessSuccessResponse(start.TaskId(), start.ProcessId())
				} else {
					res := res.(task_responses.StartProcessFailureTaskResponse)
					output <- responses.NewStartProcessFailureResponse(start.TaskId(), start.ProcessId(), res.Reason())
				}
			}
		}
	}

}
