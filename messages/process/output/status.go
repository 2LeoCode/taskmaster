package output

import "taskmaster/messages/helpers"

type Status interface {
	Message
	helpers.Global
	isStatus() bool
	ProcessId() uint
	Value() string
}

type status struct {
	message
	helpers.BaseGlobal
	processId uint
	value     string
}

func (*status) isStatus() bool       { return true }
func (this *status) ProcessId() uint { return this.processId }
func (this *status) Value() string   { return this.value }

func NewStatus(processId uint, value string) Status {
	return &status{
		processId: processId,
		value:     value,
	}
}
