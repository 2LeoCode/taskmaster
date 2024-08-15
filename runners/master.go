package runners

import (
	"taskmaster/config"
	"taskmaster/helpers"
	"taskmaster/messages/master/input"
	"taskmaster/messages/master/output"
	taskInput "taskmaster/messages/task/input"
	taskOutput "taskmaster/messages/task/output"
	"taskmaster/utils"
)

type MasterRunner struct {
	ConfigLoader config.ConfigLoader
	Config       *config.Config

	Input  <-chan input.Message
	Output chan<- output.Message

	Tasks             []*TaskRunner
	LocalTasksOutput  <-chan utils.Pair[uint, taskOutput.Message]
	GlobalTasksOutput <-chan []taskOutput.Message
}

func NewMasterRunner(configLoader config.ConfigLoader, input <-chan input.Message, output chan<- output.Message) (*MasterRunner, error) {
	config, err := configLoader()

	if err != nil {
		return nil, err
	}

	nTasks := len(config.Tasks)

	taskInputs := make([]chan taskInput.Message, nTasks)
	taskOutputs := make([]chan taskOutput.Message, nTasks)

	localTasksOutput := make(chan utils.Pair[uint, taskOutput.Message])
	globalTasksOutput := make(chan []taskOutput.Message)

	instance := MasterRunner{
		ConfigLoader:      configLoader,
		Config:            config,
		Input:             input,
		Output:            output,
		Tasks:             make([]*TaskRunner, nTasks),
		LocalTasksOutput:  localTasksOutput,
		GlobalTasksOutput: globalTasksOutput,
	}

	for i := range instance.Tasks {
		taskInputs[i] = make(chan taskInput.Message)
		taskOutputs[i] = make(chan taskOutput.Message)
		instance.Tasks[i] = newTaskRunner(uint(i), config, taskInputs[i], taskOutputs[i])
	}

	go func() {
		globalChunk := make([]taskOutput.Message, nTasks)
		for {
			for i := 0; i < len(taskOutputs); i++ {
				value, ok := <-taskOutputs[i]
				if !ok {
					// Not sure what to do here yet
					close(localTasksOutput)
					close(globalTasksOutput)
					return
				}
				switch value.(type) {

				case helpers.Local:
					localTasksOutput <- utils.NewPair(uint(i), value)
					i--

				case helpers.Global:
					globalChunk[i] = value

				}
			}
			globalTasksOutput <- globalChunk
		}
	}()

	return &instance, nil
}

func (this *MasterRunner) Run() {
	for _, task := range this.Tasks {
		go task.Run()
	}

	go func() {
		defer close(this.Output)

	loop:
		for {
			select {

			case local, ok := <-this.LocalTasksOutput:
				if !ok {
					break loop
				}
				switch local.Second.(type) {

				case taskOutput.StartProcess:
					this.Output <- output.NewStartProcess(
						local.First,
						local.Second.(taskOutput.StartProcess),
					)

				case taskOutput.StopProcess:
					this.Output <- output.NewStopProcess(
						local.First,
						local.Second.(taskOutput.StopProcess),
					)

				case taskOutput.RestartProcess:
					this.Output <- output.NewRestartProcess(
						local.First,
						local.Second.(taskOutput.RestartProcess),
					)

				}

			case global, ok := <-this.GlobalTasksOutput:
				if !ok {
					break loop
				}
				switch global[0].(type) {

				case taskOutput.Status:
					this.Output <- output.NewStatus(
						utils.Transform(
							global,
							func(i int, elem *taskOutput.Message) taskOutput.Status {
								return (*elem).(taskOutput.Status)
							},
						),
					)

				}
			}
		}
	}()

	for req := range this.Input {
		switch req.(type) {

		case input.Status:
			for _, task := range this.Tasks {
				task.Input <- taskInput.NewStatus()
			}

		case input.StartProcess:
			req := req.(input.StartProcess)
			if req.TaskId() >= uint(len(this.Tasks)) {
				this.Output <- output.NewStartProcessFailure(
					req.TaskId(),
					req.ProcessId(),
					"Invalid task ID",
				)
				break
			}
			this.Tasks[req.TaskId()].Input <- taskInput.NewStartProcess(req.ProcessId())

		case input.StopProcess:
			req := req.(input.StopProcess)
			if req.TaskId() >= uint(len(this.Tasks)) {
				this.Output <- output.NewStopProcessFailure(
					req.TaskId(),
					req.ProcessId(),
					"Invalid task ID",
				)
				break
			}
			this.Tasks[req.TaskId()].Input <- taskInput.NewStopProcess(req.ProcessId())

		case input.RestartProcess:
			req := req.(input.RestartProcess)
			if req.TaskId() >= uint(len(this.Tasks)) {
				this.Output <- output.NewRestartProcessFailure(
					req.TaskId(),
					req.ProcessId(),
					"Invalid task ID",
				)
				break
			}
			this.Tasks[req.TaskId()].Input <- taskInput.NewRestartProcess(req.ProcessId())

		case input.Shutdown:
			for _, task := range this.Tasks {
				task.Input <- taskInput.NewShutdown()
			}

		case input.Reload:
			// TODO

		default:
			this.Output <- output.NewBadRequest()

		}
	}
}
