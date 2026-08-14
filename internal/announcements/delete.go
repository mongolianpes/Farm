package announcements

import (
	"context"
	"errors"

	pb "project-farm/internal/announcements/proto"
)

func DeleteAnnouncement(announcementID, userID int) error {
	if err := initService(); err != nil {
		return err
	}

	if err := sendReqDeleteAnnouncement(announcementID, userID); err != nil {
		return err
	}

	return nil
}

func sendReqDeleteAnnouncement(announcementID, userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	resp, err := client.service.DeleteAnnouncement(ctx, &pb.DeleteAnnouncementRequest{
		AnnouncementID: int32(announcementID),
		UserID:         int32(userID),
	})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	return nil
}
