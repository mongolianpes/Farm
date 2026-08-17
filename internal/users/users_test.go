package users

import (
	"database/sql"
	"fmt"
	"testing"

	"project-farm/internal/crypto"
)

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

type users struct {
	ids []int
}

func updateTestUsers(db *sql.DB) (users, error) {
	updUsers := users{}
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
	updUsers.ids = append(updUsers.ids, userID1)
	updUsers.ids = append(updUsers.ids, userID2)
	updUsers.ids = append(updUsers.ids, userID3)

	return updUsers, err
}

func TestRegisterUser(t *testing.T) {
	usersServiceHost = "localhost:8086"
	db := connectToDBForTest()
	if _, err := updateTestUsers(db); err != nil {
		t.Error(err)
	}

	if err := RegisterUser(nil, "1", "1", "1234567890", "interesi12345"); err != nil {
		t.Error("Не удалось зарегистрировать пользователя без аватарки: " + err.Error())
	}

	var name string
	if err := db.QueryRow("SELECT name FROM users WHERE login = $1", "1").Scan(&name); err != nil {
		t.Error("Функция не вернула ошибку, но пользователь не был зарегистрирован в БД. Ошибка при запросе: " + err.Error())
	}

	if name != "1" {
		t.Error("Функция зарегистрировала пользователя под другим именем")
	}

	if err := RegisterUser(nil, "1", "1", "1234567890", "interesi12345"); err == nil {
		t.Error("Удалось зарегистрировать уже существующего пользователя")
	}

	if err := RegisterUser(nil, "my", "1", "1234567890", "interesi12345"); err == nil {
		t.Error("Удалось зарегистрировать пользователя с недопустимым логином \"my\"")
	}

	if err := RegisterUser(nil, "2", "2", "123456789", "interesi12345"); err == nil {
		t.Error("Удалось зарегистрировать пользователя с паролем 9 символов (должно быть больше 9)")
	}

	if err := RegisterUser(nil, "2", "2", "1234567890", "2"); err == nil {
		t.Error("Удалось зарегистрировать пользователя со слишком короткими интересами")
	}

	veryLongInterests := "asfdgmdfgmfrghasdfghafhfhlo;pghmraghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghasfdgmdfgmfrghasdfghafhfhlo;pghmraghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghasfdgmdfgmfrghasdfghafhfhlo;pghmraghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghghgh"
	if err := RegisterUser(nil, "2", "2", "1234567890", veryLongInterests); err == nil {
		t.Error("Удалось зарегистрировать пользователя со слишком длинными интересами")
	}
}

func TestAuthUser(t *testing.T) {
	usersServiceHost = "localhost:8086"
	db := connectToDBForTest()
	users, err := updateTestUsers(db)
	if err != nil {
		t.Error(err)
	}

	result, err := AuthUser("11", "1234567890")
	if err != nil {
		t.Error(err)
	}
	if result.name != "1" {
		t.Error("Вернул чужое имя")
	}
	if result.ID != users.ids[0] {
		t.Error("Вернул чужой ID")
	}

	if _, err = AuthUser("11", "1234567891"); err == nil {
		t.Error("Удалось авторизоваться под неверным паролем")
	}

	if _, err = AuthUser("111", "1234567890"); err == nil {
		t.Error("Удалось авторизоваться под несуществующим пользователем")
	}
}

func TestGetUserInfo(t *testing.T) {
	usersServiceHost = "localhost:8086"
	db := connectToDBForTest()
	users, err := updateTestUsers(db)
	if err != nil {
		t.Error(err)
	}

	result, err := GetUserInfo("11")
	if err != nil {
		t.Error(err)
	}

	if result.ID != users.ids[0] {
		t.Error("Функция вернула чужой ID")
	}
	if result.name != "1" {
		t.Error("Функция вернула чужое имя")
	}
}
