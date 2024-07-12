package runners

import (
	"sync"
	"taskmaster/config"
	"taskmaster/messages/process-requests"
	"taskmaster/messages/process-responses"
	"taskmaster/messages/responses"
	"taskmaster/messages/task-requests"
	"taskmaster/messages/task-responses"
	"taskmaster/utils"
)

type TaskRunner struct {
	TaskId uint

	Processes []ProcessRunner
}

func NewTaskRunner(id, instances uint) TaskRunner {
	runner := TaskRunner{
		TaskId:    id,
		Processes: make([]ProcessRunner, instances),
	}

	for i := range runner.Processes {
		runner.Processes[i] = NewProcessRunner(uint(i))
	}
	return runner
}

func (this *TaskRunner) Run(config *config.Config, taskId uint, input <-chan task_requests.TaskRequest, output chan<- task_responses.TaskResponse) {
	var waitGroup sync.WaitGroup
	defer waitGroup.Wait()

	taskConfig := config.Tasks[taskId]

	processInputs := make([]chan process_requests.ProcessRequest, len(this.Processes))
	processOutputs := make([]chan process_responses.ProcessResponse, len(this.Processes))
	agg := make(chan process_responses.ProcessResponse)

	for i, proc := range this.Processes {
		waitGroup.Add(1)
		processInputs[i] = make(chan process_requests.ProcessRequest)
		processOutputs[i] = make(chan process_responses.ProcessResponse)
		procInput := processInputs[i]
		procOutput := processOutputs[i]

		go func() {
			defer waitGroup.Done()

			go proc.Run(config, taskId, procInput, procOutput)
			for msg := range procOutput {
				agg <- msg // forward responses to aggregator channel
			}
		}()
	}

	if taskConfig.StartAtLaunch {
		for _, ch := range processInputs {
			ch <- process_requests.NewStartProcessRequest()
		}
	}

	for {
		req := <-input
		if _, ok := req.(task_requests.StatusTaskRequest); ok {
			for _, ch := range processInputs {
				ch <- process_requests.NewStatusProcessRequest()
			}

			res := make([]process_responses.StatusProcessResponse, len(processOutputs))
			for i := range processOutputs {
				value, _ := (<-agg).(process_responses.StatusProcessResponse)
				res[i] = value
			}
			output <- task_responses.NewStatusTaskResponse(
				responses.TaskStatus{
					Id: this.TaskId,
					Processes: utils.Map(
						res,
						func(_ int, value *process_responses.StatusProcessResponse) responses.ProcessStatus {
							return (*value).Status()
						},
					),
				},
			)
		}
	}
}
