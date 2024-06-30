package requests

import "taskmaster/config"

type ReloadConfigResponse interface {
	Response
	reloadConfig()
}

type ReloadConfigSuccessResponse interface {
	ReloadConfigResponse
	success()
	NewConfig() config.Config
}

type ReloadConfigFailureResponse interface {
	ReloadConfigResponse
	failure()
	Reason() string
}
