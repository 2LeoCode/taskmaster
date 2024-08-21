package configManager

import (
	"taskmaster/config"
	"taskmaster/state"
)

type Task struct {
	Master *Master
	id     uint
}

func UseTask[R any](manager *Task, hook func(*config.Config, uint) R) R {
	return UseMaster(manager.Master, func(conf *config.Config) R {
		return hook(conf, manager.id)
	})
}

func (this *Task) Subscribe(hook func(*config.Config, *config.Config, uint) state.StateCleanupFn) {
	this.Master.state.Subscribe(func(conf, prev *config.Config) state.StateCleanupFn {
		return hook(conf, prev, this.id)
	})
}

func NewTask(master *Master, id uint) *Task {
	return &Task{master, id}
}
