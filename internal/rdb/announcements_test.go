package rdb

import (
	"context"
	"fmt"
	"project-farm/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSaveAndGetAnnouncementInfo(t *testing.T) {
	client, mr := setupTestRedis(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	data := models.AnnouncementData{
		AnnouncementID:     1001,
		AuthorName:         "Иван",
		AuthorID:           42,
		Title:              "Продам трактор",
		Description:        "Почти новый",
		Category:           "Техника",
		LinkToAnnouncement: "https://example.com/1001",
		Images:             []string{"img1.jpg", "img2.jpg"},
	}

	err := client.SaveAnnouncementInfo(ctx, data)
	require.NoError(t, err)

	require.True(t, mr.Exists(fmt.Sprintf(redisKeyPrefixAnnouncementInfo, data.AnnouncementID)))

	funcResult, err := client.GetAnnouncementInfo(ctx, data.AnnouncementID)
	require.NoError(t, err)

	require.Equal(t, data.AnnouncementID, funcResult.AnnouncementID)
	require.Equal(t, data.AuthorName, funcResult.AuthorName)
	require.Equal(t, data.AuthorID, funcResult.AuthorID)
	require.Equal(t, data.Title, funcResult.Title)
	require.Equal(t, data.Description, funcResult.Description)
	require.Equal(t, data.Category, funcResult.Category)
	require.Equal(t, data.LinkToAnnouncement, funcResult.LinkToAnnouncement)
	require.Equal(t, data.Images, funcResult.Images)
}
