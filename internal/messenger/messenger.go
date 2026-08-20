package messenger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "project-farm/internal/messenger/proto"
)

func (c *Client) SendMessege(ctx context.Context, senderID, receivedID, relatedAnnouncementID int, messageText string) error {
	_, err := c.service.SendMessage(ctx, &pb.SendMessageRequest{
		ReceivedID:            int32(receivedID),
		SenderID:              int32(senderID),
		RelatedAnnouncementID: int32(relatedAnnouncementID),
		Text:                  messageText,
	})
	if err != nil {
		slog.Warn("Ошибка при отправке сообщения во внутрененм мессенджере", "receivedID", receivedID, "senderID", senderID)
		return err
	}

	return nil
}

type ChatInfo struct {
	PartnerID             int
	PartnerName           string
	PartnerAvatarPath     string
	RelatedAnnouncementID int
}

func (c *Client) GetUserChats(ctx context.Context, userID int) ([]*ChatInfo, error) {
	resp, err := c.service.GetChats(ctx, &pb.GetChatsRequest{
		UserID: int32(userID),
	})
	if err != nil {
		slog.Warn("Ошибка при получении чатов пользователей", "error", err)
		return nil, err
	}

	chatsInfo := []*ChatInfo{}
	for _, chat := range resp.Chats {
		chatsInfo = append(chatsInfo, &ChatInfo{
			PartnerID:             int(chat.PartnerID),
			PartnerName:           fmt.Sprintf("%s (%s)", chat.PartnerName, chat.RelatedAnnouncementTitle),
			PartnerAvatarPath:     chat.PartnerAvatarPath,
			RelatedAnnouncementID: int(chat.RelatedAnnouncementID),
		})
	}

	return chatsInfo, nil
}

type message struct {
	Text      string
	MyMessage bool
	CreateAt  time.Time
}

func (c *Client) GetChatHistory(ctx context.Context, userID, partnerID, relatedAnnouncementID, offset int) ([]*message, error) {
	resp, err := c.service.GetChatHistory(ctx, &pb.GetChatHistoryRequest{
		PartnerID:             int32(partnerID),
		UserID:                int32(userID),
		RelatedAnnouncementID: int32(relatedAnnouncementID),
		Offset:                int32(offset),
	})
	if err != nil {
		slog.Warn("Ошибка при получении истории чата", "error", err)
		return nil, err
	}

	messages := []*message{}
	for _, mess := range resp.Messages {
		messages = append(messages, &message{
			Text:      mess.Text,
			MyMessage: mess.MyMessage,
			CreateAt:  mess.CreateAt.AsTime(),
		})
	}

	return messages, nil
}
