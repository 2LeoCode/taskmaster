package runners

import (
	"taskmaster/config"
	"taskmaster/config/manager"
	"taskmaster/helpers"
	"taskmaster/messages/master/input"
	"taskmaster/messages/master/output"
	taskInput "taskmaster/messages/task/input"
	taskOutput "taskmaster/messages/task/output"
	"taskmaster/state"
	"taskmaster/utils"
)

type MasterRunner struct {
	ConfigManager *configManager.MasterConfigManager

	Input  <-chan input.Message
	Output chan<- output.Message

	Tasks             []*TaskRunner
	LocalTasksOutput  <-chan utils.Pair[uint, taskOutput.Message]
	GlobalTasksOutput <-chan []taskOutput.Message
}

func NewMasterRunner(manager *configManager.MasterConfigManager, input <-chan input.Message, output chan<- output.Message) (*MasterRunner, error) {
	nTasks := configManager.UseMaster(manager, func(config *config.Config) int { return len(config.Tasks) })

	taskInputs := make([]chan taskInput.Message, nTasks)
	taskOutputs := make([]chan taskOutput.Message, nTasks)

	localTasksOutput := make(chan utils.Pair[uint, taskOutput.Message])
	globalTasksOutput := make(chan []taskOutput.Message)

	instance := &MasterRunner{
		ConfigManager:     manager,
		Input:             input,
		Output:            output,
		Tasks:             make([]*TaskRunner, nTasks),
		LocalTasksOutput:  localTasksOutput,
		GlobalTasksOutput: globalTasksOutput,
	}

	for i := range instance.Tasks {
		taskInputs[i] = make(chan taskInput.Message)
		taskOutputs[i] = make(chan taskOutput.Message)
		if task, err := newTaskRunner(configManager.NewTask(manager, uint(i)), taskInputs[i], taskOutputs[i]); err != nil {
			return nil, err
		} else {
			instance.Tasks[i] = task
		}
	}

	manager.Subscribe(func(config, prev *config.Config) state.StateCleanupFn {
		// On config reload, to implement here
		return func() {}
	})

	globalOutputs := make([]chan taskOutput.Message, len(taskOutputs))
	for i, out := range taskOutputs {
		globalOutputs[i] = make(chan taskOutput.Message)
		i := i
		out := out
		go func() {
			for msg := range out {
				switch msg.(type) {
				case helpers.Local:
					localTasksOutput <- utils.NewPair(uint(i), msg)
				case helpers.Global:
					globalOutputs[i] <- msg
				}
			}
		}()
	}
	go func() {
		chunk := make([]taskOutput.Message, len(globalOutputs))
		for {
			for i, ch := range globalOutputs {
				if value, ok := <-ch; !ok {
					return
				} else {
					chunk[i] = value
				}
			}

			globalTasksOutput <- chunk
		}
	}()

	return instance, nil
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
			go this.ConfigManager.Load()

		default:
			this.Output <- output.NewBadRequest()

		}
	}
}
