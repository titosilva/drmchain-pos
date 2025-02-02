package longtask

import "context"

type LongTask[T interface{}] struct {
	cancellation context.Context
	cancel       context.CancelFunc
	action       func(context.Context) T
	finally      func()
	resultChan   chan T
}

func Run[T any](action func(context.Context) T) *LongTask[T] {
	return RunWithContext(context.Background(), action)
}

func RunWithContext[T any](ctx context.Context, action func(context.Context) T) *LongTask[T] {
	ctx, cancel := context.WithCancel(ctx)
	lt := &LongTask[T]{
		cancellation: ctx,
		cancel:       cancel,
		action:       action,
		resultChan:   make(chan T),
	}

	started := make(chan struct{})
	go func(cStarted chan struct{}) {
		close(started)
		lt.resultChan <- action(lt.cancellation)
	}(started)

	<-started
	return lt
}

func (lt *LongTask[T]) Finally(finally func()) *LongTask[T] {
	lt.finally = finally
	return lt
}

func (lt *LongTask[T]) Cancel() {
	lt.cancel()
}

func (lt *LongTask[T]) Await() (T, bool) {
	var result T
	var ok bool
	finished := false

	for {
		select {
		case r := <-lt.resultChan:
			result = r
			ok = true
			finished = true
		case <-lt.cancellation.Done():
			result = *new(T)
			ok = false
			finished = true
		}

		if finished {
			break
		}
	}

	if lt.finally != nil {
		lt.finally()
	}

	return result, ok
}
