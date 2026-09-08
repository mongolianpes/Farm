package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"project-farm/internal/announcements"
	"project-farm/internal/rdb"
	"project-farm/internal/session"
	"project-farm/internal/users"
)

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	data, err := getAnnouncementsByParameters(h.DB, h.RedisDB, w, r)
	if err != nil {
		if err.Error() == "User have not session" {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		data = []*announcements.AnnouncementData{}
	}

	if r.URL.Query().Get("offset") == "" {
		dataAnnouncements := AnnouncementsData{
			Announcements:     data,
			SearchString:      r.URL.Query().Get("search"),
			SearchCategory:    r.URL.Query().Get("category"),
			ManyAnnouncements: true,
		}
		h.Tmpl.ExecuteTemplate(w, "announcements.html", dataAnnouncements)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func getAnnouncementsByParameters(db *sql.DB, rdb rdb.DB, w http.ResponseWriter, r *http.Request) ([]*announcements.AnnouncementData, error) {
	var offsetInt int
	var err error
	data := []*announcements.AnnouncementData{}
	offset := r.URL.Query().Get("offset")
	if offset == "" {
		offsetInt = 0
	} else {
		offsetInt, err = strconv.Atoi(offset)
		if err != nil {
			return data, err
		}
	}

	userIDInt := 0
	userID := r.URL.Query().Get("userid")
	if userID == "" {
		if login := r.URL.Query().Get("login"); login != "" {
			userIDInt, _ = users.GetUserID(login)
		} else {
			userIDInt, _ = session.GetUserID(db, rdb, w, r)
		}
	} else {
		userIDInt, _ = strconv.Atoi(userID)
	}

	searchString := r.URL.Query().Get("search")

	category := r.URL.Query().Get("category")

	orderBy := r.URL.Query().Get("orderby")

	authorID := r.URL.Query().Get("authorid")
	if authorID == "" {
		if login := r.URL.Query().Get("login"); login == "my" {
			authorIDInt, err := session.GetUserID(db, rdb, w, r)
			if err != nil {
				return data, err
			}
			authorID = strconv.Itoa(authorIDInt)
		}
	}

	data, err = announcements.SearchAnnouncements(offsetInt, userIDInt, searchString, category, orderBy, authorID)
	if err != nil {
		return data, err
	}

	return data, nil
}
