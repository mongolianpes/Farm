package handlers

import (
	"net/http"
	"strconv"
	"time"

	"project-farm/internal/images"
	"project-farm/internal/session"
)

const (
	limitMessagesToShow = 5
)

type MessengerData struct {
	Name          string
	PartnerAvatar string
	PartnerID     string
	Partners      []PartnerInfo
}

type PartnerInfo struct {
	Name       string
	AvatarPath string
	ID         int
}

type MessageData struct {
	CreateAt  time.Time
	Text      string
	MyMessage bool
}

type SendMessage struct {
	Text               string `json:"text"`
	ReceivedUsed       string `json:"receiveduser"`
	InitAnnouncementID int    `json:"initannouncementid"`
}

func (h *Handler) MessengerHandler(w http.ResponseWriter, r *http.Request) {
	pageData := MessengerData{}

	partnerID := r.URL.Query().Get("partnerid")
	if partnerID == "" {
		userID, err := session.GetUserID(h.DB, w, r)
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}

		query := `SELECT DISTINCT 
	CASE 
        WHEN sender_id = $1 THEN received_id 
        ELSE sender_id 
    END AS partner_id
FROM messages 
WHERE sender_id = $1 OR received_id = $1;`
		rows, err := h.DB.Query(query, userID)
		if err != nil {
			http.Error(w, "Ошибка получения ID партнера", http.StatusInternalServerError)
			return
		}

		var receivedID int
		var receivedName string
		var receivedAvatarPath string
		for rows.Next() {
			if err := rows.Scan(&receivedID); err != nil {
				http.Error(w, "Не удалось получить партнеров", http.StatusInternalServerError)
				return
			}

			if err := h.DB.QueryRow("SELECT name, avatar_path FROM users WHERE user_id = $1", receivedID).Scan(&receivedName, &receivedAvatarPath); err != nil {
				http.Error(w, "Не получили имя и аватар партнера", http.StatusInternalServerError)
				return
			}

			receivedAvatarPath = images.MakeCurrentPathToImage(receivedAvatarPath)

			pageData.Partners = append(pageData.Partners, PartnerInfo{
				Name:       receivedName,
				AvatarPath: receivedAvatarPath,
				ID:         receivedID,
			})
		}
	} else {
		pageData.PartnerID = partnerID

		if err := h.DB.QueryRow("SELECT name, avatar_path FROM users WHERE user_id = $1", partnerID).Scan(&pageData.Name, &pageData.PartnerAvatar); err != nil {
			http.Error(w, "Не удалось получить партнера", http.StatusInternalServerError)
			return
		}

		// initAnnouncementID := r.URL.Query().Get("announcementid")

		pageData.PartnerAvatar = images.MakeCurrentPathToImage(pageData.PartnerAvatar)
	}

	h.Tmpl.ExecuteTemplate(w, "messanger.html", pageData)
}

func (h *Handler) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/messenger", http.StatusSeeOther)
		return
	}

	userID, err := session.GetUserID(h.DB, w, r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	var sendMessage SendMessage
	if err := json.NewDecoder(r.Body).Decode(&sendMessage); err != nil {
		http.Error(w, "Сообщение не отправлено", http.StatusBadRequest)
		return
	}

	if sendMessage.Text == "" {
		http.Error(w, "Сообщение не должно быть пустым", http.StatusBadRequest)
		return
	}

	if _, err := h.DB.Exec("INSERT INTO messages (sender_id, received_id, message) VALUES ($1, $2, $3)", userID, sendMessage.ReceivedUsed, sendMessage.Text); err != nil {
		http.Error(w, "Сообщение не отправлено", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := session.GetUserID(h.DB, w, r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	partnerID := r.URL.Query().Get("partnerid")
	if partnerID == "" {
		http.Error(w, "Необходимо укзать параметр partnerid", http.StatusBadRequest)
		return
	}

	var offsetInt int
	offset := r.URL.Query().Get("offset")
	if offset == "" {
		offsetInt = 0
	} else {
		offsetInt, err = strconv.Atoi(offset)
		if err != nil {
			http.Error(w, "Неверный параметр offset", http.StatusBadRequest)
			return
		}
	}

	query := "SELECT create_at, message, sender_id FROM messages WHERE (sender_id = $1 AND received_id = $2) OR (sender_id = $2 AND received_id = $1) ORDER BY create_at DESC OFFSET $3 LIMIT $4"
	rows, err := h.DB.Query(query, userID, partnerID, offsetInt*limitMessagesToShow, limitMessagesToShow)
	if err != nil {
		http.Error(w, "Не удалось получить данные из БД", http.StatusInternalServerError)
		return
	}

	var messages []MessageData

	var myMessage bool
	var createAt time.Time
	var text string
	var senderID int
	for rows.Next() {
		if err := rows.Scan(&createAt, &text, &senderID); err != nil {
			http.Error(w, "Не удалось просканировать данные из БД", http.StatusInternalServerError)
			return
		}

		if senderID == userID {
			myMessage = true
		} else {
			myMessage = false
		}

		messages = append(messages, MessageData{
			CreateAt:  createAt,
			Text:      text,
			MyMessage: myMessage,
		})
	}

	json.NewEncoder(w).Encode(messages)
}
