package shell

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"taskmaster/atom"
	"taskmaster/messages/helpers"
	"taskmaster/messages/master/input"
	"taskmaster/messages/master/output"
	"taskmaster/terminal"
	"taskmaster/utils"
	"unicode"
)

const HELP_MESSAGE = `help: show this help help message
status: see the status of every program
start <id>: start a program
stop <id>: stop a program
restart <id>: restart a program
reload: reload configuration file (restart programs only if needed)
shutdown: stop all processes and taskmaster`

type SpecialKey uint

const (
	KEY_CTRL_D    SpecialKey = 4
	KEY_ENTER                = 10
	KEY_BACKSPACE            = 127
	KEY_UP                   = 4283163
	KEY_LEFT                 = 4479771
	KEY_RIGHT                = 4414235
	KEY_DOWN                 = 4348699
)

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

		history := []string{}
		cmdBackup := ""
		historySelection := 0

		for !shouldStop.Get() {

			handleKeystroke := func(code uint32) (tokens []string, execute bool) {
				commandLock.Lock()
				defer commandLock.Unlock()
				if unicode.IsPrint(rune(code)) {
					command = command[:cursor] + string(code) + command[cursor:]
					cursor++
					return nil, false
				}
				switch SpecialKey(code) {
				case KEY_BACKSPACE:
					if len(command) != 0 && cursor != 0 {
						command = command[:cursor-1] + command[cursor:]
						cursor--
					}

				case KEY_ENTER:
					tokens := strings.Split(command, " ")
					utils.Filter(&tokens, func(_ int, token *string) bool { return len(*token) != 0 })
					if len(tokens) != 0 {
						command = strings.Trim(command, " \t\r\f\v\n")
						if len(history) == 0 || history[len(history)-1] != command {
							history = append(history, command)
						}
						command = ""
						cmdBackup = ""
						historySelection = 0
						cursor = 0
						fmt.Print("\n")
						return tokens, true
					}
					return nil, false
				case KEY_UP:
					if historySelection == 0 {
						cmdBackup = command
					}
					if historySelection < len(history) {
						historySelection++
						command = history[len(history)-historySelection]
						cursor = len(command)
					}
				case KEY_DOWN:
					if historySelection == 0 {
						break
					}
					historySelection--
					if historySelection == 0 {
						command = cmdBackup
					} else {
						command = history[len(history)-historySelection]
					}
					cursor = len(command)
				case KEY_LEFT:
					if cursor != 0 {
						cursor--
					}
				case KEY_RIGHT:
					if cursor != len(command) {
						cursor++
					}
				case KEY_CTRL_D:
					return []string{"shutdown"}, true
				}
				return nil, false
			}

			reloadInProgress.Wait()
			DisplayCommand()
			input := make([]byte, 4)
			if _, err := io.ReadAtLeast(reader, input, 1); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
				executeCommand([]string{"shutdown"})
				return
			} else {
				keyCode := binary.NativeEndian.Uint32(input)
				if tokens, ok := handleKeystroke(keyCode); ok {
					executeCommand(tokens)
				}
			}
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
			fmt.Print("\033[2K\r")
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
			DisplayCommand()
		}
	}
}
