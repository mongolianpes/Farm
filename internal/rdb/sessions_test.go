package rdb

import (
	"context"
	"database/sql"
	"fmt"
	"project-farm/internal/crypto"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*Client, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Предполагаем, что у тебя есть способ создать Client с уже готовым *redis.Client
	// Если нет — добавь конструктор NewClientFromRedis(rdb *redis.Client)
	rdbClient := &Client{rdb: client} // или через экспортированный конструктор

	t.Cleanup(func() {
		client.Close()
		mr.Close()
	})

	return rdbClient, mr
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

func updateTestUsers(db *sql.DB) ([]int, error) {
	updUsers := []int{}
	if _, err := db.Exec("DELETE FROM users"); err != nil {
		return updUsers, err
	}

	hashedPass, err := crypto.HashString("1234567890")
	if err != nil {
		return updUsers, err
	}

	rowsUserIDs, err := db.Query(`INSERT INTO users (login, name, password) VALUES
		(11, '1', $1),
		(22, '2', $1),
		(33, '3', $1)
		RETURNING user_id`, hashedPass)
	if err != nil {
		return updUsers, err
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
	updUsers = append(updUsers, userID1)
	updUsers = append(updUsers, userID2)
	updUsers = append(updUsers, userID3)

	return updUsers, err
}

func TestSetAndGetUserID(t *testing.T) {
	client, mr := setupTestRedis(t)
	users, err := updateTestUsers(connectToDBForTest())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	session := "seskey"
	_, err = client.GetUserID(ctx, session)
	require.Error(t, err)

	err = client.SetSession(ctx, session, strconv.Itoa(users[0]))
	require.NoError(t, err)

	id, err := client.GetUserID(ctx, session)
	require.NoError(t, err)
	mrID, err := mr.Get(sessionPrefix + session)
	require.NoError(t, err)
	if mrID != id {
		t.Error("Отправленный id не равен сохраненому в БД")
	}
}
