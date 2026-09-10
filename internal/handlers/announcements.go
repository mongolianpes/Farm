package handlers

import (
	"context"
	"net/http"
	"strconv"

	"project-farm/internal/announcements"
	"project-farm/internal/images"
	"project-farm/internal/models"
	"project-farm/internal/session"
)

type CreateAnnouncementData struct {
	Title       string
	Description string
	Category    string
	Error       string
}

type AnnouncementsData struct {
	ManyAnnouncements bool
	SearchString      string
	SearchCategory    string
	Announcements     []*models.AnnouncementData
}

func (h *Handler) CreateAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	var data CreateAnnouncementData
	if r.Method != http.MethodPost {
		h.Tmpl.ExecuteTemplate(w, "create-announcement.html", data)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	sessionID, err := session.GetCookie(r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	newSession, userID, err := session.GetUserID(ctx, h.DB, h.RedisDB, sessionID)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	if newSession != "" {
		session.SetCookie(w, newSession)
	}

	r.ParseMultipartForm(20 << 20)
	title := r.FormValue("title")
	description := r.FormValue("description")
	category := r.FormValue("category")
	images := r.MultipartForm.File["images"]

	if err := announcements.CreateAnnouncement(userID, title, description, category, images); err != nil {
		data.Description = description
		data.Title = title
		data.Category = category
		data.Error = err.Error()
		h.Tmpl.ExecuteTemplate(w, "create-announcement.html", data)
		return
	}

	http.Redirect(w, r, "/announcements", http.StatusSeeOther)
}

func (h *Handler) AnnouncementsPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	if announcementIDStr := r.URL.Query().Get("id"); announcementIDStr != "" {
		if announcementID, err := strconv.Atoi(announcementIDStr); err == nil {
			h.showOneAnnouncement(w, r, ctx, announcementID)
			return
		}
	}

	var err error
	data := AnnouncementsData{}
	data.Announcements, err = getAnnouncementsByParameters(ctx, h.DB, h.RedisDB, w, r)
	if err != nil {
		if err == session.ErrUserHaveNotSession {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		http.Error(w, "Не удалось получить объявления, попробуйте позже", http.StatusInternalServerError)
		return
	}

	data.SearchString = r.URL.Query().Get("search")
	data.SearchCategory = r.URL.Query().Get("category")
	data.ManyAnnouncements = true
	h.Tmpl.ExecuteTemplate(w, "announcements.html", data)
}

func (h *Handler) showOneAnnouncement(w http.ResponseWriter, r *http.Request, ctx context.Context, announcementID int) {
	sessionID, err := session.GetCookie(r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	newSession, userID, err := session.GetUserID(ctx, h.DB, h.RedisDB, sessionID)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	if newSession != "" {
		session.SetCookie(w, newSession)
	}

	announcementInfo, err := announcements.GetAnnouncementInfo(h.RedisDB, announcementID, userID)
	if err != nil {
		announcementInfo.Description = "Произошла ошибка " + err.Error()
	}

	data := AnnouncementsData{
		Announcements: []*models.AnnouncementData{},
	}

	var imagesWithCurrentPath []string
	for _, image := range announcementInfo.Images {
		imagesWithCurrentPath = append(imagesWithCurrentPath, images.MakeCurrentPathToImage(image))
	}

	data.Announcements = append(data.Announcements, &models.AnnouncementData{
		AuthorName:     announcementInfo.AuthorName,
		AuthorID:       announcementInfo.AuthorID,
		Title:          announcementInfo.Title,
		Description:    announcementInfo.Description,
		Category:       announcementInfo.Category,
		Images:         imagesWithCurrentPath,
		AnnouncementID: announcementInfo.AnnouncementID,
	})

	data.ManyAnnouncements = false

	h.Tmpl.ExecuteTemplate(w, "announcements.html", data)
}

func (h *Handler) DeleteAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/profile?login=my", http.StatusSeeOther)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/profile?login=my", http.StatusSeeOther)
		return
	}

	sessionID, err := session.GetCookie(r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	newSession, userID, err := session.GetUserID(ctx, h.DB, h.RedisDB, sessionID)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
	}
	if newSession != "" {
		session.SetCookie(w, newSession)
	}

	if err := announcements.DeleteAnnouncement(id, userID); err != nil {
		http.Redirect(w, r, "/profile?login=my", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/profile?login=my", http.StatusSeeOther)
}
