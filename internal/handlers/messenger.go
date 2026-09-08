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
	PartnerID             string
	RelatedAnnouncementID string
	Partners              []*messenger.ChatInfo
}

type SendMessage struct {
	Text                  string `json:"text"`
	ReceivedUsedID        int    `json:"receiveduserid"`
	RelatedAnnouncementID int    `json:"relatedannouncementid"`
}

func (h *Handler) MessengerHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()
	pageData := MessengerData{}

	partnerID := r.URL.Query().Get("partnerid")
	if partnerID == "" {
		sessionID, err := session.GetCookie(r)
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		newSession, userID, err := session.GetUserID(ctx, h.DB, h.RedisDB, sessionID)
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		if newSession != "" {
			session.SetCookie(w, newSession)
		}

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
		pageData.PartnerID = partnerID

		if err := h.DB.QueryRow("SELECT name, avatar_path FROM users WHERE user_id = $1", partnerID).Scan(&pageData.Name, &pageData.PartnerAvatar); err != nil {
			http.Error(w, "Не удалось получить партнера", http.StatusInternalServerError)
			return
		}

		relatedAnnouncementID := r.URL.Query().Get("relatedannouncementid")
		pageData.RelatedAnnouncementID = relatedAnnouncementID

		pageData.PartnerAvatar = images.MakeCurrentPathToImage(pageData.PartnerAvatar)
	}

	h.Tmpl.ExecuteTemplate(w, "messanger.html", pageData)
}

func (h *Handler) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/messenger", http.StatusSeeOther)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	sessionID, err := session.GetCookie(r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	newSession, userID, err := session.GetUserID(ctx, h.DB, h.RedisDB, sessionID)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	if newSession != "" {
		session.SetCookie(w, newSession)
	}

	var sendMessage SendMessage
	if err := json.NewDecoder(r.Body).Decode(&sendMessage); err != nil {
		http.Error(w, "Не удалось разобрать полученный данные", http.StatusBadRequest)
		return
	}

	if err := h.Messenger.SendMessege(ctx, userID, sendMessage.ReceivedUsedID, sendMessage.RelatedAnnouncementID, sendMessage.Text); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	sessionID, err := session.GetCookie(r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	newSession, userID, err := session.GetUserID(ctx, h.DB, h.RedisDB, sessionID)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	if newSession != "" {
		session.SetCookie(w, newSession)
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

	chat, err := h.Messenger.GetChatHistory(ctx, userID, partnerID, relatedAnnouncementID, offsetInt)
	if err != nil {
		http.Error(w, "Ошибка получения чата", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(chat)
}
