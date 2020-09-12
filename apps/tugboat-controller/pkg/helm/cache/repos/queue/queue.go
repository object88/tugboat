package queue

import (
	"context"
	"sync"
	"time"
)

// Worker is a callback used to perform some long-running task.
// Worker is guarenteed not to be called concurrently with the same key. It may
// be called concurrently with _different_ keys.
type Worker func(ctx context.Context, key string) error

// Queue manages invocations of a Worker func.
type Queue struct {
	Config

	lines map[string]*Line
	lock  sync.Mutex
}

// Config describes the configuration of a Queue and Line.  It is shared among
// all Lines owned by a Queue.
type Config struct {
	Cooldown time.Duration
	Timeout  time.Duration

	Worker Worker
}

// New returns a new instance of a Queue.
// Because a Queue contains an instance of a `sync.Mutex`, it cannot be copied
// after its first use.
// func New(worker Worker) Queue {
// 	return Queue{
// 		Config: Config{
// 			Cooldown: 5 * time.Second,
// 			Timeout:  15 * time.Second,
// 			Worker:   worker,
// 		},
// 		lines: map[string]*Line{},
// 	}
// }

func (q *Queue) Initialize(worker Worker) {
	if q.Cooldown == 0 {
		q.Cooldown = 5 * time.Second
	}
	if q.Timeout == 0 {
		q.Timeout = 15 * time.Second
	}
	q.lines = map[string]*Line{}
	q.Worker = worker
}

// Work requests that the Worker associated with the Queue is invoked with
// `key`.
// If the Worker is already invoked with `key`, Work will return with the error
// returned by work, but will not start a new invocation.
func (q *Queue) Work(ctx context.Context, key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	errch := make(chan error, 1)

	err := q.enqueue(ctx, key, errch)
	if err != nil {
		return err
	}

	return <-errch
}

func (q *Queue) enqueue(ctx context.Context, key string, errch chan<- error) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	l, ok := q.lines[key]
	if !ok {
		var err error
		l, err = NewLine(&q.Config, key)
		if err != nil {
			return err
		}
		q.lines[key] = l
	}

	l.Enqueue(ctx, errch)

	return nil
}
