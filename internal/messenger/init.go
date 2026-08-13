package messenger

import (
	"os"
	pb "project-farm/internal/messenger/proto"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const timeToCompleteRequest = 30 * time.Second

type messengerClient struct {
	sync.Mutex
	service pb.MessengerClient
	conn    *grpc.ClientConn
}

var client messengerClient
var messengerServiceHost = os.Getenv("MESSENGER_SERVICE_HOST_GRPC_PORT")

func initService() error {
	client.Lock()
	defer client.Unlock()
	if client.service != nil {
		return nil
	}

	conn, err := grpc.NewClient(messengerServiceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	client.service = pb.NewMessengerClient(conn)
	client.conn = conn
	return nil
}

func CloseConnectionToService() error {
	return client.conn.Close()
}
