package session

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"project-farm/internal/crypto"

	"github.com/lib/pq"
)

const (
	SessionIDCookieName         = "session_id"
	sqlQueryToDeleteOldSessions = "DELETE FROM sessions WHERE NOW() - create_at >= INTERVAL '3 days';"
)

func SetSessionID(db *sql.DB, userID int, w http.ResponseWriter) error {
	var sessionID string
	var err error
	for {
		sessionID, err = crypto.RandHex32()
		if err != nil {
			return errors.New("Не удалось сгенерировать сессию")
		}

		if _, err := db.Exec("INSERT INTO sessions (user_id, session_id) VALUES ($1, $2)", userID, sessionID); err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23505":
					continue
				default:
					return errors.New("Неизвестная внутреняя ошибка сервера")
				}
			}
		} else {
			break
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SessionIDCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(4 * 24 * time.Hour),
	})

	return nil
}

func GetUserID(db *sql.DB, w http.ResponseWriter, r *http.Request) (int, error) {
	var userID int
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

	if createAt.Add(24 * time.Hour).After(time.Now()) {
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

	if createAt.Add(24 * time.Hour).After(time.Now()) {
		SetSessionIDStr(db, userID, w)
	}

	return userID, nil
}

func SetSessionIDStr(db *sql.DB, userID string, w http.ResponseWriter) error {
	var sessionID string
	var err error
	for {
		sessionID, err = crypto.RandHex32()
		if err != nil {
			return errors.New("Не удалось сгенерировать сессию")
		}

		if _, err := db.Exec("INSERT INTO sessions (user_id, session_id) VALUES ($1, $2)", userID, sessionID); err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23505":
					continue
				default:
					return errors.New("Неизвестная внутреняя ошибка сервера")
				}
			}
		} else {
			break
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SessionIDCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(4 * 24 * time.Hour),
	})

	return nil
}

func DeleteOldSessions(db *sql.DB) {
	if _, err := db.Exec(sqlQueryToDeleteOldSessions); err != nil {
		fmt.Println(err.Error())
	}
}
