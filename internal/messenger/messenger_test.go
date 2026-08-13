package messenger

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func deleteTestMessage(db *sql.DB, messageText string) error {
	if _, err := db.Exec("DELETE FROM messages WHERE message = $1", messageText); err != nil {
		return err
	}

	return nil
}

type updMessagesInfo struct {
	usersIDs         []int
	announcementsIDs []int
}

func updateTestUsers(db *sql.DB) (updMessagesInfo, error) {
	updMessages := updMessagesInfo{}
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

func updateTestAnnouncements(db *sql.DB, updMessages updMessagesInfo) (updMessagesInfo, error) {
	rowsAnnouncementIDs, err := db.Query(`INSERT INTO announcements (title, description, category, announcement_author_id) VALUES
	('1', '1', 'Участки', $1),
	('2', '2', 'Участки', $2),
	('22', '22', 'Участки', $2),
	('222', '222', 'Участки', $2)
	RETURNING announcement_id`, updMessages.usersIDs[0], updMessages.usersIDs[1])
	if err != nil {
		return updMessages, err
	}

	var announcementID1 int
	var announcementID2 int
	var announcementID3 int
	var announcementID4 int
	rowsAnnouncementIDs.Next()
	rowsAnnouncementIDs.Scan(&announcementID1)
	rowsAnnouncementIDs.Next()
	rowsAnnouncementIDs.Scan(&announcementID2)
	rowsAnnouncementIDs.Next()
	rowsAnnouncementIDs.Scan(&announcementID3)
	rowsAnnouncementIDs.Next()
	rowsAnnouncementIDs.Scan(&announcementID4)
	updMessages.announcementsIDs = append(updMessages.announcementsIDs, announcementID1)
	updMessages.announcementsIDs = append(updMessages.announcementsIDs, announcementID2)
	updMessages.announcementsIDs = append(updMessages.announcementsIDs, announcementID3)
	updMessages.announcementsIDs = append(updMessages.announcementsIDs, announcementID4)

	return updMessages, nil
}

func updateTestMessages(db *sql.DB) (updMessagesInfo, error) {
	db.Exec("DELETE FROM messages")

	updMessages, err := updateTestUsers(db)
	if err != nil {
		return updMessages, err
	}

	updMessages, err = updateTestAnnouncements(db, updMessages)
	if err != nil {
		return updMessages, err
	}

	messages := []struct {
		senderID, receivedID, announcementID int
		text                                 string
	}{
		{updMessages.usersIDs[0], updMessages.usersIDs[1], updMessages.announcementsIDs[0], "Привет по объявлению 10"},
		{updMessages.usersIDs[1], updMessages.usersIDs[0], updMessages.announcementsIDs[0], "Ответ по объявлению 10"},
		{updMessages.usersIDs[0], updMessages.usersIDs[1], updMessages.announcementsIDs[0], "Привет по объявлению 10 (2)"},
		{updMessages.usersIDs[0], updMessages.usersIDs[1], updMessages.announcementsIDs[0], "Привет по объявлению 10 (3)"},
		{updMessages.usersIDs[1], updMessages.usersIDs[0], updMessages.announcementsIDs[0], "Ответ по объявлению 10 (2)"},
		{updMessages.usersIDs[1], updMessages.usersIDs[0], updMessages.announcementsIDs[0], "Ответ по объявлению 10 (3)"},
		{updMessages.usersIDs[0], updMessages.usersIDs[1], updMessages.announcementsIDs[1], "Привет по объявлению 20"},
		{updMessages.usersIDs[0], updMessages.usersIDs[1], updMessages.announcementsIDs[2], "Прив по 3 для 2"},
		{updMessages.usersIDs[0], updMessages.usersIDs[2], updMessages.announcementsIDs[2], "Прив по 3 для 3"},
		{updMessages.usersIDs[0], updMessages.usersIDs[2], updMessages.announcementsIDs[3], "Прив по 4 для 3"},
		{updMessages.usersIDs[0], updMessages.usersIDs[2], updMessages.announcementsIDs[0], "Сообщение пользователю 3 по объявлению 10"},
		{updMessages.usersIDs[2], updMessages.usersIDs[0], updMessages.announcementsIDs[0], "Ответ пользователю 1 по объявлению 10"},
	}

	for _, m := range messages {
		if _, err := db.Exec(`INSERT INTO messages (sender_id, received_id, related_announcement_id, message)
							VALUES ($1, $2, $3, $4)`,
			m.senderID, m.receivedID, m.announcementID, m.text); err != nil {
			return updMessages, err
		}
		time.Sleep(50 * time.Millisecond) // задержка между вставками
	}

	return updMessages, nil
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

func TestGetChatHistory(t *testing.T) {
	db := connectToDBForTest()
	messengerServiceHost = "localhost:8086"

	updMessages, err := updateTestMessages(db)
	if err != nil {
		t.Error(err)
	}

	userID := updMessages.usersIDs[0]
	partnerID := updMessages.usersIDs[1]
	relatedAnnouncementID := updMessages.announcementsIDs[0]
	offset := 0
	chat, err := GetChatHistory(partnerID, userID, relatedAnnouncementID, offset)
	if err != nil {
		t.Error(err)
	}

	if len(chat) != 5 {
		t.Errorf("Количество полученных сообщений не соответствует нужному. Получено сообщений: %v, offset %v", len(chat), offset)
	}

	offset++
	chat, err = GetChatHistory(partnerID, userID, relatedAnnouncementID, offset)
	if err != nil {
		t.Error(err)
	}

	if len(chat) != 1 {
		t.Errorf("Количество полученных сообщений не соответствует нужному. Получено сообщений: %v, offset %v, userID %v, partnerID %v, relatedAnnouncementID %v", len(chat), offset, userID, partnerID, relatedAnnouncementID)
	}

	chat, err = GetChatHistory(userID, userID, relatedAnnouncementID, offset)
	if err == nil {
		t.Error("Получил историю чата с самим собой")
	}
}

func TestSendMessage(t *testing.T) {
	db := connectToDBForTest()
	messengerServiceHost = "localhost:8086"

	updMessages, err := updateTestUsers(db)
	if err != nil {
		t.Error(err)
	}

	updMessages, err = updateTestAnnouncements(db, updMessages)
	if err != nil {
		t.Error(err)
	}

	receivedID := updMessages.usersIDs[0]
	senderID := updMessages.usersIDs[1]
	relatedAnnouncementID := updMessages.announcementsIDs[0]
	messegeText := "Это тестовое сообщение"
	defer deleteTestMessage(db, messegeText)

	if err := SendMessege(receivedID, senderID, relatedAnnouncementID, ""); err == nil {
		t.Error("Получилось отправить пустое сообщение")
	}

	if err := SendMessege(senderID, senderID, relatedAnnouncementID, messegeText); err == nil {
		t.Error("Получилось отправить сообщение самому себе")
	}

	if err := SendMessege(senderID, receivedID, relatedAnnouncementID, messegeText); err != nil {
		t.Error(err)
	}

	rows, err := db.Query("SELECT received_id, sender_id, related_announcement_id FROM messages WHERE message = $1", messegeText)
	if err != nil {
		t.Error(err)
	}

	var queryReceivedID int
	var querySenderID int
	var queryRelatedAnnounID int
	rows.Next()
	rows.Scan(&queryReceivedID, &querySenderID, &queryRelatedAnnounID)
	if rows.Next() {
		t.Error("Было отправлено одно сообщение, но записано больше одного")
	}

	if queryReceivedID != receivedID || querySenderID != senderID || queryRelatedAnnounID != relatedAnnouncementID {
		t.Error("Один из ID не равен тому, что был отправлен")
	}
}

func TestGetChats(t *testing.T) {
	db := connectToDBForTest()
	messengerServiceHost = "localhost:8086"

	updMessages, err := updateTestMessages(db)
	if err != nil {
		t.Error(err)
	}

	userID := updMessages.usersIDs[0]

	chats, err := GetUserChats(userID)
	if err != nil {
		t.Error(err)
	}

	if len(chats) != 6 {
		t.Errorf("Количество чатов не соответствуюет нужному. UserID %v", userID)
	}
}
