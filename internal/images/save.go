package images

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "project-farm/internal/images/proto"
)

type imagesClient struct {
	sync.Mutex
	service pb.ImageServiceClient
	conn    *grpc.ClientConn
}

var client imagesClient
var imagesServiceHost = os.Getenv("IMAGES_SERVICE_HOST_GRPC_PORT")

const timeToCompleteRequest = 30 * time.Second

func initService() error {
	client.Lock()
	defer client.Unlock()
	if client.service != nil {
		return nil
	}

	conn, err := grpc.NewClient(imagesServiceHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	client.conn = conn
	client.service = pb.NewImageServiceClient(conn)
	return nil
}

func CloseConnectionToService() error {
	return client.conn.Close()
}

func SaveImage(width, height int32, file multipart.File) (string, error) {
	if err := initService(); err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	stream, err := client.service.DownloadImages(ctx)
	if err != nil {
		return "", err
	}

	imageData, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	req := &pb.DownloadImagesRequest{
		Info: &pb.ImageInfo{
			Compress: "low",
			Format:   "webp",
		},
		Image: imageData,
	}

	if width != 0 && height != 0 {
		req.Info.Width = []int32{width}
		req.Info.Height = []int32{height}
	}

	if err := stream.Send(req); err != nil {
		return "", err
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return "", err
	}

	if resp.Error != "" {
		return "", SaveImageError{
			Message:           resp.Error,
			CountStoragePaths: len(resp.StoragePath),
		}
	}

	if len(resp.StoragePath) == 0 {
		return "", SaveImageError{
			Message:           resp.Error,
			CountStoragePaths: len(resp.StoragePath),
		}
	}

	return strings.Replace(resp.StoragePath[len(resp.StoragePath)-1], "./files/", "", 1), nil
}

type SaveImageError struct {
	Message           string
	CountStoragePaths int
}

func (saveImage SaveImageError) Error() string {
	if saveImage.CountStoragePaths == 0 {
		return "Не были получены пути к изображениям"
	}

	return saveImage.Message
}

func CheckCurrentFileExtansion(header *multipart.FileHeader) bool {
	filenameSplitted := strings.Split(header.Filename, ".")
	fileExtansion := filenameSplitted[len(filenameSplitted)-1]

	fileExtansion = strings.ToLower(fileExtansion)
	switch fileExtansion {
	case "png", "jpg", "jpeg", "webp":
		return true
	default:
		return false
	}
}
