package notification

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/internal/generated/notifier"
	"github.com/object88/tugboat/internal/slack"
	"google.golang.org/grpc"
)

type Listener struct {
	notifier.UnimplementedListenerServer
	logger logr.Logger

	bot *slack.Bot
}

func New(logger logr.Logger, bot *slack.Bot) *Listener {
	return &Listener{
		bot:    bot,
		logger: logger,
	}
}

func (l *Listener) Register(s *grpc.Server, logger logr.Logger) error {
	notifier.RegisterListenerServer(s, l)
	return nil
}

func (l *Listener) OpenDeployment(ctx context.Context, req *notifier.StartDeploymentRequest) (*notifier.StartDeploymentResponse, error) {
	l.logger.Info("Got OpenDeployment rpc")
	if err := l.bot.SendMessage("general", "deployment started"); err != nil {
		l.logger.Error(err, "failed to send message to Slack", "error", err)
	}
	return &notifier.StartDeploymentResponse{}, nil
}
