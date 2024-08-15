package output

import taskOutput "taskmaster/messages/task/output"

type Status interface {
	Message
	isStatus() bool
	Tasks() []taskOutput.Status
}

type status struct {
	message
	tasks []taskOutput.Status
}

func (*status) isStatus() bool                  { return true }
func (this *status) Tasks() []taskOutput.Status { return this.tasks }

func NewStatus(tasks []taskOutput.Status) Status { return &status{tasks: tasks} }
