package messenger

import (
	"context"
	"log/slog"
	"os"

	pb "project-farm/internal/messenger/proto"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	service pb.MessengerClient
	conn    *grpc.ClientConn
}

type Messenger interface {
	SendMessege(ctx context.Context, senderID, receivedID, relatedAnnouncementID int, messageText string) error
	GetUserChats(ctx context.Context, userID int) ([]*ChatInfo, error)
	GetChatHistory(ctx context.Context, userID, partnerID, relatedAnnouncementID, offset int) ([]*message, error)
}

var messengerServiceHost = os.Getenv("MESSENGER_SERVICE_HOST_GRPC_PORT")

func NewClient() (*Client, error) {
	client := &Client{}
	if client.service != nil {
		return client, nil
	}

	conn, err := grpc.NewClient(messengerServiceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Не удалось создать подключение к микросервису Announcements", "error", err)
		return client, err
	}

	client.service = pb.NewMessengerClient(conn)
	client.conn = conn
	return client, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
