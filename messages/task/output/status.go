package output

import (
	"taskmaster/messages/helpers"
	processOutput "taskmaster/messages/process/output"
)

type Status interface {
	Message
	helpers.Global
	isStatus() bool
	TaskId() uint
	Name() string
	Processes() []processOutput.Status
}

type status struct {
	message
	helpers.BaseGlobal
	taskId    uint
	name      string
	processes []processOutput.Status
}

func (*status) isStatus() bool                         { return true }
func (this *status) TaskId() uint                      { return this.taskId }
func (this *status) Name() string                      { return this.name }
func (this *status) Processes() []processOutput.Status { return this.processes }

func NewStatus(taskId uint, name string, processes []processOutput.Status) Status {
	return &status{
		taskId:    taskId,
		name:      name,
		processes: processes,
	}
}
