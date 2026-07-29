package announcements

import (
	"context"
	"errors"

	pb "project-farm/internal/announcements/proto"
)

func DeleteAnnouncement(id int) error {
	resp, err := client.DeleteAnnouncement(context.Background(), &pb.DeleteAnnouncementRequest{
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
