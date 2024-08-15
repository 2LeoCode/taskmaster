package runners

import (
	"fmt"
	"os"
	"os/exec"
	"taskmaster/config"
	"taskmaster/messages/process/input"
	"taskmaster/messages/process/output"
	"taskmaster/process-events"
	"time"
)

type ProcessRunner struct {
	Id uint

	Input  chan input.Message
	Output chan output.Message

	TaskConfig config.Task

	StartTime    *time.Time
	StartedTime  *time.Time
	StartRetries *uint
	StopTime     *time.Time
	StoppedTime  *time.Time
	ExitStatus   *int

	HasBeenStopped bool
	HasBeenKilled  bool
}

func newProcessRunner(id uint, taskConfig *config.Task, input chan input.Message, output chan output.Message) *ProcessRunner {
	return &ProcessRunner{
		Id:         id,
		Input:      input,
		Output:     output,
		TaskConfig: *taskConfig,
	}
}

func (this *ProcessRunner) Run(config *config.Config, taskId uint, input <-chan process_requests.ProcessRequest, output chan<- process_responses.ProcessResponse) {
	events := make(chan process_events.ProcessEvent)

	taskConfig := config.Tasks[taskId]

	stdoutLogFile, err := os.OpenFile(
		fmt.Sprintf(
			"%s/%d-%d_%s-stdout.log",
			config.LogDir,
			taskId,
			this.ProcId,
			time.Now().Format("060102_030405"),
		),
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		os.ModeAppend.Perm(),
	)
	if err != nil {
		// send InitFailureProcessResponse
		return
	}
	defer stdoutLogFile.Close()

	stderrLogFile, err := os.OpenFile(
		fmt.Sprintf(
			"%s/%d-%d_%s-stderr.log",
			config.LogDir,
			taskId,
			this.ProcId,
			time.Now().Format("060102_030405"),
		),
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		os.ModeAppend.Perm(),
	)
	if err != nil {
		// send InitFailureProcessResponse
		return
	}
	defer stderrLogFile.Close()
	// send InitSuccessProcessResponse

	cmd := exec.Command(*taskConfig.Command, taskConfig.Arguments...)
	cmd.Stdout = stdoutLogFile
	cmd.Stderr = stderrLogFile

	for {
		select {
		case event := <-events:
			if _, ok := event.(process_events.ExitProcessEvent); ok {
				if this.StartedTime == nil {

				}
				this.ExitStatus = new(int)
				*this.ExitStatus = cmd.ProcessState.ExitCode()
				this.StoppedTime = new(time.Time)
				*this.StoppedTime = time.Now()
			} else if _, ok := event.(process_events.StartedProcessEvent); ok {
				if this.ExitStatus == nil {
					this.StartedTime = new(time.Time)
					*this.StartedTime = time.Now()
				} else {
					// TODO: Handle restart attempts
				}
			} else if event, ok := event.(process_events.StartProcessEvent); ok {
				if event, ok := event.(process_events.StartSuccessProcessEvent); ok {
					output <- process_responses.NewStartSuccessProcessResponse()
				} else if event, ok := event.(process_events.StartFailureProcessEvent); ok {
					output <- process_responses.NewStartFailureProcessResponse(event.Reason())
				}
			}
		case req := <-input:
			if _, ok := req.(process_requests.StatusProcessRequest); ok {
				res := responses.ProcessStatus{Id: this.ProcId}
				switch {
				case this.StoppedTime != nil:
					if *this.ExitStatus == taskConfig.ExpectedExitStatus {
						res.Status = "SUCCESS "
					} else {
						res.Status = "FAILURE "
					}
					res.Status += fmt.Sprint(*this.ExitStatus)
					if this.HasBeenKilled {
						res.Status += " KILLED"
					} else if this.HasBeenStopped {
						res.Status += " STOPPED"
					}
				case this.StartTime == nil:
					res.Status = "NOT_STARTED"
				case this.StartedTime == nil:
					res.Status = "STARTING"
				default:
					res.Status = "RUNNING"
				}
				output <- process_responses.NewStatusProcessResponse(res)
			} else if _, ok := req.(process_requests.StartProcessRequest); ok {
				if this.StartTime != nil {
					events <- process_events.NewStartFailureProcessEvent("Process already started")
					break
				}
				this.StartTime = new(time.Time)
				*this.StartTime = time.Now()
				if err := cmd.Run(); err != nil {
					events <- process_events.NewStartFailureProcessEvent(err.Error())
				} else {
					events <- process_events.NewStartSuccessProcessEvent()
					go func() {
						cmd.Wait()
						events <- process_events.NewExitProcessEvent()
					}()
					go func() {
						time.Sleep(time.Duration(taskConfig.StartTime) * time.Millisecond)
						events <- process_events.NewStartedProcessEvent()
					}()
				}
			}
		}
	}
}
