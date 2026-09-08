package rdb

import (
	"context"
	"time"
)

const sessionPrefix string = "session_"

func (c *Client) SetSession(ctx context.Context, sessionKey, userID string) error {
	err := c.rdb.Set(ctx, sessionPrefix+sessionKey, userID, time.Hour).Err()
	return err
}

func (c *Client) GetUserID(ctx context.Context, sessionKey string) (string, error) {
	userID, err := c.rdb.Get(ctx, sessionPrefix+sessionKey).Result()
	if err != nil {
		return "", err
	}
	return userID, nil
}
