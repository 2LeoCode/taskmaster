package output

import (
	"taskmaster/config"
	"taskmaster/messages/helpers"
)

type Reload interface {
	Message
	isReload() bool
}

type ReloadSuccess interface {
	Reload
	helpers.Success
	NewConfig() *config.Config
}

type ReloadFailure interface {
	Reload
	helpers.Failure
}

type reload struct{ message }

type reloadSuccess struct {
	reload
	helpers.BaseSuccess
	newConfig *config.Config
}

type reloadFailure struct {
	reload
	helpers.BaseFailure
}

func (*reload) isReload() bool { return true }

func (this *reloadSuccess) NewConfig() *config.Config { return this.newConfig }

func NewReloadSuccess(newConfig *config.Config) ReloadSuccess {
	return &reloadSuccess{newConfig: newConfig}
}

func NewReloadFailure(reason string) ReloadFailure {
	instance := reloadFailure{}
	instance.SetReason(reason)
	return &instance
}
