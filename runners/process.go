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
const ERROR_PREVIOUSLY_FAILED = "Process previously failed to start, check your configuration"
const ERROR_STOP_STOPPED = "Process already stopped"

type processState struct {
	startRetries    atom.Atom[uint]
	userStartTime   atom.Atom[*time.Time]
	startTime       atom.Atom[*time.Time]
	userStopTime    atom.Atom[*time.Time]
	stopTime        atom.Atom[*time.Time]
	exitStatus      atom.Atom[*int]
	hasBeenKilled   atom.Atom[bool]
	command         atom.Atom[*exec.Cmd]
	hasBeenShutdown atom.Atom[bool]
	stoppedEarly    atom.Atom[bool]
	failedToStart   atom.Atom[bool]
	isRestarting    atom.Atom[bool]
}

func (this *processState) Reset() {
	*this = processState{
		userStartTime:   atom.NewAtom[*time.Time](nil),
		startTime:       atom.NewAtom[*time.Time](nil),
		userStopTime:    atom.NewAtom[*time.Time](nil),
		stopTime:        atom.NewAtom[*time.Time](nil),
		exitStatus:      atom.NewAtom[*int](nil),
		command:         atom.NewAtom[*exec.Cmd](nil),
		failedToStart:   atom.NewAtom(false),
		hasBeenKilled:   atom.NewAtom(false),
		isRestarting:    this.isRestarting,
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
		startRetries:    atom.NewAtom[uint](0),
		failedToStart:   atom.NewAtom(false),
		hasBeenKilled:   atom.NewAtom(false),
		hasBeenShutdown: atom.NewAtom(false),
		stoppedEarly:    atom.NewAtom(false),
		isRestarting:    atom.NewAtom(false),
	}
}

type ProcessResponse uint

const (
	STARTING ProcessResponse = iota
	STARTING_ERROR
	RESTARTING_ERROR
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
	stopSignal     chan StopSignal
}

func (this *ProcessRunner) close() {
	this.State.hasBeenShutdown.Set(true)

	if !this.State.failedToStart.Get() {
		this.StopProcess()
		<-this.stopSignal
	}
	close(this.Output)
	this.StdoutLogFile.Close()
	this.StderrLogFile.Close()
}

func (this *ProcessRunner) initCommand() {
	command := exec.Command(*this.TaskConfig.Command, this.TaskConfig.Arguments...)

	for k, v := range maps.All(this.TaskConfig.Environment) {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
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
		stopSignal:     make(chan StopSignal),
	}
	if stdoutLogFile, err := getLogFile(STDOUT, taskConf.Stdout, conf, taskId, id); err != nil {
		return nil, err
	} else if stderrLogFile, err := getLogFile(STDERR, taskConf.Stderr, conf, taskId, id); err != nil {
		return nil, err
	} else {
		instance.StdoutLogFile = stdoutLogFile
		instance.StderrLogFile = stderrLogFile
	}
	return instance, nil
}

func (this *ProcessRunner) StartProcess() error {
	this.initCommand()

	this.State.startTime.Set(utils.New(time.Now()))
	this.State.exitStatus.Set(nil)
	command := this.State.command.Get()

	oldUmask := syscall.Umask(int(*this.TaskConfig.Permissions))
	err := command.Start()
	syscall.Umask(oldUmask)

	if err != nil {
		this.State.failedToStart.Set(true)
		return err
	} else {
		go func() {
			retry := func() {
				this.State.startRetries.Update(func(old uint) uint { return old + 1 })
				this.State.Reset()
				this.internalOutput <- RETRYING
				this.StartProcess()
			}

			state, _ := command.Process.Wait()
			exitCode := state.ExitCode()
			this.State.exitStatus.Set(utils.New(
				exitCode,
			))
			this.State.stopTime.Set(utils.New(time.Now()))
			hasAttempts := this.TaskConfig.RestartAttempts == 0 ||
				this.State.startRetries.Get() < this.TaskConfig.RestartAttempts
			if exitCode == this.TaskConfig.ExpectedExitStatus {
				this.internalOutput <- STOPPED_SUCCESSFULLY
			} else {
				this.internalOutput <- STOPPED_UNSUCCESSFULLY
			}
			if !this.State.isRestarting.Get() && !this.State.hasBeenShutdown.Get() {
				if this.State.userStartTime.Get() == nil {
					this.State.stoppedEarly.Set(true)
					if hasAttempts && this.TaskConfig.Restart != "never" {
						retry()
					}
				} else if hasAttempts && (this.TaskConfig.Restart == "always" ||
					(this.TaskConfig.Restart == "unless-stopped" &&
						this.State.userStopTime.Get() == nil &&
						exitCode != this.TaskConfig.ExpectedExitStatus)) {
					retry()
				}
			}
			this.State.isRestarting.Set(false)
			this.stopSignal <- StopSignal{}
		}()

		go func() {
			time.Sleep(time.Duration(this.TaskConfig.StartTime) * time.Millisecond)
			this.State.userStartTime.Set(utils.New(time.Now()))
			if this.State.exitStatus.Get() == nil {
				this.State.stoppedEarly.Set(false)
				this.internalOutput <- STARTED
			}
		}()
	}
	return nil
}

func (this *ProcessRunner) StopProcess() {
	if process := this.State.command.Get().Process; process != nil {
		this.State.userStopTime.Set(utils.New(time.Now()))
		process.Signal(SIGNAL_TABLE[this.TaskConfig.StopSignal])
		go func() {
			time.Sleep(time.Duration(this.TaskConfig.StopTime) * time.Millisecond)
			if this.State.exitStatus.Get() == nil {
				this.State.hasBeenKilled.Set(true)
				process.Kill()
			}
		}()
	}
}

func (this *ProcessRunner) RestartProcess() {
	this.StopProcess()
	go func() {
		<-this.stopSignal
		this.State.Reset()
		if err := this.StartProcess(); err != nil {
			this.internalOutput <- RESTARTING_ERROR
			this.commandErrors <- err
		}
	}()
}

func (this *ProcessRunner) StatusProcess() {
	status := ""
	stopTime := this.State.stopTime.Get()
	exitStatus := this.State.exitStatus.Get()
	switch {
	case this.State.failedToStart.Get():
		status += "FAILED_TO_START"
	case stopTime != nil && exitStatus != nil:
		expectedExitStatus := this.TaskConfig.ExpectedExitStatus
		status += ""
		if *exitStatus == expectedExitStatus {
			status += "SUCCESS "
		} else {
			status += "FAILURE "
		}
		status += fmt.Sprint(*exitStatus)
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
	retries := this.State.startRetries.Get()
	if this.TaskConfig.RestartAttempts != 0 && retries != 0 {
		status += fmt.Sprintf(" [Restarted %d/%d]", retries, this.TaskConfig.RestartAttempts)
	}
	if this.State.stoppedEarly.Get() {
		status += " (stopped early)"
	}
	this.Output <- output.NewStatus(this.Id, status)
}

func (this *ProcessRunner) Run() {
	defer this.close()
	go func() {
		actionError := func(action string) string {
			err := <-this.commandErrors
			return fmt.Sprintf("failed to %s: %s", action, err.Error())
		}
		for req := range this.internalOutput {
			msg := fmt.Sprintf("[%d - %d] ", this.TaskId, this.Id)
			switch req {
			case STARTING:
				msg += "starting"
			case RESTARTING:
				msg += "restarting"
			case STARTING_ERROR:
				msg += actionError("start")
			case RESTARTING_ERROR:
				msg += actionError("restart")
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
			this.internalOutput <- STARTING_ERROR
			this.commandErrors <- err
		}
	}
	for req := range this.Input {
		switch req.(type) {

		case input.Status:
			this.StatusProcess()

		case input.Start:
			var err *string
			if this.State.failedToStart.Get() {
				err = utils.New(ERROR_PREVIOUSLY_FAILED)
			} else if this.State.hasBeenKilled.Get() {
				err = utils.New(ERROR_START_KILLED)
			} else if this.State.startTime.Get() != nil {
				err = utils.New(ERROR_START_ALREADY_STARTED)
			} else if this.State.userStopTime.Get() != nil {
				err = utils.New(ERROR_START_STOPPED)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "\r \rCannot start process: %s\n> ", *err)
				break
			}
			this.internalOutput <- STARTING
			if err := this.StartProcess(); err != nil {
				this.internalOutput <- STARTING_ERROR
				this.commandErrors <- err
			}

		case input.Stop:
			var err *string
			if this.State.failedToStart.Get() {
				err = utils.New(ERROR_PREVIOUSLY_FAILED)
			} else if this.State.startTime.Get() == nil || this.State.exitStatus.Get != nil {
				err = utils.New(ERROR_STOP_STOPPED)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "\r \rCannot stop process: %s\n> ", *err)
				break
			}
			this.internalOutput <- STOPPING
			this.StopProcess()
			go func() { <-this.stopSignal }()

		case input.Restart:
			if this.State.failedToStart.Get() {
				fmt.Fprintf(os.Stderr, "\r \rCannot restart process: %s\n> ", ERROR_PREVIOUSLY_FAILED)
				break
			}
			this.internalOutput <- RESTARTING
			this.State.isRestarting.Set(true)
			this.RestartProcess()

		case input.Shutdown:
			return
		}
	}
}
