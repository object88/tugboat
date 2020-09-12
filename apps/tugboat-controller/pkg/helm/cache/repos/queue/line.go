package queue

import (
	"context"
	"sync"
	"time"
)

type Line struct {
	*Config

	key string

	cooling sync.WaitGroup

	invoked bool

	respondents []chan<- error
	lock        sync.Mutex
}

func NewLine(config *Config, key string) (*Line, error) {
	if key == "" {
		return nil, ErrEmptyKey
	}
	l := &Line{
		Config: config,
		key:    key,
	}
	return l, nil
}

// Enqueue invokes the Worker as appropriate, and will send its reponse to the
// provided channel.
// Enqueue is non-blocking.
func (l *Line) Enqueue(ctx context.Context, respondent chan<- error) error {
	// Handle a nil context.
	if ctx == nil {
		ctx = context.Background()
	}

	// Do not allow a nil respondent
	if respondent == nil {
		return ErrNilRespondent
	}

	// Ensure that we can manipulate the respondents list
	l.lock.Lock()
	defer l.lock.Unlock()

	// Add the provided respondent to the slice of respondents.
	l.respondents = append(l.respondents, respondent)

	// If there is already an invocation in progress, return.
	if l.invoked {
		return nil
	}

	// Mark that an invocation is in progress
	l.invoked = true

	go func() {
		// Be very careful with err; it is going to get sent to multiple respondents
		// below.
		err := l.cancelableDo(ctx)

		// Aquire the lock, as we are about to start updating respondents and
		// and resetting state.
		l.lock.Lock()
		defer l.lock.Unlock()

		// Inform all respondents
		for _, respondent := range l.respondents {
			go func(r chan<- error) {
				r <- err
			}(respondent)
		}

		// Clear out the respondents list and allow ourself to be invoked again
		l.invoked = false
		l.respondents = nil
	}()

	return nil
}

// cancelableDo is only invoked from do, and must not hold the lock.
func (l *Line) cancelableDo(ctx context.Context) error {
	// Wait for any cooling state to complete
	l.cooling.Wait()

	// Set up our own timeout to ensure that this request doesn't run
	// indefinately
	timeoutCtx, cancel := context.WithTimeout(ctx, l.Timeout)
	defer cancel()

	// Set up the response
	c := make(chan error, 1)

	// Invoke the line asynchronously
	go func() {
		c <- l.Worker(timeoutCtx, l.key)
	}()

	// Enter the cooling state and set up to exit the cooling state
	defer func() {
		l.cooling.Add(1)
		go func() {
			select {
			case <-time.After(l.Cooldown):
				// Exit the cooling state
				l.cooling.Done()
			}
		}()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		// Return the response from the line
		return err
	}
}
