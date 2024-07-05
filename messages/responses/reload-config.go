package responses

import "taskmaster/config"

type ReloadConfigResponse interface {
	Response
	reloadConfigTag()
}

type reloadConfigResponse struct{ response }

func (*reloadConfigResponse) reloadConfigTag()

type ReloadConfigSuccessResponse interface {
	ReloadConfigResponse
	successTag()
	NewConfig() config.Config
}

type reloadConfigSuccessResponse struct {
	reloadConfigResponse
	newConfig config.Config
}

func (*reloadConfigSuccessResponse) successTag()

func (this *reloadConfigSuccessResponse) NewConfig() config.Config {
	return this.newConfig
}

func NewReloadConfigSuccessResponse(newConfig config.Config) ReloadConfigSuccessResponse {
	return &reloadConfigSuccessResponse{newConfig: newConfig}
}

type ReloadConfigFailureResponse interface {
	ReloadConfigResponse
	failureTag()
	Reason() string
}

type reloadConfigFailureResponse struct {
	reloadConfigResponse
	reason string
}

func (*reloadConfigFailureResponse) failureTag()

func (this *reloadConfigFailureResponse) Reason() string {
	return this.reason
}

func NewReloadConfigFailureResponse(reason string) ReloadConfigFailureResponse {
	return &reloadConfigFailureResponse{reason: reason}
}
