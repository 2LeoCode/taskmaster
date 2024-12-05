package shell

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"taskmaster/atom"
	"taskmaster/messages/helpers"
	"taskmaster/messages/master/input"
	"taskmaster/messages/master/output"
	"taskmaster/terminal"
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

	terminal.DisableEchoMode()
	terminal.DisableCannonicalMode()
	defer func() {
		shouldStop.Set(true)
		close(out)
		terminal.EnableEchoMode()
		terminal.EnableCannonicalMode()
	}()
	commands := make(chan []string)
	reader := bufio.NewReader(os.Stdin)

	go func() {
		executeCommand := func(command []string) {
			commandOk.Add(1)
			commands <- command
			commandOk.Wait()
		}

		for !shouldStop.Get() {
			fmt.Print("> ")
			cmd := ""
			cursor := 0

			handleKeystroke := func(code uint32) bool {
				// TODO: Handle keystrokes
				return true
			}

			reloadInProgress.Wait()
			var input []byte
			if _, err := reader.Read(input); err != nil {
				executeCommand([]string{"shutdown"})
				return
			}
			keyCode := binary.NativeEndian.Uint32(input)
			if handleKeystroke(keyCode) {
				tokens := strings.Split(cmd, " ")
				utils.Filter(&tokens, func(_ int, token *string) bool { return len(*token) != 0 })
				if len(tokens) == 0 {
					continue
				}
				executeCommand(tokens)
			}
			fmt.Printf("\033[s\033[2K> %s\033[u", cmd)
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

			default:
				fmt.Printf("invalid command: %s (type `help` to get a list of available commands)\n", cmd[0])
			}
			commandOk.Done()

		case res, ok := <-in:
			if !ok {
				return
			}
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
