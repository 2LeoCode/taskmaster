package runners

import (
	"fmt"
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
	GlobalProcessesOutput chan []processOutput.Message

	processInputs  []chan processInput.Message
	processOutputs []chan processOutput.Message

	globalOutputPipes []chan processOutput.Message
}

type buildConfig struct {
	id        uint
	instances uint
}

func newTaskRunner(manager config.Manager, id uint, input <-chan input.Message, output chan<- output.Message) (*TaskRunner, error) {
	conf := manager.Get().Tasks[id]
	processInputs := make([]chan processInput.Message, conf.Instances)
	processOutputs := make([]chan processOutput.Message, conf.Instances)

	globalProcessesOutput := make(chan []processOutput.Message)

	instance := &TaskRunner{
		Config:                conf,
		Id:                    id,
		Input:                 input,
		Output:                output,
		Processes:             make([]*ProcessRunner, conf.Instances),
		GlobalProcessesOutput: globalProcessesOutput,
		processInputs:         processInputs,
		processOutputs:        processOutputs,
		globalOutputPipes:     make([]chan processOutput.Message, len(processOutputs)),
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

	for i, out := range processOutputs {
		instance.globalOutputPipes[i] = make(chan processOutput.Message)
		i := i
		out := out
		go func() {
			for msg := range out {
				switch msg.(type) {
				case helpers.Global:
					instance.globalOutputPipes[i] <- msg
				}
			}
		}()
	}

	go func() {
		chunk := make([]processOutput.Message, len(instance.globalOutputPipes))
		for {
			for i, ch := range instance.globalOutputPipes {
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
	close(this.Output)
	for _, ch := range this.globalOutputPipes {
		close(ch)
	}
	close(this.GlobalProcessesOutput)
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

	for _, proc := range this.Processes {
		processClosed.Add(1)
		go func() {
			proc.Run()
			processClosed.Done()
		}()
	}

	go func() {

		for global := range this.GlobalProcessesOutput {
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
	}()

	for req := range this.Input {
		switch req.(type) {

		case input.Status:
			this.forwardGlobalMessage(processInput.NewStatus())

		case input.StartProcess:
			req := req.(input.StartProcess)
			if req.ProcessId() >= uint(len(this.Processes)) {
				fmt.Printf("\r \rTask %d failed to start process: invalid process id: %d.\n> ", this.Id, req.ProcessId())
				break
			}
			this.processInputs[req.ProcessId()] <- processInput.NewStart()

		case input.StopProcess:
			req := req.(input.StopProcess)
			if req.ProcessId() >= uint(len(this.Processes)) {
				fmt.Printf("\r \rTask %d failed to stop process: invalid process id: %d.\n> ", this.Id, req.ProcessId())
				break
			}
			this.processInputs[req.ProcessId()] <- processInput.NewStop()

		case input.RestartProcess:
			req := req.(input.RestartProcess)
			if req.ProcessId() >= uint(len(this.Processes)) {
				fmt.Printf("\r \rTask %d failed to restart process: invalid process id: %d.\n> ", this.Id, req.ProcessId())
				break
			}
			this.processInputs[req.ProcessId()] <- processInput.NewRestart()

		case input.Shutdown:
			this.forwardGlobalMessage(processInput.NewShutdown())
			return

		}
	}
}
