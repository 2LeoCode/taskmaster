package configManager

import (
	"taskmaster/config"
	"taskmaster/state"
)

type MasterConfigManager struct {
	state *state.State[*config.Config]
	path  string
}

type TaskConfigManager struct {
	Master *MasterConfigManager
	id     uint
}

func UseMaster[R any](manager *MasterConfigManager, hook state.StateUseFn[*config.Config, R]) R {
	return state.Use(manager.state, hook)
}

func (this *MasterConfigManager) Subscribe(hook state.StateSubscribeFn[*config.Config]) {
	this.state.Subscribe(hook)
}

func (this *MasterConfigManager) Load() error {
	if config, err := config.Parse(this.path); err != nil {
		return err
	} else {
		this.state.Set(config)
	}
	return nil
}

func UseTask[R any](manager *TaskConfigManager, hook func(*config.Config, uint) R) R {
	return UseMaster(manager.Master, func(conf *config.Config) R {
		return hook(conf, manager.id)
	})
}

func (this *TaskConfigManager) Subscribe(hook func(*config.Config, *config.Config, uint) state.StateCleanupFn) {
	this.Master.state.Subscribe(func(conf, prev *config.Config) state.StateCleanupFn {
		return hook(conf, prev, this.id)
	})
}

func NewMaster(path string) *MasterConfigManager {
	instance := &MasterConfigManager{path: path}
	instance.Load()
	return instance
}

func NewTask(master *MasterConfigManager, id uint) *TaskConfigManager {
	return &TaskConfigManager{master, id}
}
