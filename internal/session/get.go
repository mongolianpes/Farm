package session

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"project-farm/internal/rdb"
)

func GetUserID(db *sql.DB, rdb rdb.DB, w http.ResponseWriter, r *http.Request) (int, error) {
	var userID int = -1
	var createAt time.Time
	cookie, err := r.Cookie(SessionIDCookieName)
	if err != nil {
		return userID, errors.New("Не имеется текущей ID сессии в Cookie")
	}

	userIDStr, err := rdb.GetUserID(cookie.Value)
	if err != nil {
		userID, createAt, err = getUserIDFromSQLDB(db, cookie.Value)
		if err != nil {
			if err == sql.ErrNoRows {
				slog.Warn("Попытка аутентификации по не существующей ID сесии", "sessionID", cookie.Value, "userIP", r.RemoteAddr)
			}
			return userID, err
		}
	}
	userID, err = strconv.Atoi(userIDStr)
	if err != nil {
		userID, createAt, err = getUserIDFromSQLDB(db, cookie.Value)
		if err != nil {
			if err == sql.ErrNoRows {
				slog.Warn("Попытка аутентификации по не существующей ID сесии", "sessionID", cookie.Value, "userIP", r.RemoteAddr)
			}
			return userID, err
		}
	}

	if createAt.Add(24 * time.Hour).Before(time.Now()) {
		SetSessionID(db, rdb, userID, w)
	}

	return userID, nil
}

func getUserIDFromSQLDB(db *sql.DB, session string) (int, time.Time, error) {
	userID := -1
	var createAt time.Time
	if err := db.QueryRow("SELECT user_id, create_at FROM sessions WHERE session_id = $1", session).Scan(&userID, &createAt); err != nil {
		return userID, createAt, errors.New("Внутреняя ошибка БД, попробуйте позже")
	}

	return userID, createAt, nil
}
