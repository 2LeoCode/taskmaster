package runners

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"taskmaster/atom"
	"taskmaster/config"
	"taskmaster/messages/process/input"
	"taskmaster/messages/process/output"
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

type processState struct {
	startRetries  atom.Atom[*uint]
	userStartTime atom.Atom[*time.Time]
	startTime     atom.Atom[*time.Time]
	userStopTime  atom.Atom[*time.Time]
	stopTime      atom.Atom[*time.Time]
	exitStatus    atom.Atom[*int]
	hasBeenKilled atom.Atom[bool]
	command       atom.Atom[*exec.Cmd]
}

func (this *processState) Reset() {
	*this = *NewProcessState()
}

func NewProcessState() *processState {
	return &processState{
		startRetries:  atom.NewAtom[*uint](nil),
		userStartTime: atom.NewAtom[*time.Time](nil),
		startTime:     atom.NewAtom[*time.Time](nil),
		userStopTime:  atom.NewAtom[*time.Time](nil),
		stopTime:      atom.NewAtom[*time.Time](nil),
		exitStatus:    atom.NewAtom[*int](nil),
		hasBeenKilled: atom.NewAtom[bool](false),
		command:       atom.NewAtom[*exec.Cmd](nil),
	}
}

type ProcessRunner struct {
	ConfigManager config.Manager
	TaskConfig    config.Task

	TaskId uint
	Id     uint

	StdoutLogFile *os.File
	StderrLogFile *os.File

	Input  <-chan input.Message
	Output chan<- output.Message

	State *processState
}

func (this *ProcessRunner) close() {
	this.StopProcess()
	this.State.command.Get().Wait()
	this.StdoutLogFile.Close()
	this.StderrLogFile.Close()
}

func (this *ProcessRunner) initCommand(conf config.Task) {
	command := exec.Command(*conf.Command, conf.Arguments...)
	command.Dir = this.TaskConfig.WorkingDirectory
	command.Stdout = this.StdoutLogFile
	command.Stderr = this.StderrLogFile
	this.State.command.Set(command)
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

func newProcessRunner(manager config.Manager, taskId, id uint, input <-chan input.Message, output chan<- output.Message) (*ProcessRunner, error) {
	conf := manager.Get()
	taskConf := conf.Tasks[taskId]
	instance := &ProcessRunner{
		ConfigManager: manager,
		TaskConfig:    taskConf,
		TaskId:        taskId,
		Id:            id,
		Input:         input,
		Output:        output,
		State:         NewProcessState(),
	}
	if stdoutLogFile, err := getLogFile(STDOUT, taskConf.Stdout, conf, taskId, id); err != nil {
		return nil, err
	} else if stderrLogFile, err := getLogFile(STDERR, taskConf.Stderr, conf, taskId, id); err != nil {
		return nil, err
	} else {
		instance.StdoutLogFile = stdoutLogFile
		instance.StderrLogFile = stderrLogFile
	}
	instance.initCommand(taskConf)
	return instance, nil
}

func (this *ProcessRunner) RestartProcess() {
	command := this.State.command.Get()
	// TODO: do not forget umask for restartProcess
	command.Process.Signal(SIGNAL_TABLE[this.TaskConfig.StopSignal])
	go func() { //Proceed with the function when we know the process has stopped
		time.Sleep(time.Duration(this.TaskConfig.StopTime) * time.Millisecond)
		if exitStatus := this.State.exitStatus.Get(); exitStatus == nil {
			this.State.hasBeenKilled.Set(true)
			command.Process.Kill()
			command.Process.Wait()
		}
		//Remove trace of previous run to prevent conflict with current functions
		this.State.Reset()
		this.initCommand(this.TaskConfig)
		this.StartProcess()
	}()
}

func (this *ProcessRunner) StartProcess() {
	configStartTime := this.TaskConfig.StartTime
	this.State.startTime.Set(utils.New(time.Now()))
	command := this.State.command.Get()
	//Maybe restore this after running the process?
	// TODO: TaskCong.Permissions is NULL, uncomment this when this is fixed
	//syscall.Umask(int(*this.TaskConfig.Permissions))
	if err := command.Start(); err != nil {
		this.Output <- output.NewStartFailure(
			fmt.Sprintf("Command failed to run (%s)", err.Error()),
		)
	} else {
		go func() {
			command.Wait()
			this.State.exitStatus.Set(utils.New(
				this.State.command.Get().ProcessState.ExitCode(),
			))
			this.State.stopTime.Set(utils.New(time.Now()))
			if this.State.userStopTime.Get() != nil {
				this.Output <- output.NewStopSuccess(this.State.hasBeenKilled.Get())
			}
		}()

		go func() {
			time.Sleep(time.Duration(configStartTime) * time.Millisecond)
			if this.State.exitStatus.Get() == nil {
				this.Output <- output.NewStartSuccess()
				this.State.userStartTime.Set(utils.New(time.Now()))
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
	case this.State.stopTime.Get() != nil:
		expectedExitStatus := this.TaskConfig.ExpectedExitStatus
		actualExitStatus := this.State.exitStatus.Get()
		result := ""
		if *actualExitStatus == expectedExitStatus {
			result += "SUCCESS "
		} else {
			result += "FAILURE "
		}
		result += fmt.Sprint(*actualExitStatus)
		if this.State.hasBeenKilled.Get() {
			status += " KILLED"
		} else if this.State.userStopTime.Get() != nil {
			status += " STOPPED"
		}
	case this.State.startTime.Get() == nil:
		status += "NOT_STARTED"
	case this.State.userStartTime.Get() == nil:
		status += "STARTING"
	default:
		status += "RUNNING"
	}
	this.Output <- output.NewStatus(this.Id, status)
}

func (this *ProcessRunner) StopProcess() {
	this.State.userStopTime.Set(utils.New(time.Now()))
	command := this.State.command.Get()
	command.Process.Signal(SIGNAL_TABLE[this.TaskConfig.StopSignal]) // crash here
	go func() {
		println("waiting for process")
		time.Sleep(time.Duration(this.TaskConfig.StopTime) * time.Millisecond)
		if this.State.exitStatus.Get() != nil {
			return
		}
		this.State.hasBeenKilled.Set(true)
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
				println("PROCESS: received STATUS")
				this.StatusProcess()

			case input.Start:
				if this.State.userStopTime.Get() != nil {
					this.Output <- output.NewStartFailure(ERROR_START_STOPPED)
					break
				}
				if this.State.hasBeenKilled.Get() {
					this.Output <- output.NewStartFailure(ERROR_START_KILLED)
					break
				}
				if this.State.startTime.Get() != nil {
					this.Output <- output.NewStartFailure(ERROR_START_ALREADY_STARTED)
					break
				}
				this.StartProcess()

			case input.Stop:
				this.StopProcess()

			case input.Restart:
				this.RestartProcess()

			case input.Shutdown:
				defer func() { this.Output <- output.NewShutdown() }()
				return
			}
		}
	}
}
