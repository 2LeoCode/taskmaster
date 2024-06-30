package shell

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"taskmaster/config"
	"taskmaster/requests"
)

type Shell struct {
	Input  <-chan requests.Response
	Output chan<- requests.Request
}

const HELP_MESSAGE = `help: show this help help message
status: see the status of every program
start <id>: start a program
stop <id>: stop a program
restart <id>: restart a program
reload: reload configuration file (restart programs only if needed)
shutdown: stop all processes and taskmaster`

func StartShell(config config.Config, input <-chan requests.Response, output chan<- requests.Request) {
	commands := make(chan string)
	reader := bufio.NewReader(os.Stdin)

	go func() {
		for {
			cmd, _ := reader.ReadString('\n')
			commands <- cmd
		}
	}()

	for {
		select {

		case cmd := <-commands:
			tokens := strings.Split(cmd, " ")
			if len(tokens) == 0 {
				continue
			}
			switch tokens[0] {
			case "help":
				println(HELP_MESSAGE)
			case "status":
				output <- requests.NewStatusRequest()
			case "start":
				if len(tokens) != 2 {
					println("usage: start <id>")
				}
				output <- requests.NewStartProcessRequest(tokens[1])
			case "stop":
				if len(tokens) != 2 {
					println("usage: stop <id>")
				}
				output <- requests.NewStopProcessRequest(tokens[1])
			case "restart":
				if len(tokens) != 2 {
					println("usage: restart <id>")
				}
				output <- requests.NewRestartProcessRequest(tokens[1])
			case "reload":
				output <- requests.NewReloadConfigRequest()
			case "shutdown":
				output <- requests.NewShutdownRequest()
			default:
				fmt.Printf("invalid command: %s (type `help` to get a list of available commands)", tokens[0])
			}

		case res := <-input:
			if res, ok := res.(requests.StatusResponse); ok {
				for i, task := range res.Tasks() {
					fmt.Printf("%s -- %s\n", task.Id, *config.Tasks[i].Name)
					for _, proc := range task.Processes {
						fmt.Printf("  %s -- %s\n", proc.Id, proc.Status)
					}
				}
			} else if res, ok := res.(requests.StartProcessResponse); ok {
				if _, ok := res.(requests.StartProcessSuccesResponse); ok {
					println("Successfully started program.")
				} else if failure, ok := res.(requests.StartProcessFailureResponse); ok {
					println(failure.Reason())
				}
			} else if res, ok := res.(requests.StopProcessResponse); ok {
				if _, ok := res.(requests.StopProcessSuccesResponse); ok {
					println("Successfully stopped program.")
				} else if failure, ok := res.(requests.StopProcessFailureResponse); ok {
					println(failure.Reason())
				}
			} else if res, ok := res.(requests.RestartProcessResponse); ok {
				if _, ok := res.(requests.RestartProcessSuccesResponse); ok {
					println("Successfully restarted program.")
				} else if failure, ok := res.(requests.RestartProcessFailureResponse); ok {
					println(failure.Reason())
				}
			} else if res, ok := res.(requests.ReloadConfigResponse); ok {
				if _, ok := res.(requests.ReloadConfigSuccessResponse); ok {
					println("Successfully reloaded configuration.")
				} else if failure, ok := res.(requests.ReloadConfigFailureResponse); ok {
					println(failure.Reason())
				}
			} else if _, ok := res.(requests.ShutdownResponse); ok {
				return
			}
		}
	}
}
