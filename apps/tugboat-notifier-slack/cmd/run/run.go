package run

import (
	"context"

	"github.com/object88/tugboat/apps/tugboat-notifier-slack/pkg/notification"
	v1 "github.com/object88/tugboat/apps/tugboat-slack/pkg/http/router/v1"
	"github.com/object88/tugboat/internal/cmd/common"
	"github.com/object88/tugboat/internal/slack"
	slackcliflags "github.com/object88/tugboat/internal/slack/cliflags"
	grpccliflags "github.com/object88/tugboat/pkg/grpc/cliflags"
	"github.com/object88/tugboat/pkg/grpc/server"
	"github.com/object88/tugboat/pkg/http"
	httpcliflags "github.com/object88/tugboat/pkg/http/cliflags"
	"github.com/object88/tugboat/pkg/http/probes"
	"github.com/object88/tugboat/pkg/http/router"
	"github.com/spf13/cobra"
)

type command struct {
	cobra.Command
	*common.CommonArgs

	grpcFlagMgr  *grpccliflags.FlagManager
	httpFlagMgr  *httpcliflags.FlagManager
	slackFlagMgr *slackcliflags.FlagManager

	bot   *slack.Bot
	probe *probes.Probe
}

// CreateCommand returns the `run` Command
func CreateCommand(ca *common.CommonArgs) *cobra.Command {
	var c command
	c = command{
		Command: cobra.Command{
			Use:   "run",
			Short: "run observes the state of tugboat.lauches",
			Args:  cobra.NoArgs,
			PreRunE: func(cmd *cobra.Command, args []string) error {
				return c.preexecute(cmd, args)
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.execute(cmd, args)
			},
		},
		CommonArgs:   ca,
		grpcFlagMgr:  grpccliflags.New(),
		httpFlagMgr:  httpcliflags.New(),
		slackFlagMgr: slackcliflags.New(),
	}

	flags := c.Flags()

	c.grpcFlagMgr.ConfigureGrpcPortFlag(flags)
	c.httpFlagMgr.ConfigureHttpFlag(flags)
	c.slackFlagMgr.ConfigureFlags(flags)

	return common.TraverseRunHooks(&c.Command)
}

func (c *command) preexecute(cmd *cobra.Command, args []string) error {
	cfg := c.slackFlagMgr.Config()

	c.bot = slack.New(&cfg)
	c.bot.Logger = c.Log

	c.probe = probes.New()

	return nil
}

func (c *command) execute(cmd *cobra.Command, args []string) error {
	return common.Multiblock(c.Log, c.probe, c.startGRPCServer, c.startHTTPServer)
}

func (c *command) startGRPCServer(ctx context.Context, r probes.Reporter) error {
	g, err := server.New(c.Log, c.grpcFlagMgr.GRPCPort(), notification.New(c.Log, c.bot))
	if err != nil {
		return err
	}

	return g.Serve(ctx, r)
}

func (c *command) startHTTPServer(ctx context.Context, r probes.Reporter) error {
	m, err := router.New(c.Log).Route(router.LoggingDefaultRoute, router.Defaults(c.probe, v1.Defaults(c.Log, c.bot)))
	if err != nil {
		return err
	}

	cf, err := c.httpFlagMgr.HttpsCertFile()
	if err != nil {
		return err
	}
	kf, err := c.httpFlagMgr.HttpsKeyFile()
	if err != nil {
		return err
	}

	h := http.New(c.Log, m, c.httpFlagMgr.HttpPort())
	if p := c.httpFlagMgr.HttpsPort(); p != 0 {
		if err = h.ConfigureTLS(p, cf, kf); err != nil {
			return err
		}
	}
	h.Serve(ctx, r)
	return nil
}
