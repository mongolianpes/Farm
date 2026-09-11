package session

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"project-farm/internal/rdb"

	"github.com/redis/go-redis/v9"
)

type updUsersInfo struct {
	usersIDs []int
}

func connectToRedisForTest() (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return redisClient, nil
}

func TestSetSessionID(t *testing.T) {
	redisDB, err := rdb.NewClient("localhost:6379")
	if err != nil {
		t.Fatal(err)
	}

	testUserID := rand.Intn(100)
	ctx := context.Background()
	sessionID, err := SetSessionID(ctx, redisDB, testUserID)
	if err != nil {
		t.Error(err)
	}

	redisUserID, err := redisDB.GetUserID(ctx, sessionID)
	if err != nil {
		t.Error(err)
	}
	redisUserIDInt, err := strconv.Atoi(redisUserID)
	if err != nil {
		t.Error(err)
	}
	if redisUserIDInt != testUserID {
		t.Errorf("Функция вернула id пользователя, которое не соответствует заданному изначально, redis: %s, должно быть: %v", redisUserID, testUserID)
	}
}
