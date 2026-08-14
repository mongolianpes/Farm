package images

import (
	"context"
	"io"
	"log/slog"
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
		slog.Warn("Не удалось создать подключение к микросервису Images")
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
		slog.Warn("Не создать стрим с микросервисом Images", "error", err)
		return "", err
	}

	imageData, err := io.ReadAll(file)
	if err != nil {
		slog.Warn("Не удалось прочать байты переданного файла для передачи микросервису Images", "error", err)
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
		slog.Warn("Не удалось совершить запрос к микросервису Images", "error", err)
		return "", err
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		slog.Warn("Ошибка при закрытии стрима к микросервису Images", "error", err)
		return "", err
	}

	if resp.Error != "" {
		err := SaveImageError{
			Message:           resp.Error,
			CountStoragePaths: len(resp.StoragePath),
		}
		slog.Warn("Ошибка при закрытии стрима к микросервису Images", "error", err)
		return "", err
	}

	if len(resp.StoragePath) == 0 {
		err := SaveImageError{
			Message:           resp.Error,
			CountStoragePaths: len(resp.StoragePath),
		}
		slog.Warn("Ошибка при закрытии стрима к микросервису Images", "error", err)
		return "", err
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
