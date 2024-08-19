package state

import "taskmaster/utils"

type StateSubscribeFn[T any] func(T, T) StateCleanupFn
type StateUseFn[T, R any] func(T) R
type StateCleanupFn func()

type AcquireEvent struct{}
type ReleaseEvent struct{}
type LockEvent struct{}

type State[T any] struct {
	acquirer chan<- AcquireEvent
	releaser chan<- ReleaseEvent
	lock     <-chan LockEvent

	value    T
	hooks    []StateSubscribeFn[T]
	cleanups []StateCleanupFn
}

func Use[T, R any](state *State[T], hook StateUseFn[T, R]) R {
	return withLock(state, func() R { return hook(state.value) })
}

func withLock[T, R any](state *State[T], callback func() R) R {
	defer func() { state.releaser <- ReleaseEvent{} }()
	state.acquirer <- AcquireEvent{}
	<-state.lock
	return callback()
}

func (this *State[T]) Set(value T) {
	withLock(this, utils.NoReturn(func() {
		for _, cleanup := range this.cleanups {
			if cleanup != nil {
				cleanup()
			}
		}
		this.cleanups = make([]StateCleanupFn, len(this.hooks))
		for i, hook := range this.hooks {
			this.cleanups[i] = hook(value, this.value)
		}
		this.value = value
	}))
}

func (this *State[T]) Subscribe(hook StateSubscribeFn[T]) {
	withLock(this, utils.NoReturn(func() {
		this.hooks = append(this.hooks, hook)
	}))
}

func (this *State[T]) Close(state *State[T]) {
	withLock(state, utils.NoReturn(func() {
		for _, cleanup := range state.cleanups {
			cleanup()
		}
	}))
	close(state.acquirer)
	close(state.releaser)
}

func NewState[T any](value T) *State[T] {
	acquirer := make(chan AcquireEvent)
	releaser := make(chan ReleaseEvent)
	lock := make(chan LockEvent)

	instance := &State[T]{
		acquirer: acquirer,
		releaser: releaser,
		lock:     lock,
	}

	go func() {
		defer close(lock)
		for {
			if _, ok := <-acquirer; !ok {
				break
			}
			lock <- LockEvent{}
			if _, ok := <-releaser; !ok {
				break
			}
		}
	}()

	return instance
}
