package queue

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test_Line_CancelableDo(t *testing.T) {
	count := 0

	l := Line{
		Config: &Config{
			Timeout: 10 * time.Millisecond,
			Worker: func(ctx context.Context, key string) error {
				time.Sleep(5 * time.Millisecond)
				count++
				return nil
			},
		},
	}

	err := l.cancelableDo(context.Background())
	if err != nil {
		t.Errorf("unexpected err: %s", err.Error())
	}
	if count != 1 {
		t.Errorf("unexpected state: call count is %d", count)
	}
}

func Test_Line_CancelableDo_Error(t *testing.T) {
	expectedErr := fmt.Errorf("NOTOK")
	l := Line{
		Config: &Config{
			Timeout: 10 * time.Millisecond,
			Worker: func(ctx context.Context, key string) error {
				time.Sleep(5 * time.Millisecond)
				return expectedErr
			},
		},
	}

	err := l.cancelableDo(context.Background())
	if err != expectedErr {
		t.Errorf("unexpected result: did not get expected err; actual: %s", err.Error())
	}
}

func Test_Line_CancelableDo_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	completed := false

	// Use an unreasonably long-running worker
	l := Line{
		Config: &Config{
			Timeout: 10 * time.Millisecond,
			Worker: func(ctx context.Context, key string) error {
				time.Sleep(1 * time.Second)
				completed = true
				return nil
			},
		},
	}

	err := l.cancelableDo(ctx)
	if err == nil {
		t.Errorf("unexpected result: did not get err")
	} else if err != context.DeadlineExceeded {
		t.Errorf("unexpected result: return error: %s", err.Error())
	}
	if completed {
		t.Errorf("unexpected state: the worker function completed")
	}
}

func Test_Line_CancelableDo_Cooldown(t *testing.T) {
	var invocationTime time.Time

	l := Line{
		Config: &Config{
			Cooldown: 10 * time.Millisecond,
			Timeout:  5 * time.Millisecond,
			Worker: func(ctx context.Context, key string) error {
				invocationTime = time.Now()
				return nil
			},
		},
	}

	// Perform the first invocation
	l.cancelableDo(context.Background())

	// Perform the second invocation
	coolingTime := time.Now()
	l.cancelableDo(context.Background())

	elapsedCoolingTime := invocationTime.Sub(coolingTime)
	if elapsedCoolingTime < 10*time.Millisecond {
		t.Errorf("unexpected wait: cooldown laster %s", elapsedCoolingTime)
	}
}

func Test_Line_Enqueue(t *testing.T) {
	count := 0

	l := Line{
		Config: &Config{
			Timeout: 10 * time.Millisecond,
			Worker: func(ctx context.Context, key string) error {
				time.Sleep(5 * time.Millisecond)
				count++
				return nil
			},
		},
	}

	resp := make(chan error, 1)
	l.Enqueue(context.Background(), resp)

	err := <-resp
	if err != nil {
		t.Errorf("unexpected result: received error %s", err.Error())
	}
	if count != 1 {
		t.Errorf("unexpected state: call count is %d", count)
	}
}

func Test_Line_Enqueue_NilRespondent(t *testing.T) {
	l := Line{
		Config: &Config{
			Worker: func(ctx context.Context, key string) error {
				return nil
			},
		},
	}

	err := l.Enqueue(context.Background(), nil)
	if err == nil {
		t.Errorf("unexpected result: did not get err")
	} else if err != ErrNilRespondent {
		t.Errorf("unexpected result: expected ErrNilRespondent, got %s", err.Error())
	}
}

func Test_Line_Enqueue_ManyCalls(t *testing.T) {
	// The delay needs to creep up because the processing starts as soon as the
	// first request is queued.  A small delay can be overwhelmed with a large
	// number of requests; we aren't able to queue up all requests, likely due
	// to the locking around the queue map.
	tcs := []struct {
		name string
		max  int
	}{
		{
			name: "one",
			max:  1,
		},
		{
			name: "ten",
			max:  10,
		},
		{
			name: "one hundred",
			max:  100,
		},
		{
			name: "one thousand",
			max:  1000,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			count := 0
			countingWorker := func(_ context.Context, key string) error {
				time.Sleep(25 * time.Millisecond)
				count++
				return nil
			}

			l := Line{
				Config: &Config{
					Worker: countingWorker,
				},
			}

			var wg sync.WaitGroup
			wg.Add(tc.max)

			// Start queueing
			for i := 0; i < tc.max; i++ {
				waiter := make(chan error, 1)

				// Asynchronously wait for our waiter.
				go func() {
					l.Enqueue(context.Background(), waiter)
				}()

				go func() {
					defer wg.Done()

					err := <-waiter
					if err != nil {
						t.Errorf("Got error from waiter: %s", err.Error())
					}
				}()
			}

			// Wait for the work to be completed (i.e., all callbacks have fired)
			wg.Wait()

			// Validate that the worker was called exactly once, that the line is no
			// longer invoked, and that the respondents collection has been cleaned
			if count != 1 {
				t.Errorf("unexpected result: `Worker` called %d times", count)
			}

			if l.invoked {
				t.Errorf("unexpected state: still invoked")
			}

			if l.respondents != nil {
				t.Errorf("unexpected state: did not clean up queue")
			}
		})
	}
}
