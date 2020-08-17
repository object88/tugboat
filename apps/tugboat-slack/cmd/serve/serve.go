package serve

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	v1 "github.com/object88/tugboat/apps/tugboat-slack/pkg/http/router/v1"
	"github.com/object88/tugboat/apps/tugboat-slack/pkg/slack"
	"github.com/object88/tugboat/apps/tugboat-slack/pkg/slack/cliflags"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/object88/tugboat/pkg/http"
	"github.com/object88/tugboat/pkg/http/router"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	slackFlagMgr *cliflags.FlagManager

	bot *slack.Bot
}

func CreateCommand(cmmn *common.CommonArgs) *cobra.Command {
	var c *command

	c = &command{
		Command: cobra.Command{
			Use:   "serve",
			Short: "start the slack server",
			Args:  cobra.NoArgs,
			PreRunE: func(cmd *cobra.Command, args []string) error {
				return c.preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.execute(cmd, args)
			},
		},
		CommonArgs:   cmmn,
		slackFlagMgr: cliflags.New(),
	}

	flags := c.Flags()

	c.slackFlagMgr.ConfigureFlags(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	cfg := c.slackFlagMgr.Config()

	c.bot = slack.New(&cfg)
	c.bot.Logger = c.Log

	return nil
}

func (c *command) execute(cmd *cobra.Command, args []string) error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		// extra handling here
		cancel()

		wg.Wait()
	}()

	go func() {
		defer wg.Done()

		rtr := router.New(c.Log)
		m, err := rtr.Route(router.Defaults(v1.Defaults(c.Log, c.bot)))
		if err != nil {
			return
		}

		s := http.New(c.Log, m, 3000)
		s.Serve(ctx)
	}()

	c.Log.Infof("Server Started")

	// Wait for the interrupt
	<-done

	return nil
}
