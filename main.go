package main

import (
	"flag"

	"taskmaster/config"
	"taskmaster/requests"
	"taskmaster/runner"
	"taskmaster/shell"
	"taskmaster/utils"
)

func main() {
	configPath := flag.String("config", "./tmconfig.json", "path/to/taskmaster/configuration/file.json")
	flag.Parse()

	config := utils.Must(config.Parse(*configPath))

	req := make(chan requests.Request)
	res := make(chan requests.Response)
	go runner.StartRunner(*configPath, config, req, res)
	shell.StartShell(*config, res, req)
}
