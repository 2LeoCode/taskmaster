package runner

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"taskmaster/config"
	"taskmaster/requests"
	"taskmaster/utils"
	"time"
)

type MasterRunner struct {
	Tasks []TaskRunner
}

type TaskRunner struct {
	TaskId string

	Processes []ProcessRunner
}

type ProcessRunner struct {
	ProcId string

	StartTime    *time.Time
	StartedTime  *time.Time
	StartRetries *uint
	StopTime     *time.Time
	StoppedTime  *time.Time
	ExitStatus   *int

	HasBeenStopped bool
	HasBeenKilled  bool
}

func NewMasterRunner(config *config.Config) MasterRunner {
	runner := MasterRunner{
		Tasks: make([]TaskRunner, len(config.Tasks)),
	}

	for i, task := range config.Tasks {
		runner.Tasks[i] = NewTaskRunner(i, &task)
	}
	return runner
}

func NewTaskRunner(idx int, taskConfig *config.Task) TaskRunner {
	runner := TaskRunner{
		TaskId:    fmt.Sprint(idx),
		Processes: make([]ProcessRunner, taskConfig.Instances),
	}

	for i := range runner.Processes {
		runner.Processes[i] = NewProcessRunner(i)
	}
	return runner
}

func NewProcessRunner(idx int) ProcessRunner {
	return ProcessRunner{ProcId: fmt.Sprint(idx)}
}

func (this *ProcessRunner) Run(config config.Task, input <-chan ProcessRequest, output chan<- ProcessResponse) {
	var waitGroup sync.WaitGroup
	defer waitGroup.Wait()

	cmd := exec.Command(*config.Command, config.Arguments...)

	for {
		req := <-input
		if _, ok := req.(StatusProcessRequest); ok {
			res := requests.ProcessStatus{Id: this.ProcId}
			switch {
			case this.StoppedTime != nil:
				if *this.ExitStatus == config.ExpectedExitStatus {
					res.Status = "SUCCESS "
				} else {
					res.Status = "FAILURE "
				}
				res.Status += fmt.Sprint(*this.ExitStatus)
				if this.HasBeenKilled {
					res.Status += " KILLED"
				} else if this.HasBeenStopped {
					res.Status += " STOPPED"
				}
			case cmd.Process == nil:
				res.Status = "NOT_STARTED"
			case this.StartedTime == nil:
				res.Status = "STARTING"
			default:
				res.Status = "RUNNING"
			}
			output <- NewStatusProcessResponse(res)
		}
	}
}

func (this *TaskRunner) Run(config config.Task, input <-chan TaskRequest, output chan<- TaskResponse) {
	var waitGroup sync.WaitGroup
	defer waitGroup.Wait()

	processInputs := make([]chan ProcessRequest, len(this.Processes))
	processOutputs := make([]chan ProcessResponse, len(this.Processes))
	agg := make(chan ProcessResponse)

	for i, proc := range this.Processes {
		waitGroup.Add(1)
		procInput := processInputs[i]
		procOutput := processOutputs[i]

		go func() {
			defer waitGroup.Done()

			go proc.Run(config, procInput, procOutput)
			for msg := range procOutput {
				agg <- msg // forward responses to aggregator channel
			}
		}()
	}

	for {
		req := <-input
		if _, ok := req.(StatusTaskRequest); ok {
			for _, ch := range processInputs {
				ch <- NewStatusProcessRequest()
			}

			res := make([]StatusProcessResponse, len(processOutputs))
			for i := range processOutputs {
				value, _ := (<-agg).(StatusProcessResponse)
				res[i] = value
			}
			output <- NewTaskStatusResponse(
				requests.TaskStatus{
					Id: this.TaskId,
					Processes: utils.Map(
						res,
						func(_ int, value *StatusProcessResponse) requests.ProcessStatus {
							return (*value).Status()
						},
					),
				},
			)
		}
	}
}

func StartRunner(
	configPath string,
	config *config.Config,
	input <-chan requests.Request,
	output chan<- requests.Response,
) {
	var waitGroup sync.WaitGroup
	defer waitGroup.Wait()

	runner := NewMasterRunner(config)
	taskInputs := make([]chan TaskRequest, len(runner.Tasks))
	taskOutputs := make([]chan TaskResponse, len(runner.Tasks))
	agg := make(chan TaskResponse)
	for i, task := range runner.Tasks {
		waitGroup.Add(1)
		i := i

		go func() {
			defer waitGroup.Done()

			go task.Run(config.Tasks[i], taskInputs[i], taskOutputs[i])
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
