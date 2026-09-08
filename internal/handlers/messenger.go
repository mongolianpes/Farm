package handlers

import (
	"context"
	"net/http"
	"strconv"

	"project-farm/internal/images"
	"project-farm/internal/messenger"
	"project-farm/internal/session"
)

const (
	limitMessagesToShow = 5
)

type MessengerData struct {
	Name                  string
	PartnerAvatar         string
	PartnerID             int
	RelatedAnnouncementID int
	Partners              []*messenger.ChatInfo
}

type SendMessage struct {
	Text                  string `json:"text"`
	ReceivedUsedID        int    `json:"receiveduserid"`
	RelatedAnnouncementID int    `json:"relatedannouncementid"`
}

func (h *Handler) MessengerHandler(w http.ResponseWriter, r *http.Request) {
	pageData := MessengerData{}

	partnerID := r.URL.Query().Get("partnerid")
	if partnerID == "" {
		userID, err := session.GetUserID(h.DB, h.RedisDB, w, r)
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
		defer cancel()

		chats, err := h.Messenger.GetUserChats(ctx, userID)
		if err != nil {
			http.Error(w, "Ошибка при получении чатов", http.StatusInternalServerError)
			return
		}

		for _, chat := range chats {
			pageData.Partners = append(pageData.Partners, &messenger.ChatInfo{
				PartnerID:             chat.PartnerID,
				PartnerName:           chat.PartnerName,
				PartnerAvatarPath:     images.MakeCurrentPathToImage(chat.PartnerAvatarPath),
				RelatedAnnouncementID: chat.RelatedAnnouncementID,
			})
		}
	} else {
		var err error
		pageData.PartnerID, err = strconv.Atoi(partnerID)
		if err != nil {
			http.Error(w, "Неверный partnerID", http.StatusBadRequest)
			return
		}

		if err := h.DB.QueryRow("SELECT name, avatar_path FROM users WHERE user_id = $1", partnerID).Scan(&pageData.Name, &pageData.PartnerAvatar); err != nil {
			http.Error(w, "Не удалось получить партнера", http.StatusInternalServerError)
			return
		}

		relatedAnnouncementID := r.URL.Query().Get("relatedannouncementid")
		pageData.RelatedAnnouncementID, err = strconv.Atoi(relatedAnnouncementID)
		if err != nil {
			http.Error(w, "Неверное relatedAnnouncementID", http.StatusBadRequest)
			return
		}

		pageData.PartnerAvatar = images.MakeCurrentPathToImage(pageData.PartnerAvatar)
	}

	h.Tmpl.ExecuteTemplate(w, "messanger.html", pageData)
}

func (h *Handler) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/messenger", http.StatusSeeOther)
		return
	}

	userID, err := session.GetUserID(h.DB, h.RedisDB, w, r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	var sendMessage SendMessage
	if err := json.NewDecoder(r.Body).Decode(&sendMessage); err != nil {
		http.Error(w, "Не удалось разобрать полученный данные", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	if err := h.Messenger.SendMessege(ctx, userID, sendMessage.ReceivedUsedID, sendMessage.RelatedAnnouncementID, sendMessage.Text); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := session.GetUserID(h.DB, h.RedisDB, w, r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	partnerIDStr := r.URL.Query().Get("partnerid")
	if partnerIDStr == "" {
		http.Error(w, "Необходимо укзать параметр partnerid", http.StatusBadRequest)
		return
	}
	partnerID, err := strconv.Atoi(partnerIDStr)
	if err != nil {
		http.Error(w, "Должно быть числом partnerid", http.StatusBadRequest)
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

	relatedAnnouncementIDStr := r.URL.Query().Get("relatedannouncementid")
	if relatedAnnouncementIDStr == "" {
		http.Error(w, "Необходимо укзать параметр relatedannouncementid", http.StatusBadRequest)
		return
	}
	relatedAnnouncementID, err := strconv.Atoi(relatedAnnouncementIDStr)
	if err != nil {
		http.Error(w, "Должно быть числом relatedannouncementid", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	chat, err := h.Messenger.GetChatHistory(ctx, userID, partnerID, relatedAnnouncementID, offsetInt)
	if err != nil {
		http.Error(w, "Ошибка получения чата", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(chat)
}
