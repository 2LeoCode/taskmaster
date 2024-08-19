package runners

import (
	"fmt"
	"os"
	"os/exec"
	"taskmaster/config"
	"taskmaster/config/manager"
	"taskmaster/messages/process/input"
	"taskmaster/messages/process/output"
	"taskmaster/state"
	"time"
)

type ProcessRunner struct {
	Id uint

	StdoutLogFile *os.File
	StderrLogFile *os.File

	Input  chan input.Message
	Output chan output.Message

	Command exec.Cmd

	Events chan string

	ConfigManager *configManager.TaskConfigManager
	StartTime     *time.Time
	StartedTime   *time.Time
	StartRetries  *uint
	StopTime      *time.Time
	StoppedTime   *time.Time
	ExitStatus    *int

	HasBeenStopped bool
	HasBeenKilled  bool
}

func (this *ProcessRunner) close() {
	this.StdoutLogFile.Close()
	this.StderrLogFile.Close()
}

func newProcessRunner(manager *configManager.TaskConfigManager, id uint, input chan input.Message, output chan output.Message) (*ProcessRunner, error) {
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

func (this *ProcessRunner) Run() {
	defer this.close()

	for {
		select {

		case event := <-this.Events:
			switch event {

			case "START":
				if this.ExitStatus != nil {
					this.Output <- output.NewStartSuccess()
				} else {
					this.Output <- output.NewStartFailure("Process exited before its start time")
					this.StartedTime = new(time.Time)
					*this.StartedTime = time.Now()
				}

			case "STOP":
				this.ExitStatus = new(int)
				*this.ExitStatus = this.Command.ProcessState.ExitCode()
				this.StoppedTime = new(time.Time)
				*this.StoppedTime = time.Now()

			}

		case req := <-this.Input:
			switch req.(type) {

			case input.Status:
				status := ""
				switch {
				case this.StoppedTime != nil:
					expectedExitStatus := configManager.UseMaster(this.ConfigManager.Master, func(conf *config.Config) int { return conf.Tasks[this.Id].ExpectedExitStatus })
					if *this.ExitStatus == expectedExitStatus {
						status += "SUCCESS "
					} else {
						status += "FAILURE "
					}
					status += fmt.Sprint(*this.ExitStatus)
					if this.HasBeenKilled {
						status += " KILLED"
					} else if this.HasBeenStopped {
						status += " STOPPED"
					}
				case this.StartTime == nil:
					status = "NOT_STARTED"
				case this.StartedTime == nil:
					status = "STARTING"
				default:
					status = "RUNNING"
				}
				this.Output <- output.NewStatus(this.Id, status)

			case input.Start:
				if this.StartTime != nil {
					this.Output <- output.NewStartFailure("Process already started")
					break
				}
				this.StartTime = new(time.Time)
				*this.StartTime = time.Now()
				if err := this.Command.Run(); err != nil {
					this.Output <- output.NewStartFailure(
						fmt.Sprintf("Command failed to run (%s)", err.Error()),
					)
				} else {
					go func() {
						this.Command.Wait()
						this.Events <- "STOP"
					}()

					go func() {
						startTime := configManager.UseMaster(this.ConfigManager.Master, func(conf *config.Config) uint {
							return conf.Tasks[this.Id].StartTime
						})
						time.Sleep(time.Duration(startTime) * time.Millisecond)
						this.Events <- "START"
					}()
				}

			case input.Stop:
			case input.Restart:
			case input.Shutdown:
			}
		}
	}
}
