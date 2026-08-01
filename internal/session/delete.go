package session

import (
	"database/sql"
	"fmt"
)

const sqlQueryToDeleteOldSessions = "DELETE FROM sessions WHERE NOW() - create_at >= INTERVAL '3 days';"

func DeleteOldSessions(db *sql.DB) {
	if _, err := db.Exec(sqlQueryToDeleteOldSessions); err != nil {
		fmt.Println(err.Error())
	}
}
