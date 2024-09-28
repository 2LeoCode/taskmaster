package main

import (
	"flag"
	"log"

	"taskmaster/config"
	"taskmaster/messages/master/input"
	"taskmaster/messages/master/output"
	"taskmaster/runners"
	"taskmaster/shell"
)

func main() {
	configPath := flag.String("config", "./tmconfig.json", "path/to/taskmaster/configuration/file.json")
	flag.Parse()

	configManager, err := config.NewManager(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %s\n", err)
	}

	req := make(chan input.Message)
	res := make(chan output.Message)
	runner, err := runners.NewMasterRunner(configManager, req, res)
	if err != nil {
		log.Fatalf("Failed to initialize runner: %s", err)
	}
	go runner.Run()
	shell.StartShell(res, req)
}
