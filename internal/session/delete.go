package session

import (
	"database/sql"
	"log/slog"
	"time"
)

const sqlQueryToDeleteOldSessions = "DELETE FROM sessions WHERE NOW() - create_at >= INTERVAL '3 days';"

func OldSessionsRemover(db *sql.DB) {
	for {
		if _, err := db.Exec(sqlQueryToDeleteOldSessions); err != nil {
			slog.Error("Ошибка при совершении запроса для удаления старых сессий")
		}
		time.Sleep(time.Hour * 24)
	}
}
