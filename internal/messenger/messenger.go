package messenger

import (
	"context"
	"fmt"
	"time"

	pb "project-farm/internal/messenger/proto"
)

func SendMessege(senderID, receivedID, relatedAnnouncementID int, messageText string) error {
	if err := initService(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	_, err := client.service.SendMessage(ctx, &pb.SendMessageRequest{
		ReceivedID:            int32(receivedID),
		SenderID:              int32(senderID),
		RelatedAnnouncementID: int32(relatedAnnouncementID),
		Text:                  messageText,
	})
	if err != nil {
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

func GetUserChats(userID int) ([]*ChatInfo, error) {
	if err := initService(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	resp, err := client.service.GetChats(ctx, &pb.GetChatsRequest{
		UserID: int32(userID),
	})
	if err != nil {
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

func GetChatHistory(userID, partnerID, relatedAnnouncementID, offset int) ([]*message, error) {
	if err := initService(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	resp, err := client.service.GetChatHistory(ctx, &pb.GetChatHistoryRequest{
		PartnerID:             int32(partnerID),
		UserID:                int32(userID),
		RelatedAnnouncementID: int32(relatedAnnouncementID),
		Offset:                int32(offset),
	})
	if err != nil {
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
