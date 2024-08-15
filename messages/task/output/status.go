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
	Processes() []processOutput.Status
}

type status struct {
	message
	helpers.BaseGlobal
	taskId    uint
	processes []processOutput.Status
}

func (*status) isStatus() bool                         { return true }
func (this *status) TaskId() uint                      { return this.taskId }
func (this *status) Processes() []processOutput.Status { return this.processes }

func NewStatus(taskId uint, processes []processOutput.Status) Status {
	return &status{
		taskId:    taskId,
		processes: processes,
	}
}
