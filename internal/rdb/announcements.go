package rdb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"project-farm/internal/models"
)

const (
	redisKeyAnnouncementID     = "announcementID"
	redisKeyAuthorName         = "authorName"
	redisKeyAuthorID           = "authorID"
	redisKeyTitle              = "Title"
	redisKeyDescription        = "description"
	redisKeyCategory           = "category"
	redisKeyLinkToAnnouncement = "linkToAnnouncement"
	redisKeyImages             = "imagesJSON"
)

const (
	redisKeyPrefixAnnouncementInfo = "ann_%v"
	timeToSaveAnnouncementsInRedis = time.Minute * 30
)

func (c *Client) SaveAnnouncementIDsForUser(ctx context.Context, announcements []interface{}, userID string) error {
	if err := c.rdb.RPush(ctx, userID, announcements...).Err(); err != nil {
		return err
	}

	if err := c.rdb.Expire(ctx, userID, timeToSaveSessionInRedis).Err(); err != nil {
		c.rdb.Del(ctx, userID)
		return err
	}

	return nil
}

func (c *Client) GetAnnouncementsIDsForUser(ctx context.Context, userID string) ([]string, error) {
	result, err := c.rdb.LRange(ctx, userID, 0, -1).Result()
	if err != nil {
		return []string{}, err
	}

	return result, nil
}

func (c *Client) SaveAnnouncementInfo(ctx context.Context, announcementInfo models.AnnouncementData) error {
	imagesJSON, err := json.Marshal(announcementInfo.Images)
	if err != nil {
		return err
	}

	key := fmt.Sprintf(redisKeyPrefixAnnouncementInfo, announcementInfo.AnnouncementID)
	if err := c.rdb.HSet(ctx, key, map[string]interface{}{
		redisKeyAuthorName:         announcementInfo.AuthorName,
		redisKeyAuthorID:           announcementInfo.AuthorID,
		redisKeyTitle:              announcementInfo.Title,
		redisKeyDescription:        announcementInfo.Description,
		redisKeyCategory:           announcementInfo.Category,
		redisKeyLinkToAnnouncement: announcementInfo.LinkToAnnouncement,
		redisKeyImages:             imagesJSON,
	}).Err(); err != nil {
		return err
	}

	if err := c.rdb.Expire(ctx, key, timeToSaveSessionInRedis).Err(); err != nil {
		c.rdb.Del(ctx, key)
		return err
	}

	return nil
}

func (c *Client) GetAnnouncementInfo(ctx context.Context, announcementID int) (*models.AnnouncementData, error) {
	info, err := c.rdb.HGetAll(ctx, fmt.Sprintf(redisKeyPrefixAnnouncementInfo, announcementID)).Result()
	if err != nil {
		return nil, err
	}

	var images []string
	if err := json.Unmarshal([]byte(info[redisKeyImages]), &images); err != nil {
		return nil, err
	}

	authorID, err := strconv.Atoi(info[redisKeyAuthorID])
	if err != nil {
		return nil, err
	}

	return &models.AnnouncementData{
		AuthorName:         info[redisKeyAuthorName],
		AnnouncementID:     announcementID,
		AuthorID:           authorID,
		Title:              info[redisKeyTitle],
		Description:        info[redisKeyDescription],
		Category:           info[redisKeyCategory],
		LinkToAnnouncement: info[redisKeyLinkToAnnouncement],
		Images:             images,
	}, nil
}

func (c *Client) DelAnnouncementInfo(ctx context.Context, announcementID int) error {
	return c.rdb.Del(ctx, fmt.Sprintf(redisKeyPrefixAnnouncementInfo, announcementID)).Err()
}
