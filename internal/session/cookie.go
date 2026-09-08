package session

import (
	"net/http"
	"time"
)

const IDCookieName = "session_id"

func GetCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(IDCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func SetCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     IDCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(4 * 24 * time.Hour),
	})
}
