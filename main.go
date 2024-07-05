package main

import (
	"flag"

	"taskmaster/config"
	"taskmaster/messages/requests"
	"taskmaster/messages/responses"
	"taskmaster/runners"
	"taskmaster/shell"
	"taskmaster/utils"
)

func main() {
	configPath := flag.String("config", "./tmconfig.json", "path/to/taskmaster/configuration/file.json")
	flag.Parse()

	config := utils.Must(config.Parse(*configPath))

	req := make(chan requests.Request)
	res := make(chan responses.Response)
	runner := runners.NewMasterRunner(*configPath, uint(len(config.Tasks)))
	go runner.Run(config, req, res)
	shell.StartShell(*config, res, req)
}
