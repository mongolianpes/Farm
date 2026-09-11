package session

import (
	"context"
	"math/rand"
	"project-farm/internal/crypto"
	"project-farm/internal/rdb"
	"testing"
)

func TestGetUserID(t *testing.T) {
	redisDB, err := rdb.NewClient("localhost:6379")
	if err != nil {
		t.Fatal(err)
	}

	sessionID, err := crypto.RandHex32()
	if err != nil {
		t.Fatal("Не удалось сгенерировать сессию")
	}

	testUserID := rand.Intn(100)
	ctx := context.Background()
	_, returnedUserID, err := GetUserID(ctx, redisDB, sessionID)
	if err == nil {
		t.Errorf("Не вернул ошибки, когда нету сессии, вернул ID: %v, реальный ID: %v", returnedUserID, testUserID)
	}

	sessionID, err = SetSessionID(ctx, redisDB, testUserID)
	if err != nil {
		t.Error(err)
	}

	_, returnedUserID, err = GetUserID(ctx, redisDB, sessionID)
	if returnedUserID != testUserID {
		t.Errorf("Вернул ID пользователя чужой при наличии данных в redis, вернул ID: %v, реальный ID: %v", returnedUserID, testUserID)
	}
}
