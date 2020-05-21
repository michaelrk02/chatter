package main

import (
    "time"
)

type client struct {
    Nickname string
    JoinTimestamp int64
}

func (c *client) Init(id uint32, nickname string) *client {
    c.Nickname = nickname
    c.JoinTimestamp = time.Now().UTC().Unix()
    return c
}

