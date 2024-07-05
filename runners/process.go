package runners

import (
	"fmt"
	"os"
	"os/exec"
	"taskmaster/config"
	"taskmaster/messages/process-events"
	"taskmaster/messages/process-requests"
	"taskmaster/messages/process-responses"
	"taskmaster/messages/responses"
	"time"
)

type ProcessRunner struct {
	ProcId uint

	StartTime    *time.Time
	StartedTime  *time.Time
	StartRetries *uint
	StopTime     *time.Time
	StoppedTime  *time.Time
	ExitStatus   *int

	HasBeenStopped bool
	HasBeenKilled  bool
}

func NewProcessRunner(id uint) ProcessRunner {
	return ProcessRunner{ProcId: id}
}

func (this *ProcessRunner) Run(config *config.Config, taskId uint, input <-chan process_requests.ProcessRequest, output chan<- process_responses.ProcessResponse) {
	processEvents := make(chan process_events.ProcessEvent)

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
		// TODO: Handle error
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
	}
	defer stderrLogFile.Close()
	// send InitSuccessProcessResponse

	cmd := exec.Command(*taskConfig.Command, taskConfig.Arguments...)
	cmd.Stdout = stdoutLogFile
	cmd.Stderr = stderrLogFile

	for {
		select {
		case event := <-processEvents:
			if _, ok := event.(process_events.ExitProcessEvent); ok {
				this.ExitStatus = new(int)
				*this.ExitStatus = cmd.ProcessState.ExitCode()
				this.StoppedTime = new(time.Time)
				*this.StoppedTime = time.Now()
			} else if _, ok := event.(process_events.StartedProcessEvent); ok {
				if this.ExitStatus == nil {
					this.StartedTime = new(time.Time)
					*this.StartedTime = time.Now()
				}
			} else if event, ok := event.(process_events.FailedToStartProcessEvent); ok {
				output <- process_responses.NewStartFailureProcessResponse(event.Reason())
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
				this.StartTime = new(time.Time)
				*this.StartTime = time.Now()
				if err := cmd.Run(); err != nil {
					processEvents <- process_events.NewFailedToStartProcessEvent(err.Error())
				} else {
					go func() {
						cmd.Wait()
						processEvents <- process_events.NewExitProcessEvent()
					}()
					go func() {
						time.Sleep(time.Duration(taskConfig.StartTime) * time.Millisecond)
						processEvents <- process_events.NewStartedProcessEvent()
					}()
				}
			}
		}
	}
}
