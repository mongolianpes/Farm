package announcements

import (
	"context"
	"errors"

	pb "project-farm/internal/announcements/proto"
)

func DeleteAnnouncement(id int) error {
	if err := initService(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	resp, err := client.service.DeleteAnnouncement(ctx, &pb.DeleteAnnouncementRequest{
		AnnouncementID: int32(id),
	})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	return nil
}
