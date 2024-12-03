package atom

import (
	"sync"
	"taskmaster/utils"
)

type AtomSubscriberFunc[T any] func(T, T) error

// Wrapper around atomic.Value that allows strong typing
// and subscription
type Atom[T any] interface {
	Get() (value T)
	Set(value T) (old T, err error)
	Update(updater func(old T) (new T)) (old T, err error)
	Subscribe(subscriber AtomSubscriberFunc[T])
}

type atom[T any] struct {
	subscribers []AtomSubscriberFunc[T]
	lock        *sync.Mutex
	value       *T
}

// Get the underlying value
func (this *atom[T]) Get() T {
	this.lock.Lock()
	defer this.lock.Unlock()
	return *this.value
}

func (this *atom[T]) triggerSubscribers(old *T, new *T) error {
	for _, subscriber := range this.subscribers {
		if err := subscriber(*old, *new); err != nil {
			return err
		}
	}
	return nil
}

// Set the underlying value
func (this *atom[T]) Set(value T) (T, error) {
	this.lock.Lock()
	old := *this.value
	*this.value = value
	this.lock.Unlock()
	if err := this.triggerSubscribers(&old, &value); err != nil {
		return old, err
	}
	return old, nil
}

func (this *atom[T]) Update(updater func(T) T) (T, error) {
	this.lock.Lock()
	old := *this.value
	*this.value = updater(*this.value)
	new := *this.value
	this.lock.Unlock()
	if err := this.triggerSubscribers(&old, &new); err != nil {
		return old, err
	}
	return old, nil
}

// Trigger a function when Set is called
func (this *atom[T]) Subscribe(subscriber AtomSubscriberFunc[T]) {
	this.subscribers = append(this.subscribers, subscriber)
}

func NewAtom[T any](initialValue T) Atom[T] {
	instance := &atom[T]{
		subscribers: []AtomSubscriberFunc[T]{},
		value:       utils.New(initialValue),
		lock:        new(sync.Mutex),
	}
	return instance
}
