package announcements

import (
	"context"
	"errors"
	"mime/multipart"

	pb "project-farm/internal/announcements/proto"
	imagesService "project-farm/internal/images"
)

func CreateAnnouncement(title, description, category, authorID string, images []*multipart.FileHeader) error {
	stream, err := client.CreateAnnouncement(context.Background())
	if err != nil {
		return err
	}

	req := &pb.CreateAnnouncementRequest{
		Title:       title,
		Description: description,
		Category:    category,
		AuthorID:    authorID,
	}

	if err := stream.Send(req); err != nil {
		return err
	}

	resp, err := stream.Recv()
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	if len(images) >= 1 {
		imagesPath, err := saveImages(images)
		if err != nil {
			return err
		}

		if err := stream.Send(&pb.CreateAnnouncementRequest{
			ImagesPath: imagesPath,
		}); err != nil {
			return err
		}
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}

	final, err := stream.Recv()
	if err != nil {
		return err
	}
	if final.Error != "" {
		return errors.New(final.Error)
	}

	return nil
}

func saveImages(images []*multipart.FileHeader) ([]string, error) {
	var imagesForDataBase []string
	for _, fileHeader := range images {
		file, err := fileHeader.Open()
		if err != nil {
			return imagesForDataBase, errors.New("Ошибка открытия одного из файлов")
		}
		defer file.Close()

		if !imagesService.CheckCurrentFileExtansion(fileHeader) {
			return imagesForDataBase, errors.New("Попробуйте загрузить картинку в другом формате (png, jpg, webp)")
		}

		filenameForDataBase, err := imagesService.SaveImage(0, 0, file)
		if err != nil {
			return imagesForDataBase, errors.New("Ошибка сохранеия одного из файлов")
		}

		imagesForDataBase = append(imagesForDataBase, filenameForDataBase)
	}

	return imagesForDataBase, nil
}
