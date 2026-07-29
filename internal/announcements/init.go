package announcements

import (
	"os"
	pb "project-farm/internal/announcements/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var client pb.AnnouncementsClient
var announcementsServiceHost = os.Getenv("ANNOUNCEMENTS_SERVICE_HOST_GRPC_PORT")

func InitService() error {
	conn, err := grpc.NewClient(announcementsServiceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	client = pb.NewAnnouncementsClient(conn)
	return nil
}
