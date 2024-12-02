package runners

import (
	"sync"

	"taskmaster/config"
	"taskmaster/messages/helpers"
	processInput "taskmaster/messages/process/input"
	processOutput "taskmaster/messages/process/output"
	"taskmaster/messages/task/input"
	"taskmaster/messages/task/output"
	"taskmaster/utils"
)

type TaskRunner struct {
	Config config.Task

	Id uint

	Input  <-chan input.Message
	Output chan<- output.Message

	Processes             []*ProcessRunner
	LocalProcessesOutput  <-chan utils.Pair[uint, processOutput.Message]
	GlobalProcessesOutput <-chan []processOutput.Message

	processInputs  []chan processInput.Message
	processOutputs []chan processOutput.Message
}

type buildConfig struct {
	id        uint
	instances uint
}

func newTaskRunner(manager config.Manager, id uint, input <-chan input.Message, output chan<- output.Message) (*TaskRunner, error) {
	conf := manager.Get().Tasks[id]
	processInputs := make([]chan processInput.Message, conf.Instances)
	processOutputs := make([]chan processOutput.Message, conf.Instances)

	localProcessesOutput := make(chan utils.Pair[uint, processOutput.Message])
	globalProcessesOutput := make(chan []processOutput.Message)

	instance := &TaskRunner{
		Config:                conf,
		Id:                    id,
		Input:                 input,
		Output:                output,
		Processes:             make([]*ProcessRunner, conf.Instances),
		LocalProcessesOutput:  localProcessesOutput,
		GlobalProcessesOutput: globalProcessesOutput,
		processInputs:         processInputs,
		processOutputs:        processOutputs,
	}

	for i := range instance.Processes {
		processInputs[i] = make(chan processInput.Message)
		processOutputs[i] = make(chan processOutput.Message)
		if process, err := newProcessRunner(manager, id, uint(i), processInputs[i], processOutputs[i]); err != nil {
			return nil, err
		} else {
			instance.Processes[i] = process
		}
	}

	globalOutputs := make([]chan processOutput.Message, len(processOutputs))

	var wg sync.WaitGroup
	wg.Add(len(processOutputs))

	go func() {
		wg.Wait()
		close(localProcessesOutput)
	}()

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
			close(globalOutputs[i])
			wg.Done()
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

func (this *TaskRunner) close(processClosed *sync.WaitGroup) {
	for _, ch := range this.processInputs {
		close(ch)
	}
	processClosed.Wait()
}

func (this *TaskRunner) forwardGlobalMessage(message interface {
	helpers.Global
	processInput.Message
}) {
	for _, ch := range this.processInputs {
		ch <- message
	}
}

func (this *TaskRunner) Run() {
	var processClosed sync.WaitGroup
	defer this.close(&processClosed)

	for i, proc := range this.Processes {
		processClosed.Add(1)
		go func() {
			proc.Run()
			close(this.processOutputs[i])
			processClosed.Done()
		}()
	}

	go func() {

		for {
			select {

			case local, ok := <-this.LocalProcessesOutput:
				if !ok {
					return
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
					return
				}
				switch global[0].(type) {

				case processOutput.Status:
					this.Output <- output.NewStatus(
						this.Id,
						*this.Config.Name,
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
			this.forwardGlobalMessage(processInput.NewStatus())

		case input.StartProcess:
			req := req.(input.StartProcess)
			if req.ProcessId() >= uint(len(this.Processes)) {
				this.Output <- output.NewStartProcessFailure(
					req.ProcessId(),
					"Invalid process ID",
				)
				break
			}
			this.processInputs[req.ProcessId()] <- processInput.NewStart()

		case input.StopProcess:
			req := req.(input.StopProcess)
			if req.ProcessId() >= uint(len(this.Processes)) {
				this.Output <- output.NewStopProcessFailure(
					req.ProcessId(),
					"Invalid process ID",
				)
				break
			}
			this.processInputs[req.ProcessId()] <- processInput.NewStop()

		case input.RestartProcess:
			req := req.(input.RestartProcess)
			if req.ProcessId() >= uint(len(this.Processes)) {
				this.Output <- output.NewRestartProcessFailure(
					req.ProcessId(),
					"Invalid process ID",
				)
				break
			}
			this.processInputs[req.ProcessId()] <- processInput.NewRestart()

		case input.Shutdown:
			this.forwardGlobalMessage(processInput.NewShutdown())
			return

		}
	}
}
