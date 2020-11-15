package common

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-logr/logr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/object88/tugboat/pkg/http/probes"
)

// Blocker is a long-running func that can be canceled with the provided
// context
type Blocker func(ctx context.Context, r probes.Reporter) error

// Block will run f until SIGINT or SIGTERM is caught
func Block(log logr.Logger, p *probes.Probe, f Blocker) error {
	return Multiblock(log, p, f)
}

func Multiblock(log logr.Logger, p *probes.Probe, fs ...Blocker) error {
	if err := p.SetCapacity(len(fs) + 1); err != nil {
		return err
	}

	// First reporter is for us.
	r := p.Reporter(0)
	r.Ready()

	// Trap OS system signals.
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
			log.Info("Starting Blocker func...", "blocker", i)
			defer wg.Done()

			err := f(ctx, p.Reporter(i+1))
			if err != nil {
				errs = multierror.Append(errs, err)
				log.Error(err, err.Error(), "blocker", i)
			} else {
				log.Info("Exited Blocker func without error", "blocker", i)
			}

			// Closing the channel will allow the wait to finish, and we no longer need
			// to wait on an `os.Signal`.
			close(done)
		}(k, f)
	}

	// Wait for an signal to exit
	log.Info("Waiting on any Blocker func to exit")
	<-done

	// We have received a signal; set our liveness probe to Down
	r.Kill()

	return errs.ErrorOrNil()
}
