package session

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"project-farm/internal/crypto"
	"project-farm/internal/rdb"

	"github.com/lib/pq"
)

func SetSessionID(db *sql.DB, rdb rdb.DB, userID int, w http.ResponseWriter) error {
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

	if err := rdb.SetSession(sessionID, strconv.Itoa(userID)); err != nil {
		return err
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

func SetSessionIDStr(db *sql.DB, rdb rdb.DB, userID string, w http.ResponseWriter) error {
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

	if err := rdb.SetSession(sessionID, userID); err != nil {
		return err
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
