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

	streamCreate, err := client.service.CreateAnnouncement(ctx)
	if err != nil {
		return err
	}

	req := &pb.CreateAnnouncementRequest{
		Title:       title,
		Description: description,
		Category:    category,
		AuthorID:    authorID,
	}

	if err := streamCreate.Send(req); err != nil {
		return err
	}

	respCreate, err := streamCreate.CloseAndRecv()
	if err != nil {
		return err
	}
	if respCreate.Error != "" {
		return errors.New(respCreate.Error)
	}

	if len(images) >= 1 {
		imagesPath, err := saveImages(images)
		if err != nil {
			return err
		}

		streamAddImages, err := client.service.AddImages(ctx)
		if err != nil {
			DeleteAnnouncement(int(respCreate.AnnouncementID))
			return err
		}

		if err := streamAddImages.Send(&pb.AddImagesRequest{
			ImagesPath:     imagesPath,
			AnnouncementID: respCreate.AnnouncementID,
		}); err != nil {
			return err
		}

		respAddImages, err := streamAddImages.CloseAndRecv()
		if err != nil {
			DeleteAnnouncement(int(respCreate.AnnouncementID))
			return err
		}
		if respAddImages.Error != "" {
			return errors.New(respAddImages.Error)
		}
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

		if !imagesService.CheckCurrentFileExtansion(fileHeader) {
			continue
		}

		filenameForDataBase, err := imagesService.SaveImage(0, 0, file)
		if err != nil {
			continue
		}

		imagesForDataBase = append(imagesForDataBase, filenameForDataBase)
		file.Close()
	}

	if len(imagesForDataBase) == 0 {
		return imagesForDataBase, errors.New("Загружайте картинки в форматах: png, jpg, webp")
	}
	return imagesForDataBase, nil
}
