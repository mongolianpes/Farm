package session

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"project-farm/internal/rdb"

	"github.com/redis/go-redis/v9"
)

type updUsersInfo struct {
	usersIDs []int
}

func updateTestUsers(db *sql.DB) (updUsersInfo, error) {
	updMessages := updUsersInfo{}
	if _, err := db.Exec("DELETE FROM users"); err != nil {
		return updMessages, err
	}

	rowsUserIDs, err := db.Query(`INSERT INTO users (login, name, password) VALUES
		(1, '1', '1234567890'),
		(2, '2', '1234567890'),
		(3, '3', '1234567890')
		RETURNING user_id`)
	if err != nil {
		return updMessages, err
	}

	var userID1 int
	var userID2 int
	var userID3 int
	rowsUserIDs.Next()
	rowsUserIDs.Scan(&userID1)
	rowsUserIDs.Next()
	rowsUserIDs.Scan(&userID2)
	rowsUserIDs.Next()
	rowsUserIDs.Scan(&userID3)
	updMessages.usersIDs = append(updMessages.usersIDs, userID1)
	updMessages.usersIDs = append(updMessages.usersIDs, userID2)
	updMessages.usersIDs = append(updMessages.usersIDs, userID3)

	return updMessages, err
}

func connectToDBForTest() *sql.DB {
	host := "localhost"
	port := "5432"
	user := "postgres"
	password := "123"
	dbname := "project_farm"

	// host := os.Getenv("DB_HOST")
	// port := os.Getenv("DB_PORT")
	// user := os.Getenv("DB_USER")
	// password := os.Getenv("DB_PASSWORD")
	// dbname := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	return db
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
	sqlDB := connectToDBForTest()
	redisDB, err := rdb.NewClient("localhost:6379")
	if err != nil {
		t.Fatal(err)
	}
	users, err := updateTestUsers(sqlDB)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	sessionID, err := SetSessionID(ctx, sqlDB, redisDB, users.usersIDs[0])
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
	if redisUserIDInt != users.usersIDs[0] {
		t.Errorf("Функция вернула id пользователя, которое не соответствует заданному изначально, redis: %s, должно быть: %v", redisUserID, users.usersIDs[0])
	}

	var sqlValue string
	if err := sqlDB.QueryRow("SELECT session_id FROM sessions WHERE user_id = $1", users.usersIDs[0]).Scan(&sqlValue); err != nil {
		t.Error(err)
	}
	if sqlValue != sessionID {
		t.Errorf("Функция вернула значение, которое не соответствует сохраненому в SQL, sql: %s, должно быть: %s", sqlValue, sessionID)
	}
}
