package runners

import (
	"sync"

	"taskmaster/config"
	"taskmaster/messages/helpers"
	"taskmaster/messages/master/input"
	"taskmaster/messages/master/output"
	taskInput "taskmaster/messages/task/input"
	taskOutput "taskmaster/messages/task/output"
	"taskmaster/state"
	"taskmaster/utils"
)

type MasterRunner struct {
	ConfigManager *config.Manager

	Input  <-chan input.Message
	Output chan<- output.Message

	Tasks             []*TaskRunner
	LocalTasksOutput  <-chan utils.Pair[uint, taskOutput.Message]
	GlobalTasksOutput <-chan []taskOutput.Message

	taskInputs []chan<- taskInput.Message
	stopSignal chan StopSignal
}

type StopSignal struct{}

func NewMasterRunner(manager *config.Manager, in <-chan input.Message, out chan<- output.Message) (*MasterRunner, error) {
	nTasks := config.Use(manager, func(config *config.Config) int { return len(config.Tasks) })

	taskInputs := make([]chan taskInput.Message, nTasks)
	taskOutputs := make([]chan taskOutput.Message, nTasks)

	localTasksOutput := make(chan utils.Pair[uint, taskOutput.Message])
	globalTasksOutput := make(chan []taskOutput.Message)

	instance := &MasterRunner{
		ConfigManager:     manager,
		Input:             in,
		Output:            out,
		Tasks:             make([]*TaskRunner, nTasks),
		LocalTasksOutput:  localTasksOutput,
		GlobalTasksOutput: globalTasksOutput,
		taskInputs:        utils.Transform(taskInputs, func(_ int, ch *chan taskInput.Message) chan<- taskInput.Message { return *ch }),
		stopSignal:        make(chan StopSignal),
	}

	for i := range instance.Tasks {
		taskInputs[i] = make(chan taskInput.Message)
		taskOutputs[i] = make(chan taskOutput.Message)
		if task, err := newTaskRunner(manager, uint(i), taskInputs[i], taskOutputs[i]); err != nil {
			return nil, err
		} else {
			instance.Tasks[i] = task
		}
	}

	globalOutputs := make([]chan taskOutput.Message, len(taskOutputs))

	var wg sync.WaitGroup
	wg.Add(len(taskOutputs))

	go func() {
		wg.Wait()
		close(localTasksOutput)
	}()

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
			close(globalOutputs[i])
			wg.Done()
		}()
	}

	manager.Subscribe(func(conf, prevConf *config.Config) (state.StateCleanupFn, error) {
		if conf.LogDir != prevConf.LogDir || len(conf.Tasks) != len(prevConf.Tasks) {
			// Reload master
			instance.forwardGlobalMessage(taskInput.NewShutdown())
			instance.stopSignal <- StopSignal{}
			if newInstance, err := NewMasterRunner(manager, in, out); err != nil {
				return nil, err
			} else {
				*instance = *newInstance
			}
		} else {
			// Check which task needs to be reloaded

			for i := range conf.Tasks {
				if conf.Tasks[i].String() != prevConf.Tasks[i].String() {
					taskInputs[i] <- taskInput.NewShutdown()
					if task, err := newTaskRunner(manager, uint(i), taskInputs[i], taskOutputs[i]); err != nil {
						return nil, err
					} else {
						instance.Tasks[i] = task
					}
				}
			}
		}
		return nil, nil
	})

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

func (this *MasterRunner) close() {
	for _, ch := range this.taskInputs {
		close(ch)
	}
	close(this.Output)
	close(this.stopSignal)
}

func (this *MasterRunner) forwardGlobalMessage(message interface {
	helpers.Global
	taskInput.Message
}) {
	for _, ch := range this.taskInputs {
		ch <- message
	}
}

func (this *MasterRunner) Run() {
	defer this.close()
	for _, task := range this.Tasks {
		go task.Run()
	}

	go func() {
		for {
			select {

			case local, ok := <-this.LocalTasksOutput:
				if !ok {
					return
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
					return
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

	for {
		select {
		case <-this.stopSignal:
			return

		case req, ok := <-this.Input:
			if !ok {
				return
			}

			switch req.(type) {

			case input.Status:
				this.forwardGlobalMessage(taskInput.NewStatus())

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
				this.taskInputs[req.TaskId()] <- taskInput.NewStartProcess(req.ProcessId())

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
				this.taskInputs[req.TaskId()] <- taskInput.NewStopProcess(req.ProcessId())

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
				this.taskInputs[req.TaskId()] <- taskInput.NewRestartProcess(req.ProcessId())

			case input.Shutdown:
				this.forwardGlobalMessage(taskInput.NewShutdown())
				return

			case input.Reload:
				go func() {
					if err := this.ConfigManager.Load(); err != nil {
						this.Output <- output.NewReloadFailure(err.Error())
					}
				}()

			default:
				this.Output <- output.NewBadRequest()

			}
		}

	}
}
