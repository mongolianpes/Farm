package session

import (
	"database/sql"
	"fmt"
	"time"
)

const sqlQueryToDeleteOldSessions = "DELETE FROM sessions WHERE NOW() - create_at >= INTERVAL '3 days';"

func OldSessionsRemover(db *sql.DB) {
	for {
		if _, err := db.Exec(sqlQueryToDeleteOldSessions); err != nil {
			fmt.Println(err.Error())
		}
		time.Sleep(time.Hour * 24)
	}
}
