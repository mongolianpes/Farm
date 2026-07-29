package handlers

import "net/http"

const htmlPagesPath = "htmlPages/"

func (h *Handler) HomePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.ServeFile(w, r, htmlPagesPath+"404.html")
		return
	}
	http.ServeFile(w, r, htmlPagesPath+"home.html")
}
