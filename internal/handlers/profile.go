package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/lib/pq"

	"project-farm/internal/announcements"
	"project-farm/internal/crypto"
	"project-farm/internal/embedding"
	"project-farm/internal/images"
	"project-farm/internal/session"
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
	Avatar    []byte
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

	userLogin := r.FormValue("login")
	if userLogin == myLoginAlias {
		data.Error = "Вы не можете зарегистрироваться с данным логином"
		h.Tmpl.ExecuteTemplate(w, "register.html", data)
		return
	}

	if len(r.FormValue("password")) <= 9 {
		data.Error = "Длина пароля должна быть больше 9 символов"
		h.Tmpl.ExecuteTemplate(w, "register.html", data)
		return
	}

	password, err := crypto.HashString(r.FormValue("password"))
	if err != nil {
		data.Error = "Не удалось вас зарегистрировать с данным паролем, попробуйте другой"
		h.Tmpl.ExecuteTemplate(w, "register.html", data)
		return
	}

	interests := r.FormValue("interests")
	if len(interests) > allowedInterestsLen {
		data.Error = "Допустимая длина интересов: 250 символов"
		h.Tmpl.ExecuteTemplate(w, "register.html", data)
		return
	}

	var pathToImages string
	avatar, header, err := r.FormFile("avatar")
	if err == nil {
		if !images.CheckCurrentFileExtansion(header) {
			data.Error = "Попробуйте загрузить картинку в другом формате (png, jpg, webp)"
			h.Tmpl.ExecuteTemplate(w, "register.html", data)
			return
		}

		pathToImages, err = images.SaveImage(200, 200, avatar)
		if err != nil {
			data.Error = "Попробуйте загрузить картинку в другом формате: " + err.Error()
			h.Tmpl.ExecuteTemplate(w, "register.html", data)
			return
		}
	}

	var userID int
	if err := h.DB.QueryRow("INSERT INTO users (login, name, password, avatar_path) VALUES ($1, $2, $3, $4) RETURNING user_id", userLogin, r.FormValue("name"), password, pathToImages).Scan(&userID); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				data.Error = "Пользователь с таким логин уже существует"
			default:
				data.Error = "Неизвестная ошибка, попробуйте позже"
			}
		} else {
			data.Error = "Неизвестная ошибка, попробуйте позже"
		}

		data.Login = r.FormValue("login")
		data.Name = r.FormValue("name")
		data.Password = r.FormValue("password")
		data.Interests = r.FormValue("interests")
		h.Tmpl.ExecuteTemplate(w, "register.html", data)

		return
	}

	go func() {
		err := embedding.InsertEmbedding(h.DB, userID, interests, "UPDATE users SET embedding = $1::float8[] WHERE user_id = $1")
		if err != nil {
			slog.Error("Ошибка при создании и вставки эмбеддинга для пользователя", "error", err)
		}
	}()

	http.Redirect(w, r, "/auth", http.StatusSeeOther)
}

func (h *Handler) AuthHandler(w http.ResponseWriter, r *http.Request) {
	var data AuthPageData
	if r.Method != http.MethodPost {
		h.Tmpl.ExecuteTemplate(w, "auth.html", data)
		return
	}

	data.Login = r.FormValue("login")
	data.Password = r.FormValue("password")

	var userID int
	var currentPassword string
	var userName string
	var avatarPath string
	if err := h.DB.QueryRow("SELECT user_id, password, name, avatar_path FROM users WHERE login = $1", r.FormValue("login")).Scan(&userID, &currentPassword, &userName, &avatarPath); err != nil {
		switch err {
		case sql.ErrNoRows:
			data.Error = "Неверный логин или пароль"
			h.Tmpl.ExecuteTemplate(w, "auth.html", data)
			return
		default:
			data.Error = "Внутреняя ошибка БД, попробуйте позже"
			h.Tmpl.ExecuteTemplate(w, "auth.html", data)
			return
		}
	}

	if !crypto.VerifyHash(data.Password, currentPassword) {
		data.Error = "Неверный логин или пароль"
		h.Tmpl.ExecuteTemplate(w, "auth.html", data)
		return
	}

	if err := session.SetSessionID(h.DB, userID, w); err != nil {
		data.Error = err.Error()
		h.Tmpl.ExecuteTemplate(w, "auth.html", data)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     userNameCookieName,
		Value:    userName,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(10 * 365 * 24 * time.Hour),
	})

	if avatarPath != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     userAvatarPathName,
			Value:    images.MakeCurrentPathToImage(avatarPath),
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(10 * 365 * 24 * time.Hour),
		})
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query()) == 0 {
		http.Redirect(w, r, "/profile?login=my", http.StatusSeeOther)
		return
	}
	var userID int
	var name string
	var avatarPath string
	var err error
	login := r.URL.Query().Get("login")
	if login == myLoginAlias || login == "" {
		login = myLoginAlias
		userID, err = session.GetUserID(h.DB, w, r)
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}

		if err := h.DB.QueryRow("SELECT name, avatar_path FROM users WHERE user_id = $1", userID).Scan(&name, &avatarPath); err != nil {
			http.Error(w, "Не удалось найти данного пользователя", http.StatusBadRequest)
			return
		}
	} else {
		if err := h.DB.QueryRow("SELECT name, user_id, avatar_path FROM users WHERE login = $1", login).Scan(&name, &userID, &avatarPath); err != nil {
			http.Error(w, "Не удалось найти данного пользователя", http.StatusBadRequest)
			return
		}
	}

	avatarPath = images.MakeCurrentPathToImage(avatarPath)

	data := ProfileData{
		Name:         name,
		Login:        login,
		ID:           userID,
		AvatarPath:   avatarPath,
		SearchString: r.URL.Query().Get("search"),
	}

	announcementsData, err := getAnnouncementsByParameters(h.DB, w, r)
	if err != nil {
		if err.Error() == "User have not session" {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
	}

	data.Announcements = announcementsData

	h.Tmpl.ExecuteTemplate(w, "profile.html", data)
}

func (h *Handler) GetHeaderCookieHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := session.GetUserID(h.DB, w, r)
	if err != nil {
		http.Error(w, "Данной сессии не существует", http.StatusBadRequest)
		return
	}

	var userName string
	var avatarPath string
	if err := h.DB.QueryRow("SELECT name, avatar_path FROM users WHERE user_id = $1", userID).Scan(&userName, &avatarPath); err != nil {
		switch err {
		case sql.ErrNoRows:
			http.Error(w, "Неверный логин", http.StatusBadRequest)
			return
		default:
			http.Error(w, "Неизвестная ошибка", http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     userNameCookieName,
		Value:    userName,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(10 * 365 * 24 * time.Hour),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     userAvatarPathName,
		Value:    images.MakeCurrentPathToImage(avatarPath),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(10 * 365 * 24 * time.Hour),
	})

	w.WriteHeader(http.StatusOK)
}
