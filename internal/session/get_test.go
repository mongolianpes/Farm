package session

import (
	"context"
	"project-farm/internal/crypto"
	"project-farm/internal/rdb"
	"testing"
)

func TestGetUserID(t *testing.T) {
	sqlDB := connectToDBForTest()
	redisDB, err := rdb.NewClient("localhost:6379")
	if err != nil {
		t.Fatal(err)
	}
	users, err := updateTestUsers(sqlDB)
	if err != nil {
		t.Fatal(err)
	}

	sessionID, err := crypto.RandHex32()
	if err != nil {
		t.Fatal("Не удалось сгенерировать сессию")
	}

	if _, err = sqlDB.Exec("INSERT INTO sessions (user_id, session_id) VALUES ($1, $2)", users.usersIDs[0], sessionID); err != nil {
		t.Error(err)
	}

	ctx := context.Background()
	_, returnedUserID, err := GetUserID(ctx, sqlDB, redisDB, sessionID)
	if returnedUserID != users.usersIDs[0] {
		t.Errorf("Вернул ID пользователя чужой при наличии данных только в SQL, вернул ID: %v, реальный ID: %v", returnedUserID, users.usersIDs[0])
	}

	sessionID, err = SetSessionID(ctx, sqlDB, redisDB, users.usersIDs[1])
	if err != nil {
		t.Error(err)
	}

	_, returnedUserID, err = GetUserID(ctx, sqlDB, redisDB, sessionID)
	if returnedUserID != users.usersIDs[1] {
		t.Errorf("Вернул ID пользователя чужой при наличии данных в redis, вернул ID: %v, реальный ID: %v", returnedUserID, users.usersIDs[1])
	}
}
