package models

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
