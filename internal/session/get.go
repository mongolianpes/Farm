package session

import (
	"database/sql"
	"errors"
	"net/http"
	"time"
)

func GetUserID(db *sql.DB, w http.ResponseWriter, r *http.Request) (int, error) {
	var userID int = -1
	var createAt time.Time
	cookie, err := r.Cookie(SessionIDCookieName)
	if err != nil {
		return userID, errors.New("Не имеется текущей ID сессии в Cookie")
	}

	if err := db.QueryRow("SELECT user_id, create_at FROM sessions WHERE session_id = $1", cookie.Value).Scan(&userID, &createAt); err != nil {
		switch err {
		case sql.ErrNoRows:
			return userID, errors.New("Не имеется текущей ID сессии в БД")
		default:
			return userID, errors.New("Внутреняя ошибка БД, попробуйте позже")
		}
	}

	if createAt.Add(24 * time.Hour).Before(time.Now()) {
		SetSessionID(db, userID, w)
	}

	return userID, nil
}

func GetUserIDStr(db *sql.DB, w http.ResponseWriter, r *http.Request) (string, error) {
	var userID string
	var createAt time.Time
	cookie, err := r.Cookie(SessionIDCookieName)
	if err != nil {
		return userID, errors.New("Не имеется текущей ID сессии в Cookie")
	}

	if err := db.QueryRow("SELECT user_id, create_at FROM sessions WHERE session_id = $1", cookie.Value).Scan(&userID, &createAt); err != nil {
		switch err {
		case sql.ErrNoRows:
			return userID, errors.New("Не имеется текущей ID сессии в БД")
		default:
			return userID, errors.New("Внутреняя ошибка БД, попробуйте позже")
		}
	}

	if createAt.Add(24 * time.Hour).Before(time.Now()) {
		SetSessionIDStr(db, userID, w)
	}

	return userID, nil
}
