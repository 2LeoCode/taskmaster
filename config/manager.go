package config

import (
	"taskmaster/state"
)

type Manager struct {
	state *state.State[*Config]
	path  string
}

func Use[R any](manager *Manager, hook state.StateUseFn[*Config, R]) R {
	return state.Use(manager.state, hook)
}

func (this *Manager) Subscribe(hook state.StateSubscribeFn[*Config]) {
	this.state.Subscribe(hook)
}

func (this *Manager) Load() error {
	if config, err := Parse(this.path); err != nil {
		return err
	} else {
		return this.state.Set(config)
	}
}

func NewManager(path string) (*Manager, error) {
	instance := &Manager{path: path, state: state.NewState[*Config](nil)}
	if err := instance.Load(); err != nil {
		return nil, err
	}
	return instance, nil
}
