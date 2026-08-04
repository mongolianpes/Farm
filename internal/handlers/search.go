package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"project-farm/internal/announcements"
	"project-farm/internal/session"
)

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	data, err := getAnnouncementsByParameters(h.DB, w, r)
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

func getAnnouncementsByParameters(db *sql.DB, w http.ResponseWriter, r *http.Request) ([]*announcements.AnnouncementData, error) {
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

	userID := r.URL.Query().Get("userid")
	if userID == "" {
		if login := r.URL.Query().Get("login"); login != "" {
			userIDInt, err := getUserIDByLogin(db, w, r)
			if err != nil {
				return data, err
			}
			userID = strconv.Itoa(userIDInt)
		} else {
			userID, err = session.GetUserIDStr(db, w, r)
			if err != nil {
				return data, errors.New("User have not session")
			}
		}
	}

	searchString := r.URL.Query().Get("search")

	category := r.URL.Query().Get("category")

	orderBy := r.URL.Query().Get("orderby")

	authorID := r.URL.Query().Get("authorid")
	if authorID == "" {
		if login := r.URL.Query().Get("login"); login == "my" {
			authorID, err = session.GetUserIDStr(db, w, r)
			if err != nil {
				return data, err
			}
		}
	}

	data, err = announcements.SearchAnnouncements(offsetInt, 0, userID, searchString, category, orderBy, authorID)
	if err != nil {
		return data, err
	}

	return data, nil
}

func getUserIDByLogin(db *sql.DB, w http.ResponseWriter, r *http.Request) (userID int, err error) {
	login := r.URL.Query().Get("login")
	if login == myLoginAlias {
		userID, err = session.GetUserID(db, w, r)
	} else {
		err = db.QueryRow("SELECT user_id FROM users WHERE login = $1", login).Scan(&userID)
	}
	return
}
