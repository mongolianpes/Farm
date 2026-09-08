package handlers

import (
	"context"
	"net/http"
	"time"

	"project-farm/internal/announcements"
	"project-farm/internal/images"
	"project-farm/internal/session"
	"project-farm/internal/users"
)

const (
	userNameCookieName  = "username"
	userAvatarPathName  = "useravatar"
	myLoginAlias        = "my"
	allowedInterestsLen = 250
)

type AuthPageData struct {
	Login    string
	Password string
	Error    string
}

type RegPageData struct {
	Login     string
	Name      string
	Password  string
	Interests string
	Error     string
}

type ProfileData struct {
	Name          string
	Login         string
	ID            int
	SearchString  string
	AvatarPath    string
	Announcements []*announcements.AnnouncementData
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var data RegPageData
	if r.Method != http.MethodPost {
		h.Tmpl.ExecuteTemplate(w, "register.html", data)
		return
	}

	login := r.FormValue("login")
	name := r.FormValue("name")
	password := r.FormValue("password")
	interests := r.FormValue("interests")

	_, header, err := r.FormFile("avatar")
	if err != nil {
		if err.Error() != "http: no such file" {
			data.Error = err.Error()
			data.Login = login
			data.Name = name
			data.Password = password
			data.Interests = interests
			h.Tmpl.ExecuteTemplate(w, "register.html", data)
			return
		}
	}

	if err := users.RegisterUser(header, login, name, password, interests); err != nil {
		data.Error = err.Error()
		data.Login = login
		data.Name = name
		data.Password = password
		data.Interests = interests
		h.Tmpl.ExecuteTemplate(w, "register.html", data)
		return
	}

	http.Redirect(w, r, "/auth", http.StatusSeeOther)
}

func (h *Handler) AuthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	var data AuthPageData
	if r.Method != http.MethodPost {
		h.Tmpl.ExecuteTemplate(w, "auth.html", data)
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	userInfo, err := users.AuthUser(login, password)
	if err != nil {
		data.Error = err.Error()
		h.Tmpl.ExecuteTemplate(w, "auth.html", data)
		return
	}

	sessionID, err := session.SetSessionID(ctx, h.DB, h.RedisDB, userInfo.ID)
	if err != nil {
		data.Error = err.Error()
		h.Tmpl.ExecuteTemplate(w, "auth.html", data)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     session.IDCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(4 * 24 * time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     userNameCookieName,
		Value:    userInfo.Name,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(10 * 365 * 24 * time.Hour),
	})

	if userInfo.AvatarPath != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     userAvatarPathName,
			Value:    images.MakeCurrentPathToImage(userInfo.AvatarPath),
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(10 * 365 * 24 * time.Hour),
		})
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	if len(r.URL.Query()) == 0 {
		http.Redirect(w, r, "/profile?login=my", http.StatusSeeOther)
		return
	}

	var userID int
	var err error
	login := r.URL.Query().Get("login")
	if login == myLoginAlias {
		sessionID, err := session.GetCookie(r)
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		var newSession string
		newSession, userID, err = session.GetUserID(ctx, h.DB, h.RedisDB, sessionID)
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
		}
		if newSession != "" {
			session.SetCookie(w, newSession)
		}
	}

	userInfo, err := users.GetUserInfo(userID, login)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	data := ProfileData{
		Name:         userInfo.Name,
		Login:        login,
		ID:           userInfo.ID,
		AvatarPath:   images.MakeCurrentPathToImage(userInfo.AvatarPath),
		SearchString: r.URL.Query().Get("search"),
	}

	announcementsData, err := getAnnouncementsByParameters(ctx, h.DB, h.RedisDB, w, r)
	if err != nil {
		if err == session.ErrUserHaveNotSession {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
	}

	data.Announcements = announcementsData

	h.Tmpl.ExecuteTemplate(w, "profile.html", data)
}

func (h *Handler) GetHeaderCookieHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), timeToCompleteRequest)
	defer cancel()

	sessionID, err := session.GetCookie(r)
	if err != nil {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	newSession, userID, err := session.GetUserID(ctx, h.DB, h.RedisDB, sessionID)
	if err != nil {
		http.Error(w, "Данной сессии не существует", http.StatusBadRequest)
		return
	}
	if newSession != "" {
		session.SetCookie(w, newSession)
	}

	userInfo, err := users.GetUserInfo(userID, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     userNameCookieName,
		Value:    userInfo.Name,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(10 * 365 * 24 * time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     userAvatarPathName,
		Value:    images.MakeCurrentPathToImage(userInfo.AvatarPath),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(10 * 365 * 24 * time.Hour),
	})

	w.WriteHeader(http.StatusOK)
}
