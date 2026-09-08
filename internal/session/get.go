package session

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"project-farm/internal/rdb"
)

var ErrUserHaveNotSession = errors.New("User have not session")

func GetUserID(ctx context.Context, db *sql.DB, rdb rdb.DB, session string) (string, int, error) {
	var userID int = -1
	var createAt time.Time

	userIDStr, err := rdb.GetUserID(ctx, session)
	if err != nil {
		userID, createAt, err = getUserIDFromSQLDB(db, session)
		if err != nil {
			if err == sql.ErrNoRows {
				slog.Warn("Попытка аутентификации по не существующей ID сесии", "sessionID", session)
			}
			return "", userID, ErrUserHaveNotSession
		}
	}
	userID, err = strconv.Atoi(userIDStr)
	if err != nil {
		userID, createAt, err = getUserIDFromSQLDB(db, session)
		if err != nil {
			if err == sql.ErrNoRows {
				slog.Warn("Попытка аутентификации по не существующей ID сесии", "sessionID", session)
			}
			return "", userID, ErrUserHaveNotSession
		}
	}

	newSession := ""
	if createAt.Add(24 * time.Hour).Before(time.Now()) {
		newSession, _ = SetSessionID(ctx, db, rdb, userID)
	}

	return newSession, userID, nil
}

func getUserIDFromSQLDB(db *sql.DB, session string) (int, time.Time, error) {
	userID := -1
	var createAt time.Time
	if err := db.QueryRow("SELECT user_id, create_at FROM sessions WHERE session_id = $1", session).Scan(&userID, &createAt); err != nil {
		return userID, createAt, errors.New("Внутреняя ошибка БД, попробуйте позже")
	}

	return userID, createAt, nil
}
