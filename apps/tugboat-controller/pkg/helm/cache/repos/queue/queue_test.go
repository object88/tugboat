package queue

import (
	"context"
	"testing"
	"time"
)

func Test_Queue_Work(t *testing.T) {
	count := 0
	worker := func(ctx context.Context, key string) error {
		time.Sleep(5 * time.Millisecond)
		count++
		return nil
	}
	q := Queue{}
	q.Initialize(worker)

	err := q.Work(context.Background(), "foo")
	if err != nil {
		t.Errorf("unexpected result: received error %s", err.Error())
	}

	if count != 1 {
		t.Errorf("unexpected state: call count is %d", count)
	}
}

func Test_Queue_Work_BadKey(t *testing.T) {
	q := Queue{}
	q.Initialize(func(ctx context.Context, key string) error {
		return nil
	})

	err := q.Work(context.Background(), "")
	if err == nil {
		t.Errorf("unexpected result: did not receive error")
	}
	if err != ErrEmptyKey {
		t.Errorf("unexpected result: expected ErrEmptyKey, got %s", err.Error())
	}
}
