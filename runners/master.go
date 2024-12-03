package runners

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"taskmaster/atom"
	"taskmaster/config"
	"taskmaster/messages/helpers"
	"taskmaster/messages/master/input"
	"taskmaster/messages/master/output"
	taskInput "taskmaster/messages/task/input"
	taskOutput "taskmaster/messages/task/output"
	"taskmaster/utils"
	"time"
)

type MasterRunner struct {
	ConfigManager config.Manager

	Input  <-chan input.Message
	Output chan<- output.Message

	Tasks             []*TaskRunner
	GlobalTasksOutput chan []taskOutput.Message

	taskInputs  *[]chan taskInput.Message
	taskOutputs *[]chan taskOutput.Message
	stopSignal  chan StopSignal

	tasksClosed        *sync.WaitGroup
	specificTaskClosed []*sync.WaitGroup
	masterClosed       *sync.WaitGroup

	shouldCloseOutput atom.Atom[bool]
	globalOutputPipes []chan taskOutput.Message

	reloadSignal chan os.Signal
}

type StopSignal struct{}

var TaskmasterLogFile = atom.NewAtom[*os.File](nil)

func NewMasterRunner(manager config.Manager, in <-chan input.Message, out chan<- output.Message) (*MasterRunner, error) {
	conf := manager.Get()

	taskInputs := make([]chan taskInput.Message, len(conf.Tasks))
	taskOutputs := make([]chan taskOutput.Message, len(conf.Tasks))

	globalTasksOutput := make(chan []taskOutput.Message)

	tasks := make([]*TaskRunner, len(conf.Tasks))

	instance := &MasterRunner{
		ConfigManager:      manager,
		Input:              in,
		Output:             out,
		Tasks:              tasks,
		GlobalTasksOutput:  globalTasksOutput,
		taskInputs:         &taskInputs,
		taskOutputs:        &taskOutputs,
		stopSignal:         make(chan StopSignal),
		tasksClosed:        new(sync.WaitGroup),
		specificTaskClosed: make([]*sync.WaitGroup, len(conf.Tasks)),
		masterClosed:       new(sync.WaitGroup),
		shouldCloseOutput:  atom.NewAtom(true),
		globalOutputPipes:  make([]chan taskOutput.Message, len(taskOutputs)),
		reloadSignal:       make(chan os.Signal),
	}
	if globalLogFile, err := os.OpenFile(fmt.Sprintf("%s/taskmaster_%s.log", conf.LogDir, time.Now().Format("060102_150405")), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666); err != nil {
		return nil, err
	} else {
		TaskmasterLogFile.Set(globalLogFile)
	}

	signal.Notify(instance.reloadSignal, syscall.SIGHUP)

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

	linkTaskOutput := func(idx uint, out <-chan taskOutput.Message) {
		for msg := range out {
			switch msg.(type) {
			case helpers.Global:
				instance.globalOutputPipes[idx] <- msg
			}
		}
	}

	for i, out := range taskOutputs {
		instance.globalOutputPipes[i] = make(chan taskOutput.Message)
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
				go instance.Run()
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
		chunk := make([]taskOutput.Message, len(instance.globalOutputPipes))
		for {
			for i, ch := range instance.globalOutputPipes {
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
	for _, ch := range this.globalOutputPipes {
		close(ch)
	}
	close(this.GlobalTasksOutput)
	TaskmasterLogFile.Get().Close()
	TaskmasterLogFile.Set(nil)
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
			this.specificTaskClosed[i].Done()
			this.tasksClosed.Done()
		}()
	}

	go func() {
		for global := range this.GlobalTasksOutput {
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
	}()

	reloadConfig := func() {
		if err := this.ConfigManager.Load(); err != nil {
			this.Output <- output.NewReloadFailure(err.Error())
		} else {
			this.Output <- output.NewReloadSuccess()
		}
	}

	for {
		select {

		case <-this.reloadSignal:
			go reloadConfig()

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
					break
				}
				(*this.taskInputs)[req.TaskId()] <- taskInput.NewStartProcess(req.ProcessId())

			case input.StopProcess:
				req := req.(input.StopProcess)
				if req.TaskId() >= uint(len(this.Tasks)) {
					break
				}
				(*this.taskInputs)[req.TaskId()] <- taskInput.NewStopProcess(req.ProcessId())

			case input.RestartProcess:
				req := req.(input.RestartProcess)
				if req.TaskId() >= uint(len(this.Tasks)) {
					break
				}
				(*this.taskInputs)[req.TaskId()] <- taskInput.NewRestartProcess(req.ProcessId())

			case input.Shutdown:
				this.forwardGlobalMessage(taskInput.NewShutdown())
				return

			case input.Reload:
				go reloadConfig()

			default:
				this.Output <- output.NewBadRequest()

			}
		}

	}
}
