package common

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

// Blocker is a long-running func that can be canceled with the provided
// context
type Blocker func(ctx context.Context) error

// Block will run f until SIGINT or SIGTERM is caught
func Block(f Blocker) error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		// Cancel any running context.
		cancel()

		// And wait for the below go func to complete.
		wg.Wait()
	}()

	var finalerr error
	go func() {
		defer wg.Done()

		finalerr = f(ctx)

		// Closing the channel will allow the wait to finish, and we no longer need
		// to wait on an `os.Signal`.
		close(done)
	}()

	// Wait for an signal to exit
	<-done

	return finalerr
}

func Multiblock(log *logrus.Logger, fs ...Blocker) error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(len(fs))

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		// Cancel any running context.
		cancel()

		// And wait for the below go funcs to complete.
		wg.Wait()
	}()

	var errs *multierror.Error
	for k, f := range fs {
		go func(i int, f Blocker) {
			log.Infof("Starting Blocker func (%d)...", i)
			defer wg.Done()

			err := f(ctx)
			if err != nil {
				errs = multierror.Append(errs, err)
				log.Infof("Exited Blocker func (%d) without err: %s", i, err.Error())
			} else {
				log.Infof("Exited Blocker func (%d) without error", i)
			}

			// Closing the channel will allow the wait to finish, and we no longer need
			// to wait on an `os.Signal`.
			close(done)
		}(k, f)
	}

	// Wait for an signal to exit
	log.Infof("Waiting on any Blocker func to exit")
	<-done

	return errs.ErrorOrNil()
}
