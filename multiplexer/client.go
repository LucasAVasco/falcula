package multiplexer

import (
	"github.com/fatih/color"
)

// Client is a multiplexer client that can write to the multiplexer. Multiple clients can write to the same multiplexer
type Client struct {
	multi *Multiplexer // Multiplexer that created this client

	name  string
	level string
	clr   *color.Color
	id    uint
}

func (c *Client) GetName() string {
	return c.name
}

func (c *Client) GetLevel() string {
	return c.level
}

func (c *Client) GetId() uint {
	return c.id
}

func (c *Client) GetColor() *color.Color {
	return c.clr
}

func (c *Client) Write(p []byte) (n int, err error) {
	c.multi.mutex.Lock()
	defer c.multi.mutex.Unlock()

	return c.multi.callback(c, p)
}
