package main

import (
	"flag"
	"log"

	configManager "taskmaster/config/manager"
	"taskmaster/messages/master/input"
	"taskmaster/messages/master/output"
	"taskmaster/runners"
	"taskmaster/shell"
)

func main() {
	configPath := flag.String("config", "./tmconfig.json", "path/to/taskmaster/configuration/file.json")
	flag.Parse()

	configManager := configManager.NewMaster(*configPath)

	req := make(chan input.Message)
	res := make(chan output.Message)
	runner, err := runners.NewMasterRunner(configManager, req, res)
	if err != nil {
		log.Fatalf("Failed to initialize runner: %s", err)
	}
	go runner.Run()
	shell.StartShell(configManager, res, req)
}
