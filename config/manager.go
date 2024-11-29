package config

import (
	"taskmaster/atom"
)

type Manager interface {
	Get() (value *Config)
	Subscribe(hook atom.AtomSubscriberFunc[*Config])
	Load() error
}

type manager struct {
	atom atom.Atom[*Config]
	path string
}

func (this *manager) Get() *Config {
	return this.atom.Get()
}

func (this *manager) Subscribe(hook atom.AtomSubscriberFunc[*Config]) {
	this.atom.Subscribe(hook)
}

func (this *manager) Load() error {
	if config, err := Parse(this.path); err != nil {
		return err
	} else {
		if _, err := this.atom.Set(config); err != nil {
			return err
		}
	}
	return nil
}

func NewManager(path string) (Manager, error) {
	instance := &manager{path: path, atom: atom.NewAtom[*Config](nil)}
	if err := instance.Load(); err != nil {
		return nil, err
	}
	return instance, nil
}
