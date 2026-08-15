package announcements

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	pb "project-farm/internal/announcements/proto"
)

type AnnouncementData struct {
	AuthorName         string
	AuthorID           int
	Title              string
	Description        string
	Category           string
	LinkToAnnouncement string
	AnnouncementID     int
	Images             []string
}

func SearchAnnouncements(offset int, userID, SearchString, category, orderBy, authorID string) ([]*AnnouncementData, error) {
	if err := initService(); err != nil {
		return []*AnnouncementData{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	resp, err := client.service.SearchAnnouncements(ctx, &pb.SearchAnnouncementsRequest{
		Offset:       int32(offset),
		UserID:       userID,
		SearchString: SearchString,
		Category:     category,
		Orderby:      orderBy,
		AuthorID:     authorID,
	})
	if err != nil {
		slog.Warn("Не удалось получить список объявлений", "error", err)
		return nil, err
	}

	result := []*AnnouncementData{}
	for _, announcement := range resp.AnnouncementsData {
		result = append(result, &AnnouncementData{
			AuthorName:         announcement.AuthorName,
			AuthorID:           int(announcement.AuthorID),
			Title:              announcement.Title,
			Description:        announcement.Description,
			Category:           announcement.Category,
			LinkToAnnouncement: announcement.LinkToAnnouncement,
			AnnouncementID:     int(announcement.AnnouncementID),
			Images:             announcement.Images,
		})
	}

	return result, nil
}

func GetAnnouncementInfo(announcementID, userID int) (AnnouncementData, error) {
	if err := initService(); err != nil {
		return AnnouncementData{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	resp, err := client.service.SearchAnnouncements(ctx, &pb.SearchAnnouncementsRequest{
		AnnouncementID: int32(announcementID),
		UserID:         strconv.Itoa(userID),
	})
	if err != nil {
		slog.Warn("Не удалось получить объявление", "announcementID", announcementID, "userID", userID, "error", err)
		return AnnouncementData{}, err
	}

	for _, announcement := range resp.AnnouncementsData {
		return AnnouncementData{
			AuthorName:         announcement.AuthorName,
			AuthorID:           int(announcement.AuthorID),
			Title:              announcement.Title,
			Description:        announcement.Description,
			Category:           announcement.Category,
			LinkToAnnouncement: announcement.LinkToAnnouncement,
			AnnouncementID:     int(announcement.AnnouncementID),
			Images:             announcement.Images,
		}, nil
	}

	return AnnouncementData{}, errors.New("Данного объявления не существует")
}
