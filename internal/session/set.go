package session

import (
	"context"
	"errors"
	"strconv"

	"project-farm/internal/crypto"
	"project-farm/internal/rdb"

	"github.com/redis/go-redis/v9"
)

func SetSessionID(ctx context.Context, rdb rdb.DB, userID int) (string, error) {
	var sessionID string
	var err error
	for {
		sessionID, err = crypto.RandHex32()
		if err != nil {
			return "", errors.New("Не удалось сгенерировать сессию")
		}

		haveSession, err := rdb.GetUserID(ctx, sessionID)
		if err == redis.Nil || haveSession == "" {
			break
		}
	}
	if err := rdb.SetSession(ctx, sessionID, strconv.Itoa(userID)); err != nil {
		return "", err
	}

	return sessionID, nil
}
