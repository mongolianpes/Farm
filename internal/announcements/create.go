package announcements

import (
	"context"
	"errors"
	"log/slog"
	"mime/multipart"

	pb "project-farm/internal/announcements/proto"
	imagesService "project-farm/internal/images"
)

const maxImagesSize5MB = 5 * 1024 * 1024

func CreateAnnouncement(userID int, title, description, category string, images []*multipart.FileHeader) error {
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

	announcementID, err := sendReqCreateAnnouncement(int32(userID), title, description, category)
	if err != nil {
		slog.Warn("Не удалось создать объявление", "announcementID", announcementID, "userID", userID, "error", err)
		return err
	}

	if len(images) >= 1 {
		if err := sendReqAddImages(images, announcementID, int32(userID)); err != nil {
			slog.Warn("Не удалось создать объявление, поскольку не удалось загрузить картинки", "announcementID", announcementID, "userID", userID, "error", err)
			return err
		}
	}

	slog.Warn("Успешно создано объявление", "announcementID", announcementID, "userID", userID)

	return nil
}

func sendReqCreateAnnouncement(authorID int32, title, description, category string) (int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	respCreate, err := client.service.CreateAnnouncement(ctx, &pb.CreateAnnouncementRequest{
		Title:       title,
		Description: description,
		Category:    category,
		AuthorID:    authorID,
	})
	if err != nil {
		return 0, err
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

	_, err = client.service.AddImages(ctx, &pb.AddImagesRequest{
		ImagesPath:     imagesPath,
		AnnouncementID: announcementdID,
	})
	if err != nil {
		DeleteAnnouncement(int(announcementdID), int(userID))
		return err
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
