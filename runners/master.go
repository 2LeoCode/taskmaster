package runners

import (
	"sync"
	"taskmaster/atom"
	"taskmaster/config"
	"taskmaster/messages/helpers"
	"taskmaster/messages/master/input"
	"taskmaster/messages/master/output"
	taskInput "taskmaster/messages/task/input"
	taskOutput "taskmaster/messages/task/output"
	"taskmaster/utils"
)

type MasterRunner struct {
	ConfigManager config.Manager

	Input  <-chan input.Message
	Output chan<- output.Message

	Tasks             []*TaskRunner
	LocalTasksOutput  chan utils.Pair[uint, taskOutput.Message]
	GlobalTasksOutput chan []taskOutput.Message

	taskInputs  *[]chan taskInput.Message
	taskOutputs *[]chan taskOutput.Message
	stopSignal  chan StopSignal

	tasksClosed        *sync.WaitGroup
	specificTaskClosed []*sync.WaitGroup
	masterClosed       *sync.WaitGroup

	shouldCloseOutput atom.Atom[bool]
}

type StopSignal struct{}

func NewMasterRunner(manager config.Manager, in <-chan input.Message, out chan<- output.Message) (*MasterRunner, error) {
	conf := manager.Get()

	taskInputs := make([]chan taskInput.Message, len(conf.Tasks))
	taskOutputs := make([]chan taskOutput.Message, len(conf.Tasks))

	localTasksOutput := make(chan utils.Pair[uint, taskOutput.Message])
	globalTasksOutput := make(chan []taskOutput.Message)

	tasks := make([]*TaskRunner, len(conf.Tasks))

	instance := &MasterRunner{
		ConfigManager:      manager,
		Input:              in,
		Output:             out,
		Tasks:              tasks,
		LocalTasksOutput:   localTasksOutput,
		GlobalTasksOutput:  globalTasksOutput,
		taskInputs:         &taskInputs,
		taskOutputs:        &taskOutputs,
		stopSignal:         make(chan StopSignal),
		tasksClosed:        new(sync.WaitGroup),
		specificTaskClosed: make([]*sync.WaitGroup, len(conf.Tasks)),
		masterClosed:       new(sync.WaitGroup),
		shouldCloseOutput:  atom.NewAtom(true),
	}

	instance.masterClosed.Add(1)

	for i := range instance.Tasks {
		instance.specificTaskClosed[i] = new(sync.WaitGroup)
		taskInputs[i] = make(chan taskInput.Message)
		taskOutputs[i] = make(chan taskOutput.Message)
		if task, err := newTaskRunner(manager, uint(i), taskInputs[i], taskOutputs[i]); err != nil {
			return nil, err
		} else {
			instance.Tasks[i] = task
		}
	}

	globalOutputs := make([]chan taskOutput.Message, len(taskOutputs))

	linkTaskOutput := func(idx uint, out <-chan taskOutput.Message) {
		for msg := range out {
			switch msg.(type) {
			case helpers.Local:
				localTasksOutput <- utils.NewPair(idx, msg)
			case helpers.Global:
				globalOutputs[idx] <- msg
			}
		}
	}

	for i, out := range taskOutputs {
		globalOutputs[i] = make(chan taskOutput.Message)
		go linkTaskOutput(uint(i), out)
	}

	manager.Subscribe(func(conf, prevConf *config.Config) error {
		if conf.LogDir != prevConf.LogDir || len(conf.Tasks) != len(prevConf.Tasks) {
			// Reload master
			instance.forwardGlobalMessage(taskInput.NewShutdown())
			instance.tasksClosed.Wait()
			instance.shouldCloseOutput.Set(false)
			instance.stopSignal <- StopSignal{}
			instance.masterClosed.Wait()
			if newInstance, err := NewMasterRunner(manager, in, out); err != nil {
				return err
			} else {
				*instance = *newInstance
			}
		} else {
			// Check which task needs to be reloaded

			for i := range conf.Tasks {
				if conf.Tasks[i].String() != prevConf.Tasks[i].String() {
					close(taskInputs[i])
					instance.tasksClosed.Add(1)
					instance.specificTaskClosed[i].Wait()
					instance.specificTaskClosed[i].Add(1)
					taskInputs[i] = make(chan taskInput.Message)
					taskOutputs[i] = make(chan taskOutput.Message)
					go linkTaskOutput(uint(i), taskOutputs[i])

					if task, err := newTaskRunner(manager, uint(i), taskInputs[i], taskOutputs[i]); err != nil {
						return err
					} else {
						instance.Tasks[i] = task
						go func() {
							task.Run()
							close(taskOutputs[i])
							instance.specificTaskClosed[i].Done()
							instance.tasksClosed.Done()
						}()
					}
				}
			}
		}
		return nil
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
	for _, ch := range *this.taskInputs {
		close(ch)
	}
	if this.shouldCloseOutput.Get() {
		close(this.Output)
	}
	close(this.stopSignal)
	this.tasksClosed.Wait()
	close(this.GlobalTasksOutput)
	close(this.LocalTasksOutput)
	this.masterClosed.Done()
}

func (this *MasterRunner) forwardGlobalMessage(message interface {
	helpers.Global
	taskInput.Message
}) {
	for _, ch := range *this.taskInputs {
		ch <- message
	}
}

func (this *MasterRunner) Run() {
	defer this.close()
	for i, task := range this.Tasks {
		this.tasksClosed.Add(1)
		this.specificTaskClosed[i].Add(1)
		go func() {
			task.Run()
			close((*this.taskOutputs)[i])
			this.specificTaskClosed[i].Done()
			this.tasksClosed.Done()
		}()
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
				(*this.taskInputs)[req.TaskId()] <- taskInput.NewStartProcess(req.ProcessId())

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
				(*this.taskInputs)[req.TaskId()] <- taskInput.NewStopProcess(req.ProcessId())

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
				(*this.taskInputs)[req.TaskId()] <- taskInput.NewRestartProcess(req.ProcessId())

			case input.Shutdown:
				this.forwardGlobalMessage(taskInput.NewShutdown())
				return

			case input.Reload:
				go func() {
					if err := this.ConfigManager.Load(); err != nil {
						this.Output <- output.NewReloadFailure(err.Error())
					} else {
						this.Output <- output.NewReloadSuccess()
					}
				}()

			default:
				this.Output <- output.NewBadRequest()

			}
		}

	}
}
