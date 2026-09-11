package handlers

import (
	"context"
	"net/http"
	"strconv"

	"project-farm/internal/announcements"
	"project-farm/internal/models"
	"project-farm/internal/rdb"
	"project-farm/internal/session"
	"project-farm/internal/users"
)

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	data, err := getAnnouncementsByParameters(ctx, h.RedisDB, w, r)
	if err != nil {
		if err == session.ErrUserHaveNotSession {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		data = []*models.AnnouncementData{}
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

func getAnnouncementsByParameters(ctx context.Context, rdb rdb.DB, w http.ResponseWriter, r *http.Request) ([]*models.AnnouncementData, error) {
	var offsetInt int
	var err error
	data := []*models.AnnouncementData{}
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
			sessionID, err := session.GetCookie(r)
			if err != nil {
				return data, err
			}
			var newSession string
			newSession, userIDInt, _ = session.GetUserID(ctx, rdb, sessionID)
			if newSession != "" {
				session.SetCookie(w, newSession)
			}
		}
	} else {
		userIDInt, _ = strconv.Atoi(userID)
	}

	searchString := r.URL.Query().Get("search")

	category := r.URL.Query().Get("category")

	orderBy := r.URL.Query().Get("orderby")

	authorID := r.URL.Query().Get("authorid")
	authorIDInt := 0
	if authorID == "" {
		if login := r.URL.Query().Get("login"); login == "my" {
			sessionID, err := session.GetCookie(r)
			if err != nil {
				return data, err
			}
			var newSession string
			newSession, authorIDInt, err = session.GetUserID(ctx, rdb, sessionID)
			if err != nil {
				return data, err
			}
			if newSession != "" {
				session.SetCookie(w, newSession)
			}
		}
	} else {
		authorIDInt, err = strconv.Atoi(authorID)
	}

	data, err = announcements.SearchAnnouncements(rdb, offsetInt, userIDInt, authorIDInt, searchString, category, orderBy)
	if err != nil {
		return data, err
	}

	return data, nil
}
