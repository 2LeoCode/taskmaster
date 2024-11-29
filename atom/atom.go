package atom

import (
	"sync/atomic"
)

type AtomSubscriberFunc[T any] func(T, T) error

// Wrapper around atomic.Value that allows strong typing
// and subscription
type Atom[T any] interface {
	Get() (value T)
	Set(value T) (old T, err error)
	Subscribe(subscriber AtomSubscriberFunc[T])
}

type atom[T any] struct {
	subscribers []AtomSubscriberFunc[T]
	value       *atomic.Value
}

// Get the underlying value
func (this *atom[T]) Get() T {
	return this.value.Load().(T)
}

// Set the underlying value
func (this *atom[T]) Set(value T) (T, error) {
	old := this.value.Swap(value).(T)
	for _, subscriber := range this.subscribers {
		if err := subscriber(old, value); err != nil {
			return old, err
		}
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
		value:       new(atomic.Value),
	}
	instance.value.Store(initialValue)
	return instance
}
