package rdb

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
	ctx context.Context
}

type DB interface {
	SetSession(key, value string) error
	GetUserID(key string) (string, error)
	Close() error
}

func NewClient() *Client {
	return &Client{
		rdb: redis.NewClient(&redis.Options{
			Addr: "rdb:6379",
		}),
		ctx: context.Background(),
	}
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
