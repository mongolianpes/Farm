package handlers

import "net/http"

const HTMLPagesPath = "htmlPages/"

func (h *Handler) HomePageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, HTMLPagesPath+"home.html")
}
