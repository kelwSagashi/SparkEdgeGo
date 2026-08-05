package mqtt

import "context"

type Client struct {
	BrokerURL string
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(ctx context.Context) error {
	// TODO: Implement EMQX connection, subscription, heartbeat, and queue retry.
	return ctx.Err()
}
