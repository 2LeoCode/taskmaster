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

	configLoader := config.NewConfigLoader(*configPath)

	req := make(chan input.Message)
	res := make(chan output.Message)
	runner, err := runners.NewMasterRunner(configLoader, req, res)
	if err != nil {
		log.Fatalf("Failed to initialize runner: %s", err)
	}
	go runner.Run()
	shell.StartShell(res, req)
}
