package shell

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"taskmaster/atom"
	"taskmaster/messages/helpers"
	"taskmaster/messages/master/input"
	"taskmaster/messages/master/output"
	"taskmaster/utils"
)

const HELP_MESSAGE = `help: show this help help message
status: see the status of every program
start <id>: start a program
stop <id>: stop a program
restart <id>: restart a program
reload: reload configuration file (restart programs only if needed)
shutdown: stop all processes and taskmaster`

func StartShell(in <-chan output.Message, out chan<- input.Message) {
	var reloadInProgress sync.WaitGroup
	var commandOk sync.WaitGroup

	shouldStop := atom.NewAtom(false)

	defer func() {
		shouldStop.Set(true)
		close(out)
	}()
	commands := make(chan []string)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanLines)

	go func() {
		for !shouldStop.Get() {
			fmt.Print("> ")
			reloadInProgress.Wait()
			if ok := scanner.Scan(); !ok {
				commands <- []string{"shutdown"}
				break
			}
			cmd := scanner.Text()
			tokens := strings.Split(cmd, " ")
			utils.Filter(&tokens, func(_ int, token *string) bool { return len(*token) != 0 })
			if len(tokens) == 0 {
				continue
			}
			commandOk.Add(1)
			commands <- tokens
			commandOk.Wait()
		}
	}()

	for {
		select {

		case cmd := <-commands:
			switch cmd[0] {
			case "help":
				fmt.Println(HELP_MESSAGE)

			case "status":
				out <- input.NewStatus()

			case "start":
				if len(cmd) != 3 {
					fmt.Fprintln(os.Stderr, "usage: start <task-id> <process-id>")
					break
				}
				taskId, taskIdErr := strconv.ParseUint(cmd[1], 10, 64)
				processId, processIdErr := strconv.ParseUint(cmd[2], 10, 64)
				if taskIdErr != nil || processIdErr != nil {
					fmt.Fprintln(os.Stderr, "Error: task-id and process-id must be valid positive integers")
					break
				}
				out <- input.NewStartProcess(uint(taskId), uint(processId))

			case "stop":
				if len(cmd) != 3 {
					fmt.Fprintln(os.Stderr, "usage: stop <task-id> <process-id>")
					break
				}
				taskId, taskIdErr := strconv.ParseUint(cmd[1], 10, 64)
				processId, processIdErr := strconv.ParseUint(cmd[2], 10, 64)
				if taskIdErr != nil || processIdErr != nil {
					fmt.Fprintln(os.Stderr, "Error: task-id and process-id must be valid positive integers")
					break
				}
				out <- input.NewStopProcess(uint(taskId), uint(processId))

			case "restart":
				if len(cmd) != 3 {
					fmt.Fprintln(os.Stderr, "usage: restart <task-id> <process-id>")
					break
				}
				taskId, taskIdErr := strconv.ParseUint(cmd[1], 10, 64)
				processId, processIdErr := strconv.ParseUint(cmd[2], 10, 64)
				if taskIdErr != nil || processIdErr != nil {
					fmt.Fprintln(os.Stderr, "Error: task-id and process-id must be valid positive integers")
				}
				out <- input.NewRestartProcess(uint(taskId), uint(processId))

			case "reload":
				reloadInProgress.Add(1)
				out <- input.NewReload()

			case "shutdown":
				out <- input.NewShutdown()
				return

			default:
				fmt.Printf("invalid command: %s (type `help` to get a list of available commands)\n", cmd[0])
			}
			commandOk.Done()

		case res := <-in:
			fmt.Print("\r \r")
			switch res.(type) {
			case output.Status:
				res := res.(output.Status)
				for _, task := range res.Tasks() {
					fmt.Printf("%d -- %s\n", task.TaskId(), task.Name())
					for _, proc := range task.Processes() {
						fmt.Printf("  %d -- %s\n", proc.ProcessId(), proc.Value())
					}
				}

			case output.Reload:
				reloadInProgress.Done()
				switch res.(type) {
				case helpers.Success:
					fmt.Println("Successfully reloaded configuration.")
				case helpers.Failure:
					fmt.Printf("Failed to reload configuration: %s.\n", res.(output.ReloadFailure).Reason())
				}

			case output.BadRequest:
				fmt.Fprintln(os.Stderr, "Invalid request.")
			}
			fmt.Print("> ")
		}
	}
}
