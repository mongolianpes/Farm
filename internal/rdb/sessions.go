package rdb

func (c *Client) SetSession(sessionKey, userID string) error {
	err := c.rdb.Set(c.ctx, sessionKey, userID, 0).Err()
	return err
}

func (c *Client) GetUserID(sessionKey string) (string, error) {
	userID, err := c.rdb.Get(c.ctx, sessionKey).Result()
	if err != nil {
		return "", err
	}
	return userID, nil
}
