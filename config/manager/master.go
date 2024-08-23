package configManager

import (
	"taskmaster/config"
	"taskmaster/state"
)

type Master struct {
	state *state.State[*config.Config]
	path  string
}

func UseMaster[R any](manager *Master, hook state.StateUseFn[*config.Config, R]) R {
	return state.Use(manager.state, hook)
}

func (this *Master) Subscribe(hook state.StateSubscribeFn[*config.Config]) {
	this.state.Subscribe(hook)
}

func (this *Master) Load() error {
	if config, err := config.Parse(this.path); err != nil {
		return err
	} else {
		this.state.Set(config)
	}
	return nil
}

func NewMaster(path string) (*Master, error) {
	instance := &Master{path: path, state: state.NewState[*config.Config](nil)}
	if err := instance.Load(); err != nil {
		return nil, err
	}
	return instance, nil
}
