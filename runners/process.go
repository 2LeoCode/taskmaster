package runners

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"taskmaster/config"
	"taskmaster/config/manager"
	"taskmaster/messages/process/input"
	"taskmaster/messages/process/output"
	"taskmaster/state"
	"time"
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

type ProcessRunner struct {
	Id uint

	StdoutLogFile *os.File
	StderrLogFile *os.File

	Input  chan input.Message
	Output chan output.Message

	Command exec.Cmd

	Events chan string

	ConfigManager *configManager.Task
	StartRetries  *uint
	UserStartTime *time.Time
	StartTime     *time.Time
	UserStopTime  *time.Time
	StopTime      *time.Time
	ExitStatus    *int

	HasBeenKilled bool
}

func (this *ProcessRunner) close() {
	this.StdoutLogFile.Close()
	this.StderrLogFile.Close()
}

func newProcessRunner(manager *configManager.Task, id uint, input chan input.Message, output chan output.Message) (*ProcessRunner, error) {
	instance := &ProcessRunner{
		Id:            id,
		Input:         input,
		Output:        output,
		ConfigManager: manager,
		Events:        make(chan string),
	}
	if err := configManager.UseTask(manager, func(config *config.Config, taskId uint) error {
		if stdoutLogFile, err := os.OpenFile(
			fmt.Sprintf(
				"%s/%d-%d_%s-stdout.log",
				config.LogDir,
				taskId,
				instance.Id,
				time.Now().Format("060102_030405"),
			),
			os.O_WRONLY|os.O_APPEND|os.O_CREATE,
			os.ModeAppend.Perm(),
		); err != nil {
			return err
		} else if stderrLogFile, err := os.OpenFile(
			fmt.Sprintf(
				"%s/%d-%d_%s-stderr.log",
				config.LogDir,
				taskId,
				instance.Id,
				time.Now().Format("060102_030405"),
			),
			os.O_WRONLY|os.O_APPEND|os.O_CREATE,
			os.ModeAppend.Perm(),
		); err != nil {
			return err
		} else {
			instance.StdoutLogFile = stdoutLogFile
			instance.StderrLogFile = stderrLogFile
		}
		instance.Command = *exec.Command(*config.Tasks[taskId].Command, config.Tasks[taskId].Arguments...)
		instance.Command.Stdout = instance.StdoutLogFile
		instance.Command.Stderr = instance.StderrLogFile
		return nil
	}); err != nil {
		return nil, err
	}

	manager.Master.Subscribe(func(config, prev *config.Config) state.StateCleanupFn {
		// On config reload, process specific actions, to implement here
		return func() {}
	})

	return instance, nil
}

func (this *ProcessRunner) RestartProcess() {
	configStopSignal, configStopTime := func() (string, uint) {
		type fieldsType struct {
			stopSignal string
			stopTime   uint
		}
		fields := configManager.UseTask(this.ConfigManager, func(config *config.Config, taskId uint) fieldsType {
			return fieldsType{
				config.Tasks[taskId].StopSignal,
				config.Tasks[taskId].StopTime,
			}
		})
		return fields.stopSignal, fields.stopTime
	}()
	this.Command.Process.Signal(SIGNAL_TABLE[configStopSignal])
	go func() {
		time.Sleep(time.Duration(configStopTime) * time.Millisecond)
		if this.ExitStatus == nil {
		this.HasBeenKilled = true
		this.Command.Process.Kill()
		}
	this.Command.Wait()
	println("Process stopped")
	//Remove trace of previous run to prevent conflict with current functions
	//This will need to be fixed
	this.UserStopTime = nil
	this.HasBeenKilled = false
	this.StopTime = nil  
	this.ExitStatus = nil
	this.UserStartTime = nil
	this.StartTime = nil
	configManager.UseTask(this.ConfigManager, func(conf *config.Config, taskId uint)  int {
		task := conf.Tasks[taskId]
		this.Command = *exec.Command(*task.Command, task.Arguments...)
		return 0
	})
	this.StartProcess()

	}()
}
func (this *ProcessRunner) StartProcess() {
	configStartTime := configManager.UseTask(this.ConfigManager, func(conf *config.Config, taskId uint) uint {
		return conf.Tasks[this.Id].StartTime
	})
	this.StartTime = new(time.Time)
	*this.StartTime = time.Now()
	if err := this.Command.Start(); err != nil {
		this.Output <- output.NewStartFailure(
			fmt.Sprintf("Command failed to run (%s)", err.Error()),
		)
	} else {
		go func() {
			this.Command.Wait()
			this.Events <- "STOP"
		}()

		go func() {
			time.Sleep(time.Duration(configStartTime) * time.Millisecond)
			this.Events <- "START"
		}()
	}

}

func (this *ProcessRunner) StopProcess() {
	configStopSignal, configStopTime := func() (string, uint) {
		type fieldsType struct {
			stopSignal string
			stopTime   uint
		}
		fields := configManager.UseTask(this.ConfigManager, func(config *config.Config, taskId uint) fieldsType {
			return fieldsType{
				config.Tasks[taskId].StopSignal,
				config.Tasks[taskId].StopTime,
			}
		})
		return fields.stopSignal, fields.stopTime
	}()
	this.UserStopTime = new(time.Time)
	*this.UserStopTime = time.Now()
	this.Command.Process.Signal(SIGNAL_TABLE[configStopSignal])
	go func() {
		time.Sleep(time.Duration(configStopTime) * time.Millisecond)
		this.Events <- "KILL_IF_ALIVE"
	}()

}

func (this *ProcessRunner) Run() {
	defer this.close()

	for {
		select {

		case event := <-this.Events:
			switch event {

			case "START":
				if this.ExitStatus == nil {
					println("Should start?")
					this.Output <- output.NewStartSuccess()
					this.UserStartTime = new(time.Time)
					*this.UserStartTime = time.Now()
				} else {
					this.Output <- output.NewStartFailure("Process exited before its start time")
				}

			case "STOP":
				this.ExitStatus = new(int)
				*this.ExitStatus = this.Command.ProcessState.ExitCode()
				this.StopTime = new(time.Time)
				*this.StopTime = time.Now()
				if this.UserStopTime != nil {
					this.Output <- output.NewStopSuccess(this.HasBeenKilled)
				}

			case "KILL_IF_ALIVE":
				if this.ExitStatus != nil {
					break
				}
				this.HasBeenKilled = true
				this.Command.Process.Kill()
			}

		case req := <-this.Input:
			switch req.(type) {

			case input.Status:
				status := ""
				switch {
				case this.StopTime != nil:
					expectedExitStatus := configManager.UseTask(this.ConfigManager, func(conf *config.Config, taskId uint) int { return conf.Tasks[taskId].ExpectedExitStatus })
					if *this.ExitStatus == expectedExitStatus {
						status += "SUCCESS "
					} else {
						status += "FAILURE "
					}
					status += fmt.Sprint(*this.ExitStatus)
					if this.HasBeenKilled {
						status += " KILLED"
					} else if this.UserStopTime != nil {
						status += " STOPPED"
					}
				case this.StartTime == nil:
					status = "NOT_STARTED"
				case this.UserStartTime == nil:
					status = "STARTING"
				default:
					status = "RUNNING"
				}
				this.Output <- output.NewStatus(this.Id, status)

			case input.Start:
				if this.UserStopTime != nil {
					this.Output <- output.NewStartFailure("Process already completed. Use [restart] instead if you really want to start it.")
					break
				} else if this.HasBeenKilled {
					this.Output <- output.NewStartFailure("Process was killed. Use [restart] instead if you really want to start it.")
					break
				} else if this.StartTime != nil  {
					this.Output <- output.NewStartFailure("Process already started")
					break
				}
				this.StartProcess()
			case input.Stop:
				this.StopProcess()
			case input.Restart:
				this.RestartProcess()
			case input.Shutdown:
			}
		}
	}
}
