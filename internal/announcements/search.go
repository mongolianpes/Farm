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

func SearchAnnouncements(offset, announcementID int, userID, SearchString, category, orderBy, authorID string) ([]*AnnouncementData, error) {
	if err := initService(); err != nil {
		return []*AnnouncementData{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	stream, err := client.service.SearchAnnouncements(ctx)
	if err != nil {
		slog.Warn("Не удалось выполнить поиск по обявлениям", "error", err)
		return []*AnnouncementData{}, err
	}

	req := &pb.SearchAnnouncementsRequest{
		Offset:         int32(offset),
		UserID:         userID,
		SearchString:   SearchString,
		Category:       category,
		Orderby:        orderBy,
		AnnouncementID: int32(announcementID),
		AuthorID:       authorID,
	}

	if err := stream.Send(req); err != nil {
		slog.Warn("Не удалось выполнить поиск по обявлениям", "error", err)
		return []*AnnouncementData{}, err
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		slog.Warn("Не удалось выполнить поиск по обявлениям", "error", err)
		return []*AnnouncementData{}, err
	}
	if resp.Error != "" {
		slog.Warn("Не удалось выполнить поиск по обявлениям", "error", err)
		return []*AnnouncementData{}, errors.New(resp.Error)
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

	stream, err := client.service.SearchAnnouncements(ctx)
	if err != nil {
		slog.Warn("Не удалось получить информацию по объявлению", "error", err)
		return AnnouncementData{}, err
	}

	req := &pb.SearchAnnouncementsRequest{
		AnnouncementID: int32(announcementID),
		UserID:         strconv.Itoa(userID),
	}

	if err := stream.Send(req); err != nil {
		return AnnouncementData{}, err
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return AnnouncementData{}, err
	}
	if resp.Error != "" {
		return AnnouncementData{}, errors.New(resp.Error)
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
