package runners

import (
	"sync"
	"taskmaster/config"
	"taskmaster/messages/requests"
	"taskmaster/messages/responses"
	"taskmaster/messages/task-requests"
	"taskmaster/messages/task-responses"
	"taskmaster/utils"
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
	taskInputs := make([]chan task_requests.TaskRequest, len(runner.Tasks))
	taskOutputs := make([]chan TaskResponse, len(runner.Tasks))
	agg := make(chan TaskResponse)
	for i, task := range runner.Tasks {
		waitGroup.Add(1)
		i := i

		go func() {
			defer waitGroup.Done()

			go task.Run(config, i, taskInputs[i], taskOutputs[i])
			for msg := range taskOutputs[i] {
				agg <- msg // forward responses to aggregator channel
			}
		}()
	}

	for {
		req := <-input
		if _, ok := req.(requests.StatusRequest); ok {
			for _, ch := range taskInputs {
				ch <- NewStatusTaskRequest() // forward to each task
			}

			res := make([]StatusTaskResponse, len(taskOutputs))
			for i := range taskOutputs {
				value, _ := (<-agg).(StatusTaskResponse)
				res[i] = value
			}
			output <- requests.NewStatusResponse(utils.Map(
				res, func(i int, value *StatusTaskResponse) requests.TaskStatus {
					return (*value).Status()
				},
			))
		}
	}

}
