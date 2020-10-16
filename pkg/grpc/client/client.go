package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type Client struct {
	logger logr.Logger
	cc     *grpc.ClientConn
}

func New(logger logr.Logger) *Client {
	return &Client{
		logger: logger,
	}
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) Connect(address *url.URL) error {
	cc, err := grpc.Dial(address.String(), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to dial client: %w", err)
	}

	state := cc.GetState()
	for state != connectivity.Ready {
		// continue checking for state change
		// until one of break states is found
		change := cc.WaitForStateChange(context.Background(), state)
		if !change {
			// ctx is done, return
			// something upstream is cancelling
			return fmt.Errorf("NAH")
		}

		state = cc.GetState()
		c.logger.Info("State change", "state", state, "target", cc.Target())
	}

	c.cc = cc

	return nil
}

func (c *Client) ClientConnection() *grpc.ClientConn {
	return c.cc
}
