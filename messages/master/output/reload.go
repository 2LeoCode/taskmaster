package output

import (
	"taskmaster/messages/helpers"
)

type Reload interface {
	Message
	isReload() bool
}

type ReloadSuccess interface {
	Reload
	helpers.Success
}

type ReloadFailure interface {
	Reload
	helpers.Failure
}

type reload struct{ message }

type reloadSuccess struct {
	reload
	helpers.BaseSuccess
}

type reloadFailure struct {
	reload
	helpers.BaseFailure
}

func (*reload) isReload() bool { return true }

func NewReloadSuccess() ReloadSuccess {
	return &reloadSuccess{}
}

func NewReloadFailure(reason string) ReloadFailure {
	instance := reloadFailure{}
	instance.SetReason(reason)
	return &instance
}
