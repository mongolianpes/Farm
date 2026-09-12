package announcements

import (
	"context"
	"log/slog"

	pb "project-farm/internal/announcements/proto"
	"project-farm/internal/rdb"
)

func DeleteAnnouncement(rdb rdb.DB, announcementID, userID int) error {
	if err := initService(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	if err := sendReqDeleteAnnouncement(ctx, announcementID, userID); err != nil {
		slog.Warn("Не удалось удалить объявления", "announcementID", announcementID, "userID", userID, "error", err)
		return err
	}

	if err := rdb.DelAnnouncementInfo(ctx, announcementID); err != nil {
		slog.Warn("Не удалось удалить объявления из Redis", "announcementID", announcementID, "userID", userID, "error", err)
	}

	slog.Warn("Успешное удаление объявления", "announcementID", announcementID, "userID", userID)
	return nil
}

func sendReqDeleteAnnouncement(ctx context.Context, announcementID, userID int) error {
	_, err := client.service.DeleteAnnouncement(ctx, &pb.DeleteAnnouncementRequest{
		AnnouncementID: int64(announcementID),
		UserID:         int64(userID),
	})
	if err != nil {
		return err
	}

	return nil
}
