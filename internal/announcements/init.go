package announcements

import (
	"log/slog"
	"os"
	pb "project-farm/internal/announcements/proto"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const timeToCompleteRequest = 30 * time.Second

type announcementsClient struct {
	sync.Mutex
	service pb.AnnouncementsClient
	conn    *grpc.ClientConn
}

var client announcementsClient
var announcementsServiceHost = os.Getenv("ANNOUNCEMENTS_SERVICE_HOST_GRPC_PORT")

func initService() error {
	client.Lock()
	defer client.Unlock()
	if client.service != nil {
		return nil
	}

	conn, err := grpc.NewClient(announcementsServiceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Не удалось создать подключение к микросервису Announcements", "error", err)
		return err
	}

	client.service = pb.NewAnnouncementsClient(conn)
	client.conn = conn
	return nil
}

func CloseConnectionToService() error {
	if client.conn == nil {
		return nil
	}
	return client.conn.Close()
}
