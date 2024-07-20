package shell

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"taskmaster/config"
	"taskmaster/messages/requests"
	"taskmaster/messages/responses"
)

type Shell struct {
	Input  <-chan responses.Response
	Output chan<- requests.Request
}

const HELP_MESSAGE = `help: show this help help message
status: see the status of every program
start <id>: start a program
stop <id>: stop a program
restart <id>: restart a program
reload: reload configuration file (restart programs only if needed)
shutdown: stop all processes and taskmaster`

func StartShell(config config.Config, input <-chan responses.Response, output chan<- requests.Request) {
	commands := make(chan []string)
	reader := bufio.NewReader(os.Stdin)

	go func() {
		for {
			cmd, _ := reader.ReadString('\n')
			if len(cmd) != 0 && cmd[len(cmd)-1] == '\n' {
				// This check is needed because in case of CTRL+D input,
				// the \n is not included in the returned string
				cmd = cmd[:len(cmd)-1]
			}
			if len(cmd) == 0 {
				// input is empty, don't send
				continue
			}
			tokens := strings.Split(cmd, " ")
			if len(tokens) == 1 && len(tokens[0]) == 0 {
				// string contains only spaces, ignore
				continue
			}
			commands <- tokens
		}
	}()

	for {
		select {

		case cmd := <-commands:
			switch cmd[0] {
			case "help":
				println(HELP_MESSAGE)
			case "status":
				output <- requests.NewStatusRequest()
			case "start":
				if len(cmd) != 3 {
					println("usage: start <task-id> <process-id>")
				}
				taskId, taskIdErr := strconv.ParseUint(cmd[1], 10, 64)
				processId, processIdErr := strconv.ParseUint(cmd[2], 10, 64)
				if taskIdErr != nil || processIdErr != nil {
					println("Error: task-id and process-id must be valid positive integers")
				}
				output <- requests.NewStartProcessRequest(uint(taskId), uint(processId))
			case "stop":
				if len(cmd) != 2 {
					println("usage: stop <id>")
				}
				output <- requests.NewStopProcessRequest(cmd[1])
			case "restart":
				if len(cmd) != 2 {
					println("usage: restart <id>")
				}
				output <- requests.NewRestartProcessRequest(cmd[1])
			case "reload":
				output <- requests.NewReloadConfigRequest()
			case "shutdown":
				output <- requests.NewShutdownRequest()
			default:
				fmt.Printf("invalid command: %s (type `help` to get a list of available commands)\n", cmd[0])
			}

		case res := <-input:
			if res, ok := res.(responses.StatusResponse); ok {
				for i, task := range res.Tasks() {
					fmt.Printf("%d -- %s\n", task.Id, *config.Tasks[i].Name)
					for _, proc := range task.Processes {
						fmt.Printf("  %d -- %s\n", proc.Id, proc.Status)
					}
				}
			} else if res, ok := res.(responses.StartProcessResponse); ok {
				if success, ok := res.(responses.StartProcessSuccesResponse); ok {
					println("Successfully started program %d in task %d.", success.ProcessId(), success.TaskId())
				} else if failure, ok := res.(responses.StartProcessFailureResponse); ok {
					println("Failed to start program %d in task %d: %s.", failure.ProcessId(), failure.TaskId(), failure.Reason())
				}
			} else if res, ok := res.(responses.StopProcessResponse); ok {
				if _, ok := res.(responses.StopProcessSuccesResponse); ok {
					println("Successfully stopped program.")
				} else if failure, ok := res.(responses.StopProcessFailureResponse); ok {
					println(failure.Reason())
				}
			} else if res, ok := res.(responses.RestartProcessResponse); ok {
				if _, ok := res.(responses.RestartProcessSuccesResponse); ok {
					println("Successfully restarted program.")
				} else if failure, ok := res.(responses.RestartProcessFailureResponse); ok {
					println(failure.Reason())
				}
			} else if res, ok := res.(responses.ReloadConfigResponse); ok {
				if _, ok := res.(responses.ReloadConfigSuccessResponse); ok {
					println("Successfully reloaded configuration.")
				} else if failure, ok := res.(responses.ReloadConfigFailureResponse); ok {
					println(failure.Reason())
				}
			} else if _, ok := res.(responses.ShutdownResponse); ok {
				return
			}
		}
	}
}
