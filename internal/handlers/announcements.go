package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"project-farm/internal/announcements"
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
	Announcements     []*announcements.AnnouncementData
}

func (h *Handler) CreateAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	var data CreateAnnouncementData
	if r.Method != http.MethodPost {
		h.Tmpl.ExecuteTemplate(w, "create-announcement.html", data)
		return
	}

	userID, err := session.GetUserID(h.DB, w, r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	r.ParseMultipartForm(20 << 20)
	title := r.FormValue("title")
	description := r.FormValue("description")
	category := r.FormValue("category")
	if category == "" || title == "" || description == "" {
		data.Description = description
		data.Title = title
		data.Category = category
		data.Error = "Введите информацию во всех полях"
		h.Tmpl.ExecuteTemplate(w, "create-announcement.html", data)
		return
	}

	images := r.MultipartForm.File["images"]
	if len(images) > 10 {
		data.Description = description
		data.Title = title
		data.Category = category
		data.Error = "Загружено больше 10 изображений"
		h.Tmpl.ExecuteTemplate(w, "create-announcement.html", data)
		return
	}

	if err := announcements.CreateAnnouncement(title, description, category, strconv.Itoa(userID), images); err != nil {
		data.Description = description
		data.Title = title
		data.Category = category
		data.Error = "Неизвестная ошибка"
		h.Tmpl.ExecuteTemplate(w, "create-announcement.html", data)
		return
	}

	http.Redirect(w, r, "/announcements", http.StatusSeeOther)
}

func (h *Handler) AnnouncementsPageHandler(w http.ResponseWriter, r *http.Request) {
	if announcementIDStr := r.URL.Query().Get("id"); announcementIDStr != "" {
		if announcementID, err := strconv.Atoi(announcementIDStr); err == nil {
			h.showOneAnnouncement(w, r, announcementID)
			return
		}
	}

	var err error
	data := AnnouncementsData{}
	data.Announcements, err = getAnnouncementsByParameters(h.DB, w, r)
	if err != nil {
		if err.Error() == "User have not session" {
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

func (h *Handler) showOneAnnouncement(w http.ResponseWriter, r *http.Request, announcementID int) {
	userID, err := session.GetUserID(h.DB, w, r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	announcementInfo, err := announcements.GetAnnouncementInfo(announcementID, userID)

	data := AnnouncementsData{
		Announcements: []*announcements.AnnouncementData{},
	}

	data.Announcements = append(data.Announcements, &announcements.AnnouncementData{
		AuthorName:     announcementInfo.AuthorName,
		AuthorID:       announcementInfo.AuthorID,
		Title:          announcementInfo.Title,
		Description:    announcementInfo.Description,
		Category:       announcementInfo.Category,
		Images:         announcementInfo.Images,
		AnnouncementID: announcementInfo.AnnouncementID,
	})

	data.ManyAnnouncements = false

	h.Tmpl.ExecuteTemplate(w, "announcements.html", data)
}

func (h *Handler) DeleteAnnouncementHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Println("not post")
		http.Redirect(w, r, "/profile?login=my", http.StatusSeeOther)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/profile?login=my", http.StatusSeeOther)
		return
	}
	fmt.Println("str id an to del: ", idStr)
	fmt.Println("int id an to del: ", id)

	if err := announcements.DeleteAnnouncement(id); err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/profile?login=my", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/profile?login=my", http.StatusSeeOther)
}
