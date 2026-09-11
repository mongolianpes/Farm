package session

import (
	"context"
	"errors"
	"strconv"
	"time"

	"project-farm/internal/rdb"
)

var ErrUserHaveNotSession = errors.New("User have not session")

func GetUserID(ctx context.Context, rdb rdb.DB, session string) (string, int, error) {
	var userID int = -1
	var createAt time.Time

	userIDStr, err := rdb.GetUserID(ctx, session)
	if err != nil {
		return "", userID, err
	}
	userID, err = strconv.Atoi(userIDStr)
	if err != nil {
		return "", userID, err
	}

	newSession := ""
	if createAt.Add(24 * time.Hour).Before(time.Now()) {
		newSession, _ = SetSessionID(ctx, rdb, userID)
	}

	return newSession, userID, nil
}
