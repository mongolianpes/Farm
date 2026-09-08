package session

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"project-farm/internal/crypto"
	"project-farm/internal/rdb"

	"github.com/lib/pq"
)

func SetSessionID(ctx context.Context, db *sql.DB, rdb rdb.DB, userID int) (string, error) {
	var sessionID string
	var err error
	for {
		sessionID, err = crypto.RandHex32()
		if err != nil {
			return "", errors.New("Не удалось сгенерировать сессию")
		}

		if _, err := db.Exec("INSERT INTO sessions (user_id, session_id) VALUES ($1, $2)", userID, sessionID); err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23505":
					continue
				default:
					return "", errors.New("Неизвестная внутреняя ошибка сервера")
				}
			}
		} else {
			break
		}
	}

	if err := rdb.SetSession(ctx, sessionID, strconv.Itoa(userID)); err != nil {
		return "", err
	}

	return sessionID, nil
}
