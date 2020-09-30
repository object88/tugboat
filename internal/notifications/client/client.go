package client

import (
	"context"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	"github.com/object88/tugboat/internal/generated/notifier"
	grpcclient "github.com/object88/tugboat/pkg/grpc/client"
	"google.golang.org/grpc"
)

type Client struct {
	logger logr.Logger

	listeners []notifier.ListenerClient
}

func New(logger logr.Logger) *Client {
	return &Client{
		logger: logger,
	}
}

func (c *Client) Connect(targets []*url.URL) error {
	lcs := make([]notifier.ListenerClient, len(targets))
	for k, v := range targets {
		cc := grpcclient.New(c.logger)
		c.logger.Info("connecting to gRPC target", "target", v)
		if err := cc.Connect(v); err != nil {
			return err
		}

		lcs[k] = notifier.NewListenerClient(cc.ClientConnection())
	}

	c.listeners = lcs

	return nil
}

func (c *Client) DeploymentStarted() error {
	for _, v := range c.listeners {
		err := func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, err := v.OpenDeployment(ctx, &notifier.StartDeploymentRequest{}, grpc.WaitForReady(true))
			return err
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
