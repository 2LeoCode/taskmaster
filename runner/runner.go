package runner

import (
	"fmt"
	"taskmaster/config"
	"taskmaster/requests"
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

	StartTime   *time.Time
	StartedTime *time.Time
	StopTime    *time.Time
	StoppedTime *time.Time
	WasKilled   *bool
	ExitStatus  *uint
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
		runner.Processes[i] = NewProcessRunner(idx, i, taskConfig)
	}
	return runner
}

func NewProcessRunner(taskIdx, idx int, taskConfig *config.Task) ProcessRunner {
	return ProcessRunner{ProcId: fmt.Sprintf("%d%d", taskIdx, idx)}
}

func StartRunner(
	configPath string,
	config *config.Config,
	input <-chan requests.Request,
	output chan<- requests.Response,
) {

}
