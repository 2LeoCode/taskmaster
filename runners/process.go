package runners

import (
	"errors"
	"fmt"
	"maps"
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
	startRetries    atom.Atom[*uint]
	userStartTime   atom.Atom[*time.Time]
	startTime       atom.Atom[*time.Time]
	userStopTime    atom.Atom[*time.Time]
	stopTime        atom.Atom[*time.Time]
	exitStatus      atom.Atom[*int]
	hasBeenKilled   atom.Atom[bool]
	command         atom.Atom[*exec.Cmd]
	hasBeenShutdown atom.Atom[bool]
	stoppedEarly    atom.Atom[bool]
}

func (this *processState) Reset() {
	*this = processState{
		userStartTime:   atom.NewAtom[*time.Time](nil),
		startTime:       atom.NewAtom[*time.Time](nil),
		userStopTime:    atom.NewAtom[*time.Time](nil),
		stopTime:        atom.NewAtom[*time.Time](nil),
		exitStatus:      atom.NewAtom[*int](nil),
		command:         atom.NewAtom[*exec.Cmd](nil),
		hasBeenKilled:   atom.NewAtom(false),
		startRetries:    this.startRetries,
		hasBeenShutdown: this.hasBeenShutdown,
		stoppedEarly:    this.stoppedEarly,
	}
}

func NewProcessState() *processState {
	return &processState{
		userStartTime:   atom.NewAtom[*time.Time](nil),
		startTime:       atom.NewAtom[*time.Time](nil),
		userStopTime:    atom.NewAtom[*time.Time](nil),
		stopTime:        atom.NewAtom[*time.Time](nil),
		exitStatus:      atom.NewAtom[*int](nil),
		command:         atom.NewAtom[*exec.Cmd](nil),
		hasBeenKilled:   atom.NewAtom(false),
		startRetries:    atom.NewAtom[*uint](nil),
		hasBeenShutdown: atom.NewAtom(false),
		stoppedEarly:    atom.NewAtom(false),
	}
}

type ProcessResponse uint

const (
	STARTING ProcessResponse = iota
	STARTING_ERROR
	RESTARTING
	RETRYING
	STARTED
	STOPPED_EARLY
	STOPPING
	STOPPED_SUCCESSFULLY
	STOPPED_UNSUCCESSFULLY
)

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

	internalOutput chan ProcessResponse
	commandErrors  chan error
}

func (this *ProcessRunner) close() {
	this.State.hasBeenShutdown.Set(true)
	this.StopProcess()
	close(this.Output)
	this.State.command.Get().Wait()
	this.StdoutLogFile.Close()
	this.StderrLogFile.Close()
}

func (this *ProcessRunner) initCommand() {
	command := exec.Command(*this.TaskConfig.Command, this.TaskConfig.Arguments...)

	command.Env = make([]string, len(this.TaskConfig.Environment))
	i := 0
	for k, v := range maps.All(this.TaskConfig.Environment) {
		command.Env[i] = fmt.Sprintf("%s=%s", k, v)
	}

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
	getSourceName := func() string {
		switch source {
		case STDOUT:
			return "stdout"
		case STDERR:
			return "stderr"
		}
		return ""
	}

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
			"%s/%d-%d_%s-%s.log",
			conf.LogDir,
			taskId,
			processId,
			time.Now().Format("060102_150405"),
			getSourceName(),
		)
	case "ignore":
		logFileName = "/dev/null"
	default:
		return nil, fmt.Errorf("Invalid log configuration %s", logConfig)
	}

	return os.OpenFile(
		logFileName,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0o666,
	)
}

func newProcessRunner(manager config.Manager, taskId, id uint, input <-chan input.Message, output chan<- output.Message) (*ProcessRunner, error) {
	conf := manager.Get()
	taskConf := conf.Tasks[taskId]
	instance := &ProcessRunner{
		ConfigManager:  manager,
		TaskConfig:     taskConf,
		TaskId:         taskId,
		Id:             id,
		Input:          input,
		Output:         output,
		State:          NewProcessState(),
		internalOutput: make(chan ProcessResponse),
		commandErrors:  make(chan error),
	}
	if stdoutLogFile, err := getLogFile(STDOUT, taskConf.Stdout, conf, taskId, id); err != nil {
		return nil, err
	} else if stderrLogFile, err := getLogFile(STDERR, taskConf.Stderr, conf, taskId, id); err != nil {
		return nil, err
	} else {
		instance.StdoutLogFile = stdoutLogFile
		instance.StderrLogFile = stderrLogFile
	}
	instance.State.startRetries.Set(utils.New(uint(0)))
	return instance, nil
}

func (this *ProcessRunner) StartProcess() error {
	this.initCommand()
	configStartTime := this.TaskConfig.StartTime
	if this.State.hasBeenKilled.Get() {
		return errors.New(ERROR_START_KILLED)
	} else if this.State.startTime.Get() != nil {
		return errors.New(ERROR_START_ALREADY_STARTED)
	} else if this.State.userStopTime.Get() != nil {
		return errors.New(ERROR_START_STOPPED)
	}

	this.State.startTime.Set(utils.New(time.Now()))
	this.State.exitStatus.Set(nil)
	command := this.State.command.Get()

	oldUmask := syscall.Umask(int(*this.TaskConfig.Permissions))
	err := command.Start()
	syscall.Umask(oldUmask)

	if err != nil {
		this.internalOutput <- STARTING_ERROR
		this.commandErrors <- err
	} else {
		go func() {
			state, _ := command.Process.Wait()
			exitCode := state.ExitCode()
			this.State.exitStatus.Set(utils.New(
				exitCode,
			))
			this.State.stopTime.Set(utils.New(time.Now()))
			if exitCode == this.TaskConfig.ExpectedExitStatus {
				this.internalOutput <- STOPPED_SUCCESSFULLY
			} else {
				this.internalOutput <- STOPPED_UNSUCCESSFULLY
			}
			if this.State.userStartTime.Get() == nil {
				this.State.stoppedEarly.Set(true)
				if this.TaskConfig.Restart != "never" && (this.TaskConfig.RestartAttempts == 0 || *this.State.startRetries.Get() < this.TaskConfig.RestartAttempts) {
					this.State.startRetries.Update(func(old *uint) *uint { return utils.New(*old + 1) })
					this.State.Reset()
					this.internalOutput <- RETRYING
					this.StartProcess()
					return
				}
			}
			if (*this.State.exitStatus.Get() != this.TaskConfig.ExpectedExitStatus && this.TaskConfig.Restart != "never") || this.TaskConfig.Restart == "always" {

				if !(this.TaskConfig.Restart == "unless-stopped" && this.State.userStopTime.Get() != nil) {
					if this.TaskConfig.RestartAttempts == 0 || *this.State.startRetries.Get() < this.TaskConfig.RestartAttempts {
						this.State.startRetries.Update(func(old *uint) *uint { return utils.New(*old + 1) })
						this.State.Reset()
						this.internalOutput <- RETRYING
						this.StartProcess()
						return
					}
				}
			}
		}()

		go func() {
			time.Sleep(time.Duration(configStartTime) * time.Millisecond)
			this.State.userStartTime.Set(utils.New(time.Now()))
			if this.State.exitStatus.Get() == nil {
				this.internalOutput <- STARTED
				this.State.stoppedEarly.Set(false)
			}
		}()
	}
	return nil
}

func (this *ProcessRunner) StopProcess() error {
	command := this.State.command.Get()
	if this.State.startTime.Get() == nil || command.Process == nil {
		return errors.New("process is not started")
	}
	this.State.userStopTime.Set(utils.New(time.Now()))
	command.Process.Signal(SIGNAL_TABLE[this.TaskConfig.StopSignal])
	go func() {
		time.Sleep(time.Duration(this.TaskConfig.StopTime) * time.Millisecond)
		if this.State.exitStatus.Get() != nil {
			return
		}
		this.State.hasBeenKilled.Set(true)
		command.Process.Kill()
	}()
	return nil
}

func (this *ProcessRunner) RestartProcess() error {
	if err := this.StopProcess(); err != nil {
		return err
	}
	command := this.State.command.Get()
	go func() {
		command.Wait()
		this.State.Reset()
		this.StartProcess()
	}()
	return nil
}

func (this *ProcessRunner) StatusProcess() {
	status := ""
	switch {
	case this.State.stopTime.Get() != nil:
		expectedExitStatus := this.TaskConfig.ExpectedExitStatus
		actualExitStatus := this.State.exitStatus.Get()
		status += ""
		if *actualExitStatus == expectedExitStatus {
			status += "SUCCESS "
		} else {
			status += "FAILURE "
		}
		status += fmt.Sprint(*actualExitStatus)
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
	if this.TaskConfig.RestartAttempts > 0 {
		status += fmt.Sprintf(" [Restarted %d/%d]", *this.State.startRetries.Get(), this.TaskConfig.RestartAttempts)
	}
	if this.State.stoppedEarly.Get() {
		status += " (stopped early)"
	}
	this.Output <- output.NewStatus(this.Id, status)
}

func (this *ProcessRunner) Run() {
	defer this.close()
	go func() {
		for req := range this.internalOutput {
			msg := fmt.Sprintf("[%d - %d] ", this.TaskId, this.Id)
			switch req {
			case STARTING:
				msg += "starting"
			case RESTARTING:
				msg += "restarting"
			case STARTING_ERROR:
				err := <-this.commandErrors
				msg += fmt.Sprintf("failed to start (%s)", err.Error())
			case RETRYING:
				msg += "retrying"
			case STARTED:
				msg += "started"
			case STOPPING:
				msg += "stopping"
			case STOPPED_SUCCESSFULLY:
				msg += "stopped successfully"
			case STOPPED_UNSUCCESSFULLY:
				msg += "stopped unsuccessfully"
			}
			fmt.Printf("\r \r%s\n> ", msg)
			fmt.Fprintf(TaskmasterLogFile.Get(), "%s: %s\n", time.Now().Format("06/01/02 15:04:05"), msg)
		}
	}()
	if this.TaskConfig.StartAtLaunch {
		this.internalOutput <- STARTING
		if err := this.StartProcess(); err != nil {
			this.Output <- output.NewStartFailure(err.Error())
		} else {
			this.Output <- output.NewStartSuccess()
		}
	}
	for req := range this.Input {
		switch req.(type) {

		case input.Status:
			this.StatusProcess()

		case input.Start:
			this.internalOutput <- STARTING
			if err := this.StartProcess(); err != nil {
				this.Output <- output.NewStartFailure(err.Error())
			} else {
				this.Output <- output.NewStartSuccess()
			}

		case input.Stop:
			this.internalOutput <- STOPPING
			if err := this.StopProcess(); err != nil {
				this.Output <- output.NewStopFailure(err.Error())
			} else {
				this.Output <- output.NewStopSuccess()
			}

		case input.Restart:
			this.internalOutput <- RESTARTING
			if err := this.RestartProcess(); err != nil {
				this.Output <- output.NewRestartFailure(err.Error())
			} else {
				this.Output <- output.NewRestartSuccess()
			}

		case input.Shutdown:
			return
		}
	}
}
