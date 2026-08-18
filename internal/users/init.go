package users

import (
	"log/slog"
	"os"
	"sync"
	"time"

	pb "project-farm/internal/users/proto"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const timeToCompleteRequest = 30 * time.Second

type usersClient struct {
	sync.Mutex
	service pb.UsersClient
	conn    *grpc.ClientConn
}

var client usersClient

var usersServiceHost = os.Getenv("USERS_SERVICE_HOST_GRPC_PORT")

func initService() error {
	client.Lock()
	defer client.Unlock()
	if client.service != nil {
		return nil
	}

	conn, err := grpc.NewClient(usersServiceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Не удалось создать подключение к микросервису Users", "error", err)
		return err
	}

	client.service = pb.NewUsersClient(conn)
	client.conn = conn
	return nil
}

func CloseConnectionToService() error {
	if client.conn == nil {
		return nil
	}
	return client.conn.Close()
}
