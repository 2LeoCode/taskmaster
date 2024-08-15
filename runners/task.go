package runners

import (
	"taskmaster/config"
	"taskmaster/helpers"
	processInput "taskmaster/messages/process/input"
	processOutput "taskmaster/messages/process/output"
	"taskmaster/messages/task/input"
	"taskmaster/messages/task/output"
	"taskmaster/utils"
)

type TaskRunner struct {
	Id     uint
	Config config.Task

	Input  chan input.Message
	Output chan output.Message

	Processes             []*ProcessRunner
	LocalProcessesOutput  <-chan utils.Pair[uint, processOutput.Message]
	GlobalProcessesOutput <-chan []processOutput.Message
}

func newTaskRunner(id uint, config *config.Config, input chan input.Message, output chan output.Message) *TaskRunner {
	taskConfig := config.Tasks[id]

	processInputs := make([]chan processInput.Message, taskConfig.Instances)
	processOutputs := make([]chan processOutput.Message, taskConfig.Instances)

	localProcessesOutput := make(chan utils.Pair[uint, processOutput.Message])
	globalProcessesOutput := make(chan []processOutput.Message)

	instance := TaskRunner{
		Id:                    id,
		Config:                taskConfig,
		Input:                 input,
		Output:                output,
		Processes:             make([]*ProcessRunner, taskConfig.Instances),
		LocalProcessesOutput:  localProcessesOutput,
		GlobalProcessesOutput: globalProcessesOutput,
	}

	for i := range instance.Processes {
		processInputs[i] = make(chan processInput.Message)
		processOutputs[i] = make(chan processOutput.Message)
		instance.Processes[i] = newProcessRunner(uint(i), &taskConfig, processInputs[i], processOutputs[i])
	}

	go func() {
		globalChunk := make([]processOutput.Message, taskConfig.Instances)
		for {
			for i := 0; i < len(processOutputs); i++ {
				value, ok := <-processOutputs[i]
				if !ok {
					// Not sure what to do here yet
					close(localProcessesOutput)
					close(globalProcessesOutput)
					return
				}
				switch value.(type) {

				case helpers.Local:
					localProcessesOutput <- utils.NewPair(uint(i), value)
					i--

				case helpers.Global:
					globalChunk[i] = value

				}
				globalProcessesOutput <- globalChunk
			}
		}
	}()

	return &instance
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
