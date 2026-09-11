package announcements

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"project-farm/internal/models"
	"project-farm/internal/rdb"

	pb "project-farm/internal/announcements/proto"
)

func SearchAnnouncements(rdb rdb.DB, offset, userID, authorID int, SearchString, category, orderBy string) ([]*models.AnnouncementData, error) {
	if err := initService(); err != nil {
		return []*models.AnnouncementData{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	result := []*models.AnnouncementData{}

	if offset == 0 {
		savedAnnouncementIDs, err := rdb.GetAnnouncementsIDsForUser(ctx, strconv.Itoa(userID))
		if err == nil && len(savedAnnouncementIDs) != 0 {
			for _, id := range savedAnnouncementIDs {
				idStr, err := strconv.Atoi(id)
				if err != nil {
					break
				}
				announcement, err := rdb.GetAnnouncementInfo(ctx, idStr)
				if err != nil {
					break
				}

				result = append(result, announcement)
			}

			return result, nil
		}
	}

	resp, err := client.service.SearchAnnouncements(ctx, &pb.SearchAnnouncementsRequest{
		Offset:       int64(offset),
		UserID:       int64(userID),
		SearchString: SearchString,
		Category:     category,
		Orderby:      orderBy,
		AuthorID:     int64(authorID),
	})
	if err != nil {
		slog.Warn("Не удалось получить список объявлений", "error", err)
		return nil, err
	}

	var announcementIDs []interface{}
	for _, announcement := range resp.AnnouncementsData {
		result = append(result, &models.AnnouncementData{
			AuthorName:         announcement.AuthorName,
			AuthorID:           int(announcement.AuthorID),
			Title:              announcement.Title,
			Description:        announcement.Description,
			Category:           announcement.Category,
			LinkToAnnouncement: announcement.LinkToAnnouncement,
			AnnouncementID:     int(announcement.AnnouncementID),
			Images:             announcement.Images,
		})

		announcementIDs = append(announcementIDs, announcement.AnnouncementID)
	}

	if err := rdb.SaveAnnouncementIDsForUser(ctx, announcementIDs, strconv.Itoa(userID)); err != nil {
		slog.Error("Не удалось сохранить список объявлений в redis")
	} else {
		for _, announcement := range result {
			if err := rdb.SaveAnnouncementInfo(ctx, *announcement); err != nil {
				slog.Error("Не удалось сохранить объявлениe в redis")
			}
		}
	}

	return result, nil
}

func GetAnnouncementInfo(rdb rdb.DB, announcementID, userID int) (models.AnnouncementData, error) {
	if err := initService(); err != nil {
		return models.AnnouncementData{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	savedAnnouncement, err := rdb.GetAnnouncementInfo(ctx, announcementID)
	if err == nil {
		return *savedAnnouncement, nil
	}

	resp, err := client.service.SearchAnnouncements(ctx, &pb.SearchAnnouncementsRequest{
		AnnouncementID: int64(announcementID),
		UserID:         int64(userID),
	})
	if err != nil {
		slog.Warn("Не удалось получить объявление", "announcementID", announcementID, "userID", userID, "error", err)
		return models.AnnouncementData{}, err
	}

	for _, announcement := range resp.AnnouncementsData {
		result := models.AnnouncementData{
			AuthorName:         announcement.AuthorName,
			AuthorID:           int(announcement.AuthorID),
			Title:              announcement.Title,
			Description:        announcement.Description,
			Category:           announcement.Category,
			LinkToAnnouncement: announcement.LinkToAnnouncement,
			AnnouncementID:     int(announcement.AnnouncementID),
			Images:             announcement.Images,
		}

		if err := rdb.SaveAnnouncementInfo(ctx, result); err != nil {
			slog.Error("Не удалось сохранить объявлениe в redis")
		}

		return result, nil
	}

	return models.AnnouncementData{}, errors.New("Данного объявления не существует")
}
