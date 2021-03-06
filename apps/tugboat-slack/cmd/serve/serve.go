package serve

import (
	"context"

	v1 "github.com/object88/tugboat/apps/tugboat-slack/pkg/http/router/v1"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/object88/tugboat/internal/slack"
	slackcliflags "github.com/object88/tugboat/internal/slack/cliflags"
	"github.com/object88/tugboat/pkg/http"
	httpcliflags "github.com/object88/tugboat/pkg/http/cliflags"
	"github.com/object88/tugboat/pkg/http/probes"
	"github.com/object88/tugboat/pkg/http/router"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	httpFlagMgr  *httpcliflags.FlagManager
	slackFlagMgr *slackcliflags.FlagManager

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
		httpFlagMgr:  httpcliflags.New(),
		slackFlagMgr: slackcliflags.New(),
	}

	flags := c.Flags()

	c.httpFlagMgr.ConfigureHttpFlag(flags)
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
	p := probes.New()
	rtr := router.New(c.Log)

	return common.Block(c.Log, p, func(ctx context.Context, r probes.Reporter) error {
		m, err := rtr.Route(router.LoggingDefaultRoute, router.Defaults(p, v1.Defaults(c.Log, c.bot)))
		if err != nil {
			return err
		}

		s := http.New(c.Log, m, c.httpFlagMgr.HttpPort())
		s.Serve(ctx, r)
		return nil
	})
}
