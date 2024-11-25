package runners

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"taskmaster/config"
	"taskmaster/messages/process/input"
	"taskmaster/messages/process/output"
	"taskmaster/state"
	"taskmaster/utils"
)

var SIGNAL_TABLE = map[string]os.Signal{
	"SIGHUP":    syscall.SIGHUP,
	"SIGINT":    syscall.SIGINT,
	"SIGQUIT":   syscall.SIGQUIT,
	"SIGILL":    syscall.SIGILL,
	"SIGTRAP":   syscall.SIGTRAP,
	"SIGABRT":   syscall.SIGABRT,
	"SIGFPE":    syscall.SIGFPE,
	"SIGKILL":   syscall.SIGKILL,
	"SIGBUS":    syscall.SIGBUS,
	"SIGSEGV":   syscall.SIGSEGV,
	"SIGSYS":    syscall.SIGSYS,
	"SIGPIPE":   syscall.SIGPIPE,
	"SIGALRM":   syscall.SIGALRM,
	"SIGTERM":   syscall.SIGTERM,
	"SIGUSR1":   syscall.SIGUSR1,
	"SIGUSR2":   syscall.SIGUSR2,
	"SIGCHLD":   syscall.SIGCHLD,
	"SIGPWR":    syscall.SIGPWR,
	"SIGWINCH":  syscall.SIGWINCH,
	"SIGURG":    syscall.SIGURG,
	"SIGPOLL":   syscall.SIGPOLL,
	"SIGSTOP":   syscall.SIGSTOP,
	"SIGTSTP":   syscall.SIGTSTP,
	"SIGCONT":   syscall.SIGCONT,
	"SIGTTIN":   syscall.SIGTTIN,
	"SIGTTOU":   syscall.SIGTTOU,
	"SIGVTALRM": syscall.SIGVTALRM,
	"SIGPROF":   syscall.SIGPROF,
	"SIGXCPU":   syscall.SIGXCPU,
	"SIGXFSZ":   syscall.SIGXFSZ,
}

const HINT_RESTART string = "Use [restart] instead if this is on purpose"
const ERROR_START_STOPPED string = "Process already completed. " + HINT_RESTART
const ERROR_START_KILLED string = "Process was killed. " + HINT_RESTART
const ERROR_START_ALREADY_STARTED string = "Process already started. " + HINT_RESTART
const ERROR_START_STOPPED_EARLY = "Process stopped before the configured time"

type ProcessState struct {
	StartRetries  *state.State[*uint]
	UserStartTime *state.State[*time.Time]
	StartTime     *state.State[*time.Time]
	UserStopTime  *state.State[*time.Time]
	StopTime      *state.State[*time.Time]
	ExitStatus    *state.State[*int]
	HasBeenKilled *state.State[bool]
	Command       *state.State[*exec.Cmd]
}

func (this *ProcessState) Reset() {
	*this = *NewProcessState()
}

func NewProcessState() *ProcessState {
	return &ProcessState{
		StartRetries:  state.NewState[*uint](nil),
		UserStartTime: state.NewState[*time.Time](nil),
		StartTime:     state.NewState[*time.Time](nil),
		UserStopTime:  state.NewState[*time.Time](nil),
		StopTime:      state.NewState[*time.Time](nil),
		ExitStatus:    state.NewState[*int](nil),
		HasBeenKilled: state.NewState(false),
		Command:       state.NewState[*exec.Cmd](nil),
	}
}

type ProcessRunner struct {
	ConfigManager *config.Manager
	TaskConfig    *config.Task

	TaskId uint
	Id     uint

	StdoutLogFile *os.File
	StderrLogFile *os.File

	Input  <-chan input.Message
	Output chan<- output.Message

	Events chan string

	State *ProcessState
}

func (this *ProcessRunner) close() {
	this.StopProcess()
	this.StdoutLogFile.Close()
	this.StderrLogFile.Close()
}

func (this *ProcessRunner) initCommand(conf *config.Task) {
	command := exec.Command(*conf.Command, conf.Arguments...)
	command.Dir = conf.WorkingDirectory
	command.Stdout = this.StdoutLogFile
	command.Stderr = this.StderrLogFile
	this.State.Command.Set(command)
}

type OutputSource int

const (
	STDOUT OutputSource = iota
	STDERR
)

func getLogFile(source OutputSource, logConfig string, conf *config.Config, taskId, processId uint) (*os.File, error) {
	if logConfig == "inherit" {
		switch source {
		case STDOUT:
			return os.Stdout, nil
		case STDERR:
			return os.Stderr, nil
		}
		return nil, errors.New("source must be either STDOUT or STDERR")
	}

	var logFileName string

	switch logConfig {
	case "redirect":
		logFileName = fmt.Sprintf(
			"%s/%d-%d_%s-stdout.log",
			conf.LogDir,
			taskId,
			processId,
			time.Now().Format("060102_030405"),
		)
	case "ignore":
		logFileName = "/dev/null"
	default:
		return nil, fmt.Errorf("Invalid log configuration %s", logConfig)
	}

	return os.OpenFile(
		logFileName,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		os.ModeAppend.Perm(),
	)
}

func newProcessRunner(manager *config.Manager, taskConfig *config.Task, taskId, id uint, input <-chan input.Message, output chan<- output.Message) (*ProcessRunner, error) {
	instance := &ProcessRunner{
		ConfigManager: manager,
		TaskConfig:    taskConfig,
		TaskId:        taskId,
		Id:            id,
		Input:         input,
		Output:        output,
		Events:        make(chan string),
		State:         NewProcessState(),
	}
	if err := config.Use(manager, func(conf *config.Config) error {
		if stdoutLogFile, err := getLogFile(STDOUT, taskConfig.Stdout, conf, taskId, id); err != nil {
			return err
		} else if stderrLogFile, err := getLogFile(STDERR, taskConfig.Stderr, conf, taskId, id); err != nil {
			return err
		} else {
			instance.StdoutLogFile = stdoutLogFile
			instance.StderrLogFile = stderrLogFile
		}
		instance.initCommand(taskConfig)
		return nil
	}); err != nil {
		return nil, err
	}

	return instance, nil
}

func (this *ProcessRunner) RestartProcess() {
	command := this.State.Command.Get()
	command.Process.Signal(SIGNAL_TABLE[this.TaskConfig.StopSignal])
	go func() { //Proceed with the function when we know the process has stopped
		time.Sleep(time.Duration(this.TaskConfig.StopTime) * time.Millisecond)
		if exitStatus := this.State.ExitStatus.Get(); exitStatus == nil {
			this.State.HasBeenKilled.Set(true)
			command.Process.Kill()
			command.Process.Wait()
		}
		//Remove trace of previous run to prevent conflict with current functions
		this.State.Reset()
		this.initCommand(nil)
		this.StartProcess()
	}()
}

func (this *ProcessRunner) StartProcess() {
	configStartTime := this.TaskConfig.StartTime
	this.State.StartTime.Set(utils.New(time.Now()))
	command := this.State.Command.Get()
	if err := command.Start(); err != nil {
		this.Output <- output.NewStartFailure(
			fmt.Sprintf("Command failed to run (%s)", err.Error()),
		)
	} else {
		go func() {
			command.Wait()
			this.State.ExitStatus.Set(utils.New(
				this.State.Command.Get().ProcessState.ExitCode(),
			))
			this.State.StopTime.Set(utils.New(time.Now()))
			if this.State.UserStopTime.Get() != nil {
				this.Output <- output.NewStopSuccess(state.Use(this.State.HasBeenKilled, utils.Get))
			}
		}()

		go func() {
			time.Sleep(time.Duration(configStartTime) * time.Millisecond)
			if this.State.ExitStatus.Get() == nil {
				this.Output <- output.NewStartSuccess()
				this.State.UserStartTime.Set(utils.New(time.Now()))
			} else {
				// TODO: Handle retry attempts
				this.Output <- output.NewStartFailure(ERROR_START_STOPPED_EARLY)
			}
		}()
	}

}

func (this *ProcessRunner) StatusProcess() {
	status := ""
	switch {
	case this.State.StopTime.Get() != nil:
		expectedExitStatus := this.TaskConfig.ExpectedExitStatus
		status += state.Use(this.State.ExitStatus, func(value *int) string {
			result := ""
			if *value == expectedExitStatus {
				result += "SUCCESS "
			} else {
				result += "FAILURE "
			}
			result += fmt.Sprint(*value)
			return result
		})
		if this.State.HasBeenKilled.Get() {
			status += " KILLED"
		} else if this.State.UserStopTime.Get() != nil {
			status += " STOPPED"
		}
	case this.State.StartTime.Get() == nil:
		status += "NOT_STARTED"
	case this.State.UserStartTime.Get() == nil:
		status += "STARTING"
	default:
		status += "RUNNING"
	}
	this.Output <- output.NewStatus(this.Id, status)
}

func (this *ProcessRunner) StopProcess() {
	this.State.UserStopTime.Set(utils.New(time.Now()))
	command := this.State.Command.Get()
	command.Process.Signal(SIGNAL_TABLE[this.TaskConfig.StopSignal])
	go func() {
		time.Sleep(time.Duration(this.TaskConfig.StopTime) * time.Millisecond)
		if this.State.ExitStatus.Get() != nil {
			return
		}
		this.State.HasBeenKilled.Set(true)
		command.Process.Kill()
	}()
}

func (this *ProcessRunner) Run() {
	defer this.close()
	for {
		if req, ok := <-this.Input; !ok {
			// Input channel has been closed
			return
		} else {
			switch req.(type) {

			case input.Status:
				this.StatusProcess()

			case input.Start:
				if this.State.UserStopTime.Get() != nil {
					this.Output <- output.NewStartFailure(ERROR_START_STOPPED)
					break
				}
				if this.State.HasBeenKilled.Get() {
					this.Output <- output.NewStartFailure(ERROR_START_KILLED)
					break
				}
				if this.State.StartTime.Get() != nil {
					this.Output <- output.NewStartFailure(ERROR_START_ALREADY_STARTED)
					break
				}
				this.StartProcess()
			case input.Stop:
				this.StopProcess()
			case input.Restart:
				this.RestartProcess()
			case input.Shutdown:
				return
			}
		}
	}
}
