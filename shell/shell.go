package shell

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"taskmaster/config"
	configManager "taskmaster/config/manager"
	"taskmaster/messages/helpers"
	"taskmaster/messages/master/input"
	"taskmaster/messages/master/output"
)

const HELP_MESSAGE = `help: show this help help message
status: see the status of every program
start <id>: start a program
stop <id>: stop a program
restart <id>: restart a program
reload: reload configuration file (restart programs only if needed)
shutdown: stop all processes and taskmaster`

func StartShell(manager *configManager.Master, in <-chan output.Message, out chan<- input.Message) {
	defer close(out)
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
				out <- input.NewStatus()
			case "start":
				if len(cmd) != 3 {
					println("usage: start <task-id> <process-id>")
				}
				taskId, taskIdErr := strconv.ParseUint(cmd[1], 10, 64)
				processId, processIdErr := strconv.ParseUint(cmd[2], 10, 64)
				if taskIdErr != nil || processIdErr != nil {
					println("Error: task-id and process-id must be valid positive integers")
				}
				out <- input.NewStartProcess(uint(taskId), uint(processId))
			case "stop":
				if len(cmd) != 3 {
					println("usage: stop <task-id> <process-id>")
				}
				taskId, taskIdErr := strconv.ParseUint(cmd[1], 10, 64)
				processId, processIdErr := strconv.ParseUint(cmd[2], 10, 64)
				if taskIdErr != nil || processIdErr != nil {
					println("Error: task-id and process-id must be valid positive integers")
				}
				out <- input.NewStopProcess(uint(taskId), uint(processId))
			case "restart":
				if len(cmd) != 2 {
					println("usage: restart <id>")
				}
				taskId, taskIdErr := strconv.ParseUint(cmd[1], 10, 64)
				processId, processIdErr := strconv.ParseUint(cmd[2], 10, 64)
				if taskIdErr != nil || processIdErr != nil {
					println("Error: task-id and process-id must be valid positive integers")
				}
				out <- input.NewRestartProcess(uint(taskId), uint(processId))
			case "reload":
				out <- input.NewReload()
			case "shutdown":
				out <- input.NewShutdown()
			default:
				fmt.Printf("invalid command: %s (type `help` to get a list of available commands)\n", cmd[0])
			}

		case res := <-in:
			switch res.(type) {
			case output.Status:
				res := res.(output.Status)
				for i, task := range res.Tasks() {
					name := configManager.UseMaster(manager, func(conf *config.Config) string { return *conf.Tasks[i].Name })
					fmt.Printf("%d -- %s\n", task.TaskId(), name)
					for _, proc := range task.Processes() {
						fmt.Printf("  %d -- %s\n", proc.ProcessId(), proc.Value())
					}
				}
			case output.StartProcess:
				res := res.(output.StartProcess)
				switch res.(type) {
				case helpers.Success:
					fmt.Printf("Successfully started program %d in task %d.\n", res.ProcessId(), res.TaskId())
				case helpers.Failure:
					reason := res.(helpers.Failure).Reason()
					fmt.Printf("Failed to start program %d in task %d: %s.\n", res.ProcessId(), res.TaskId(), reason)
				}

			case output.StopProcess:
				res := res.(output.StopProcess)
				switch res.(type) {
				case helpers.Success:
					fmt.Printf("Successfully stopped program %d in task %d.\n", res.ProcessId(), res.TaskId())
				case helpers.Failure:
					reason := res.(helpers.Failure).Reason()
					fmt.Printf("Failed to stop program %d in task %d: %s.\n", res.ProcessId(), res.TaskId(), reason)
				}

			case output.RestartProcess:
				res := res.(output.RestartProcess)
				switch res.(type) {
				case helpers.Success:
					fmt.Printf("Successfully restarted program %d in task %d.\n", res.ProcessId(), res.TaskId())
				case helpers.Failure:
					reason := res.(helpers.Failure).Reason()
					fmt.Printf("Failed to restart program %d in task %d: %s.\n", res.ProcessId(), res.TaskId(), reason)
				}

			case output.Reload:
				switch res.(type) {
				case helpers.Success:
					println("Successfully reloaded configuration.")
				case helpers.Failure:
				}

			case output.BadRequest:
				println("Invalid request.")
			}
		}
	}
}
