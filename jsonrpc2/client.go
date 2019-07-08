package jsonrpc2

import (
	"encoding/json"
	"sync/atomic"
)

type Requester interface {
	// Request takes call inputs and creates a valid request Message.
	Request(method string, params ...interface{}) (*Message, error)
}

var _ Requester = &Client{}

// Client is responsible for making request messages.
type Client struct {
	id int32
}

func (c *Client) NextID() int {
	return int(atomic.AddInt32(&c.id, 1))
}

func (c *Client) Request(method string, params ...interface{}) (*Message, error) {
	msg := &Message{
		Request: &Request{
			Method: method,
		},
	}
	var err error
	if msg.ID, err = json.Marshal(c.NextID()); err != nil {
		return nil, err
	}
	if msg.Request.Params, err = json.Marshal(params); err != nil {
		return nil, err
	}
	return msg, nil
}
