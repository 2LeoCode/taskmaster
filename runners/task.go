package runners

import (
	"taskmaster/config"
	"taskmaster/config/manager"
	"taskmaster/messages/helpers"
	processInput "taskmaster/messages/process/input"
	processOutput "taskmaster/messages/process/output"
	"taskmaster/messages/task/input"
	"taskmaster/messages/task/output"
	"taskmaster/state"
	"taskmaster/utils"
)

type TaskRunner struct {
	Id uint

	Input  chan input.Message
	Output chan output.Message

	Processes             []*ProcessRunner
	LocalProcessesOutput  <-chan utils.Pair[uint, processOutput.Message]
	GlobalProcessesOutput <-chan []processOutput.Message
}

type buildConfig struct {
	id        uint
	instances uint
}

func newTaskRunner(manager *configManager.Task, input chan input.Message, output chan output.Message) (*TaskRunner, error) {
	conf := configManager.UseTask(manager,
		func(config *config.Config, taskId uint) *buildConfig {
			return &buildConfig{taskId, config.Tasks[taskId].Instances}
		},
	)

	processInputs := make([]chan processInput.Message, conf.instances)
	processOutputs := make([]chan processOutput.Message, conf.instances)

	localProcessesOutput := make(chan utils.Pair[uint, processOutput.Message])
	globalProcessesOutput := make(chan []processOutput.Message)

	instance := &TaskRunner{
		Input:                 input,
		Output:                output,
		Processes:             make([]*ProcessRunner, conf.instances),
		LocalProcessesOutput:  localProcessesOutput,
		GlobalProcessesOutput: globalProcessesOutput,
	}

	for i := range instance.Processes {
		processInputs[i] = make(chan processInput.Message)
		processOutputs[i] = make(chan processOutput.Message)
		if process, err := newProcessRunner(manager, uint(i), processInputs[i], processOutputs[i]); err != nil {
			return nil, err
		} else {
			instance.Processes[i] = process
		}
	}

	manager.Master.Subscribe(func(newConf, prevConf *config.Config) state.StateCleanupFn {
		// On config reload, task specific actions, to implement here
		return func() {}
	})

	globalOutputs := make([]chan processOutput.Message, len(processOutputs))
	for i, out := range processOutputs {
		globalOutputs[i] = make(chan processOutput.Message)
		i := i
		out := out
		go func() {
			for msg := range out {
				switch msg.(type) {
				case helpers.Local:
					localProcessesOutput <- utils.NewPair(uint(i), msg)
				case helpers.Global:
					globalOutputs[i] <- msg
				}
			}
		}()
	}
	go func() {
		chunk := make([]processOutput.Message, len(globalOutputs))
		for {
			for i, ch := range globalOutputs {
				if value, ok := <-ch; !ok {
					return
				} else {
					chunk[i] = value
				}
			}
			globalProcessesOutput <- chunk
		}
	}()

	return instance, nil
}

func (this *TaskRunner) Run() {
	for _, proc := range this.Processes {
		go proc.Run()
	}

	go func() {
		defer close(this.Output)

	loop:
		for {
			select {

			case local, ok := <-this.LocalProcessesOutput:
				if !ok {
					break loop
				}
				switch local.Second.(type) {

				case processOutput.Start:
					this.Output <- output.NewStartProcess(
						local.First,
						local.Second.(processOutput.Start),
					)

				case processOutput.Stop:
					this.Output <- output.NewStopProcess(
						local.First,
						local.Second.(processOutput.Stop),
					)

				case processOutput.Restart:
					this.Output <- output.NewRestartProcess(
						local.First,
						local.Second.(processOutput.Restart),
					)
				}

			case global, ok := <-this.GlobalProcessesOutput:
				if !ok {
					break loop
				}
				switch global[0].(type) {

				case processOutput.Status:
					this.Output <- output.NewStatus(
						this.Id,
						utils.Transform(
							global,
							func(i int, elem *processOutput.Message) processOutput.Status {
								return (*elem).(processOutput.Status)
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
			for _, proc := range this.Processes {
				proc.Input <- processInput.NewStatus()
			}

		case input.StartProcess:
			req := req.(input.StartProcess)
			if req.ProcessId() >= uint(len(this.Processes)) {
				this.Output <- output.NewStartProcessFailure(
					req.ProcessId(),
					"Invalid process ID",
				)
				break
			}
			this.Processes[req.ProcessId()].Input <- processInput.NewStart()

		case input.StopProcess:
			req := req.(input.StopProcess)
			if req.ProcessId() >= uint(len(this.Processes)) {
				this.Output <- output.NewStopProcessFailure(
					req.ProcessId(),
					"Invalid process ID",
				)
				break
			}
			this.Processes[req.ProcessId()].Input <- processInput.NewStop()

		case input.RestartProcess:
			req := req.(input.RestartProcess)
			if req.ProcessId() >= uint(len(this.Processes)) {
				this.Output <- output.NewRestartProcessFailure(
					req.ProcessId(),
					"Invalid process ID",
				)
				break
			}
			this.Processes[req.ProcessId()].Input <- processInput.NewRestart()

		case input.Shutdown:
			for _, proc := range this.Processes {
				proc.Input <- processInput.NewShutdown()
			}

		}
	}
}
