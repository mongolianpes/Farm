package announcements

import (
	"context"
	"errors"
	"mime/multipart"

	pb "project-farm/internal/announcements/proto"
	imagesService "project-farm/internal/images"
)

const maxImagesSize5MB = 5 * 1024 * 1024

func CreateAnnouncement(userID int, title, description, category, authorID string, images []*multipart.FileHeader) error {
	if err := initService(); err != nil {
		return errors.New("Попробуйте создать чуть позже")
	}

	if category == "" || title == "" || description == "" {
		return errors.New("Введите информацию во всех полях")
	}

	if len(description) > 100 || len(title) > 10 {
		return errors.New("Название должно быть не больше 10 символов, а описание не больше 100")
	}

	if len(images) > 10 {
		return errors.New("Можно загрузить до 10 изображений")
	}

	for _, image := range images {
		if image.Size > maxImagesSize5MB {
			return errors.New("Одно из картинок весом больше 5 MB")
		}
	}

	announcementID, err := sendReqCreateAnnouncement(title, description, category, authorID)
	if err != nil {
		return err
	}

	if len(images) >= 1 {
		if err := sendReqAddImages(images, announcementID, int32(userID)); err != nil {
			return err
		}
	}

	return err
}

func sendReqCreateAnnouncement(title, description, category, authorID string) (int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	streamCreate, err := client.service.CreateAnnouncement(ctx)
	if err != nil {
		return 0, errors.New("Попробуйте создать чуть позже")
	}

	req := &pb.CreateAnnouncementRequest{
		Title:       title,
		Description: description,
		Category:    category,
		AuthorID:    authorID,
	}

	if err := streamCreate.Send(req); err != nil {
		return 0, errors.New("Попробуйте создать чуть позже")
	}

	respCreate, err := streamCreate.CloseAndRecv()
	if err != nil {
		return 0, err
	}
	if respCreate.Error != "" {
		return 0, errors.New(respCreate.Error)
	}

	return respCreate.AnnouncementID, nil
}

func sendReqAddImages(images []*multipart.FileHeader, announcementdID, userID int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	imagesPath, err := saveImages(images)
	if err != nil {
		DeleteAnnouncement(int(announcementdID), int(userID))
		return err
	}

	streamAddImages, err := client.service.AddImages(ctx)
	if err != nil {
		DeleteAnnouncement(int(announcementdID), int(userID))
		return err
	}

	if err := streamAddImages.Send(&pb.AddImagesRequest{
		ImagesPath:     imagesPath,
		AnnouncementID: announcementdID,
	}); err != nil {
		return err
	}

	respAddImages, err := streamAddImages.CloseAndRecv()
	if err != nil {
		DeleteAnnouncement(int(announcementdID), int(userID))
		return err
	}
	if respAddImages.Error != "" {
		return errors.New(respAddImages.Error)
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
