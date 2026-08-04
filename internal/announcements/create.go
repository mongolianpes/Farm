package announcements

import (
	"context"
	"errors"
	"mime/multipart"

	pb "project-farm/internal/announcements/proto"
	imagesService "project-farm/internal/images"
)

func CreateAnnouncement(title, description, category, authorID string, images []*multipart.FileHeader) error {
	if err := initService(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	stream, err := client.service.CreateAnnouncement(ctx)
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
			continue
		}
		defer file.Close()

		if !imagesService.CheckCurrentFileExtansion(fileHeader) {
			continue
		}

		filenameForDataBase, err := imagesService.SaveImage(0, 0, file)
		if err != nil {
			continue
		}

		imagesForDataBase = append(imagesForDataBase, filenameForDataBase)
	}

	if len(imagesForDataBase) == 0 {
		return imagesForDataBase, errors.New("Загружайте картинки в форматах: png, jpg, webp")
	}
	return imagesForDataBase, nil
}
